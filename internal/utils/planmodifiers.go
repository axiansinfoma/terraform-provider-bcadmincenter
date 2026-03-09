// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package utils

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// NoDowngradeVersion returns a plan modifier that prevents an application version
// attribute from being lowered below the value already stored in state.
func NoDowngradeVersion() planmodifier.String {
	return noDowngradeVersionModifier{}
}

type noDowngradeVersionModifier struct{}

func (m noDowngradeVersionModifier) Description(_ context.Context) string {
	return "Prevents downgrading the application version to a lower value than the current state."
}

func (m noDowngradeVersionModifier) MarkdownDescription(_ context.Context) string {
	return "Prevents downgrading `application_version` to a lower value than the current state."
}

func (m noDowngradeVersionModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Skip when state or plan value is null/unknown (e.g. create, or version not set).
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}
	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}

	stateVer := req.StateValue.ValueString()
	planVer := req.PlanValue.ValueString()

	if planVer == stateVer {
		return
	}

	stateMajor, stateMinor, err := parseAppVersion(stateVer)
	if err != nil {
		// Cannot parse state version — do not block the plan.
		return
	}

	planMajor, planMinor, err := parseAppVersion(planVer)
	if err != nil {
		// Cannot parse plan version — do not block the plan.
		return
	}

	if planMajor < stateMajor || (planMajor == stateMajor && planMinor < stateMinor) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Version downgrade not allowed",
			fmt.Sprintf(
				"The application_version cannot be downgraded from %q to %q. "+
					"Only upgrades are supported. Remove the version pin or set a version equal to or higher than the current one.",
				stateVer, planVer,
			),
		)
	}
}

func parseAppVersion(v string) (int, int, error) {
	parts := strings.SplitN(v, ".", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid version format %q: expected major.minor", v)
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid major in version %q: %w", v, err)
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid minor in version %q: %w", v, err)
	}
	return major, minor, nil
}

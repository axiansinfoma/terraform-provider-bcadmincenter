// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package utils

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestNoDowngradeVersion(t *testing.T) {
	tests := []struct {
		name      string
		stateVer  string
		planVer   string
		wantError bool
	}{
		{
			name:      "same version allowed",
			stateVer:  "26.1",
			planVer:   "26.1",
			wantError: false,
		},
		{
			name:      "upgrade minor allowed",
			stateVer:  "26.1",
			planVer:   "26.2",
			wantError: false,
		},
		{
			name:      "upgrade major allowed",
			stateVer:  "26.1",
			planVer:   "27.0",
			wantError: false,
		},
		{
			name:      "downgrade minor blocked",
			stateVer:  "26.2",
			planVer:   "26.1",
			wantError: true,
		},
		{
			name:      "downgrade major blocked",
			stateVer:  "27.0",
			planVer:   "26.5",
			wantError: true,
		},
		{
			name:      "unparseable state passes through",
			stateVer:  "not-a-version",
			planVer:   "26.1",
			wantError: false,
		},
		{
			name:      "unparseable plan passes through",
			stateVer:  "26.1",
			planVer:   "not-a-version",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modifier := NoDowngradeVersion()

			req := planmodifier.StringRequest{
				StateValue: types.StringValue(tt.stateVer),
				PlanValue:  types.StringValue(tt.planVer),
			}
			resp := &planmodifier.StringResponse{}

			modifier.PlanModifyString(context.Background(), req, resp)

			if resp.Diagnostics.HasError() != tt.wantError {
				t.Errorf("PlanModifyString() HasError = %v, wantError %v (diagnostics: %v)",
					resp.Diagnostics.HasError(), tt.wantError, resp.Diagnostics)
			}
		})
	}
}

func TestNoDowngradeVersion_NullUnknown(t *testing.T) {
	modifier := NoDowngradeVersion()

	// Null state should pass through.
	req := planmodifier.StringRequest{
		StateValue: types.StringNull(),
		PlanValue:  types.StringValue("26.1"),
	}
	resp := &planmodifier.StringResponse{}
	modifier.PlanModifyString(context.Background(), req, resp)
	if resp.Diagnostics.HasError() {
		t.Error("Null state should not cause an error")
	}

	// Null plan should pass through.
	req = planmodifier.StringRequest{
		StateValue: types.StringValue("26.1"),
		PlanValue:  types.StringNull(),
	}
	resp = &planmodifier.StringResponse{}
	modifier.PlanModifyString(context.Background(), req, resp)
	if resp.Diagnostics.HasError() {
		t.Error("Null plan should not cause an error")
	}
}

func TestNoDowngradeAppVersion(t *testing.T) {
	tests := []struct {
		name      string
		stateVer  string
		planVer   string
		wantError bool
	}{
		{
			name:      "same full version allowed",
			stateVer:  "26.5.42000.0",
			planVer:   "26.5.42000.0",
			wantError: false,
		},
		{
			name:      "upgrade by revision allowed",
			stateVer:  "26.5.42000.0",
			planVer:   "26.5.42000.1",
			wantError: false,
		},
		{
			name:      "upgrade by build allowed",
			stateVer:  "26.5.42000.0",
			planVer:   "26.5.42001.0",
			wantError: false,
		},
		{
			name:      "upgrade by minor allowed",
			stateVer:  "26.5.42000.0",
			planVer:   "26.6.0.0",
			wantError: false,
		},
		{
			name:      "upgrade by major allowed",
			stateVer:  "26.5.42000.0",
			planVer:   "27.0.0.0",
			wantError: false,
		},
		{
			name:      "downgrade by major blocked",
			stateVer:  "27.0.0.0",
			planVer:   "26.5.42000.0",
			wantError: true,
		},
		{
			name:      "downgrade by minor blocked",
			stateVer:  "26.5.42000.0",
			planVer:   "26.4.42000.0",
			wantError: true,
		},
		{
			name:      "downgrade by build blocked",
			stateVer:  "26.5.42001.0",
			planVer:   "26.5.42000.0",
			wantError: true,
		},
		{
			name:      "downgrade by revision blocked",
			stateVer:  "26.5.42000.1",
			planVer:   "26.5.42000.0",
			wantError: true,
		},
		{
			name:      "unparseable state (2-part) passes through",
			stateVer:  "26.5",
			planVer:   "26.4.0.0",
			wantError: false,
		},
		{
			name:      "unparseable plan (2-part) passes through",
			stateVer:  "26.5.42000.0",
			planVer:   "26.4",
			wantError: false,
		},
		{
			name:      "unparseable state (non-numeric) passes through",
			stateVer:  "not-a-version",
			planVer:   "26.4.0.0",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modifier := NoDowngradeAppVersion()

			req := planmodifier.StringRequest{
				StateValue: types.StringValue(tt.stateVer),
				PlanValue:  types.StringValue(tt.planVer),
			}
			resp := &planmodifier.StringResponse{}

			modifier.PlanModifyString(context.Background(), req, resp)

			if resp.Diagnostics.HasError() != tt.wantError {
				t.Errorf("PlanModifyString() HasError = %v, wantError %v (diagnostics: %v)",
					resp.Diagnostics.HasError(), tt.wantError, resp.Diagnostics)
			}
		})
	}
}

func TestNoDowngradeAppVersion_NullUnknown(t *testing.T) {
	modifier := NoDowngradeAppVersion()

	// Null state should pass through.
	req := planmodifier.StringRequest{
		StateValue: types.StringNull(),
		PlanValue:  types.StringValue("26.5.42000.1"),
	}
	resp := &planmodifier.StringResponse{}
	modifier.PlanModifyString(context.Background(), req, resp)
	if resp.Diagnostics.HasError() {
		t.Error("Null state should not cause an error")
	}

	// Null plan should pass through.
	req = planmodifier.StringRequest{
		StateValue: types.StringValue("26.5.42000.0"),
		PlanValue:  types.StringNull(),
	}
	resp = &planmodifier.StringResponse{}
	modifier.PlanModifyString(context.Background(), req, resp)
	if resp.Diagnostics.HasError() {
		t.Error("Null plan should not cause an error")
	}
}

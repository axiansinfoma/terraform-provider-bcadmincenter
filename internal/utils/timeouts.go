// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package utils

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// DefaultOperationTimeout bounds a long-running Admin Center operation when the resource
// does not specify one. It matches the "Defaults to 60 minutes" documented on every
// resource that exposes a `timeouts` block.
const DefaultOperationTimeout = 60 * time.Minute

// OperationTimeout reads one duration from a resource's optional `timeouts` block,
// falling back to DefaultOperationTimeout when the block is absent, the key is missing,
// the value is unset, or the value cannot be parsed.
//
// Resources that schema a `timeouts` block must actually read it: bcadmincenter_environment
// and bcadmincenter_environment_app both advertised `create`/`delete` and documented the
// 60 minute default, then hardcoded 60 minutes and never looked at the configured value.
// A user raising the timeout for a large environment or app still had the operation cut
// off at an hour — and the half-created resource was then left out of state entirely.
func OperationTimeout(ctx context.Context, timeouts types.Object, key string) time.Duration {
	if timeouts.IsNull() || timeouts.IsUnknown() {
		return DefaultOperationTimeout
	}

	raw, ok := timeouts.Attributes()[key]
	if !ok {
		return DefaultOperationTimeout
	}

	value, ok := raw.(types.String)
	if !ok || value.IsNull() || value.IsUnknown() || value.ValueString() == "" {
		return DefaultOperationTimeout
	}

	parsed, err := time.ParseDuration(value.ValueString())
	if err != nil || parsed <= 0 {
		tflog.Warn(ctx, "Ignoring unparseable timeout value", map[string]interface{}{
			"timeout": key,
			"value":   value.ValueString(),
		})
		return DefaultOperationTimeout
	}

	return parsed
}

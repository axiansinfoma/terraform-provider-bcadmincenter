// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package pertenantextensions

import (
	"fmt"
	"strings"

	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/constants"
)

// BuildPerTenantExtensionID creates an ARM-like resource ID for a per-tenant extension.
// Format: /tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/applications/{applicationFamily}/environments/{environmentName}/perTenantExtensions/{appId}.
func BuildPerTenantExtensionID(tenantID, applicationFamily, environmentName, appID string) string {
	return fmt.Sprintf("/tenants/%s/providers/%s/applications/%s/environments/%s/perTenantExtensions/%s",
		tenantID, constants.ProviderNamespace, applicationFamily, environmentName, appID)
}

// ParsePerTenantExtensionID parses a per-tenant extension resource ID.
// Returns: tenantID, applicationFamily, environmentName, appID, error.
func ParsePerTenantExtensionID(id string) (string, string, string, string, error) {
	parts := strings.Split(strings.TrimPrefix(id, "/"), "/")

	// Expected: tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/applications/{applicationFamily}/environments/{environmentName}/perTenantExtensions/{appId}.
	if len(parts) != 10 {
		return "", "", "", "", fmt.Errorf(
			"invalid per-tenant extension ID format: expected '/tenants/{tenantId}/providers/%s/applications/{applicationFamily}/environments/{environmentName}/perTenantExtensions/{appId}', got: %s",
			constants.ProviderNamespace, id,
		)
	}

	if parts[0] != "tenants" {
		return "", "", "", "", fmt.Errorf("invalid per-tenant extension ID: expected 'tenants' segment, got: %s", parts[0])
	}

	if parts[2] != "providers" {
		return "", "", "", "", fmt.Errorf("invalid per-tenant extension ID: expected 'providers' segment, got: %s", parts[2])
	}

	if parts[3] != constants.ProviderNamespace {
		return "", "", "", "", fmt.Errorf("invalid per-tenant extension ID: expected provider namespace '%s', got: %s", constants.ProviderNamespace, parts[3])
	}

	if parts[4] != "applications" {
		return "", "", "", "", fmt.Errorf("invalid per-tenant extension ID: expected 'applications' segment, got: %s", parts[4])
	}

	if parts[6] != "environments" {
		return "", "", "", "", fmt.Errorf("invalid per-tenant extension ID: expected 'environments' segment, got: %s", parts[6])
	}

	if parts[8] != "perTenantExtensions" {
		return "", "", "", "", fmt.Errorf("invalid per-tenant extension ID: expected 'perTenantExtensions' segment, got: %s", parts[8])
	}

	return parts[1], parts[5], parts[7], parts[9], nil
}

// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package environmentsettings

import (
	"fmt"
	"strings"

	"github.com/vllni/terraform-provider-bcadmincenter/internal/constants"
)

// BuildEnvironmentSettingsID creates an ARM-like resource ID for environment settings.
// Format: /tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/applications/{applicationFamily}/environments/{environmentName}/settings.
func BuildEnvironmentSettingsID(tenantID, applicationFamily, environmentName string) string {
	return fmt.Sprintf("/tenants/%s/providers/%s/applications/%s/environments/%s/settings",
		tenantID, constants.ProviderNamespace, applicationFamily, environmentName)
}

// ParseEnvironmentSettingsID parses an environment settings resource ID.
// Returns: tenantID, applicationFamily, environmentName, error.
func ParseEnvironmentSettingsID(id string) (string, string, string, error) {
	parts := strings.Split(strings.TrimPrefix(id, "/"), "/")

	// Expected format: tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/applications/{applicationFamily}/environments/{environmentName}/settings.
	if len(parts) != 9 {
		return "", "", "", fmt.Errorf("invalid environment settings ID format: expected '/tenants/{tenantId}/providers/%s/applications/{applicationFamily}/environments/{environmentName}/settings', got: %s", constants.ProviderNamespace, id)
	}

	if parts[0] != "tenants" {
		return "", "", "", fmt.Errorf("invalid environment settings ID: expected 'tenants' segment, got: %s", parts[0])
	}

	if parts[2] != "providers" {
		return "", "", "", fmt.Errorf("invalid environment settings ID: expected 'providers' segment, got: %s", parts[2])
	}

	if parts[3] != constants.ProviderNamespace {
		return "", "", "", fmt.Errorf("invalid environment settings ID: expected provider namespace '%s', got: %s", constants.ProviderNamespace, parts[3])
	}

	if parts[4] != "applications" {
		return "", "", "", fmt.Errorf("invalid environment settings ID: expected 'applications' segment, got: %s", parts[4])
	}

	if parts[6] != "environments" {
		return "", "", "", fmt.Errorf("invalid environment settings ID: expected 'environments' segment, got: %s", parts[6])
	}

	if parts[8] != "settings" {
		return "", "", "", fmt.Errorf("invalid environment settings ID: expected 'settings' segment, got: %s", parts[8])
	}

	return parts[1], parts[5], parts[7], nil
}

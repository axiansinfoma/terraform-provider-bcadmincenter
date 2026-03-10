// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package environmentapps

import (
	"fmt"
	"strings"

	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/constants"
)

// BuildEnvironmentAppID creates an ARM-like resource ID for an environment app.
// Format: /tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/applications/{applicationFamily}/environments/{environmentName}/apps/{appId}.
func BuildEnvironmentAppID(tenantID, applicationFamily, environmentName, appID string) string {
	return fmt.Sprintf("/tenants/%s/providers/%s/applications/%s/environments/%s/apps/%s",
		tenantID, constants.ProviderNamespace, applicationFamily, environmentName, appID)
}

// ParseEnvironmentAppID parses an environment app resource ID.
// Returns: tenantID, applicationFamily, environmentName, appID, error.
func ParseEnvironmentAppID(id string) (string, string, string, string, error) {
	parts := strings.Split(strings.TrimPrefix(id, "/"), "/")

	// Expected format: tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/applications/{applicationFamily}/environments/{environmentName}/apps/{appId}.
	if len(parts) != 10 {
		return "", "", "", "", fmt.Errorf(
			"invalid environment app ID format: expected '/tenants/{tenantId}/providers/%s/applications/{applicationFamily}/environments/{environmentName}/apps/{appId}', got: %s",
			constants.ProviderNamespace, id,
		)
	}

	if parts[0] != "tenants" {
		return "", "", "", "", fmt.Errorf("invalid environment app ID: expected 'tenants' segment, got: %s", parts[0])
	}

	if parts[2] != "providers" {
		return "", "", "", "", fmt.Errorf("invalid environment app ID: expected 'providers' segment, got: %s", parts[2])
	}

	if parts[3] != constants.ProviderNamespace {
		return "", "", "", "", fmt.Errorf("invalid environment app ID: expected provider namespace '%s', got: %s", constants.ProviderNamespace, parts[3])
	}

	if parts[4] != "applications" {
		return "", "", "", "", fmt.Errorf("invalid environment app ID: expected 'applications' segment, got: %s", parts[4])
	}

	if parts[6] != "environments" {
		return "", "", "", "", fmt.Errorf("invalid environment app ID: expected 'environments' segment, got: %s", parts[6])
	}

	if parts[8] != "apps" {
		return "", "", "", "", fmt.Errorf("invalid environment app ID: expected 'apps' segment, got: %s", parts[8])
	}

	return parts[1], parts[5], parts[7], parts[9], nil
}

// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package environments

import (
	"fmt"
	"strings"

	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/constants"
)

// BuildEnvironmentID creates an ARM-like resource ID for an environment.
// Format: /tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/applications/{applicationFamily}/environments/{environmentName}.
func BuildEnvironmentID(tenantID, applicationFamily, environmentName string) string {
	return fmt.Sprintf("/tenants/%s/providers/%s/applications/%s/environments/%s",
		tenantID, constants.ProviderNamespace, applicationFamily, environmentName)
}

// ParseEnvironmentID parses an environment resource ID.
// Returns: tenantID, applicationFamily, environmentName, error.
func ParseEnvironmentID(id string) (string, string, string, error) {
	parts := strings.Split(strings.TrimPrefix(id, "/"), "/")

	// Expected format: tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/applications/{applicationFamily}/environments/{environmentName}.
	if len(parts) != 8 {
		return "", "", "", fmt.Errorf("invalid environment ID format: expected '/tenants/{tenantId}/providers/%s/applications/{applicationFamily}/environments/{environmentName}', got: %s", constants.ProviderNamespace, id)
	}

	if parts[0] != "tenants" {
		return "", "", "", fmt.Errorf("invalid environment ID: expected 'tenants' segment, got: %s", parts[0])
	}

	if parts[2] != "providers" {
		return "", "", "", fmt.Errorf("invalid environment ID: expected 'providers' segment, got: %s", parts[2])
	}

	if parts[3] != constants.ProviderNamespace {
		return "", "", "", fmt.Errorf("invalid environment ID: expected provider namespace '%s', got: %s", constants.ProviderNamespace, parts[3])
	}

	if parts[4] != "applications" {
		return "", "", "", fmt.Errorf("invalid environment ID: expected 'applications' segment, got: %s", parts[4])
	}

	if parts[6] != "environments" {
		return "", "", "", fmt.Errorf("invalid environment ID: expected 'environments' segment, got: %s", parts[6])
	}

	return parts[1], parts[5], parts[7], nil
}

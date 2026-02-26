// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package environmentsupportcontact

import (
	"fmt"
	"strings"

	"github.com/vllni/terraform-provider-bcadmincenter/internal/constants"
)

// BuildEnvironmentSupportContactID creates an ARM-like resource ID for environment support contact.
// Format: /tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/applications/{applicationFamily}/environments/{environmentName}/supportContact.
func BuildEnvironmentSupportContactID(tenantID, applicationFamily, environmentName string) string {
	return fmt.Sprintf("/tenants/%s/providers/%s/applications/%s/environments/%s/supportContact",
		tenantID, constants.ProviderNamespace, applicationFamily, environmentName)
}

// ParseEnvironmentSupportContactID parses an environment support contact resource ID.
// Returns: tenantID, applicationFamily, environmentName, error.
func ParseEnvironmentSupportContactID(id string) (string, string, string, error) {
	parts := strings.Split(strings.TrimPrefix(id, "/"), "/")

	// Expected format: tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/applications/{applicationFamily}/environments/{environmentName}/supportContact.
	if len(parts) != 9 {
		return "", "", "", fmt.Errorf("invalid environment support contact ID format: expected '/tenants/{tenantId}/providers/%s/applications/{applicationFamily}/environments/{environmentName}/supportContact', got: %s", constants.ProviderNamespace, id)
	}

	if parts[0] != "tenants" {
		return "", "", "", fmt.Errorf("invalid environment support contact ID: expected 'tenants' segment, got: %s", parts[0])
	}

	if parts[2] != "providers" {
		return "", "", "", fmt.Errorf("invalid environment support contact ID: expected 'providers' segment, got: %s", parts[2])
	}

	if parts[3] != constants.ProviderNamespace {
		return "", "", "", fmt.Errorf("invalid environment support contact ID: expected provider namespace '%s', got: %s", constants.ProviderNamespace, parts[3])
	}

	if parts[4] != "applications" {
		return "", "", "", fmt.Errorf("invalid environment support contact ID: expected 'applications' segment, got: %s", parts[4])
	}

	if parts[6] != "environments" {
		return "", "", "", fmt.Errorf("invalid environment support contact ID: expected 'environments' segment, got: %s", parts[6])
	}

	if parts[8] != "supportContact" {
		return "", "", "", fmt.Errorf("invalid environment support contact ID: expected 'supportContact' segment, got: %s", parts[8])
	}

	return parts[1], parts[5], parts[7], nil
}

// Copyright (c) 2025 Michael Villani
// SPDX-License-Identifier: MPL-2.0

package authorized_entra_apps

import (
	"fmt"
	"strings"

	"github.com/vllni/terraform-provider-bcadmincenter/internal/constants"
)

// BuildAuthorizedEntraAppID creates an ARM-like resource ID for an authorized Entra app.
// Format: /tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/authorizedEntraApps/{appId}.
func BuildAuthorizedEntraAppID(tenantID, appID string) string {
	return fmt.Sprintf("/tenants/%s/providers/%s/authorizedEntraApps/%s",
		tenantID, constants.ProviderNamespace, appID)
}

// ParseAuthorizedEntraAppID parses an authorized Entra app resource ID.
// Returns: tenantID, appID, error.
func ParseAuthorizedEntraAppID(id string) (string, string, error) {
	parts := strings.Split(strings.TrimPrefix(id, "/"), "/")

	// Expected format: tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/authorizedEntraApps/{appId}.
	if len(parts) != 6 {
		return "", "", fmt.Errorf("invalid authorized Entra app ID format: expected '/tenants/{tenantId}/providers/%s/authorizedEntraApps/{appId}', got: %s", constants.ProviderNamespace, id)
	}

	if parts[0] != "tenants" {
		return "", "", fmt.Errorf("invalid authorized Entra app ID: expected 'tenants' segment, got: %s", parts[0])
	}

	if parts[2] != "providers" {
		return "", "", fmt.Errorf("invalid authorized Entra app ID: expected 'providers' segment, got: %s", parts[2])
	}

	if parts[3] != constants.ProviderNamespace {
		return "", "", fmt.Errorf("invalid authorized Entra app ID: expected provider namespace '%s', got: %s", constants.ProviderNamespace, parts[3])
	}

	if parts[4] != "authorizedEntraApps" {
		return "", "", fmt.Errorf("invalid authorized Entra app ID: expected 'authorizedEntraApps' resource type, got: %s", parts[4])
	}

	return parts[1], parts[5], nil
}

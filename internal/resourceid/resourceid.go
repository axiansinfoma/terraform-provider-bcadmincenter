// Copyright (c) 2025 Michael Villani
// SPDX-License-Identifier: MPL-2.0

package resourceid

import (
	"fmt"
	"strings"
)

const (
	// ProviderNamespace for Business Central Admin Center resources
	ProviderNamespace = "Microsoft.Dynamics365.BusinessCentral"
)

// BuildNotificationRecipientID creates an ARM-like resource ID for a notification recipient
// Format: /tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/notificationRecipients/{recipientId}
func BuildNotificationRecipientID(tenantID, recipientID string) string {
	return fmt.Sprintf("/tenants/%s/providers/%s/notificationRecipients/%s",
		tenantID, ProviderNamespace, recipientID)
}

// ParseNotificationRecipientID parses a notification recipient resource ID
// Returns: tenantID, recipientID, error
func ParseNotificationRecipientID(id string) (string, string, error) {
	parts := strings.Split(strings.TrimPrefix(id, "/"), "/")

	// Expected format: tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/notificationRecipients/{recipientId}
	if len(parts) != 6 {
		return "", "", fmt.Errorf("invalid notification recipient ID format: expected '/tenants/{tenantId}/providers/%s/notificationRecipients/{recipientId}', got: %s", ProviderNamespace, id)
	}

	if parts[0] != "tenants" {
		return "", "", fmt.Errorf("invalid notification recipient ID: expected 'tenants' segment, got: %s", parts[0])
	}

	if parts[2] != "providers" {
		return "", "", fmt.Errorf("invalid notification recipient ID: expected 'providers' segment, got: %s", parts[2])
	}

	if parts[3] != ProviderNamespace {
		return "", "", fmt.Errorf("invalid notification recipient ID: expected provider namespace '%s', got: %s", ProviderNamespace, parts[3])
	}

	if parts[4] != "notificationRecipients" {
		return "", "", fmt.Errorf("invalid notification recipient ID: expected 'notificationRecipients' resource type, got: %s", parts[4])
	}

	return parts[1], parts[5], nil
}

// BuildEnvironmentID creates an ARM-like resource ID for an environment
// Format: /tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/applications/{applicationFamily}/environments/{environmentName}
func BuildEnvironmentID(tenantID, applicationFamily, environmentName string) string {
	return fmt.Sprintf("/tenants/%s/providers/%s/applications/%s/environments/%s",
		tenantID, ProviderNamespace, applicationFamily, environmentName)
}

// ParseEnvironmentID parses an environment resource ID
// Returns: tenantID, applicationFamily, environmentName, error
func ParseEnvironmentID(id string) (string, string, string, error) {
	parts := strings.Split(strings.TrimPrefix(id, "/"), "/")

	// Expected format: tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/applications/{applicationFamily}/environments/{environmentName}
	if len(parts) != 8 {
		return "", "", "", fmt.Errorf("invalid environment ID format: expected '/tenants/{tenantId}/providers/%s/applications/{applicationFamily}/environments/{environmentName}', got: %s", ProviderNamespace, id)
	}

	if parts[0] != "tenants" {
		return "", "", "", fmt.Errorf("invalid environment ID: expected 'tenants' segment, got: %s", parts[0])
	}

	if parts[2] != "providers" {
		return "", "", "", fmt.Errorf("invalid environment ID: expected 'providers' segment, got: %s", parts[2])
	}

	if parts[3] != ProviderNamespace {
		return "", "", "", fmt.Errorf("invalid environment ID: expected provider namespace '%s', got: %s", ProviderNamespace, parts[3])
	}

	if parts[4] != "applications" {
		return "", "", "", fmt.Errorf("invalid environment ID: expected 'applications' segment, got: %s", parts[4])
	}

	if parts[6] != "environments" {
		return "", "", "", fmt.Errorf("invalid environment ID: expected 'environments' segment, got: %s", parts[6])
	}

	return parts[1], parts[5], parts[7], nil
}

// BuildEnvironmentSettingsID creates an ARM-like resource ID for environment settings
// Format: /tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/applications/{applicationFamily}/environments/{environmentName}/settings
func BuildEnvironmentSettingsID(tenantID, applicationFamily, environmentName string) string {
	return fmt.Sprintf("/tenants/%s/providers/%s/applications/%s/environments/%s/settings",
		tenantID, ProviderNamespace, applicationFamily, environmentName)
}

// ParseEnvironmentSettingsID parses an environment settings resource ID
// Returns: tenantID, applicationFamily, environmentName, error
func ParseEnvironmentSettingsID(id string) (string, string, string, error) {
	parts := strings.Split(strings.TrimPrefix(id, "/"), "/")

	// Expected format: tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/applications/{applicationFamily}/environments/{environmentName}/settings
	if len(parts) != 9 {
		return "", "", "", fmt.Errorf("invalid environment settings ID format: expected '/tenants/{tenantId}/providers/%s/applications/{applicationFamily}/environments/{environmentName}/settings', got: %s", ProviderNamespace, id)
	}

	if parts[0] != "tenants" {
		return "", "", "", fmt.Errorf("invalid environment settings ID: expected 'tenants' segment, got: %s", parts[0])
	}

	if parts[2] != "providers" {
		return "", "", "", fmt.Errorf("invalid environment settings ID: expected 'providers' segment, got: %s", parts[2])
	}

	if parts[3] != ProviderNamespace {
		return "", "", "", fmt.Errorf("invalid environment settings ID: expected provider namespace '%s', got: %s", ProviderNamespace, parts[3])
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

// BuildEnvironmentSupportContactID creates an ARM-like resource ID for environment support contact
// Format: /tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/applications/{applicationFamily}/environments/{environmentName}/supportContact
func BuildEnvironmentSupportContactID(tenantID, applicationFamily, environmentName string) string {
	return fmt.Sprintf("/tenants/%s/providers/%s/applications/%s/environments/%s/supportContact",
		tenantID, ProviderNamespace, applicationFamily, environmentName)
}

// ParseEnvironmentSupportContactID parses an environment support contact resource ID
// Returns: tenantID, applicationFamily, environmentName, error
func ParseEnvironmentSupportContactID(id string) (string, string, string, error) {
	parts := strings.Split(strings.TrimPrefix(id, "/"), "/")

	// Expected format: tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/applications/{applicationFamily}/environments/{environmentName}/supportContact
	if len(parts) != 9 {
		return "", "", "", fmt.Errorf("invalid environment support contact ID format: expected '/tenants/{tenantId}/providers/%s/applications/{applicationFamily}/environments/{environmentName}/supportContact', got: %s", ProviderNamespace, id)
	}

	if parts[0] != "tenants" {
		return "", "", "", fmt.Errorf("invalid environment support contact ID: expected 'tenants' segment, got: %s", parts[0])
	}

	if parts[2] != "providers" {
		return "", "", "", fmt.Errorf("invalid environment support contact ID: expected 'providers' segment, got: %s", parts[2])
	}

	if parts[3] != ProviderNamespace {
		return "", "", "", fmt.Errorf("invalid environment support contact ID: expected provider namespace '%s', got: %s", ProviderNamespace, parts[3])
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

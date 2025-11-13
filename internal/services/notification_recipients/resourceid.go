// Copyright (c) 2025 Michael Villani
// SPDX-License-Identifier: MPL-2.0

package notificationrecipients

import (
	"fmt"
	"strings"

	"github.com/vllni/terraform-provider-bcadmincenter/internal/constants"
)

// BuildNotificationRecipientID creates an ARM-like resource ID for a notification recipient
// Format: /tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/notificationRecipients/{recipientId}
func BuildNotificationRecipientID(tenantID, recipientID string) string {
	return fmt.Sprintf("/tenants/%s/providers/%s/notificationRecipients/%s",
		tenantID, constants.ProviderNamespace, recipientID)
}

// ParseNotificationRecipientID parses a notification recipient resource ID
// Returns: tenantID, recipientID, error
func ParseNotificationRecipientID(id string) (string, string, error) {
	parts := strings.Split(strings.TrimPrefix(id, "/"), "/")

	// Expected format: tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/notificationRecipients/{recipientId}
	if len(parts) != 6 {
		return "", "", fmt.Errorf("invalid notification recipient ID format: expected '/tenants/{tenantId}/providers/%s/notificationRecipients/{recipientId}', got: %s", constants.ProviderNamespace, id)
	}

	if parts[0] != "tenants" {
		return "", "", fmt.Errorf("invalid notification recipient ID: expected 'tenants' segment, got: %s", parts[0])
	}

	if parts[2] != "providers" {
		return "", "", fmt.Errorf("invalid notification recipient ID: expected 'providers' segment, got: %s", parts[2])
	}

	if parts[3] != constants.ProviderNamespace {
		return "", "", fmt.Errorf("invalid notification recipient ID: expected provider namespace '%s', got: %s", constants.ProviderNamespace, parts[3])
	}

	if parts[4] != "notificationRecipients" {
		return "", "", fmt.Errorf("invalid notification recipient ID: expected 'notificationRecipients' resource type, got: %s", parts[4])
	}

	return parts[1], parts[5], nil
}

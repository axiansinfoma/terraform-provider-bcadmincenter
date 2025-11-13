// Copyright (c) 2025 Michael Villani
// SPDX-License-Identifier: MPL-2.0

package notificationrecipients

import "github.com/hashicorp/terraform-plugin-framework/types"

// NotificationRecipient represents a notification recipient in the Business Central Admin Center
type NotificationRecipient struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// NotificationRecipientsResponse represents the API response for listing notification recipients
type NotificationRecipientsResponse struct {
	Value []NotificationRecipient `json:"value"`
}

// CreateNotificationRecipientRequest represents the request body for creating a notification recipient
type CreateNotificationRecipientRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// NotificationRecipientResourceModel represents the Terraform resource model
type NotificationRecipientResourceModel struct {
	ID    types.String `tfsdk:"id"`
	Email types.String `tfsdk:"email"`
	Name  types.String `tfsdk:"name"`
}

// NotificationSettings represents the complete notification settings for a tenant
type NotificationSettings struct {
	AADTenantID string                  `json:"aadTenantId"`
	Recipients  []NotificationRecipient `json:"recipients"`
}

// NotificationSettingsDataSourceModel represents the Terraform data source model
type NotificationSettingsDataSourceModel struct {
	ID          types.String                           `tfsdk:"id"`
	AADTenantID types.String                           `tfsdk:"aad_tenant_id"`
	Recipients  []NotificationRecipientDataSourceModel `tfsdk:"recipients"`
}

// NotificationRecipientDataSourceModel represents a recipient in the data source
type NotificationRecipientDataSourceModel struct {
	ID    types.String `tfsdk:"id"`
	Email types.String `tfsdk:"email"`
	Name  types.String `tfsdk:"name"`
}

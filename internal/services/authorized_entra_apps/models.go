// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package authorized_entra_apps

// AuthorizedApp represents an authorized Microsoft Entra app from the Admin Center API.
type AuthorizedApp struct {
	AppID                 string `json:"appId"`
	IsAdminConsentGranted bool   `json:"isAdminConsentGranted"`
}

// AuthorizedAppsResponse represents the response from the list authorized apps endpoint.
type AuthorizedAppsResponse []AuthorizedApp

// ManageableTenant represents a tenant that an app can manage.
type ManageableTenant struct {
	EntraTenantID string `json:"entraTenantId"`
}

// ManageableTenantsResponse represents the response from the manageable tenants endpoint.
type ManageableTenantsResponse struct {
	Value []ManageableTenant `json:"value"`
}

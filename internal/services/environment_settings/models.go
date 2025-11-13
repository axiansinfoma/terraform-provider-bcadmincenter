// Copyright (c) 2025 Michael Villani
// SPDX-License-Identifier: MPL-2.0

package environmentsettings

// UpdateSettings represents the update window configuration for an environment.
type UpdateSettings struct {
	PreferredStartTime    *string `json:"preferredStartTime,omitempty"`    // Start time in HH:mm format (24h)
	PreferredEndTime      *string `json:"preferredEndTime,omitempty"`      // End time in HH:mm format (24h)
	TimeZoneID            *string `json:"timeZoneId,omitempty"`            // Windows time zone identifier
	PreferredStartTimeUTC *string `json:"preferredStartTimeUtc,omitempty"` // UTC timestamp (legacy)
	PreferredEndTimeUTC   *string `json:"preferredEndTimeUtc,omitempty"`   // UTC timestamp (legacy)
}

// AppInsightsKeyRequest represents the request body for setting Application Insights key.
type AppInsightsKeyRequest struct {
	Key string `json:"key"` // Application Insights connection string or instrumentation key
}

// SecurityGroupResponse represents the response for getting a security group.
type SecurityGroupResponse struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
}

// SecurityGroupRequest represents the request body for setting a security group.
type SecurityGroupRequest struct {
	Value string `json:"Value"` // Microsoft Entra group object ID
}

// AccessWithM365LicensesResponse represents the response for getting M365 license access setting.
type AccessWithM365LicensesResponse struct {
	Enabled bool `json:"enabled"`
}

// AccessWithM365LicensesRequest represents the request body for setting M365 license access.
type AccessWithM365LicensesRequest struct {
	Enabled bool `json:"enabled"`
}

// AppUpdateCadenceRequest represents the request body for setting app update cadence.
type AppUpdateCadenceRequest struct {
	Value string `json:"value"` // "Default", "DuringMajorUpgrade", or "DuringMajorMinorUpgrade"
}

// PartnerAccessResponse represents the response for getting partner access settings.
type PartnerAccessResponse struct {
	Status                  string   `json:"status"`                            // "Disabled", "AllowAllPartnerTenants", or "AllowSelectedPartnerTenants"
	AllowedPartnerTenantIDs []string `json:"allowedPartnerTenantIds,omitempty"` // Only if status is "AllowSelectedPartnerTenants"
}

// PartnerAccessRequest represents the request body for setting partner access settings.
type PartnerAccessRequest struct {
	Status                  string   `json:"status"`                            // "Disabled", "AllowAllPartnerTenants", or "AllowSelectedPartnerTenants"
	AllowedPartnerTenantIDs []string `json:"allowedPartnerTenantIds,omitempty"` // Only if status is "AllowSelectedPartnerTenants"
}

// TimeZone represents a time zone from the API.
type TimeZone struct {
	ID                            string `json:"id"`                            // Time zone identifier (e.g., "Romance Standard Time")
	DisplayName                   string `json:"displayName"`                   // Display name
	CurrentUTCOffset              string `json:"currentUtcOffset"`              // Offset from UTC (e.g., "+01:00")
	SupportsDaylightSavingTime    bool   `json:"supportsDaylightSavingTime"`    // Whether DST is supported
	IsCurrentlyDaylightSavingTime bool   `json:"isCurrentlyDaylightSavingTime"` // Whether DST is currently active
}

// TimeZoneListResponse represents the response for listing time zones.
type TimeZoneListResponse struct {
	Value []TimeZone `json:"value"`
}

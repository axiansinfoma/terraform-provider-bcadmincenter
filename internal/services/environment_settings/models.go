// Copyright (c) 2025 Axians Infoma GmbH
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
//
// Two decoders for `applications/settings/timezones` used to exist — this one and
// timezones.TimeZone — with mutually exclusive JSON tags for the same two fields
// (supportsDaylightSavings/offsetFromUTC versus supportsDaylightSavingTime/currentUtcOffset).
// encoding/json is case-insensitive but not synonym-aware, so at most one of them could
// ever have matched the wire format, and the other silently produced false/"" for every
// entry. Rather than guess which spelling the live API uses, both are accepted and read
// through the SupportsDST and UTCOffset accessors.
type TimeZone struct {
	ID          string `json:"id"`          // Time zone identifier (e.g., "Romance Standard Time")
	DisplayName string `json:"displayName"` // Display name

	// Offset from UTC (e.g., "+01:00"), under either spelling.
	CurrentUTCOffset string `json:"currentUtcOffset"`
	OffsetFromUTC    string `json:"offsetFromUTC"`

	// Whether DST is supported, under either spelling.
	SupportsDaylightSavingTime bool `json:"supportsDaylightSavingTime"`
	SupportsDaylightSavings    bool `json:"supportsDaylightSavings"`

	IsCurrentlyDaylightSavingTime bool `json:"isCurrentlyDaylightSavingTime"` // Whether DST is currently active
}

// UTCOffset returns the UTC offset, whichever field name the API populated.
func (t TimeZone) UTCOffset() string {
	if t.CurrentUTCOffset != "" {
		return t.CurrentUTCOffset
	}
	return t.OffsetFromUTC
}

// SupportsDST reports whether the zone observes daylight saving time, whichever field
// name the API populated.
func (t TimeZone) SupportsDST() bool {
	return t.SupportsDaylightSavingTime || t.SupportsDaylightSavings
}

// TimeZoneListResponse represents the response for listing time zones.
type TimeZoneListResponse struct {
	Value []TimeZone `json:"value"`
}

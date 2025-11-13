// Copyright (c) 2025 Michael Villani
// SPDX-License-Identifier: MPL-2.0

package timezones

// TimeZoneResponse represents the API response for timezones
type TimeZoneResponse struct {
	Value []TimeZone `json:"value"`
}

// TimeZone represents a single timezone
type TimeZone struct {
	ID                      string `json:"id"`
	DisplayName             string `json:"displayName"`
	SupportsDaylightSavings bool   `json:"supportsDaylightSavings"`
	OffsetFromUTC           string `json:"offsetFromUTC"`
}

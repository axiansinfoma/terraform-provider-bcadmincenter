// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package timezones

// TimeZoneResponse represents the API response for timezones.
type TimeZoneResponse struct {
	Value []TimeZone `json:"value"`
}

// TimeZone represents a single timezone.
//
// The Admin Center has documented this payload with two different field spellings, and
// environment_settings decoded it with the other pair. encoding/json is case-insensitive
// but not synonym-aware, so whichever struct guessed wrong silently produced false and ""
// for every entry — and the offset-filtering pattern in the tutorials would have returned
// nothing. Both spellings are accepted; read them through SupportsDST and UTCOffset.
type TimeZone struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`

	SupportsDaylightSavings    bool `json:"supportsDaylightSavings"`
	SupportsDaylightSavingTime bool `json:"supportsDaylightSavingTime"`

	OffsetFromUTC    string `json:"offsetFromUTC"`
	CurrentUTCOffset string `json:"currentUtcOffset"`
}

// UTCOffset returns the UTC offset, whichever field name the API populated.
func (t TimeZone) UTCOffset() string {
	if t.OffsetFromUTC != "" {
		return t.OffsetFromUTC
	}
	return t.CurrentUTCOffset
}

// SupportsDST reports whether the zone observes daylight saving time, whichever field
// name the API populated.
func (t TimeZone) SupportsDST() bool {
	return t.SupportsDaylightSavings || t.SupportsDaylightSavingTime
}

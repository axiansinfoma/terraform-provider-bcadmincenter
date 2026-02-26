// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package available_applications

// Ring represents a logical ring grouping where environments can be created.
type Ring struct {
	Name           string `json:"name"`
	ProductionRing bool   `json:"productionRing"`
	FriendlyName   string `json:"friendlyName"`
}

// CountryRingDetails contains the country code and available rings.
type CountryRingDetails struct {
	CountryCode string `json:"countryCode"`
	Rings       []Ring `json:"rings"`
}

// ApplicationFamily represents an application family and its available countries/regions with rings.
type ApplicationFamily struct {
	ApplicationFamily    string               `json:"applicationFamily"`
	CountriesRingDetails []CountryRingDetails `json:"countriesringDetails"`
}

// AvailableApplicationsResponse represents the API response for available applications.
type AvailableApplicationsResponse struct {
	Value []ApplicationFamily `json:"value"`
}

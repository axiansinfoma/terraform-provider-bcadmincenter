# Copyright (c) 2025 Michael Villani
# SPDX-License-Identifier: MPL-2.0

# Query available applications, countries, and rings

data "bcadmincenter_available_applications" "example" {}

# Output all available application families
output "application_families" {
  value = data.bcadmincenter_available_applications.example.application_families
}

# Find production rings for a specific country
locals {
  # Get the first application family (typically BusinessCentral)
  app_family = data.bcadmincenter_available_applications.example.application_families[0]

  # Filter for US country
  us_countries = [
    for country in local.app_family.countries_ring_details :
    country if country.country_code == "US"
  ]

  # Get production rings only
  us_production_rings = length(local.us_countries) > 0 ? [
    for ring in local.us_countries[0].rings :
    ring if ring.production_ring
  ] : []
}

# Output the production ring name for US
output "us_production_ring_name" {
  value = length(local.us_production_rings) > 0 ? local.us_production_rings[0].name : "Not found"
}

# Output all available country codes
output "available_country_codes" {
  value = [
    for country in data.bcadmincenter_available_applications.example.application_families[0].countries_ring_details :
    country.country_code
  ]
}

# Create a map of country codes to their production ring names
output "production_rings_by_country" {
  value = {
    for country in data.bcadmincenter_available_applications.example.application_families[0].countries_ring_details :
    country.country_code => [
      for ring in country.rings :
      ring.name if ring.production_ring
    ][0]
  }
}

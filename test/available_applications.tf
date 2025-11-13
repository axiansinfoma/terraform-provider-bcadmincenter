# Copyright (c) Michael Villani
# SPDX-License-Identifier: MPL-2.0

# Test configuration for available applications data sources

# Get all available applications, countries, and rings
data "bcadmincenter_available_applications" "all" {}

output "available_application_families" {
  value = [for app in data.bcadmincenter_available_applications.all.application_families : app.name]
}

output "business_central_countries" {
  value = [
    for app in data.bcadmincenter_available_applications.all.application_families :
    app.countries_ring_details if app.name == "BusinessCentral"
  ]
}

# Get specific application family details
data "bcadmincenter_application_family" "business_central" {
  name = "BusinessCentral"
}

output "bc_application_family" {
  value = {
    name      = data.bcadmincenter_application_family.business_central.name
    countries = data.bcadmincenter_application_family.business_central.countries_ring_details
  }
}

# Find available rings for a specific country
locals {
  germany_rings = [
    for country in data.bcadmincenter_application_family.business_central.countries_ring_details :
    country.rings if country.country_code == "DE"
  ]
}

output "germany_available_rings" {
  value = local.germany_rings
}

# Copyright (c) 2025 Axians Infoma GmbH
# SPDX-License-Identifier: MPL-2.0

# Retrieve information about the BusinessCentral application family

terraform {
  required_providers {
    bcadmincenter = {
      source = "vllni/bcadmincenter"
    }
  }
}

data "bcadmincenter_application_family" "bc" {
  name = "BusinessCentral"
}

# Output the available countries
output "available_countries" {
  description = "List of country codes where BusinessCentral is available"
  value       = [for country in data.bcadmincenter_application_family.bc.countries_ring_details : country.country_code]
}

# Output the production rings for the US
output "us_production_rings" {
  description = "Production rings available in the US"
  value = [
    for ring in [
      for country in data.bcadmincenter_application_family.bc.countries_ring_details :
      country if country.country_code == "US"
    ][0].rings : ring.name if ring.production_ring
  ]
}

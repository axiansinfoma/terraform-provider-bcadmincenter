---
page_title: "Data Source bcadmincenter_available_applications - bcadmincenter"
subcategory: ""
description: |-
  Retrieves the list of available application families with their countries/regions and rings. Use this data source to discover what values can be used for environment creation.
---

# Data Source (bcadmincenter_available_applications)

Retrieves the list of available application families with their countries/regions and rings. Use this data source to discover what values can be used for environment creation.

This data source retrieves the catalog of available Business Central application families, supported countries/regions, and release rings (logical ring groupings). Use this data source to:

- Discover valid values for environment creation
- List supported countries/regions for each application family
- Identify production vs. preview rings
- Programmatically select the appropriate ring based on your requirements

## Example Usage

### Basic Usage

```terraform
# Copyright (c) 2025 Axians Infoma GmbH
# SPDX-License-Identifier: MPL-2.0

# Query available applications, countries, and rings

terraform {
  required_providers {
    bcadmincenter = {
      source = "axiansinfoma/bcadmincenter"
    }
  }
}

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
```

### Find Production Ring for Specific Country

```terraform
data "bcadmincenter_available_applications" "apps" {}

locals {
  # Find all rings for US
  us_countries = [
    for country in data.bcadmincenter_available_applications.apps.application_families[0].countries_ring_details :
    country if country.country_code == "US"
  ]
  
  # Filter to production rings only
  us_production_rings = [
    for ring in local.us_countries[0].rings :
    ring if ring.production_ring
  ]
  
  # Select the first production ring
  selected_ring = local.us_production_rings[0].name
}

output "us_production_ring" {
  description = "Production ring for US environments"
  value       = local.selected_ring
}

output "us_ring_details" {
  description = "Full details of available US rings"
  value       = local.us_countries[0].rings
}
```

### List All Supported Countries

```terraform
data "bcadmincenter_available_applications" "apps" {}

locals {
  # Extract all country codes
  all_countries = flatten([
    for app_family in data.bcadmincenter_available_applications.apps.application_families : [
      for country in app_family.countries_ring_details :
      country.country_code
    ]
  ])
}

output "supported_countries" {
  description = "All countries/regions supported by Business Central"
  value       = sort(distinct(local.all_countries))
}
```

### Use with Environment Resource

```terraform
data "bcadmincenter_available_applications" "apps" {}

locals {
  app_family = data.bcadmincenter_available_applications.apps.application_families[0]
  
  # Find the production ring for a specific country
  country_rings = [
    for country in local.app_family.countries_ring_details :
    country if country.country_code == var.deployment_country
  ]
  
  production_ring = [
    for ring in local.country_rings[0].rings :
    ring.name if ring.production_ring
  ][0]
}

resource "bcadmincenter_environment" "production" {
  name               = "production"
  application_family = local.app_family.name
  type               = "Production"
  country_code       = var.deployment_country
  ring_name          = local.production_ring
  azure_region       = var.azure_region
}
```

### Create Environments for Multiple Countries

```terraform
data "bcadmincenter_available_applications" "apps" {}

locals {
  app_family = data.bcadmincenter_available_applications.apps.application_families[0]
  
  # Define deployment regions
  deployment_countries = ["US", "GB", "DK", "DE"]
  
  # Map each country to its production ring
  country_rings = {
    for country_code in local.deployment_countries :
    country_code => [
      for country in local.app_family.countries_ring_details :
      country if country.country_code == country_code
    ][0]
  }
  
  # Extract production ring name for each country
  production_rings = {
    for country_code, country_data in local.country_rings :
    country_code => [
      for ring in country_data.rings :
      ring.name if ring.production_ring
    ][0]
  }
}

resource "bcadmincenter_environment" "regional_envs" {
  for_each = local.production_rings

  name               = "prod-${lower(each.key)}"
  application_family = local.app_family.name
  type               = "Production"
  country_code       = each.key
  ring_name          = each.value
  azure_region       = var.region_mapping[each.key]
}

output "regional_environment_urls" {
  value = {
    for country, env in bcadmincenter_environment.regional_envs :
    country => env.web_client_login_url
  }
}
```

### Display Available Rings with Details

```terraform
data "bcadmincenter_available_applications" "apps" {}

output "ring_catalog" {
  description = "Complete catalog of rings by country"
  value = {
    for app_family in data.bcadmincenter_available_applications.apps.application_families :
    app_family.name => {
      for country in app_family.countries_ring_details :
      country.country_code => [
        for ring in country.rings : {
          name           = ring.name
          friendly_name  = ring.friendly_name
          production     = ring.production_ring
        }
      ]
    }
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Read-Only

- `application_families` (Attributes List) List of available application families (see [below for nested schema](#nestedatt--application_families))
- `id` (String) Data source identifier

<a id="nestedatt--application_families"></a>
### Nested Schema for `application_families`

Read-Only:

- `countries_ring_details` (Attributes List) List of countries/regions with their available rings (see [below for nested schema](#nestedatt--application_families--countries_ring_details))
- `name` (String) The name of the application family (typically 'BusinessCentral')

<a id="nestedatt--application_families--countries_ring_details"></a>
### Nested Schema for `application_families.countries_ring_details`

Read-Only:

- `country_code` (String) Code for the country/region (e.g., 'US', 'GB', 'DK')
- `rings` (Attributes List) List of available rings for this country/region (see [below for nested schema](#nestedatt--application_families--countries_ring_details--rings))

<a id="nestedatt--application_families--countries_ring_details--rings"></a>
### Nested Schema for `application_families.countries_ring_details.rings`

Read-Only:

- `friendly_name` (String) The display-friendly name of the ring
- `name` (String) The API name of the ring (e.g., 'PROD', 'PREVIEW')
- `production_ring` (Boolean) Indicates whether this is a production ring

## Attribute Reference

This data source exports the following attributes:

### application_families

A list of available application families. Each application family contains:

- `name` (String) - The name of the application family (typically `BusinessCentral`)
- `countries_ring_details` (List) - Countries/regions and their available rings

### countries_ring_details

For each country/region in an application family:

- `country_code` (String) - ISO country code (e.g., `US`, `GB`, `DK`, `DE`, `FR`)
- `rings` (List) - Available release rings for this country

### rings

For each ring available in a country/region:

- `name` (String) - API name of the ring (e.g., `PROD`, `PREVIEW`)
- `friendly_name` (String) - Display-friendly name of the ring
- `production_ring` (Boolean) - Whether this is a production ring (`true`) or preview/insider ring (`false`)

## Understanding Rings

Business Central uses "rings" to manage application releases and updates:

### Production Rings

Production rings (`production_ring = true`) are for stable, generally available releases:

- **Recommended for production environments**
- Updates follow a predictable schedule
- Thoroughly tested before release
- Longer support lifecycle
- Higher stability guarantees

### Preview/Insider Rings

Preview rings (`production_ring = false`) provide early access to new features:

- **For testing and development only**
- Early access to upcoming features
- More frequent updates
- Potentially breaking changes
- Shorter support lifecycle
- Used to validate changes before production rollout

### Ring Selection Best Practices

1. **Production Environments**: Always use production rings
2. **Sandbox Environments**: Can use preview rings to test upcoming changes
3. **Multi-Environment Strategy**: 
   - Production → Production ring
   - Staging → Production ring (same as production)
   - Development → Preview ring (test future updates)

## Country/Region Codes

Common country codes supported by Business Central:

| Code | Country/Region |
|------|----------------|
| `US` | United States |
| `GB` | United Kingdom |
| `DK` | Denmark |
| `DE` | Germany |
| `FR` | France |
| `NL` | Netherlands |
| `BE` | Belgium |
| `ES` | Spain |
| `IT` | Italy |
| `SE` | Sweden |
| `NO` | Norway |
| `FI` | Finland |
| `CA` | Canada |
| `AU` | Australia |
| `NZ` | New Zealand |

-> **Note:** The actual list of supported countries may vary by application family and is returned by this data source. Always query the data source for the most current list.

## Refresh Behavior

This data source queries the Business Central Admin Center API on every Terraform run. The data is:

- **Not cached** - Always returns current catalog information
- **Read-only** - Does not modify any resources
- **Lightweight** - Minimal API overhead

Consider using this data source in combination with local values to avoid repeated calculations:

```terraform
data "bcadmincenter_available_applications" "apps" {}

locals {
  # Cache the application family reference
  bc_app_family = data.bcadmincenter_available_applications.apps.application_families[0]
  
  # Pre-calculate common values
  production_rings_by_country = {
    for country in local.bc_app_family.countries_ring_details :
    country.country_code => [
      for ring in country.rings :
      ring if ring.production_ring
    ]
  }
}
```

## Use Cases

### 1. Environment Provisioning Validation

Validate that your desired configuration is supported before attempting to create an environment:

```terraform
data "bcadmincenter_available_applications" "apps" {}

locals {
  supported_countries = flatten([
    for app in data.bcadmincenter_available_applications.apps.application_families : [
      for country in app.countries_ring_details :
      country.country_code
    ]
  ])
}

# Validation check
resource "terraform_data" "validate_country" {
  lifecycle {
    precondition {
      condition     = contains(local.supported_countries, var.deployment_country)
      error_message = "Country ${var.deployment_country} is not supported. Supported countries: ${join(", ", local.supported_countries)}"
    }
  }
}
```

### 2. Dynamic Ring Selection

Automatically select the appropriate ring based on environment type:

```terraform
data "bcadmincenter_available_applications" "apps" {}

locals {
  app_family = data.bcadmincenter_available_applications.apps.application_families[0]
  
  country_data = [
    for country in local.app_family.countries_ring_details :
    country if country.country_code == var.country_code
  ][0]
  
  # Select production ring for prod, preview for dev
  selected_ring = var.environment_type == "production" ? [
    for ring in local.country_data.rings :
    ring.name if ring.production_ring
  ][0] : [
    for ring in local.country_data.rings :
    ring.name if !ring.production_ring
  ][0]
}
```

### 3. Documentation Generation

Generate documentation of available configurations:

```terraform
data "bcadmincenter_available_applications" "apps" {}

output "configuration_guide" {
  value = <<-EOT
    Available Business Central Configurations:
    
    ${join("\n", [
      for app_family in data.bcadmincenter_available_applications.apps.application_families :
      "Application Family: ${app_family.name}\n  Supported Countries: ${
        join(", ", [
          for country in app_family.countries_ring_details :
          country.country_code
        ])
      }"
    ])}
  EOT
}
```

## Related Resources

- `bcadmincenter_environment` resource - Create environments using the ring information
- Business Central Admin Center API - [Available Applications Endpoint](https://learn.microsoft.com/en-us/dynamics365/business-central/dev-itpro/administration/administration-center-api_available_applications)

# Available Applications Service

This service handles interactions with the Business Central Admin Center API for retrieving available application families, countries/regions, and rings.

## Purpose

The available applications service provides information about:
- Available Business Central application families (e.g., "BusinessCentral")
- Supported countries/regions for each application family
- Available rings (e.g., PROD, PREVIEW) within each country/region
- Which rings are production rings

This information is essential for:
- Validating environment creation parameters
- Discovering supported deployment regions
- Selecting appropriate rings for environment provisioning

## API Endpoint

```
GET /admin/v2.24/applications/
```

## Response Structure

```json
{
  "value": [
    {
      "applicationFamily": "BusinessCentral",
      "countriesringDetails": [
        {
          "countryCode": "US",
          "rings": [
            {
              "name": "PROD",
              "productionRing": true,
              "friendlyName": "Production"
            },
            {
              "name": "PREVIEW",
              "productionRing": false,
              "friendlyName": "Preview"
            }
          ]
        }
      ]
    }
  ]
}
```

## Components

### models.go
Defines the data structures that map to the API response:
- `Ring` - Individual ring information
- `CountryRingDetails` - Country code and associated rings
- `ApplicationFamily` - Application family name and country details
- `AvailableApplicationsResponse` - Top-level API response wrapper

### service.go
Implements the service layer for API communication:
- `Service` - Service struct that wraps the API client
- `GetAvailableApplications()` - Retrieves all available applications and rings
- `GetApplicationFamily()` - Retrieves a specific application family by name

### data_source_available_applications.go
Implements the Terraform data source for listing all application families:
- Schema definition with nested attributes
- Configure method for client injection
- Read method that fetches and transforms API data

### data_source_application_family.go
Implements the Terraform data source for retrieving a single application family:
- Schema definition with required `name` parameter
- Configure method for client injection
- Read method that fetches and transforms data for a specific family

## Usage in Terraform

### List All Application Families

```hcl
data "bcadmincenter_available_applications" "apps" {}

# Access application families
output "families" {
  value = data.bcadmincenter_available_applications.apps.application_families
}

# Find production ring for a country
locals {
  us_prod_ring = [
    for country in data.bcadmincenter_available_applications.apps.application_families[0].countries_ring_details :
    [for ring in country.rings : ring.name if ring.production_ring][0]
    if country.country_code == "US"
  ][0]
}
```

### Get Specific Application Family

```hcl
data "bcadmincenter_application_family" "bc" {
  name = "BusinessCentral"
}

# Output available countries
output "available_countries" {
  value = [for country in data.bcadmincenter_application_family.bc.countries_ring_details : country.country_code]
}

# Use in environment resource
resource "bcadmincenter_environment" "prod" {
  name                = "production"
  application_family  = data.bcadmincenter_application_family.bc.name
  country_code        = data.bcadmincenter_application_family.bc.countries_ring_details[0].country_code
  ring_name           = data.bcadmincenter_application_family.bc.countries_ring_details[0].rings[0].name
  type                = "Production"
  application_version = "24.0"
}
```

## Documentation Reference

For more details on the API, see:
https://learn.microsoft.com/en-us/dynamics365/business-central/dev-itpro/administration/administration-center-api_available_applications

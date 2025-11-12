# Available Applications Data Source Example

This example demonstrates how to use the `bcadmincenter_available_applications` data source to discover available application families, countries/regions, and rings for Business Central environment creation.

## What This Example Does

1. Queries the Business Central Admin Center API to retrieve all available application families
2. Displays all application families with their supported countries and rings
3. Filters and outputs specific information like:
   - Production ring names for specific countries
   - List of all available country codes
   - A map of country codes to their production ring names

## Use Cases

- **Environment Creation Planning**: Determine valid values for `country_code` and `ring_name` before creating environments
- **Multi-Region Deployments**: Identify which countries/regions support Business Central
- **Ring Selection**: Find production vs. preview rings for different countries
- **Validation**: Verify that desired country/ring combinations are available

## Expected Output

Running this example will output:

- `application_families`: Complete list of application families with all nested details
- `us_production_ring_name`: The production ring name for the US region
- `available_country_codes`: Array of all supported country codes
- `production_rings_by_country`: Map of country codes to their production ring names

## Integration with Environment Resources

You can use this data source in combination with the `bcadmincenter_environment` resource:

```terraform
data "bcadmincenter_available_applications" "apps" {}

resource "bcadmincenter_environment" "prod" {
  name               = "production"
  application_family = data.bcadmincenter_available_applications.apps.application_families[0].name
  type              = "Production"
  country_code      = "US"
  
  # Dynamically select the production ring for US
  ring_name = [
    for country in data.bcadmincenter_available_applications.apps.application_families[0].countries_ring_details :
    [for ring in country.rings : ring.name if ring.production_ring][0]
    if country.country_code == "US"
  ][0]
  
  application_version = "24.0"
}
```

## API Reference

This data source queries:
- `GET /admin/v2.24/applications/`

See the [official documentation](https://learn.microsoft.com/en-us/dynamics365/business-central/dev-itpro/administration/administration-center-api_available_applications#applications-and-corresponding-countriesregions-with-rings) for more details.

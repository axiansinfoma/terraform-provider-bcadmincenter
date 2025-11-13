# Application Family Data Source Example

This example demonstrates how to use the `bcadmincenter_application_family` data source to retrieve information about a specific application family.

## Usage

The data source retrieves detailed information about a single application family, including:

- Available countries/regions
- Available rings per country/region
- Production vs. preview ring indicators
- Ring friendly names

## Example Outputs

The example shows how to:

1. Query the BusinessCentral application family
2. Extract a list of available countries
3. Filter for production rings in a specific country

## Integration with Resources

This data source is commonly used with the `bcadmincenter_environment` resource to dynamically configure environments based on available options.

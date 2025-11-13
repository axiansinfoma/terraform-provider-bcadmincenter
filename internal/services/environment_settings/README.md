# Environment Settings Service

This service implements Business Central environment settings management through the Admin Center API.

## Overview

The environment settings service provides comprehensive configuration management for Business Central environments, including:

- **Update Windows**: Configure when environment updates can run
- **Application Insights**: Set up telemetry for monitoring
- **Security Groups**: Restrict environment access to specific Azure AD groups
- **M365 License Access**: Enable access with Microsoft 365 licenses
- **App Update Cadence**: Control how frequently AppSource apps update
- **Partner Access**: Manage delegated administrator access

## Service Methods

### Update Settings

- `GetUpdateSettings()` - Retrieve current update window configuration
- `SetUpdateSettings()` - Configure update window with start time, end time, and timezone
- `GetTimeZones()` - List available Windows time zone identifiers

### Application Insights

- `SetAppInsightsKey()` - Configure Application Insights connection string (triggers environment restart)

### Security Groups

- `GetSecurityGroup()` - Get the currently assigned Azure AD security group
- `SetSecurityGroup()` - Assign a security group to restrict access
- `ClearSecurityGroup()` - Remove security group restrictions

### Access Controls

- `GetAccessWithM365Licenses()` - Check if M365 license access is enabled
- `SetAccessWithM365Licenses()` - Enable/disable M365 license access

### App Management

- `SetAppUpdateCadence()` - Configure AppSource app update frequency

### Partner Access

- `GetPartnerAccess()` - Get partner access configuration
- `SetPartnerAccess()` - Configure delegated administrator access

## Terraform Resource

**Resource Name**: `bcadmincenter_environment_settings`

### Required Attributes

- `application_family` - Application family (e.g., "BusinessCentral")
- `environment_name` - Name of the environment

### Optional Attributes

- `update_window_start_time` - Update window start (HH:mm format)
- `update_window_end_time` - Update window end (HH:mm format)
- `update_window_timezone` - Windows time zone identifier
- `app_insights_key` - Application Insights connection string (sensitive, triggers restart)
- `security_group_id` - Azure AD security group object ID
- `access_with_m365_licenses` - Enable M365 license access (bool)
- `app_update_cadence` - App update frequency ("Default", "DuringMajorUpgrade", "DuringMajorMinorUpgrade")
- `partner_access_status` - Partner access mode ("Disabled", "AllowAllPartnerTenants", "AllowSelectedPartnerTenants")
- `allowed_partner_tenant_ids` - List of allowed partner tenant IDs

## API Endpoints

All endpoints use the base path `/admin/v2.24/applications/{applicationFamily}/environments/{environmentName}/settings/`

- **GET/PUT** `/upgrade` - Update window settings
- **POST** `/appinsightskey` - Application Insights configuration
- **GET/POST/DELETE** `/securitygroupaccess` - Security group management
- **GET/POST** `/accesswithm365licenses` - M365 license access
- **PUT** `/appSourceAppsUpdateCadence` - App update cadence
- **GET/PUT** `/partneraccess` - Partner access settings

## Important Considerations

### Update Windows

- Must be at least 6 hours long
- Uses Windows time zone identifiers (e.g., "Pacific Standard Time")
- Can use either wall-time + timezone (recommended) or UTC parameters
- Conflicts with scheduled updates can cause `ScheduledUpgradeConstraintViolation` errors

### Application Insights

- Setting the key triggers an **automatic environment restart**
- No restart occurs if the environment is not in 'Active' status
- Connection strings are preferred over legacy instrumentation keys
- The key cannot be read back (write-only)

### Security Groups

- Uses Azure AD (Microsoft Entra) security group object IDs
- Returns 204 No Content when no group is configured
- If a group is deleted in Azure AD, the ID remains but DisplayName is empty

### M365 License Access

- Requires environment version 21.1 or later
- Feature may not be available on older environments

### Partner Access

- **Requires Global Administrator permissions**
- Delegated administrator authentication is NOT supported for this setting
- Used to control access from partner tenants and multitenant apps

### App Update Cadence

- No GET endpoint available (write-only setting)
- Values: "Default", "DuringMajorUpgrade", "DuringMajorMinorUpgrade"

## Testing

### Service Tests (`service_test.go`)

- Mock HTTP server for all API endpoints
- Tests for success scenarios
- Tests for error scenarios (404, 400, etc.)
- Tests for edge cases (null responses, missing data)

### Resource Tests (`resource_test.go`)

- Metadata validation
- Schema validation
- Configure method testing
- Attribute presence verification

All tests pass and provide comprehensive coverage of the service functionality.

## Resource Lifecycle

### Create

Applies all configured settings to the environment:
1. Sets update window if provided
2. Configures Application Insights if key provided
3. Assigns security group if specified
4. Enables/disables M365 license access if configured
5. Sets app update cadence if specified
6. Configures partner access if specified

### Read

Reads current settings from the API:
- Update window settings
- Security group assignment
- M365 license access status
- Note: Some settings cannot be read back (AppInsights key, app update cadence, partner access)

### Update

Updates changed settings:
- Compares plan vs state for each setting
- Only updates settings that have changed
- Can clear security group by setting to null

### Delete

Removes resource from Terraform state but does NOT reset environment settings.

**Important**: Settings remain as configured on the environment. Add a warning to users about this behavior.

## Example Usage

```terraform
resource "bc_admin_center_environment" "production" {
  name               = "production"
  application_family = "BusinessCentral"
  type               = "Production"
  country_code       = "US"
  ring_name          = "Production"
  application_version = "25.0"
}

resource "bc_admin_center_environment_settings" "production" {
  application_family = bc_admin_center_environment.production.application_family
  environment_name   = bc_admin_center_environment.production.name

  # Update window (Pacific Time, 10 PM - 6 AM)
  update_window_start_time = "22:00"
  update_window_end_time   = "06:00"
  update_window_timezone   = "Pacific Standard Time"

  # Telemetry
  app_insights_key = var.app_insights_connection_string

  # Access control
  security_group_id = "12345678-1234-1234-1234-123456789012"
  
  # Features
  access_with_m365_licenses = true
  app_update_cadence       = "DuringMajorUpgrade"
}
```

## Documentation

- Template: `templates/resources/environment_settings.md.tmpl`
- Generated: `docs/resources/environment_settings.md`
- Examples: `examples/resources/bc_admin_center_environment_settings/resource.tf`

## References

- [Business Central Admin Center API - Environment Settings](https://learn.microsoft.com/en-us/dynamics365/business-central/dev-itpro/administration/administration-center-api_environment_settings)
- [Managing Updates in the Admin Center](https://learn.microsoft.com/en-us/dynamics365/business-central/dev-itpro/administration/tenant-admin-center-update-management)
- [Environment Telemetry](https://learn.microsoft.com/en-us/dynamics365/business-central/dev-itpro/administration/tenant-admin-center-telemetry)

# Authorized Microsoft Entra Apps Service

This service handles interactions with the Business Central Admin Center API for managing authorized Microsoft Entra applications.

## Overview

Authorized Microsoft Entra apps are applications that have been granted permission to call the Business Central Admin Center API. This service allows you to:

- List all authorized apps for a tenant
- Authorize an app to access the Admin Center API
- Remove an app's authorization
- Get a list of manageable tenants (for apps authenticating as themselves)

## API Endpoints

All endpoints use the base path `/admin/v2.24/authorizedAadApps`

- **GET** `/authorizedAadApps` - List all authorized apps (user authentication only)
- **PUT** `/authorizedAadApps/{appId}` - Authorize an app (user authentication only)
- **DELETE** `/authorizedAadApps/{appId}` - Remove authorization
- **GET** `/authorizedAadApps/manageableTenants` - Get manageable tenants (app authentication only)

## Service Methods

### Authorization Management

- `ListAuthorizedApps()` - Get all authorized apps (requires user authentication)
- `AuthorizeApp(appID)` - Authorize an app to call the API (requires user authentication)
- `RemoveAuthorizedApp(appID)` - Remove app authorization

### Tenant Discovery (for apps)

- `GetManageableTenants()` - List tenants where the app is authorized (requires app authentication)

## Terraform Resources

### Resource: `bcadmincenter_authorized_entra_app`

Manages authorization of a Microsoft Entra app to call the Business Central Admin Center API.

**Required Attributes:**
- `app_id` - The application (client) ID of the Microsoft Entra app

**Optional Attributes:**
- `aad_tenant_id` - The Azure AD tenant ID (defaults to provider's tenant)

**Computed Attributes:**
- `id` - ARM-like resource ID
- `is_admin_consent_granted` - Whether admin consent has been granted

**Important Notes:**
- This resource does NOT grant admin consent - that must be done separately in Azure AD
- This resource does NOT assign permission sets in environments
- Cannot be used when authenticated as an app

### Data Source: `bcadmincenter_authorized_entra_apps`

Retrieves a list of all authorized Microsoft Entra apps.

**Computed Attributes:**
- `apps` - List of authorized apps, each containing:
  - `app_id` - The application (client) ID
  - `is_admin_consent_granted` - Whether admin consent has been granted

### Data Source: `bcadmincenter_manageable_tenants`

Retrieves a list of Microsoft Entra tenants where the authenticating app is authorized.

**Computed Attributes:**
- `tenants` - List of manageable tenants, each containing:
  - `entra_tenant_id` - The Microsoft Entra tenant ID

**Important Notes:**
- This data source can ONLY be used with app authentication (service principal)
- Designed for multi-tenant apps to discover customer tenants
- Cannot be used with user authentication

## Authentication Considerations

### User Authentication (Delegated Access)

These endpoints can only be used with user authentication:
- `ListAuthorizedApps()`
- `AuthorizeApp()`

They will fail if you're authenticated as an app (service principal).

### App Authentication (Application Access)

This endpoint can only be used with app authentication:
- `GetManageableTenants()`

It's designed for multi-tenant apps to discover which tenants they can manage.

## Admin Consent and Permissions

### Authorization vs Consent

**Authorizing an app** (via this service) allows it to appear in the Business Central Admin Center's authorized apps list. However, for the app to actually call the API, you must also:

1. **Grant admin consent** for the `AdminCenter.ReadWrite.All` permission in Azure AD
2. **Assign permission sets** in Business Central environments for specific operations (e.g., `D365 BACKUP/RESTORE` for database exports)

### Complete Setup Process

1. Register the app in Azure AD
2. Add the `AdminCenter.ReadWrite.All` API permission
3. **Use this service** to authorize the app in Business Central Admin Center
4. Grant admin consent in Azure AD (via Azure Portal or Admin Center UI)
5. Assign required permission sets in environments (if needed for app management, database exports, etc.)

## Resource ID Format

All resources use an ARM-like resource ID format:

```
/tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/authorizedEntraApps/{appId}
```

## Example Usage

### Authorize an App

```terraform
resource "bcadmincenter_authorized_entra_app" "partner_app" {
  app_id = "550e8400-e29b-41d4-a716-446655440000"
}
```

### List All Authorized Apps

```terraform
data "bcadmincenter_authorized_entra_apps" "all" {}

output "authorized_apps" {
  value = data.bcadmincenter_authorized_entra_apps.all.apps
}
```

## Testing

The service includes comprehensive tests:

- **Service tests** (`service_test.go`) - Tests for all service methods with mock HTTP responses
- **Resource/Data source tests** (`resource_test.go`) - Tests for Metadata, Schema, and Configure methods
- **Resource ID tests** - Tests in `internal/resourceid/resourceid_test.go` for ID parsing and building

Run tests with:
```bash
go test ./internal/services/authorized_entra_apps/... -v
```

## API Documentation

For detailed API documentation, see:
- [Business Central Admin Center API - Authorized Microsoft Entra apps](https://learn.microsoft.com/en-us/dynamics365/business-central/dev-itpro/administration/administration-center-api_authorizedaadapps)
- [Authenticate using service-to-service Microsoft Entra apps](https://learn.microsoft.com/en-us/dynamics365/business-central/dev-itpro/administration/administration-center-api#authenticate-using-service-to-service-microsoft-entra-apps-client-credentials-flow)

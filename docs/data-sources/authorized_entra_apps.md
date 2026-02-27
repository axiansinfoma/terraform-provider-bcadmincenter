---
page_title: "Data Source bcadmincenter_authorized_entra_apps - bcadmincenter"
subcategory: "Settings"
description: |-
  Retrieves a list of all Microsoft Entra apps authorized to call the Business Central Admin Center API.
---

# Data Source (bcadmincenter_authorized_entra_apps)

Retrieves a list of all Microsoft Entra apps authorized to call the Business Central Admin Center API.

Use this data source to retrieve all Microsoft Entra apps that are authorized to call the Business Central Admin Center API for your tenant. This is useful for:

- Auditing which apps have API access
- Checking admin consent status
- Making conditional decisions based on authorized apps
- Validating security configurations

~> **Important:** This data source cannot be used when the provider is authenticated as an app (service principal). It requires user authentication with delegated permissions.

## Example Usage

### List All Authorized Apps

```terraform
# Copyright (c) 2025 Axians Infoma GmbH
# SPDX-License-Identifier: MPL-2.0

# Example: List All Authorized Microsoft Entra Apps

terraform {
  required_providers {
    bcadmincenter = {
      source = "axiansinfoma/bcadmincenter"
    }
  }
}

provider "bcadmincenter" {
  # Authentication can be configured via environment variables:
  # AZURE_CLIENT_ID
  # AZURE_CLIENT_SECRET
  # AZURE_TENANT_ID
}

# Get a list of all authorized Microsoft Entra apps
data "bcadmincenter_authorized_entra_apps" "all" {}

# Output the list of authorized apps
output "authorized_apps" {
  value = [
    for app in data.bcadmincenter_authorized_entra_apps.all.apps : {
      app_id                   = app.app_id
      is_admin_consent_granted = app.is_admin_consent_granted
    }
  ]
  description = "List of all authorized Microsoft Entra apps"
}

# Filter apps by consent status
output "apps_with_consent" {
  value = [
    for app in data.bcadmincenter_authorized_entra_apps.all.apps :
    app.app_id if app.is_admin_consent_granted
  ]
  description = "App IDs that have admin consent granted"
}

output "apps_without_consent" {
  value = [
    for app in data.bcadmincenter_authorized_entra_apps.all.apps :
    app.app_id if !app.is_admin_consent_granted
  ]
  description = "App IDs that need admin consent"
}
```

### Filter Apps with Admin Consent

```terraform
data "bcadmincenter_authorized_entra_apps" "all" {}

output "apps_with_consent" {
  description = "Apps that have admin consent granted"
  value = [
    for app in data.bcadmincenter_authorized_entra_apps.all.apps :
    app.app_id if app.is_admin_consent_granted
  ]
}

output "apps_without_consent" {
  description = "Apps that need admin consent"
  value = [
    for app in data.bcadmincenter_authorized_entra_apps.all.apps :
    app.app_id if !app.is_admin_consent_granted
  ]
}
```

### Check if Specific App is Authorized

```terraform
data "bcadmincenter_authorized_entra_apps" "all" {}

locals {
  target_app_id = "550e8400-e29b-41d4-a716-446655440000"
  authorized_app_ids = [for app in data.bcadmincenter_authorized_entra_apps.all.apps : app.app_id]
  is_app_authorized = contains(local.authorized_app_ids, local.target_app_id)
}

output "app_authorization_status" {
  value = local.is_app_authorized ? "Authorized" : "Not Authorized"
}
```

### Conditional Resource Creation

```terraform
data "bcadmincenter_authorized_entra_apps" "all" {}

locals {
  monitoring_app_id = "550e8400-e29b-41d4-a716-446655440000"
  app_exists = contains(
    [for app in data.bcadmincenter_authorized_entra_apps.all.apps : app.app_id],
    local.monitoring_app_id
  )
}

# Only authorize the app if it's not already authorized
resource "bcadmincenter_authorized_entra_app" "monitoring" {
  count = local.app_exists ? 0 : 1
  
  app_id = local.monitoring_app_id
}
```

### Generate Audit Report

```terraform
data "bcadmincenter_authorized_entra_apps" "all" {}

output "authorization_audit" {
  description = "Audit report of all authorized apps"
  value = {
    total_apps = length(data.bcadmincenter_authorized_entra_apps.all.apps)
    apps_with_consent = length([
      for app in data.bcadmincenter_authorized_entra_apps.all.apps :
      app if app.is_admin_consent_granted
    ])
    apps_without_consent = length([
      for app in data.bcadmincenter_authorized_entra_apps.all.apps :
      app if !app.is_admin_consent_granted
    ])
    app_details = [
      for app in data.bcadmincenter_authorized_entra_apps.all.apps : {
        app_id        = app.app_id
        has_consent   = app.is_admin_consent_granted
        consent_status = app.is_admin_consent_granted ? "Granted" : "Pending"
      }
    ]
  }
}
```

### Export to CSV-like Format

```terraform
data "bcadmincenter_authorized_entra_apps" "all" {}

output "apps_csv" {
  description = "CSV-formatted list of authorized apps"
  value = join("\n", concat(
    ["app_id,is_admin_consent_granted"],
    [
      for app in data.bcadmincenter_authorized_entra_apps.all.apps :
      "${app.app_id},${app.is_admin_consent_granted}"
    ]
  ))
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Read-Only

- `apps` (Attributes List) List of authorized Microsoft Entra apps. (see [below for nested schema](#nestedatt--apps))

<a id="nestedatt--apps"></a>
### Nested Schema for `apps`

Read-Only:

- `app_id` (String) The application (client) ID of the Microsoft Entra app.
- `is_admin_consent_granted` (Boolean) Indicates whether admin consent has been granted for the app.

## Authentication Requirements

~> **Warning:** This data source can only be used with user authentication (delegated permissions). It will fail if the provider is authenticated as an app using client credentials flow.

If you need to discover manageable tenants from an app context, use the separate `GetManageableTenants` API endpoint.

## Admin Consent Status

The `is_admin_consent_granted` attribute indicates whether admin consent has been granted for each app. Apps without consent are authorized but cannot yet call the API successfully.

To grant consent:
- Via Azure Portal: Navigate to the app's permissions and click "Grant admin consent"
- Via Admin Center: Go to the Authorized Microsoft Entra apps page and click "Grant Consent"

## Common Use Cases

### Security Auditing

Monitor which apps have API access and their consent status:

```terraform
data "bcadmincenter_authorized_entra_apps" "all" {}

# Alert if any app lacks admin consent
check "admin_consent_check" {
  assert {
    condition = alltrue([
      for app in data.bcadmincenter_authorized_entra_apps.all.apps :
      app.is_admin_consent_granted
    ])
    error_message = "One or more authorized apps are missing admin consent"
  }
}
```

### Compliance Reporting

Generate reports for compliance audits:

```terraform
data "bcadmincenter_authorized_entra_apps" "all" {}

resource "local_file" "compliance_report" {
  filename = "bc-authorized-apps-report.json"
  content = jsonencode({
    report_date = timestamp()
    tenant_id   = var.tenant_id
    authorized_apps = data.bcadmincenter_authorized_entra_apps.all.apps
  })
}
```

### Automated Cleanup Detection

Identify apps that should be deauthorized:

```terraform
data "bcadmincenter_authorized_entra_apps" "all" {}

locals {
  # List of app IDs that should be authorized
  approved_apps = [
    "550e8400-e29b-41d4-a716-446655440000",
    "660e8400-e29b-41d4-a716-446655440001"
  ]
  
  # Find unauthorized apps
  unauthorized_apps = [
    for app in data.bcadmincenter_authorized_entra_apps.all.apps :
    app.app_id if !contains(local.approved_apps, app.app_id)
  ]
}

output "apps_to_review" {
  description = "Apps that may need to be deauthorized"
  value       = local.unauthorized_apps
}
```

## Performance Considerations

This data source retrieves all authorized apps in a single API call. For tenants with many authorized apps, consider:

- Caching the data source results using Terraform's `lifecycle` block
- Filtering in Terraform rather than making multiple data source calls
- Using outputs to process and filter the list as needed

## Related Resources

- `bcadmincenter_authorized_entra_app` - Resource to manage individual app authorizations
- `bcadmincenter_environment` - Manage Business Central environments
- `bcadmincenter_environment_settings` - Configure environment settings

## See Also

- [Business Central Admin Center API - Authorized Microsoft Entra apps](https://learn.microsoft.com/en-us/dynamics365/business-central/dev-itpro/administration/administration-center-api_authorizedaadapps)
- [Authenticate using service-to-service Microsoft Entra apps](https://learn.microsoft.com/en-us/dynamics365/business-central/dev-itpro/administration/administration-center-api#authenticate-using-service-to-service-microsoft-entra-apps-client-credentials-flow)
- [Managing Microsoft Entra applications](https://learn.microsoft.com/en-us/azure/active-directory/manage-apps/what-is-application-management)

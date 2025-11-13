# Notification Settings Data Source Example

This example demonstrates how to retrieve notification settings for a Business Central tenant.

## What This Data Source Provides

The notification settings data source returns:
- The Azure AD tenant ID
- A complete list of all notification recipients configured for the tenant

## Usage Patterns

### View Current Recipients

Use this data source to see all currently configured notification recipients without making changes.

### Conditional Logic

Use the data to make decisions in your Terraform configuration:

```terraform
data "bcadmincenter_notification_settings" "current" {
}

# Only create a new recipient if fewer than 5 exist
resource "bcadmincenter_notification_recipient" "additional" {
  count = length(data.bcadmincenter_notification_settings.current.recipients) < 5 ? 1 : 0
  
  email = "additional-admin@example.com"
  name  = "Additional Administrator"
}
```

### Reference in Other Resources

Use recipient information for validation or cross-referencing:

```terraform
data "bcadmincenter_notification_settings" "current" {
}

# Check if a specific email is already configured
locals {
  admin_email = "admin@example.com"
  is_configured = contains([
    for r in data.bcadmincenter_notification_settings.current.recipients : r.email
  ], local.admin_email)
}

output "admin_already_configured" {
  value = local.is_configured
}
```

## Important Notes

- This is a read-only data source
- All recipients are retrieved in a single API call
- The tenant ID is the same as the authenticated Azure AD tenant

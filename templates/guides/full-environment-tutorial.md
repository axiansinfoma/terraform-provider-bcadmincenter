---
page_title: "Guide: Provisioning a Business Central Environment"
subcategory: "Getting Started"
description: |-
  This guide walks you through provisioning a production-ready Business Central environment from scratch using the Business Central Admin Center Terraform provider.
---

# Tutorial: Provisioning a Business Central Environment

This tutorial walks you through provisioning a production-ready Business Central environment from scratch using the Business Central Admin Center Terraform provider. It uses data sources to discover dynamic values (application family, country, ring) so the configuration stays valid as the API evolves.

## Prerequisites

- Terraform 1.0 or later installed
- A configured Azure AD application with the required permissions (see [Authentication Setup](./service-principal-authentication.md))
- Service principal added to the **AdminAgents** group in Business Central Admin Center

## What You Will Build

By the end of this tutorial you will have:

- A **production** Business Central environment in the US region, using the latest production ring
- An **environment settings** resource that configures an update window and timezone
- An **environment support contact** resource that sets the tenant's support information
- Terraform **outputs** for key environment attributes (web client URL, version, etc.)

## Step 1: Configure Authentication

Export your service principal credentials as environment variables so they are not stored in source code:

```bash
export ARM_CLIENT_ID="00000000-0000-0000-0000-000000000000"
export ARM_CLIENT_SECRET="your-client-secret"
export ARM_TENANT_ID="00000000-0000-0000-0000-000000000000"
```

## Step 2: Initialize the Provider

Create a `versions.tf` file:

```terraform
# versions.tf
terraform {
  required_version = ">= 1.0"

  required_providers {
    bcadmincenter = {
      source  = "axiansinfoma/bcadmincenter"
      version = "~> 1.0"
    }
  }
}

provider "bcadmincenter" {
  # Credentials are read from ARM_CLIENT_ID, ARM_CLIENT_SECRET, ARM_TENANT_ID
}
```

Run `terraform init` to install the provider:

```bash
terraform init
```

## Step 3: Discover Available Application Families and Rings

Instead of hard-coding `ring_name` and `application_family`, use data sources to discover valid values at plan time. This ensures your configuration remains correct when Microsoft updates available rings or countries.

```terraform
# data.tf

# Query all available application families, countries, and rings
data "bcadmincenter_available_applications" "apps" {}

locals {
  # Select the first (and typically only) application family
  app_family = data.bcadmincenter_available_applications.apps.application_families[0]

  # Find the country configuration for "US"
  us_country = one([
    for c in local.app_family.countries_ring_details :
    c if c.country_code == "US"
  ])

  # Select the production ring (production_ring == true)
  production_ring = one([
    for r in local.us_country.rings :
    r if r.production_ring
  ])
}

# Discover available timezones for configuring environment settings
data "bcadmincenter_timezones" "available" {}

locals {
  pacific_tz = one([
    for tz in data.bcadmincenter_timezones.available.timezones :
    tz if tz.id == "Pacific Standard Time"
  ])
}
```

Verify the discovered values before proceeding:

```bash
terraform plan -target=data.bcadmincenter_available_applications.apps \
               -target=data.bcadmincenter_timezones.available
```

## Step 4: Create the Production Environment with Settings

Create a `main.tf` file that references the data source outputs:

```terraform
# main.tf

resource "bcadmincenter_environment" "production" {
  name               = "production"
  application_family = local.app_family.name
  type               = "Production"
  country_code       = local.us_country.country_code
  ring_name          = local.production_ring.name
  azure_region       = "westus2"

  # Configure environment settings inline
  settings {
    # Schedule updates during off-peak hours (window must be at least 6 hours)
    update_window_start_time = "22:00"
    update_window_end_time   = "06:00"
    update_window_timezone   = local.pacific_tz.id

    # Only accept updates during major upgrades (conservative cadence)
    app_update_cadence = "DuringMajorUpgrade"
  }

  timeouts {
    create = "90m"
    delete = "60m"
  }
}
```

~> **Warning:** Environment creation is an asynchronous operation that typically takes 15–30 minutes.
The provider polls the API and blocks until the environment reaches `Active` status or the timeout expires.

## Step 5: Set a Support Contact

```terraform
# main.tf (continued)

resource "bcadmincenter_environment_support_contact" "production" {
  application_family = bcadmincenter_environment.production.application_family
  environment_name   = bcadmincenter_environment.production.name

  name  = "Contoso IT Support"
  email = "bc-support@contoso.com"
  url   = "https://support.contoso.com"
}
```

## Step 6: Add Outputs

```terraform
# outputs.tf

output "environment_name" {
  description = "Name of the provisioned environment"
  value       = bcadmincenter_environment.production.name
}

output "web_client_url" {
  description = "URL to access the Business Central web client"
  value       = bcadmincenter_environment.production.web_client_login_url
}

output "application_version" {
  description = "Business Central application version (assigned by the API)"
  value       = bcadmincenter_environment.production.application_version
}

output "ring_used" {
  description = "Release ring the environment was created in"
  value       = local.production_ring.name
}
```

## Step 7: Apply the Configuration

Preview the planned changes:

```bash
terraform plan
```

Apply when satisfied:

```bash
terraform apply
```

After a successful apply you will see output similar to:

```
Outputs:

application_version = "25.3.0.0"
environment_name    = "production"
ring_used           = "PROD"
web_client_url      = "https://businesscentral.dynamics.com/00000000-0000-0000-0000-000000000000/production"
```

## Step 8: Import an Existing Environment

If you have an environment that was created outside Terraform, you can import it into state:

```bash
terraform import bcadmincenter_environment.production \
  "/tenants/YOUR_TENANT_ID/providers/Microsoft.Dynamics365.BusinessCentral/applications/BusinessCentral/environments/production"
```

After importing, run `terraform plan` to verify that the live configuration matches your Terraform code. Adjust any attributes that show a diff.

## Complete Example

Below is the full working configuration combining all the steps above:

```terraform
# versions.tf
terraform {
  required_version = ">= 1.0"
  required_providers {
    bcadmincenter = {
      source  = "axiansinfoma/bcadmincenter"
      version = "~> 1.0"
    }
  }
}

provider "bcadmincenter" {}

# data.tf
data "bcadmincenter_available_applications" "apps" {}
data "bcadmincenter_timezones" "available" {}

locals {
  app_family      = data.bcadmincenter_available_applications.apps.application_families[0]
  us_country      = one([for c in local.app_family.countries_ring_details : c if c.country_code == "US"])
  production_ring = one([for r in local.us_country.rings : r if r.production_ring])
  pacific_tz      = one([for tz in data.bcadmincenter_timezones.available.timezones : tz if tz.id == "Pacific Standard Time"])
}

# main.tf
resource "bcadmincenter_environment" "production" {
  name               = "production"
  application_family = local.app_family.name
  type               = "Production"
  country_code       = local.us_country.country_code
  ring_name          = local.production_ring.name
  azure_region       = "westus2"

  settings {
    update_window_start_time = "22:00"
    update_window_end_time   = "06:00"
    update_window_timezone   = local.pacific_tz.id
    app_update_cadence       = "DuringMajorUpgrade"
  }

  timeouts {
    create = "90m"
    delete = "60m"
  }
}

resource "bcadmincenter_environment_support_contact" "production" {
  application_family = bcadmincenter_environment.production.application_family
  environment_name   = bcadmincenter_environment.production.name
  name               = "Contoso IT Support"
  email              = "bc-support@contoso.com"
  url                = "https://support.contoso.com"
}

# outputs.tf
output "web_client_url"      { value = bcadmincenter_environment.production.web_client_login_url }
output "application_version" { value = bcadmincenter_environment.production.application_version }
output "ring_used"           { value = local.production_ring.name }
```

## Troubleshooting

### Environment creation timed out

Increase the `create` timeout and re-run `terraform apply`. The provider will resume polling the existing operation rather than starting a new one.

### `ring_name` is invalid

Run `terraform plan` against only the data source to print available ring names:

```bash
terraform plan -target=data.bcadmincenter_available_applications.apps
```

### Permission denied

Verify that:
1. Admin consent was granted for `AdminCenter.ReadWrite.All`
2. The service principal is a member of the AdminAgents group
3. You are authenticating to the correct tenant

## Next Steps

- [Multi-Tenant Management Tutorial](./multi-tenant-management.md) – iterate over multiple tenants
- [Environment resource](../resources/environment.md) – full attribute reference including the `settings` block
- [Environment Support Contact resource](../resources/environment_support_contact.md) – full attribute reference

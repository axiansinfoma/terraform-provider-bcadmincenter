# Tutorial: Multi-Tenant Management

This tutorial shows how to manage Business Central environments across **multiple tenants** using the Business Central Admin Center Terraform provider. You will learn how to:

- Discover all tenants your application can manage using the `bcadmincenter_manageable_tenants` data source
- Iterate over tenants to create a sandbox environment in each one
- Use provider aliases for explicit per-tenant configuration
- Import environments that were already created outside Terraform

## Prerequisites

- Terraform 1.0 or later installed
- A service principal (app registration) that has been added to the **AdminAgents** group **in every tenant** you want to manage
- `AdminCenter.ReadWrite.All` permission granted and admin-consented in each tenant (or via a multi-tenant app registration)

For setup instructions, see the [Service Principal authentication tutorial](./service-principal-authentication.md).

## Overview of Approaches

There are two patterns for multi-tenant management:

| Pattern | When to Use |
|---------|-------------|
| **Dynamic iteration** (this tutorial, Steps 1–4) | You manage an open-ended list of tenants; new tenants onboard by being added to the authorized list |
| **Explicit provider aliases** (Steps 5–6) | You have a small, fixed set of tenants and want the most readable, explicit configuration |

---

## Pattern 1: Dynamic Tenant Iteration

### Step 1: Discover Manageable Tenants

The `bcadmincenter_manageable_tenants` data source lists all tenants where your app has been granted access. This is only available when authenticating as an application (client credentials flow).

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
  # Uses AZURE_CLIENT_ID, AZURE_CLIENT_SECRET, AZURE_TENANT_ID
  # This must be the "home" tenant of the multi-tenant app registration.
}
```

```terraform
# data.tf

# Returns all tenants where this app is an AdminAgent
data "bcadmincenter_manageable_tenants" "all" {}

# Also discover the production ring once – it is the same across all tenants
data "bcadmincenter_available_applications" "apps" {}

locals {
  tenant_ids = [
    for t in data.bcadmincenter_manageable_tenants.all.tenants :
    t.entra_tenant_id
  ]

  app_family      = data.bcadmincenter_available_applications.apps.application_families[0]
  us_country      = one([for c in local.app_family.countries_ring_details : c if c.country_code == "US"])
  production_ring = one([for r in local.us_country.rings : r if r.production_ring])
}
```

Verify the list of discovered tenants:

```bash
terraform plan -target=data.bcadmincenter_manageable_tenants.all
```

Expected output:

```
Changes to Outputs:
  + manageable_tenant_ids = [
      "aaaaaaaa-0000-0000-0000-000000000001",
      "bbbbbbbb-0000-0000-0000-000000000002",
      "cccccccc-0000-0000-0000-000000000003",
    ]
```

### Step 2: Create a Sandbox in Each Tenant

Use `for_each` over the tenant list to create one sandbox environment per tenant. The `aad_tenant_id` attribute tells the provider which tenant to target for each resource.

```terraform
# main.tf

resource "bcadmincenter_environment" "sandbox" {
  for_each = toset(local.tenant_ids)

  aad_tenant_id      = each.key
  name               = "sandbox"
  application_family = local.app_family.name
  type               = "Sandbox"
  country_code       = local.us_country.country_code
  ring_name          = local.production_ring.name
  azure_region       = "westus2"

  timeouts {
    create = "90m"
    delete = "60m"
  }
}
```

### Step 3: Output the Results

```terraform
# outputs.tf

output "sandbox_urls" {
  description = "Web client URLs for the sandbox in each tenant"
  value = {
    for tenant_id, env in bcadmincenter_environment.sandbox :
    tenant_id => env.web_client_login_url
  }
}

output "sandbox_versions" {
  description = "Application version deployed in each tenant's sandbox"
  value = {
    for tenant_id, env in bcadmincenter_environment.sandbox :
    tenant_id => env.application_version
  }
}
```

Apply the configuration:

```bash
terraform apply
```

### Step 4: Import Existing Sandboxes

If a sandbox environment already exists in one or more tenants, import it instead of recreating it:

```bash
# Format: /tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/applications/{appFamily}/environments/{envName}

terraform import 'bcadmincenter_environment.sandbox["aaaaaaaa-0000-0000-0000-000000000001"]' \
  "/tenants/aaaaaaaa-0000-0000-0000-000000000001/providers/Microsoft.Dynamics365.BusinessCentral/applications/BusinessCentral/environments/sandbox"

terraform import 'bcadmincenter_environment.sandbox["bbbbbbbb-0000-0000-0000-000000000002"]' \
  "/tenants/bbbbbbbb-0000-0000-0000-000000000002/providers/Microsoft.Dynamics365.BusinessCentral/applications/BusinessCentral/environments/sandbox"
```

After importing, run `terraform plan` to verify state matches your configuration:

```bash
terraform plan
```

Adjust any attribute differences in your Terraform code, then apply:

```bash
terraform apply
```

---

## Pattern 2: Explicit Provider Aliases

Use provider aliases when you have a fixed set of tenants and prefer explicit, readable configuration.

### Step 5: Declare Provider Aliases

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

# Tenant A
provider "bcadmincenter" {
  alias     = "tenant_a"
  client_id = var.client_id
  tenant_id = var.tenant_a_id
  # client_secret read from AZURE_CLIENT_SECRET
}

# Tenant B
provider "bcadmincenter" {
  alias     = "tenant_b"
  client_id = var.client_id
  tenant_id = var.tenant_b_id
}
```

```terraform
# variables.tf

variable "client_id" {
  description = "Azure AD application (client) ID"
  type        = string
}

variable "tenant_a_id" {
  description = "Azure AD tenant ID for Tenant A"
  type        = string
}

variable "tenant_b_id" {
  description = "Azure AD tenant ID for Tenant B"
  type        = string
}
```

### Step 6: Create Resources per Tenant

```terraform
# main.tf

resource "bcadmincenter_environment" "tenant_a_production" {
  provider = bcadmincenter.tenant_a

  name               = "production"
  application_family = "BusinessCentral"
  type               = "Production"
  country_code       = "US"
  ring_name          = "PROD"
  azure_region       = "westus2"
}

resource "bcadmincenter_environment" "tenant_b_production" {
  provider = bcadmincenter.tenant_b

  name               = "production"
  application_family = "BusinessCentral"
  type               = "Production"
  country_code       = "US"
  ring_name          = "PROD"
  azure_region       = "eastus"
}
```

Import existing environments for explicit aliases:

```bash
terraform import bcadmincenter_environment.tenant_a_production \
  "/tenants/TENANT_A_ID/providers/Microsoft.Dynamics365.BusinessCentral/applications/BusinessCentral/environments/production"

terraform import bcadmincenter_environment.tenant_b_production \
  "/tenants/TENANT_B_ID/providers/Microsoft.Dynamics365.BusinessCentral/applications/BusinessCentral/environments/production"
```

---

## Security Considerations

1. **Use a multi-tenant app registration** so a single service principal can act across tenants – this avoids maintaining separate credentials per tenant.
2. **Limit the AdminAgents group** to only the service principals that need programmatic access.
3. **Store credentials in a secrets manager** (Azure Key Vault, HashiCorp Vault, or your CI/CD platform's secrets) – never commit them to source control.
4. **Use Workload Identity** where possible to eliminate the need for stored client secrets entirely. See the [Workload Identity for GitHub Actions tutorial](./workload-identity-github.md).

## Troubleshooting

### `bcadmincenter_manageable_tenants` returns an empty list

- Confirm the app is authenticating with client credentials (not delegated/user credentials).
- Verify the service principal has been added to AdminAgents in each target tenant.
- Check that admin consent was granted for `AdminCenter.ReadWrite.All`.

### Import fails with "resource not found"

- Verify the tenant ID, application family, and environment name in the import ID.
- Confirm the environment exists in the Business Central Admin Center for that tenant.
- Check that your service principal has access to the target tenant.

### `for_each` plan shows unexpected additions or deletions

- If tenants are added or removed from the `manageable_tenants` list, Terraform will plan to create or destroy the corresponding environments.
- Use `terraform state mv` to rename resources in state if a tenant ID changes.

## Next Steps

- [Full Environment Tutorial](./full-environment-tutorial.md) – provision a complete environment with settings and support contact
- [Service Principal Authentication](./service-principal-authentication.md) – set up a multi-tenant app registration
- [`bcadmincenter_manageable_tenants` data source](../docs/data-sources/manageable_tenants.md) – full attribute reference
- [`bcadmincenter_environment` resource](../docs/resources/environment.md) – full attribute reference

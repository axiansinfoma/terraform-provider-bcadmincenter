# Business Central Admin Center - Environment Examples

This directory contains examples for managing Business Central environments using Terraform.

## Examples

### Basic Production Environment
```terraform
resource "bcadmincenter_environment" "production" {
  name               = "production"
  application_family = "BusinessCentral"
  type               = "Production"
  country_code       = "US"
}
```

### Sandbox Environment with Specific Version
```terraform
resource "bcadmincenter_environment" "dev" {
  name                = "development"
  application_family  = "BusinessCentral"
  type                = "Sandbox"
  country_code        = "US"
  ring_name           = "Production"
  application_version = "24.0"
  azure_region        = "eastus"

  timeouts {
    create = "90m"
  }
}
```

### Multiple Environments
```terraform
variable "environments" {
  type = map(object({
    type         = string
    country_code = string
    azure_region = string
  }))
  default = {
    production = {
      type         = "Production"
      country_code = "US"
      azure_region = "westus2"
    }
    staging = {
      type         = "Sandbox"
      country_code = "US"
      azure_region = "westus2"
    }
    development = {
      type         = "Sandbox"
      country_code = "US"
      azure_region = "eastus"
    }
  }
}

resource "bcadmincenter_environment" "environments" {
  for_each = var.environments

  name               = each.key
  application_family = "BusinessCentral"
  type               = each.value.type
  country_code       = each.value.country_code
  azure_region       = each.value.azure_region
  ring_name          = "Production"

  timeouts {
    create = "90m"
    delete = "60m"
  }
}

output "environment_urls" {
  value = {
    for name, env in bcadmincenter_environment.environments :
    name => env.web_client_login_url
  }
  description = "Web client URLs for all environments"
}
```

## Prerequisites

1. **Azure AD Application Registration**:
   - Register an application in Azure AD
   - Add the required API permissions for Business Central Admin Center
   - Create a client secret

2. **Admin Agent Access**:
   - The application must be added to the AdminAgents group
   - This grants delegated admin access to Business Central

3. **Environment Variables**:
   ```bash
   export AZURE_CLIENT_ID="00000000-0000-0000-0000-000000000000"
   export AZURE_CLIENT_SECRET="your-client-secret"
   export AZURE_TENANT_ID="00000000-0000-0000-0000-000000000000"
   ```

## Running the Examples

1. Initialize Terraform:
   ```bash
   terraform init
   ```

2. Review the planned changes:
   ```bash
   terraform plan
   ```

3. Apply the configuration:
   ```bash
   terraform apply
   ```

4. View outputs:
   ```bash
   terraform output
   ```

## Important Notes

- **Environment Creation Time**: Creating an environment typically takes 15-30 minutes
- **Force Replacement**: Most attributes cannot be changed after creation and will force a new resource
- **Production Environments**: Have additional restrictions and cannot be easily deleted
- **Naming**: Environment names must be unique within the tenant and are 1-30 characters
- **Country Codes**: Use ISO 3166-1 alpha-2 codes (e.g., US, GB, DK, DE)
- **Azure Regions**: Choose regions close to your users for better performance

## Cleanup

To destroy the created environments:

```bash
terraform destroy
```

**Warning**: Deleting an environment is permanent and cannot be undone. Ensure you have backups if needed.

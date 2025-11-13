---
page_title: "Provider: Business Central Admin Center"
description: |-
  The Business Central Admin Center provider enables Infrastructure as Code (IaC) management of Microsoft Dynamics 365 Business Central environments.
---

# Business Central Admin Center Provider

The Business Central Admin Center provider enables Infrastructure as Code (IaC) management of Microsoft Dynamics 365 Business Central environments through the [Business Central Admin Center API](https://learn.microsoft.com/en-us/dynamics365/business-central/dev-itpro/administration/administration-center-api).

## Features

- **Environment Management**: Create, update, and delete Business Central production and sandbox environments
- **Configuration Management**: Configure environment settings, access controls, and telemetry
- **Application Management**: Manage application installations and updates
- **Administrative Operations**: Configure notifications, query operations, and monitor quotas
- **Multiple Authentication Methods**: Support for service principals, managed identities, Azure CLI, and more

## Authentication

The provider supports multiple authentication methods via the Azure SDK for Go, following the same patterns as the AzureRM provider:

### Supported Authentication Methods

1. **[Service Principal with Client Secret](https://github.com/vllni/terraform-provider-bcadmincenter/blob/main/tutorials/service-principal-authentication.md)** - For automated scenarios
2. **[Workload Identity Federation](https://github.com/vllni/terraform-provider-bcadmincenter/blob/main/tutorials/workload-identity-github.md)** - Recommended for CI/CD (GitHub Actions, Azure DevOps)
3. **Service Principal with Certificate** - For enhanced security
4. **[Managed Identity](https://github.com/vllni/terraform-provider-bcadmincenter/blob/main/tutorials/managed-identity-authentication.md)** - For Azure-hosted environments (VMs, Container Instances, App Service)
5. **[Azure CLI Authentication](https://github.com/vllni/terraform-provider-bcadmincenter/blob/main/tutorials/azure-cli-authentication.md)** - For local development
6. **Device Code Flow** - For interactive scenarios

-> **Detailed Setup Guides**: For comprehensive step-by-step tutorials on setting up each authentication method, see the [Authentication Tutorials](https://github.com/vllni/terraform-provider-bcadmincenter/tree/main/tutorials) in the GitHub repository.

### Required Permissions

To use this provider, you need:

- **AdminCenter.ReadWrite.All** permission on the "Dynamics 365 Business Central administration center" API (Application ID: `996def3d-b36c-4153-8607-a6fd3c01b89f`)
- Membership in the **AdminAgents** group for delegated admin access to Business Central tenants
- Appropriate Azure AD tenant access

### Setting Up Authentication

#### Azure AD Application Registration

1. Register an application in Azure AD:
   ```bash
   az ad app create --display-name "Terraform BC Admin Center"
   ```

2. Create a service principal:
   ```bash
   az ad sp create --id <application-id>
   ```

3. Grant the required API permissions:
   ```bash
   # Add the AdminCenter.ReadWrite.All permission
   az ad app permission add --id <application-id> \
     --api 996def3d-b36c-4153-8607-a6fd3c01b89f \
     --api-permissions 2e3cf0a5-be71-42b6-8b82-6f50da52005d=Role
   
   # Grant admin consent
   az ad app permission grant --id <application-id> \
     --api 996def3d-b36c-4153-8607-a6fd3c01b89f
   ```

4. Create a client secret:
   ```bash
   az ad app credential reset --id <application-id>
   ```

5. Add the service principal to the AdminAgents group in your Business Central Admin Center

## Example Usage

### Service Principal with Client Secret

```terraform
# Copyright (c) 2025 Michael Villani
# SPDX-License-Identifier: MPL-2.0

terraform {
  required_providers {
    bcadmincenter = {
      source = "vllni/bcadmincenter"
    }
  }
}

# Example 1: Service Principal with Client Secret
provider "bcadmincenter" {
  client_id     = "00000000-0000-0000-0000-000000000000"
  client_secret = "your-client-secret"
  tenant_id     = "00000000-0000-0000-0000-000000000000"

  # Optional: Override default settings
  environment = "public" # public, usgovernment, china
}

# Example 2: Using environment variables for authentication
# Set these environment variables:
# AZURE_CLIENT_ID
# AZURE_CLIENT_SECRET
# AZURE_TENANT_ID
# AZURE_ENVIRONMENT (optional)
#
# provider "bcadmincenter" {
#   # Configuration will be automatically picked up from environment
# }

# Example 3: Azure Workload Identity (Recommended for CI/CD in Kubernetes)
# When running in a Kubernetes cluster with Azure Workload Identity enabled,
# set these environment variables:
# AZURE_CLIENT_ID
# AZURE_TENANT_ID
# AZURE_FEDERATED_TOKEN_FILE
# AZURE_AUTHORITY_HOST
#
# provider "bcadmincenter" {
#   # Provider automatically detects and uses workload identity credentials
# }
```

### Using Environment Variables

The provider supports the following environment variables (following Azure conventions):

```bash
export AZURE_CLIENT_ID="00000000-0000-0000-0000-000000000000"
export AZURE_CLIENT_SECRET="client-secret-value"
export AZURE_TENANT_ID="00000000-0000-0000-0000-000000000000"
```

When these are set, you can use a simplified provider configuration:

```terraform
provider "bcadmincenter" {
  # Authentication will use environment variables
}
```

### Azure CLI Authentication (Local Development)

For local development, you can authenticate using the Azure CLI:

```bash
az login
```

Then use the provider without explicit credentials:

```terraform
provider "bcadmincenter" {
  tenant_id = "00000000-0000-0000-0000-000000000000"
  # Client credentials will be obtained from Azure CLI
}
```

### Managed Identity (Azure-Hosted)

When running on Azure compute resources with managed identity enabled:

```terraform
provider "bcadmincenter" {
  use_msi   = true
  tenant_id = "00000000-0000-0000-0000-000000000000"
}
```

### Workload Identity (Kubernetes/CI-CD)

For Kubernetes workloads using Azure Workload Identity:

```terraform
provider "bcadmincenter" {
  use_oidc  = true
  tenant_id = "00000000-0000-0000-0000-000000000000"
  client_id = "00000000-0000-0000-0000-000000000000"
}
```

Environment variables for workload identity:
- `AZURE_FEDERATED_TOKEN_FILE` - Path to the federated token file
- `AZURE_AUTHORITY_HOST` - Azure Active Directory authority host
- `AZURE_CLIENT_ID` - Client ID of the user-assigned managed identity

## Multi-Tenant Scenarios

If you need to manage environments across multiple Business Central tenants, use provider aliases:

```terraform
provider "bcadmincenter" {
  alias     = "tenant1"
  client_id = var.client_id
  tenant_id = var.tenant1_id
}

provider "bcadmincenter" {
  alias     = "tenant2"
  client_id = var.client_id
  tenant_id = var.tenant2_id
}

resource "bcadmincenter_environment" "tenant1_prod" {
  provider = bcadmincenter.tenant1
  name     = "production"
  # ... other configuration
}

resource "bcadmincenter_environment" "tenant2_prod" {
  provider = bcadmincenter.tenant2
  name     = "production"
  # ... other configuration
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `auxiliary_tenant_ids` (List of String) List of auxiliary tenant IDs for multi-tenant scenarios.
- `client_id` (String) The Client ID (Application ID) for Azure AD authentication. Can also be set via AZURE_CLIENT_ID environment variable.
- `client_secret` (String, Sensitive) The Client Secret for Azure AD authentication. Can also be set via AZURE_CLIENT_SECRET environment variable.
- `environment` (String) The Azure environment to use (public, usgovernment, china). Defaults to 'public'. Can also be set via AZURE_ENVIRONMENT environment variable.
- `tenant_id` (String) The Tenant ID for Azure AD authentication. Can also be set via AZURE_TENANT_ID environment variable.

## Environment Variables

The following environment variables can be used to configure the provider:

| Variable | Description |
|----------|-------------|
| `AZURE_CLIENT_ID` | The client ID for service principal authentication |
| `AZURE_CLIENT_SECRET` | The client secret for service principal authentication |
| `AZURE_TENANT_ID` | The Azure AD tenant ID |
| `AZURE_ENVIRONMENT` | The Azure cloud environment (`public`, `usgovernment`, `china`) |
| `AZURE_FEDERATED_TOKEN_FILE` | Path to federated token file (for workload identity) |
| `AZURE_AUTHORITY_HOST` | Azure AD authority host URL |
| `AZURE_CLIENT_ASSERTION` | Client assertion for federated identity credentials |

## Authentication Tutorials

For detailed step-by-step guides on setting up authentication, visit our comprehensive tutorials:

- **[Service Principal with Client Secret](https://github.com/vllni/terraform-provider-bcadmincenter/blob/main/tutorials/service-principal-authentication.md)** - Traditional authentication for automation and CI/CD
- **[Azure CLI Authentication](https://github.com/vllni/terraform-provider-bcadmincenter/blob/main/tutorials/azure-cli-authentication.md)** - Quick setup for local development
- **[Managed Identity](https://github.com/vllni/terraform-provider-bcadmincenter/blob/main/tutorials/managed-identity-authentication.md)** - Secure authentication for Azure VMs and containers
- **[Workload Identity for GitHub Actions](https://github.com/vllni/terraform-provider-bcadmincenter/blob/main/tutorials/workload-identity-github.md)** - OIDC-based authentication for GitHub workflows
- **[Workload Identity for Azure DevOps](https://github.com/vllni/terraform-provider-bcadmincenter/blob/main/tutorials/workload-identity-azure-devops.md)** - Federated credentials for Azure Pipelines

Each tutorial includes complete setup instructions, troubleshooting tips, and real-world examples.

## Additional Resources

- [Business Central Admin Center API Documentation](https://learn.microsoft.com/en-us/dynamics365/business-central/dev-itpro/administration/administration-center-api)
- [Azure AD Authentication Overview](https://learn.microsoft.com/en-us/azure/active-directory/develop/authentication-scenarios)
- [Terraform Provider Development](https://developer.hashicorp.com/terraform/plugin)
- [Provider Source Code](https://github.com/vllni/terraform-provider-bcadmincenter)

## Support

For issues, feature requests, or contributions, please visit the [GitHub repository](https://github.com/vllni/terraform-provider-bcadmincenter).

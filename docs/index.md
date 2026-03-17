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
- **Administrative Operations**: Configure notifications and monitor quotas
- **Multiple Authentication Methods**: Support for service principals, managed identities, Azure CLI, and more

## Authentication

The provider supports multiple authentication methods via the Azure SDK for Go, following the same patterns as the AzureRM provider.

-> **Step-by-step guides**: For comprehensive tutorials on setting up each authentication method, see the [Authentication Guides](guides/service-principal-authentication.md).

### Required Permissions

To use this provider, you need:

- **AdminCenter.ReadWrite.All** permission on the "Dynamics 365 Business Central administration center" API (Application ID: `996def3d-b36c-4153-8607-a6fd3c01b89f`)
- Membership in the **AdminAgents** group for delegated admin access to Business Central tenants
- Appropriate Azure AD tenant access

### Setting Up an Azure AD Application

Before authenticating, register an application in Azure AD and grant it the required permissions:

```bash
# 1. Create the application
APP_ID=$(az ad app create --display-name "Terraform BC Admin Center" --query appId --output tsv)

# 2. Create a service principal
az ad sp create --id $APP_ID

# 3. Grant AdminCenter.ReadWrite.All permission and consent
BC_API="996def3d-b36c-4153-8607-a6fd3c01b89f"
az ad app permission add --id $APP_ID \
  --api $BC_API \
  --api-permissions 2e3cf0a5-be71-42b6-8b82-6f50da52005d=Role
az ad app permission admin-consent --id $APP_ID
```

Then add the service principal to the **AdminAgents** group in the [Business Central Admin Center](https://businesscentral.dynamics.com/admin).

### Service Principal with Client Secret

```terraform
# Copyright (c) 2025 Axians Infoma GmbH
# SPDX-License-Identifier: MPL-2.0

# Service Principal with Client Secret authentication.
# Recommended for automated pipelines where Azure CLI or workload identity are not available.

terraform {
  required_providers {
    bcadmincenter = {
      source = "axiansinfoma/bcadmincenter"
    }
  }
}

provider "bcadmincenter" {
  client_id     = "00000000-0000-0000-0000-000000000000"
  client_secret = "your-client-secret"
  tenant_id     = "00000000-0000-0000-0000-000000000000"
}
```

### Environment Variables

Set credentials as environment variables and leave the provider block empty:

```bash
export AZURE_CLIENT_ID="00000000-0000-0000-0000-000000000000"
export AZURE_CLIENT_SECRET="your-client-secret"
export AZURE_TENANT_ID="00000000-0000-0000-0000-000000000000"
```

```terraform
# Copyright (c) 2025 Axians Infoma GmbH
# SPDX-License-Identifier: MPL-2.0

# Authentication via environment variables.
# Set the following variables in your shell before running Terraform:
#
#   export AZURE_CLIENT_ID="00000000-0000-0000-0000-000000000000"
#   export AZURE_CLIENT_SECRET="your-client-secret"
#   export AZURE_TENANT_ID="00000000-0000-0000-0000-000000000000"

terraform {
  required_providers {
    bcadmincenter = {
      source = "axiansinfoma/bcadmincenter"
    }
  }
}

provider "bcadmincenter" {
  # All configuration is picked up automatically from the environment variables above.
}
```

### Azure CLI (Local Development)

Authenticate with `az login` and configure the provider to use the cached credentials:

```bash
az login --tenant 00000000-0000-0000-0000-000000000000
```

```terraform
# Copyright (c) 2025 Axians Infoma GmbH
# SPDX-License-Identifier: MPL-2.0

# Azure CLI authentication – recommended for local development.
# Log in before running Terraform:
#
#   az login --tenant 00000000-0000-0000-0000-000000000000

terraform {
  required_providers {
    bcadmincenter = {
      source = "axiansinfoma/bcadmincenter"
    }
  }
}

provider "bcadmincenter" {
  tenant_id = "00000000-0000-0000-0000-000000000000"
  use_cli   = true
}
```

See the full [Azure CLI authentication guide](guides/azure-cli-authentication.md) for complete setup instructions.

### Managed Identity (Azure-Hosted)

When Terraform runs on Azure compute (VM, Container Instance, App Service) with a managed identity:

```terraform
# Copyright (c) 2025 Axians Infoma GmbH
# SPDX-License-Identifier: MPL-2.0

# Managed Identity authentication – for Terraform running on Azure compute
# (VMs, Container Instances, App Service) with a system-assigned or
# user-assigned managed identity enabled.

terraform {
  required_providers {
    bcadmincenter = {
      source = "axiansinfoma/bcadmincenter"
    }
  }
}

provider "bcadmincenter" {
  tenant_id = "00000000-0000-0000-0000-000000000000"
  use_msi   = true

  # Uncomment to use a specific user-assigned managed identity:
  # client_id = "00000000-0000-0000-0000-000000000000"
}
```

See the full [Managed Identity guide](guides/managed-identity-authentication.md).

### Workload Identity (CI/CD)

The recommended method for GitHub Actions, Azure DevOps, and Kubernetes workloads:

```terraform
# Copyright (c) 2025 Axians Infoma GmbH
# SPDX-License-Identifier: MPL-2.0

# Workload Identity (OIDC) authentication – recommended for CI/CD pipelines
# on GitHub Actions, Azure DevOps, and Kubernetes with Azure Workload Identity.
#
# The following environment variables must be set by your CI/CD platform:
#   AZURE_CLIENT_ID            – application (client) ID
#   AZURE_TENANT_ID            – Azure AD tenant ID
#   AZURE_FEDERATED_TOKEN_FILE – path to the OIDC token file

terraform {
  required_providers {
    bcadmincenter = {
      source = "axiansinfoma/bcadmincenter"
    }
  }
}

provider "bcadmincenter" {
  tenant_id = "00000000-0000-0000-0000-000000000000"
  client_id = "00000000-0000-0000-0000-000000000000"
  use_oidc  = true
}
```

- [Workload Identity for GitHub Actions guide](guides/workload-identity-github.md)
- [Workload Identity for Azure DevOps guide](guides/workload-identity-azure-devops.md)
- [Service Connection with Client Secret for Azure DevOps guide](guides/azure-devops-service-connection-secret.md)

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `ado_pipeline_service_connection_id` (String) The Azure DevOps service connection ID used when authenticating via Azure DevOps Pipeline OIDC (`SYSTEM_OIDCREQUESTURI`). When set, the provider uses the ADO OIDC endpoint protocol (POST with `serviceConnectionId` and `api-version` query parameters) instead of the GitHub Actions endpoint. Can also be set via `ARM_ADO_PIPELINE_SERVICE_CONNECTION_ID` or `ARM_OIDC_AZURE_SERVICE_CONNECTION_ID` environment variable.
- `auxiliary_tenant_ids` (List of String) List of auxiliary tenant IDs for multi-tenant scenarios.
- `base_url` (String) Override the base URL for the Business Central Admin Center API. Can also be set via BCADMINCENTER_BASE_URL environment variable. Primarily used for testing.
- `client_id` (String) The Client ID (Application ID) for Azure AD authentication. Can also be set via AZURE_CLIENT_ID environment variable.
- `client_secret` (String, Sensitive) The Client Secret for Azure AD authentication. Can also be set via AZURE_CLIENT_SECRET environment variable.
- `environment` (String) The Azure environment to use (public, usgovernment, china). Defaults to 'public'. Can also be set via AZURE_ENVIRONMENT environment variable.
- `oidc_request_token` (String, Sensitive) The bearer token used to authenticate requests to `oidc_request_url`. In GitHub Actions this is set automatically via `ACTIONS_ID_TOKEN_REQUEST_TOKEN`; in Azure DevOps via `SYSTEM_ACCESSTOKEN`. Can also be set via `ARM_OIDC_REQUEST_TOKEN`, `ACTIONS_ID_TOKEN_REQUEST_TOKEN`, or `SYSTEM_ACCESSTOKEN` environment variable.
- `oidc_request_url` (String) The URL of the OIDC token endpoint. In GitHub Actions this is set automatically via `ACTIONS_ID_TOKEN_REQUEST_URL`; in Azure DevOps via `SYSTEM_OIDCREQUESTURI`. A fresh JWT is fetched from this endpoint on every Azure AD token refresh, so short-lived tokens are automatically renewed during long Terraform runs. Can also be set via `ARM_OIDC_REQUEST_URL`, `ACTIONS_ID_TOKEN_REQUEST_URL`, or `SYSTEM_OIDCREQUESTURI` environment variable.
- `oidc_token` (String, Sensitive) A JWT bearer token to use as the OIDC client assertion. Useful when the token is provided directly by the CI/CD platform. Can also be set via `ARM_OIDC_TOKEN` or `AZURE_OIDC_TOKEN` environment variable. Setting this implies `use_oidc = true`.
- `oidc_token_file_path` (String) Path to a file containing the OIDC / federated token. The file is re-read on every Azure AD token refresh so platform-rotated tokens (e.g. Kubernetes projected volumes) are picked up automatically. Can also be set via `ARM_OIDC_TOKEN_FILE_PATH` or `AZURE_FEDERATED_TOKEN_FILE` environment variable. Used when `use_oidc = true`.
- `tenant_id` (String) The Tenant ID for Azure AD authentication. Can also be set via AZURE_TENANT_ID environment variable.
- `use_oidc` (Boolean) Force the use of OIDC / Workload Identity (federated credential) authentication. When true, the provider uses `WorkloadIdentityCredential` from the Azure SDK, which reads the federated token from the file specified by `oidc_token_file_path` (or `AZURE_FEDERATED_TOKEN_FILE`). Can also be set via `ARM_USE_OIDC` or `AZURE_USE_OIDC` environment variable.

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

## Guides

End-to-end guides for common use cases:

| Guide | Description |
|-------|-------------|
| [Provision an Environment](guides/full-environment-tutorial.md) | Step-by-step guide to provision a Business Central environment, reading ring and country information from data sources |
| [Multi-Tenant Management](guides/multi-tenant-management.md) | Manage environments across multiple Business Central tenants using iteration and import workflows |
| [Service Principal](guides/service-principal-authentication.md) | Set up service principal authentication with a client secret |
| [Azure CLI](guides/azure-cli-authentication.md) | Quick setup for local development using Azure CLI |
| [Managed Identity](guides/managed-identity-authentication.md) | Secure authentication for Azure VMs and containers |
| [Workload Identity – GitHub Actions](guides/workload-identity-github.md) | OIDC-based authentication for GitHub workflows |
| [Workload Identity – Azure DevOps](guides/workload-identity-azure-devops.md) | Federated credentials for Azure Pipelines |
| [Service Connection – Azure DevOps (Client Secret)](guides/azure-devops-service-connection-secret.md) | Service connection authentication for Azure Pipelines using a client secret |

## Additional Resources

- [Business Central Admin Center API Documentation](https://learn.microsoft.com/en-us/dynamics365/business-central/dev-itpro/administration/administration-center-api)
- [Azure AD Authentication Overview](https://learn.microsoft.com/en-us/azure/active-directory/develop/authentication-scenarios)
- [Terraform Provider Development](https://developer.hashicorp.com/terraform/plugin)
- [Provider Source Code](https://github.com/axiansinfoma/terraform-provider-bcadmincenter)

## Support

For issues, feature requests, or contributions, please visit the [GitHub repository](https://github.com/axiansinfoma/terraform-provider-bcadmincenter).

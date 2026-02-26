# Copyright (c) 2025 Axians Infoma GmbH
# SPDX-License-Identifier: MPL-2.0

terraform {
  required_providers {
    bcadmincenter = {
      source = "vllni/bcadmincenter"
    }
  }
}

# Example 1: Service Principal with Client Secret (explicit configuration)
provider "bcadmincenter" {
  client_id     = "00000000-0000-0000-0000-000000000000"
  client_secret = "your-client-secret"
  tenant_id     = "00000000-0000-0000-0000-000000000000"

  # Optional: Override default settings
  environment = "public" # public, usgovernment, china
}

# Example 2: Using environment variables for authentication
# The provider automatically reads from these environment variables if not set in configuration:
# - AZURE_CLIENT_ID       -> provider.client_id
# - AZURE_CLIENT_SECRET   -> provider.client_secret
# - AZURE_TENANT_ID       -> provider.tenant_id
# - AZURE_ENVIRONMENT     -> provider.environment
#
# Export the variables in your shell:
#   export AZURE_CLIENT_ID="00000000-0000-0000-0000-000000000000"
#   export AZURE_CLIENT_SECRET="your-client-secret"
#   export AZURE_TENANT_ID="00000000-0000-0000-0000-000000000000"
#   export AZURE_ENVIRONMENT="public"  # optional, defaults to "public"
#
# Then use an empty provider block:
# provider "bcadmincenter" {
#   # All configuration will be automatically picked up from environment variables
# }

# Example 3: Mixed configuration (provider config takes precedence)
# You can mix environment variables and explicit configuration.
# Values set in the provider block always take precedence over environment variables.
#
# export AZURE_TENANT_ID="00000000-0000-0000-0000-000000000001"
# export AZURE_CLIENT_ID="00000000-0000-0000-0000-000000000002"
# export AZURE_CLIENT_SECRET="env-secret"
#
# provider "bcadmincenter" {
#   tenant_id = "00000000-0000-0000-0000-000000000003"  # This overrides AZURE_TENANT_ID
#   # client_id and client_secret will be read from environment variables
# }

# Example 4: Azure CLI Authentication (for local development)
# When client_id and client_secret are not provided, the provider uses DefaultAzureCredential
# which tries multiple authentication methods in this order:
# 1. Environment variables (AZURE_CLIENT_ID, AZURE_CLIENT_SECRET, AZURE_TENANT_ID)
# 2. Workload Identity (if running in Azure Kubernetes Service)
# 3. Managed Identity (if running on Azure VM/Container/App Service)
# 4. Azure CLI (az login)
# 5. Azure Developer CLI (azd auth login)
#
# For local development with Azure CLI:
#   az login --tenant 00000000-0000-0000-0000-000000000000
#
# provider "bcadmincenter" {
#   tenant_id = "00000000-0000-0000-0000-000000000000"
#   # Authentication will be obtained from Azure CLI
# }

# Example 5: Azure Workload Identity (Recommended for CI/CD in Kubernetes)
# When running in a Kubernetes cluster with Azure Workload Identity enabled,
# the following environment variables are automatically set by the workload identity webhook:
# - AZURE_CLIENT_ID
# - AZURE_TENANT_ID
# - AZURE_FEDERATED_TOKEN_FILE
# - AZURE_AUTHORITY_HOST
#
# The provider automatically detects and uses these credentials via DefaultAzureCredential.
# provider "bcadmincenter" {
#   # Provider automatically detects and uses workload identity credentials from environment
# }

# Example 6: Azure Managed Identity (for Azure VMs, Container Instances, App Service)
# When running on Azure infrastructure with managed identity enabled,
# DefaultAzureCredential automatically detects and uses the managed identity.
#
# provider "bcadmincenter" {
#   tenant_id = "00000000-0000-0000-0000-000000000000"
#   # Authentication will use the system-assigned or user-assigned managed identity
# }

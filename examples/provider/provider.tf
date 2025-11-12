terraform {
  required_providers {
    bc_admin_center = {
      source = "vllni/bc-admin-center"
    }
  }
}

# Example 1: Service Principal with Client Secret
provider "bc_admin_center" {
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
# provider "bc_admin_center" {
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
# provider "bc_admin_center" {
#   # Provider automatically detects and uses workload identity credentials
# }

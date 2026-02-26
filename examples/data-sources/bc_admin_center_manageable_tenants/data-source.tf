# Copyright (c) 2025 Axians Infoma GmbH
# SPDX-License-Identifier: MPL-2.0

# Example: Get Manageable Tenants for App
# Note: This data source can only be used when authenticated as an app (service principal)

terraform {
  required_providers {
    bcadmincenter = {
      source = "vllni/bcadmincenter"
    }
  }
}

provider "bcadmincenter" {
  # Authentication must use app credentials (client credentials flow)
  # AZURE_CLIENT_ID
  # AZURE_CLIENT_SECRET
  # AZURE_TENANT_ID
}

# Get a list of all tenants where this app is authorized
data "bcadmincenter_manageable_tenants" "all" {}

# Output the list of manageable tenant IDs
output "manageable_tenant_ids" {
  description = "List of tenant IDs where this app can manage Business Central"
  value       = [for tenant in data.bcadmincenter_manageable_tenants.all.tenants : tenant.entra_tenant_id]
}

# Output the number of manageable tenants
output "tenant_count" {
  description = "Number of tenants this app can manage"
  value       = length(data.bcadmincenter_manageable_tenants.all.tenants)
}

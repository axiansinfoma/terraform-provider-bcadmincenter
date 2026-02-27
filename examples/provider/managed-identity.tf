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

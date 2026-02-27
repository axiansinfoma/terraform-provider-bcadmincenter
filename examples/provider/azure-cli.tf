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

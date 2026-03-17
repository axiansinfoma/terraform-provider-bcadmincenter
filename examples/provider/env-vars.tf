# Copyright (c) 2025 Axians Infoma GmbH
# SPDX-License-Identifier: MPL-2.0

# Authentication via environment variables.
# Set the following variables in your shell before running Terraform:
#
#   export ARM_CLIENT_ID="00000000-0000-0000-0000-000000000000"
#   export ARM_CLIENT_SECRET="your-client-secret"
#   export ARM_TENANT_ID="00000000-0000-0000-0000-000000000000"

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

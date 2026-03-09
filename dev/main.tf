# Copyright Axians Infoma GmbH 2025, 2026, 0
# SPDX-License-Identifier: MPL-2.0

# Example test configuration
terraform {
  required_providers {
    bcadmincenter = {
      source = "axiansinfoma/bcadmincenter"
    }
  }
}

provider "bcadmincenter" {
  # Authentication will be read from environment variables:
  # AZURE_TENANT_ID
  # Optional: AZURE_CLIENT_ID
  # Optional: AZURE_CLIENT_SECRET
  # If not set, will use Azure CLI authentication
  tenant_id = "c6c8fc28-a4d8-4cb3-aee2-7853710491df"
}

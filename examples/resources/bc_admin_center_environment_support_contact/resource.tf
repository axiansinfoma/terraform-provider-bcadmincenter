# Copyright (c) 2025 Axians Infoma GmbH
# SPDX-License-Identifier: MPL-2.0

# Configure the Business Central Admin Center provider
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

# Configure support contact for a Business Central environment
resource "bcadmincenter_environment_support_contact" "example" {
  application_family = "BusinessCentral"
  environment_name   = "Production"

  # Optional: specify the Azure AD tenant ID (defaults to the provider's configured tenant_id)
  # aad_tenant_id = "00000000-0000-0000-0000-000000000000"

  name  = "IT Support Team"
  email = "support@example.com"
  url   = "https://support.example.com"
}

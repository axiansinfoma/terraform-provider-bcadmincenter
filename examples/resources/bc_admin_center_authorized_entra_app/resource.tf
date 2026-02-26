# Copyright (c) 2025 Axians Infoma GmbH
# SPDX-License-Identifier: MPL-2.0

# Example: Authorize a Microsoft Entra App

terraform {
  required_providers {
    bcadmincenter = {
      source = "vllni/bcadmincenter"
    }
  }
}

provider "bcadmincenter" {
  # Authentication can be configured via environment variables:
  # AZURE_CLIENT_ID
  # AZURE_CLIENT_SECRET
  # AZURE_TENANT_ID
}

# Authorize a specific Microsoft Entra app to call the Business Central Admin Center API
resource "bcadmincenter_authorized_entra_app" "partner_app" {
  app_id = "550e8400-e29b-41d4-a716-446655440000"
}

# Note: This does not grant admin consent or assign permission sets in environments
# You must separately:
# 1. Grant admin consent for the AdminCenter.ReadWrite.All permission
# 2. Assign required permission sets in environments (e.g., D365 BACKUP/RESTORE)

output "app_id" {
  value       = bcadmincenter_authorized_entra_app.partner_app.app_id
  description = "The authorized app ID"
}

output "is_admin_consent_granted" {
  value       = bcadmincenter_authorized_entra_app.partner_app.is_admin_consent_granted
  description = "Whether admin consent has been granted"
}

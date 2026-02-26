# Copyright (c) 2025 Axians Infoma GmbH
# SPDX-License-Identifier: MPL-2.0

# Example: List All Authorized Microsoft Entra Apps

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

# Get a list of all authorized Microsoft Entra apps
data "bcadmincenter_authorized_entra_apps" "all" {}

# Output the list of authorized apps
output "authorized_apps" {
  value = [
    for app in data.bcadmincenter_authorized_entra_apps.all.apps : {
      app_id                   = app.app_id
      is_admin_consent_granted = app.is_admin_consent_granted
    }
  ]
  description = "List of all authorized Microsoft Entra apps"
}

# Filter apps by consent status
output "apps_with_consent" {
  value = [
    for app in data.bcadmincenter_authorized_entra_apps.all.apps :
    app.app_id if app.is_admin_consent_granted
  ]
  description = "App IDs that have admin consent granted"
}

output "apps_without_consent" {
  value = [
    for app in data.bcadmincenter_authorized_entra_apps.all.apps :
    app.app_id if !app.is_admin_consent_granted
  ]
  description = "App IDs that need admin consent"
}

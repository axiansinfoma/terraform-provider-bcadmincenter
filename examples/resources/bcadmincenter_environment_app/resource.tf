# Copyright (c) 2025 Axians Infoma GmbH
# SPDX-License-Identifier: MPL-2.0

# This example demonstrates how to install a Business Central app into an environment.
# The app version is pinned to a specific version; omitting the version attribute installs
# the latest available version. Changing the version to a higher value on a subsequent
# apply triggers an in-place update without recreating the resource.

terraform {
  required_providers {
    bcadmincenter = {
      source = "axiansinfoma/bcadmincenter"
    }
  }
}

provider "bcadmincenter" {
  # Authentication via Service Principal (or use environment variables)
  # client_id     = var.client_id
  # client_secret = var.client_secret
  # tenant_id     = var.tenant_id
}

resource "bcadmincenter_environment" "sandbox" {
  name               = "my-sandbox"
  application_family = "BusinessCentral"
  type               = "Sandbox"
  country_code       = "US"
  ring_name          = "PROD"
}

resource "bcadmincenter_environment_app" "contoso_app" {
  application_family = bcadmincenter_environment.sandbox.application_family
  environment_name   = bcadmincenter_environment.sandbox.name

  app_id  = "00000000-0000-0000-0000-000000000000"
  version = "1.0.0.0" # Omit to install the latest available version.

  install_or_update_needed_dependencies = true
  allow_preview_version                 = false
}

output "app_status" {
  description = "The current install status of the app."
  value       = bcadmincenter_environment_app.contoso_app.status
}

output "app_version" {
  description = "The installed version of the app."
  value       = bcadmincenter_environment_app.contoso_app.version
}

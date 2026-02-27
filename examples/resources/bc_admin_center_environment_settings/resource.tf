# Copyright (c) 2025 Axians Infoma GmbH
# SPDX-License-Identifier: MPL-2.0

# Basic environment settings with update window configuration

terraform {
  required_providers {
    bcadmincenter = {
      source = "axiansinfoma/bcadmincenter"
    }
  }
}

resource "bc_admin_center_environment_settings" "production" {
  application_family = "BusinessCentral"
  environment_name   = "production"

  # Optional: specify the Azure AD tenant ID (defaults to the provider's configured tenant_id)
  # aad_tenant_id = "00000000-0000-0000-0000-000000000000"

  # Configure update window (must be at least 6 hours)
  update_window_start_time = "22:00" # 10 PM
  update_window_end_time   = "06:00" # 6 AM
  update_window_timezone   = "Pacific Standard Time"

  # Enable Application Insights telemetry
  # Note: Setting this triggers an automatic environment restart
  app_insights_key = "InstrumentationKey=your-app-insights-key;IngestionEndpoint=https://westus2-1.in.applicationinsights.azure.com/"

  # Configure app update cadence
  app_update_cadence = "DuringMajorUpgrade"
}

# Reference an environment resource
resource "bc_admin_center_environment" "sandbox" {
  name               = "sandbox-1"
  application_family = "BusinessCentral"
  type               = "Sandbox"
  country_code       = "US"
  ring_name          = "Production"
  azure_region       = "eastus"
}

resource "bc_admin_center_environment_settings" "sandbox" {
  application_family = bc_admin_center_environment.sandbox.application_family
  environment_name   = bc_admin_center_environment.sandbox.name

  # Optional: specify the Azure AD tenant ID (defaults to the provider's configured tenant_id)
  # aad_tenant_id = "00000000-0000-0000-0000-000000000000"

  update_window_start_time = "20:00"
  update_window_end_time   = "04:00"
  update_window_timezone   = "Eastern Standard Time"

  # Restrict access to specific Azure AD security group
  security_group_id = "12345678-1234-1234-1234-123456789012"

  # Enable M365 license access (requires BC 21.1+)
  access_with_m365_licenses = true
}

# Copyright (c) 2025 Axians Infoma GmbH
# SPDX-License-Identifier: MPL-2.0

# Example: Production Environment

terraform {
  required_providers {
    bcadmincenter = {
      source = "axiansinfoma/bcadmincenter"
    }
  }
}

provider "bcadmincenter" {
  # Authentication can be configured via environment variables:
  # AZURE_CLIENT_ID
  # AZURE_CLIENT_SECRET
  # AZURE_TENANT_ID
}

resource "bcadmincenter_environment" "production" {
  name               = "production"
  application_family = "BusinessCentral"
  type               = "Production"
  country_code       = "US"
  ring_name          = "PROD"
  azure_region       = "westus2"

  timeouts {
    create = "90m"
    delete = "60m"
  }
}

# Example: Sandbox environment with inline settings

resource "bcadmincenter_environment" "sandbox" {
  name               = "sandbox-1"
  application_family = "BusinessCentral"
  type               = "Sandbox"
  country_code       = "US"
  ring_name          = "PROD"
  azure_region       = "eastus"

  settings {
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
}

output "web_client_url" {
  value       = bcadmincenter_environment.production.web_client_login_url
  description = "The URL to access the Business Central web client"
}

output "environment_status" {
  value       = bcadmincenter_environment.production.status
  description = "The current status of the environment"
}

output "application_version" {
  value       = bcadmincenter_environment.production.application_version
  description = "The application version running in the environment (read-only, assigned by the API)"
}

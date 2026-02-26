# Copyright (c) 2025 Axians Infoma GmbH
# SPDX-License-Identifier: MPL-2.0

# Example: Production Environment

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

resource "bcadmincenter_environment" "production" {
  name               = "production"
  application_family = "BusinessCentral"
  type               = "Production"
  country_code       = "US"
  ring_name          = "Production"
  azure_region       = "westus2"

  timeouts {
    create = "90m"
    delete = "60m"
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

# Copyright (c) 2025 Michael Villani
# SPDX-License-Identifier: MPL-2.0

# Get information about a specific Business Central environment

terraform {
  required_providers {
    bcadmincenter = {
      source = "vllni/bcadmincenter"
    }
  }
}

provider "bcadmincenter" {
  # Authentication configured via environment variables or provider block
}

data "bcadmincenter_environment" "example" {
  application_family = "BusinessCentral"
  name               = "my-production-env"
}

# Use the retrieved environment information
output "environment_status" {
  value = data.bcadmincenter_environment.example.status
}

output "web_client_url" {
  value = data.bcadmincenter_environment.example.web_client_login_url
}

output "application_version" {
  value = data.bcadmincenter_environment.example.application_version
}

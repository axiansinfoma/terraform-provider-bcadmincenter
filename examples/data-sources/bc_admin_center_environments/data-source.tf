# Copyright (c) 2025 Axians Infoma GmbH
# SPDX-License-Identifier: MPL-2.0

# List all Business Central environments for a given application family

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

data "bcadmincenter_environments" "all" {
  application_family = "BusinessCentral"
}

# Output all environment names
output "environment_names" {
  value = [for env in data.bcadmincenter_environments.all.environments : env.name]
}

# Filter for production environments
output "production_environments" {
  value = [for env in data.bcadmincenter_environments.all.environments : env.name if env.type == "Production"]
}

# Filter for sandbox environments
output "sandbox_environments" {
  value = [for env in data.bcadmincenter_environments.all.environments : env.name if env.type == "Sandbox"]
}

# Output environment URLs
output "environment_urls" {
  value = {
    for env in data.bcadmincenter_environments.all.environments :
    env.name => env.web_client_login_url
  }
}

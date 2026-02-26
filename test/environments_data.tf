# Copyright Axians Infoma GmbH 2025, 2026, 0
# SPDX-License-Identifier: MPL-2.0

# Test configuration for environments data sources

# List all environments in the tenant
data "bcadmincenter_environments" "all" {
  application_family = "BusinessCentral"
  depends_on         = [bcadmincenter_environment.test]
}

output "all_environments" {
  value = [for env in data.bcadmincenter_environments.all.environments : {
    name   = env.name
    type   = env.type
    status = env.status
  }]
}

output "environment_count" {
  value = {
    total      = length(data.bcadmincenter_environments.all.environments)
    production = length([for env in data.bcadmincenter_environments.all.environments : env if env.type == "Production"])
    sandbox    = length([for env in data.bcadmincenter_environments.all.environments : env if env.type == "Sandbox"])
  }
}

# Get specific environment details
data "bcadmincenter_environment" "test_details" {
  application_family = bcadmincenter_environment.test.application_family
  name               = bcadmincenter_environment.test.name

  depends_on = [bcadmincenter_environment.test]
}

output "test_environment_details" {
  value = {
    name                 = data.bcadmincenter_environment.test_details.name
    type                 = data.bcadmincenter_environment.test_details.type
    country_code         = data.bcadmincenter_environment.test_details.country_code
    version              = data.bcadmincenter_environment.test_details.application_version
    status               = data.bcadmincenter_environment.test_details.status
    web_client_login_url = data.bcadmincenter_environment.test_details.web_client_login_url
  }
}

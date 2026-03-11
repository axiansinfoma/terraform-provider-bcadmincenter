# Copyright (c) 2025 Axians Infoma GmbH
# SPDX-License-Identifier: MPL-2.0

# Tests for the bcadmincenter_environment resource using the Terraform test framework.
# These tests use mock_provider to validate resource schema and lifecycle behavior
# without requiring real Azure AD credentials or creating actual environments.

mock_provider "bcadmincenter" {
  mock_resource "bcadmincenter_environment" {
    defaults = {
      id                   = "/tenants/00000000-0000-0000-0000-000000000001/providers/Microsoft.Dynamics365.BusinessCentral/applications/BusinessCentral/environments/test-sandbox"
      application_family   = "BusinessCentral"
      ring_name            = "PROD"
      application_version  = "25.0.0.0"
      azure_region         = "westeurope"
      status               = "Active"
      web_client_login_url = "https://businesscentral.dynamics.com/00000000-0000-0000-0000-000000000001/test-sandbox"
      web_service_url      = "https://api.businesscentral.dynamics.com/v2.0/00000000-0000-0000-0000-000000000001/test-sandbox/ODataV4"
      app_insights_key     = ""
      platform_version     = "25.0.0.0"
      aad_tenant_id        = "00000000-0000-0000-0000-000000000001"
    }
  }
}

# Test: Sandbox environment resource with minimal configuration.
run "create_sandbox_environment" {
  command = apply

  module {
    source = "./modules/environments"
  }

  assert {
    condition     = bcadmincenter_environment.test.name == "test-sandbox"
    error_message = "Environment name should be 'test-sandbox', got '${bcadmincenter_environment.test.name}'."
  }

  assert {
    condition     = bcadmincenter_environment.test.type == "Sandbox"
    error_message = "Environment type should be 'Sandbox', got '${bcadmincenter_environment.test.type}'."
  }

  assert {
    condition     = bcadmincenter_environment.test.country_code == "DE"
    error_message = "Environment country_code should be 'DE', got '${bcadmincenter_environment.test.country_code}'."
  }

  assert {
    condition     = bcadmincenter_environment.test.status == "Active"
    error_message = "Expected environment status to be 'Active', got '${bcadmincenter_environment.test.status}'."
  }
}

# Test: Environment with inline settings block.
run "environment_with_settings" {
  command = apply

  module {
    source = "./modules/environments"
  }

  assert {
    condition     = bcadmincenter_environment.test.settings.update_window_start_time == "21:00"
    error_message = "Expected update_window_start_time to be '21:00'."
  }

  assert {
    condition     = bcadmincenter_environment.test.settings.access_with_m365_licenses == true
    error_message = "Expected access_with_m365_licenses to be true."
  }
}

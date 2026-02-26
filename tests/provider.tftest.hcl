# Copyright (c) 2025 Axians Infoma GmbH
# SPDX-License-Identifier: MPL-2.0

# Tests for the bcadmincenter provider schema using the Terraform test framework.
# These tests use mock_provider to validate provider schema and configuration
# without requiring real Azure AD credentials.

mock_provider "bcadmincenter" {
  mock_data "bcadmincenter_quotas" {
    defaults = {
      id                                = "quotas"
      production_environments_quota     = 1
      production_environments_allocated = 0
      production_environments_available = 1
      sandbox_environments_quota        = 1
      sandbox_environments_allocated    = 0
      sandbox_environments_available    = 1
      storage_quota_gb                  = 10
      storage_allocated_gb              = 0
      storage_available_gb              = 10
    }
  }
}

# Test: Provider schema is valid and data sources can be read.
run "provider_schema_valid" {
  command = plan

  module {
    source = "./modules/data_sources"
  }

  assert {
    condition     = data.bcadmincenter_quotas.test.id == "quotas"
    error_message = "Provider schema is invalid: failed to read quotas data source."
  }
}

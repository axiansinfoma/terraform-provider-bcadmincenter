# Copyright (c) 2025 Axians Infoma GmbH
# SPDX-License-Identifier: MPL-2.0

# Tests for the bcadmincenter_quotas data source using the Terraform test framework.
# These tests use mock_provider to validate data source schema and attribute behavior
# without requiring real Azure AD credentials.

mock_provider "bcadmincenter" {
  mock_data "bcadmincenter_quotas" {
    defaults = {
      id                                  = "quotas"
      production_environments_quota       = 3
      production_environments_allocated   = 1
      production_environments_available   = 2
      sandbox_environments_quota          = 3
      sandbox_environments_allocated      = 1
      sandbox_environments_available      = 2
      storage_quota_gb                    = 80
      storage_allocated_gb                = 20
      storage_available_gb                = 60
    }
  }
}

# Test: Quotas data source returns expected values.
run "quotas_data_source_basic" {
  command = plan

  module {
    source = "./modules/data_sources"
  }

  assert {
    condition     = data.bcadmincenter_quotas.test.production_environments_quota == 3
    error_message = "Expected production_environments_quota to be 3, got ${data.bcadmincenter_quotas.test.production_environments_quota}."
  }

  assert {
    condition     = data.bcadmincenter_quotas.test.sandbox_environments_quota == 3
    error_message = "Expected sandbox_environments_quota to be 3, got ${data.bcadmincenter_quotas.test.sandbox_environments_quota}."
  }

  assert {
    condition     = data.bcadmincenter_quotas.test.storage_quota_gb == 80
    error_message = "Expected storage_quota_gb to be 80, got ${data.bcadmincenter_quotas.test.storage_quota_gb}."
  }

  assert {
    condition     = data.bcadmincenter_quotas.test.id == "quotas"
    error_message = "Expected id to be 'quotas', got '${data.bcadmincenter_quotas.test.id}'."
  }
}

# Test: Quotas available calculation is consistent (quota - allocated = available).
run "quotas_available_consistent" {
  command = plan

  module {
    source = "./modules/data_sources"
  }

  assert {
    condition = (
      data.bcadmincenter_quotas.test.production_environments_available ==
      data.bcadmincenter_quotas.test.production_environments_quota -
      data.bcadmincenter_quotas.test.production_environments_allocated
    )
    error_message = "Production environments available count is inconsistent."
  }

  assert {
    condition = (
      data.bcadmincenter_quotas.test.sandbox_environments_available ==
      data.bcadmincenter_quotas.test.sandbox_environments_quota -
      data.bcadmincenter_quotas.test.sandbox_environments_allocated
    )
    error_message = "Sandbox environments available count is inconsistent."
  }
}

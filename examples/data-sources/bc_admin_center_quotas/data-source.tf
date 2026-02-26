# Copyright (c) 2025 Axians Infoma GmbH
# SPDX-License-Identifier: MPL-2.0

# Query tenant environment quotas

terraform {
  required_providers {
    bcadmincenter = {
      source = "vllni/bcadmincenter"
    }
  }
}

data "bcadmincenter_quotas" "tenant" {}

# Output quota information
output "environment_capacity" {
  value = {
    production = {
      quota     = data.bcadmincenter_quotas.tenant.production_environments_quota
      allocated = data.bcadmincenter_quotas.tenant.production_environments_allocated
      available = data.bcadmincenter_quotas.tenant.production_environments_available
    }
    sandbox = {
      quota     = data.bcadmincenter_quotas.tenant.sandbox_environments_quota
      allocated = data.bcadmincenter_quotas.tenant.sandbox_environments_allocated
      available = data.bcadmincenter_quotas.tenant.sandbox_environments_available
    }
    storage = {
      quota_gb     = data.bcadmincenter_quotas.tenant.storage_quota_gb
      allocated_gb = data.bcadmincenter_quotas.tenant.storage_allocated_gb
      available_gb = data.bcadmincenter_quotas.tenant.storage_available_gb
    }
  }
}

# Check if we have capacity before creating an environment
resource "bcadmincenter_environment" "conditional" {
  count = data.bcadmincenter_quotas.tenant.production_environments_available > 0 ? 1 : 0

  name               = "new-production"
  application_family = "BusinessCentral"
  type               = "Production"
  country_code       = "US"
  ring_name          = "PROD"
}

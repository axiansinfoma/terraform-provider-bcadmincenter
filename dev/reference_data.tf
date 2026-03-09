# Copyright Axians Infoma GmbH 2025, 2026, 0
# SPDX-License-Identifier: MPL-2.0

# Test configuration for timezones and quotas data sources

# Get all available timezones
data "bcadmincenter_timezones" "all" {}

output "timezone_count" {
  value = length(data.bcadmincenter_timezones.all.timezones)
}

# Find specific timezones
locals {
  pacific_tz = [
    for tz in data.bcadmincenter_timezones.all.timezones :
    tz if tz.id == "Pacific Standard Time"
  ]

  central_european_tz = [
    for tz in data.bcadmincenter_timezones.all.timezones :
    tz if tz.id == "Central European Standard Time"
  ]

  eastern_tz = [
    for tz in data.bcadmincenter_timezones.all.timezones :
    tz if tz.id == "Eastern Standard Time"
  ]
}

output "common_timezones" {
  value = {
    pacific = length(local.pacific_tz) > 0 ? local.pacific_tz[0] : null
    cet     = length(local.central_european_tz) > 0 ? local.central_european_tz[0] : null
    eastern = length(local.eastern_tz) > 0 ? local.eastern_tz[0] : null
  }
}

# Get tenant quotas
data "bcadmincenter_quotas" "tenant" {}

output "tenant_capacity" {
  value = {
    production_environments = {
      quota     = data.bcadmincenter_quotas.tenant.production_environments_quota
      allocated = data.bcadmincenter_quotas.tenant.production_environments_allocated
      available = data.bcadmincenter_quotas.tenant.production_environments_available
    }
    sandbox_environments = {
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

# Capacity warnings
output "capacity_warnings" {
  value = {
    production_low = data.bcadmincenter_quotas.tenant.production_environments_available < 1
    sandbox_low    = data.bcadmincenter_quotas.tenant.sandbox_environments_available < 3
    storage_low    = data.bcadmincenter_quotas.tenant.storage_available_gb < 10
  }
}

# Utilization percentages
output "capacity_utilization" {
  value = {
    production_pct = data.bcadmincenter_quotas.tenant.production_environments_quota > 0 ? (
      data.bcadmincenter_quotas.tenant.production_environments_allocated /
      data.bcadmincenter_quotas.tenant.production_environments_quota * 100
    ) : 0
    sandbox_pct = data.bcadmincenter_quotas.tenant.sandbox_environments_quota > 0 ? (
      data.bcadmincenter_quotas.tenant.sandbox_environments_allocated /
      data.bcadmincenter_quotas.tenant.sandbox_environments_quota * 100
    ) : 0
    storage_pct = data.bcadmincenter_quotas.tenant.storage_quota_gb > 0 ? (
      data.bcadmincenter_quotas.tenant.storage_allocated_gb /
      data.bcadmincenter_quotas.tenant.storage_quota_gb * 100
    ) : 0
  }
}

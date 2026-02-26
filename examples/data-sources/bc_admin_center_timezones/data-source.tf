# Copyright (c) 2025 Axians Infoma GmbH
# SPDX-License-Identifier: MPL-2.0

# Query available time zones

terraform {
  required_providers {
    bcadmincenter = {
      source = "axiansinfoma/bcadmincenter"
    }
  }
}

data "bcadmincenter_timezones" "available" {}

# Output all timezones
output "all_timezones" {
  value = data.bcadmincenter_timezones.available.timezones
}

# Find a specific timezone
locals {
  pacific_timezone = [
    for tz in data.bcadmincenter_timezones.available.timezones :
    tz if tz.id == "Pacific Standard Time"
  ][0]
}

output "pacific_timezone_info" {
  value = {
    id           = local.pacific_timezone.id
    display_name = local.pacific_timezone.display_name
    offset       = local.pacific_timezone.offset_from_utc
    has_dst      = local.pacific_timezone.supports_daylight_savings
  }
}

# Use in environment settings
resource "bcadmincenter_environment_settings" "example" {
  application_family = "BusinessCentral"
  environment_name   = "production"

  update_window_start_time = "22:00"
  update_window_end_time   = "06:00"
  update_window_timezone   = local.pacific_timezone.id
}

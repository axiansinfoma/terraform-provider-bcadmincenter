# Copyright (c) 2025 Axians Infoma GmbH
# SPDX-License-Identifier: MPL-2.0

# List all available version updates for a Business Central environment

terraform {
  required_providers {
    bcadmincenter = {
      source = "axiansinfoma/bcadmincenter"
    }
  }
}

provider "bcadmincenter" {
  # Authentication configured via environment variables or provider block
}

data "bcadmincenter_environment_updates" "prod" {
  application_family = "BusinessCentral"
  environment_name   = "production"
}

# Output all available updates
output "available_updates" {
  value = [
    for u in data.bcadmincenter_environment_updates.prod.updates :
    {
      version = u.target_version
      status  = u.update_status
    }
    if u.available
  ]
}

# Find the currently selected update (if any)
output "selected_update" {
  value = [
    for u in data.bcadmincenter_environment_updates.prod.updates :
    {
      version            = u.target_version
      status             = u.update_status
      scheduled_datetime = u.scheduled_datetime
    }
    if u.selected
  ]
}

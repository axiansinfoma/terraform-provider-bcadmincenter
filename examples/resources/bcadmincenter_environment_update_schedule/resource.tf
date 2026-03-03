# Copyright (c) 2025 Axians Infoma GmbH
# SPDX-License-Identifier: MPL-2.0

# Manages an explicitly scheduled upgrade for a Business Central environment.
# Use this resource when you need full control over the target version and schedule.
# Do NOT use application_version on bcadmincenter_environment for the same environment.

terraform {
  required_providers {
    bcadmincenter = {
      source = "axiansinfoma/bcadmincenter"
    }
  }
}

resource "bcadmincenter_environment" "prod" {
  name               = "my-production"
  application_family = "BusinessCentral"
  type               = "Production"
  country_code       = "US"
  ring_name          = "PROD"
  # application_version intentionally omitted — managed via update_schedule below
}

resource "bcadmincenter_environment_update_schedule" "prod_upgrade" {
  application_family   = bcadmincenter_environment.prod.application_family
  environment_name     = bcadmincenter_environment.prod.name
  target_version       = "26.2"
  scheduled_datetime   = "2026-04-01T02:00:00Z"
  ignore_update_window = false
}

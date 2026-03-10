# Copyright (c) 2025 Axians Infoma GmbH
# SPDX-License-Identifier: MPL-2.0

# Environment resources tested by environments.tftest.hcl.

resource "bcadmincenter_environment" "test" {
  name         = "test-sandbox"
  type         = "Sandbox"
  country_code = "DE"

  settings {
    update_window_start_time  = "21:00"
    update_window_end_time    = "03:00"
    update_window_timezone    = "Central European Standard Time"
    app_update_cadence        = "Default"
    access_with_m365_licenses = true
  }
}

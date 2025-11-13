# Copyright (c) Michael Villani
# SPDX-License-Identifier: MPL-2.0

# Create a test sandbox environment
resource "bcadmincenter_environment" "test" {
  name               = "test-sandbox"
  application_family = "BusinessCentral"
  type               = "Sandbox"
  country_code       = "DE"
  ring_name          = "PROD"
}

resource "bcadmincenter_environment_settings" "test" {
  application_family = bcadmincenter_environment.test.application_family
  environment_name   = bcadmincenter_environment.test.name

  update_window_start_time = "21:00"
  update_window_end_time   = "03:00"
  update_window_timezone   = "Central European Standard Time"

  app_update_cadence = "Default"
}

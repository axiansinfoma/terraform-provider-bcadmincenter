# Copyright (c) Michael Villani
# SPDX-License-Identifier: MPL-2.0

# Example notification recipient configuration
# NOTE: This is commented out by default - uncomment to test

resource "bcadmincenter_notification_recipient" "primary_admin" {
  email = "coda@axians-infoma.de"
  name  = "Primary Administrator"
}

data "bcadmincenter_notification_settings" "current" {
  aad_tenant_id = "9ff11aaa-cddc-4df5-97c9-b9e79db1ba1d"
}

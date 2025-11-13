# Copyright (c) Michael Villani
# SPDX-License-Identifier: MPL-2.0

# Example notification recipient configuration
# NOTE: This is commented out by default - uncomment to test

resource "bcadmincenter_notification_recipient" "primary_admin" {
  email = "coda@axians-infoma.de"
  name  = "Primary Administrator"
}

# Get current notification settings
# NOTE: This may return 404 if no notification recipients have been configured yet
# The API only returns settings after at least one recipient has been added
data "bcadmincenter_notification_settings" "current" {
  aad_tenant_id = "9ff11aaa-cddc-4df5-97c9-b9e79db1ba1d"
  depends_on    = [bcadmincenter_notification_recipient.primary_admin]
}

# Copyright (c) 2025 Axians Infoma GmbH
# SPDX-License-Identifier: MPL-2.0

# Notification recipient resource tested by notification_recipients.tftest.hcl.

resource "bcadmincenter_notification_recipient" "test" {
  email = "admin@example.com"
  name  = "System Administrator"
}

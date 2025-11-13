# Copyright (c) 2025 Michael Villani
# SPDX-License-Identifier: MPL-2.0

# Configure a notification recipient for the tenant
resource "bcadmincenter_notification_recipient" "admin" {
  email = "admin@example.com"
  name  = "Primary Administrator"
}

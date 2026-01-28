# Copyright Michael Villani 2025, 0
# SPDX-License-Identifier: MPL-2.0

# Example notification settings data source usage
# NOTE: This is commented out by default - uncomment to test

# data "bcadmincenter_notification_settings" "current" {
# }

# output "aad_tenant_id" {
#   value = data.bcadmincenter_notification_settings.current.aad_tenant_id
# }

# output "notification_recipients" {
#   value = data.bcadmincenter_notification_settings.current.recipients
# }

# output "recipient_count" {
#   value = length(data.bcadmincenter_notification_settings.current.recipients)
# }

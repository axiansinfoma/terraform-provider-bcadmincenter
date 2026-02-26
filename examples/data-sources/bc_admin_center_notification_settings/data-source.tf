# Copyright (c) 2025 Axians Infoma GmbH
# SPDX-License-Identifier: MPL-2.0

# Get notification settings for the tenant

terraform {
  required_providers {
    bcadmincenter = {
      source = "vllni/bcadmincenter"
    }
  }
}

data "bcadmincenter_notification_settings" "current" {
}

# Output the tenant ID and recipients
output "tenant_id" {
  value = data.bcadmincenter_notification_settings.current.aad_tenant_id
}

output "notification_recipients" {
  value = data.bcadmincenter_notification_settings.current.recipients
}

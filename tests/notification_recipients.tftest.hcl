# Copyright (c) 2025 Axians Infoma GmbH
# SPDX-License-Identifier: MPL-2.0

# Tests for the bcadmincenter_notification_recipient resource using the Terraform test framework.
# These tests use mock_provider to validate resource schema and lifecycle behavior
# without requiring real Azure AD credentials.

mock_provider "bcadmincenter" {
  mock_resource "bcadmincenter_notification_recipient" {
    defaults = {
      id           = "/tenants/00000000-0000-0000-0000-000000000001/providers/Microsoft.Dynamics365.BusinessCentral/notificationRecipients/test-recipient-id"
      aad_tenant_id = "00000000-0000-0000-0000-000000000001"
    }
  }
}

# Test: Notification recipient resource with required fields.
run "create_notification_recipient" {
  command = apply

  module {
    source = "./modules/notification_recipients"
  }

  assert {
    condition     = bcadmincenter_notification_recipient.test.email == "admin@example.com"
    error_message = "Expected email to be 'admin@example.com', got '${bcadmincenter_notification_recipient.test.email}'."
  }

  assert {
    condition     = bcadmincenter_notification_recipient.test.name == "System Administrator"
    error_message = "Expected name to be 'System Administrator', got '${bcadmincenter_notification_recipient.test.name}'."
  }

  assert {
    condition     = bcadmincenter_notification_recipient.test.aad_tenant_id == "00000000-0000-0000-0000-000000000001"
    error_message = "Expected aad_tenant_id to be set."
  }
}

# Test: Resource ID format is correct.
run "notification_recipient_id_format" {
  command = apply

  module {
    source = "./modules/notification_recipients"
  }

  assert {
    condition = startswith(
      bcadmincenter_notification_recipient.test.id,
      "/tenants/"
    )
    error_message = "Expected notification recipient ID to start with '/tenants/'."
  }
}

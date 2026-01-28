# Copyright Michael Villani 2025, 0
# SPDX-License-Identifier: MPL-2.0

# Test configuration for authorized Entra apps

# Register an Entra app for API access
# NOTE: This requires the app to be registered in Azure AD first
resource "bcadmincenter_authorized_entra_app" "test_app" {
  app_id = "00000000-0000-0000-0000-000000000000" # Replace with actual app ID
}

# List all authorized apps
data "bcadmincenter_authorized_entra_apps" "all" {
  depends_on = [bcadmincenter_authorized_entra_app.test_app]
}

output "authorized_apps" {
  value = data.bcadmincenter_authorized_entra_apps.all.apps
}

# List manageable tenants (for delegated admins)
# NOTE: This will return 403 Forbidden if the authenticated account is not a delegated admin
# Comment out this data source if you are not using delegated admin access
# data "bcadmincenter_manageable_tenants" "accessible" {}

# output "manageable_tenants" {
#   value = data.bcadmincenter_manageable_tenants.accessible.tenants
# }

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

# Output the environment details
output "environment_name" {
  value       = bcadmincenter_environment.test.name
  description = "The name of the created environment"
}

output "environment_status" {
  value       = bcadmincenter_environment.test.status
  description = "The status of the environment"
}

output "web_client_url" {
  value       = bcadmincenter_environment.test.web_client_login_url
  description = "URL to access the Business Central web client"
}

output "aad_tenant_id" {
  value       = bcadmincenter_environment.test.aad_tenant_id
  description = "The Azure AD tenant ID"
}
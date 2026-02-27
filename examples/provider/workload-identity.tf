# Copyright (c) 2025 Axians Infoma GmbH
# SPDX-License-Identifier: MPL-2.0

# Workload Identity (OIDC) authentication – recommended for CI/CD pipelines
# on GitHub Actions, Azure DevOps, and Kubernetes with Azure Workload Identity.
#
# The following environment variables must be set by your CI/CD platform:
#   AZURE_CLIENT_ID            – application (client) ID
#   AZURE_TENANT_ID            – Azure AD tenant ID
#   AZURE_FEDERATED_TOKEN_FILE – path to the OIDC token file

terraform {
  required_providers {
    bcadmincenter = {
      source = "axiansinfoma/bcadmincenter"
    }
  }
}

provider "bcadmincenter" {
  tenant_id = "00000000-0000-0000-0000-000000000000"
  client_id = "00000000-0000-0000-0000-000000000000"
  use_oidc  = true
}

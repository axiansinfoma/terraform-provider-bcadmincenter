# Copyright (c) 2025 Axians Infoma GmbH
# SPDX-License-Identifier: MPL-2.0

# Provider requirements for Terraform test framework tests.
# These tests use mock_provider to avoid needing real Azure credentials.
terraform {
  required_providers {
    bcadmincenter = {
      source  = "axiansinfoma/bcadmincenter"
      version = "0.0.1-preview.2"
    }
  }
}

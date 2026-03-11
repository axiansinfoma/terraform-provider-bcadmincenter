# Copyright (c) 2025 Axians Infoma GmbH
# SPDX-License-Identifier: MPL-2.0

# Environment resources tested by environments.tftest.hcl.

resource "bcadmincenter_environment" "test" {
  name         = "test-sandbox"
  type         = "Sandbox"
  country_code = "DE"
}

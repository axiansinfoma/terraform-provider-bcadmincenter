# BC Admin Center Provider - Terraform Test Framework Tests

This directory contains tests for the BC Admin Center Terraform provider using the
[Terraform test framework](https://developer.hashicorp.com/terraform/language/tests).

## Overview

These tests use `mock_provider` to validate provider schemas, resource configurations,
and data source behavior without requiring real Azure AD credentials or a live
Business Central tenant.

## Test Files

| File | Description |
|------|-------------|
| `provider.tftest.hcl` | Tests that the provider schema is valid |
| `data_sources.tftest.hcl` | Tests for data sources (quotas) |
| `environments.tftest.hcl` | Tests for environment resource and settings |
| `notification_recipients.tftest.hcl` | Tests for notification recipient resource |

## Module Structure

Test configurations are organized into modules to isolate each test area:

```
tests/
├── terraform.tf                          # Provider requirements
├── provider.tftest.hcl                   # Provider schema tests
├── data_sources.tftest.hcl               # Data source tests
├── environments.tftest.hcl               # Environment resource tests
├── notification_recipients.tftest.hcl    # Notification recipient tests
└── modules/
    ├── data_sources/                     # Quota data source configuration
    ├── environments/                     # Environment resource configuration
    └── notification_recipients/          # Notification recipient configuration
```

## Running Tests Locally

### Prerequisites

- [Terraform](https://developer.hashicorp.com/terraform/downloads) 1.7+ (for `mock_provider` support)

### Steps

1. Initialize the test environment:
   ```bash
   cd tests
   terraform init
   ```

2. Run all tests:
   ```bash
   terraform test
   ```

3. Run a specific test file:
   ```bash
   terraform test -filter=data_sources.tftest.hcl
   ```

4. Run tests with verbose output:
   ```bash
   terraform test -verbose
   ```

## How mock_provider Works

The `mock_provider` block in `.tftest.hcl` files replaces the actual provider with a
mock implementation during testing. This means:

- **No Azure credentials required** - the mock handles all provider operations
- **No real API calls** - responses come from the `mock_resource` and `mock_data` defaults
- **Schema validation** - the mock must conform to the actual provider schema
- **Fast execution** - no network calls or real infrastructure changes

### Example

```hcl
mock_provider "bcadmincenter" {
  mock_data "bcadmincenter_quotas" {
    defaults = {
      id                            = "quotas"
      production_environments_quota = 3
      sandbox_environments_quota    = 3
    }
  }
}

run "test_quotas" {
  command = plan
  
  assert {
    condition     = data.bcadmincenter_quotas.test.production_environments_quota == 3
    error_message = "Unexpected production quota."
  }
}
```

## CI Integration

These tests run automatically on every pull request via the
`.github/workflows/test.yml` workflow. The CI runs them against multiple Terraform
versions to ensure compatibility.

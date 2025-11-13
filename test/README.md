# Test Configuration for BC Admin Center Provider

This directory contains comprehensive test configurations for all resources and data sources in the BC Admin Center provider.

## Test Files Overview

### Core Provider Configuration
- **main.tf** - Provider configuration (reads credentials from environment variables)

### Resource Tests
- **environment.tf** - Environment resource testing
- **notification_recipients.tf** - Notification recipient resource testing
- **authorized_entra_apps.tf** - Authorized Entra app resource and data sources
- **complete_example.tf.example** - Comprehensive multi-environment example (rename to .tf to use)

### Data Source Tests
- **notification_settings.tf** - Notification settings data source
- **available_applications.tf** - Available applications and application family data sources
- **environments_data.tf** - Environment and environments data sources
- **reference_data.tf** - Timezones and quotas data sources

## Prerequisites

1. **Build the provider** (from the repository root):
   ```bash
   go build -o terraform-provider-bcadmincenter
   ```

2. **Configure dev override** in `~/.terraformrc`:
   ```hcl
   provider_installation {
     dev_overrides {
       "vllni/bcadmincenter" = "/workspaces/terraform-provider-bcadmincenter"
     }
     direct {}
   }
   ```

   Or use the helper script:
   ```bash
   ../scripts/setup-local-testing.sh
   ```

3. **Set Azure credentials**:
   ```bash
   export AZURE_CLIENT_ID="your-client-id"
   export AZURE_CLIENT_SECRET="your-client-secret"
   export AZURE_TENANT_ID="your-tenant-id"
   ```

## Usage

### Testing Specific Resources/Data Sources

You can selectively enable test files by commenting out unwanted configurations or using targeted applies:

```bash
# Test only data sources
terraform plan -target=data.bcadmincenter_quotas.tenant

# Test only specific resource
terraform plan -target=bcadmincenter_environment.test

# Test notification system
terraform plan -target=bcadmincenter_notification_recipient.test
```

### Full Test Workflow

1. Initialize Terraform:
   ```bash
   terraform init
   ```
   
   You should see a warning about development overrides - this is expected.

2. Review the plan:
   ```bash
   terraform plan
   ```

3. Apply the configuration:
   ```bash
   terraform apply
   ```

### Testing Individual Components

#### Data Sources Only (No Infrastructure Created)
```bash
# Test quotas and capacity planning
terraform plan -target=data.bcadmincenter_quotas.tenant

# Test timezone lookup
terraform plan -target=data.bcadmincenter_timezones.all

# Test available applications
terraform plan -target=data.bcadmincenter_available_applications.all
```

#### Resources (Creates Infrastructure)
**WARNING**: Resource tests will create actual environments and configurations in your BC tenant.

```bash
# Test environment creation (creates real sandbox)
terraform apply -target=bcadmincenter_environment.test

# Test notification recipient
terraform apply -target=bcadmincenter_notification_recipient.test
```

4. Clean up when done:
   ```bash
   terraform destroy
   ```

## What This Tests

This configuration will:
- Connect to the Business Central Admin Center API using your Azure credentials
- Create a sandbox environment
- Output the web client URL for accessing the environment

## Modifying Tests

Edit `main.tf` to test different scenarios:
- Different environment types (Production vs Sandbox)
- Different country codes
- Different Azure regions
- Multiple environments using `for_each`

## Debugging

Enable debug logging:
```bash
export TF_LOG=DEBUG
export TF_LOG_PATH=./terraform.log
terraform plan
```

View only provider logs:
```bash
export TF_LOG_PROVIDER=TRACE
terraform plan
```

## Quick Iteration

After making changes to the provider:
```bash
# In the repository root
go build -o terraform-provider-bcadmincenter

# In this directory
terraform plan  # Test your changes immediately
```

No need to run `terraform init` again when using dev overrides!

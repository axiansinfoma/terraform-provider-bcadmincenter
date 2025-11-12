# Test Configuration for BC Admin Center Provider

This directory contains a sample Terraform configuration for testing the BC Admin Center provider locally.

## Prerequisites

1. **Build the provider** (from the repository root):
   ```bash
   go build -o terraform-provider-bc-admin-center
   ```

2. **Configure dev override** in `~/.terraformrc`:
   ```hcl
   provider_installation {
     dev_overrides {
       "vllni/bc-admin-center" = "/workspaces/terraform-provider-bc-admin-center"
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
go build -o terraform-provider-bc-admin-center

# In this directory
terraform plan  # Test your changes immediately
```

No need to run `terraform init` again when using dev overrides!

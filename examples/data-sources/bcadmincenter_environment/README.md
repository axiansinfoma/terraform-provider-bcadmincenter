# Environment Data Source Example

This example demonstrates how to retrieve information about a specific Business Central environment using the `bcadmincenter_environment` data source.

## Usage

```bash
# Set your authentication credentials
export AZURE_CLIENT_ID="your-client-id"
export AZURE_CLIENT_SECRET="your-client-secret"
export AZURE_TENANT_ID="your-tenant-id"

# Initialize Terraform
terraform init

# View the plan
terraform plan

# Apply the configuration (this only reads data, no resources are created)
terraform apply
```

## What This Does

- Retrieves details about the specified environment
- Outputs the environment status, web client URL, and application version

## Required Permissions

Your Azure AD application must have:
- `AdminCenter.ReadWrite.All` permission
- Membership in the **AdminAgents** group for the Business Central tenant

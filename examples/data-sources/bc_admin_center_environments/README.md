# Environments Data Source Example

This example demonstrates how to list all Business Central environments using the `bcadmincenter_environments` data source.

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

- Retrieves a list of all environments in the tenant
- Filters and outputs environment names by type (Production/Sandbox)
- Creates a map of environment names to their web client URLs

## Use Cases

This data source is useful for:
- Creating an inventory of all environments
- Filtering environments by type or other attributes
- Dynamically creating resources based on existing environments
- Building monitoring and reporting configurations

## Required Permissions

Your Azure AD application must have:
- `AdminCenter.ReadWrite.All` permission
- Membership in the **AdminAgents** group for the Business Central tenant

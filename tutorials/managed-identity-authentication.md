# Authenticating with Managed Identity

This tutorial demonstrates how to authenticate the Business Central Admin Center provider using Azure Managed Identity. This method is ideal for Terraform running on Azure compute resources like Virtual Machines, Container Instances, or App Service.

## Prerequisites

- An Azure compute resource (VM, Container Instance, App Service, etc.)
- Appropriate permissions to assign managed identities
- Access to Business Central Admin Center as an admin

## What is Managed Identity?

Managed Identity is a feature of Azure Active Directory that provides Azure services with an automatically managed identity. This eliminates the need to manage credentials in your code or configuration.

### Types of Managed Identity

- **System-assigned**: Tied to the lifecycle of an Azure resource
- **User-assigned**: Standalone identity that can be shared across multiple resources

## Step 1: Enable Managed Identity

Choose the appropriate method based on your Azure resource:

### For Azure Virtual Machine

```bash
# Enable system-assigned managed identity
az vm identity assign \
  --name myVM \
  --resource-group myResourceGroup

# Get the principal ID
PRINCIPAL_ID=$(az vm identity show \
  --name myVM \
  --resource-group myResourceGroup \
  --query principalId \
  --output tsv)

echo "Managed Identity Principal ID: $PRINCIPAL_ID"
```

### For Azure Container Instance

```bash
# Create container instance with managed identity
az container create \
  --resource-group myResourceGroup \
  --name myContainerInstance \
  --image myregistry.azurecr.io/terraform-runner:latest \
  --assign-identity \
  --cpu 2 \
  --memory 4

# Get the principal ID
PRINCIPAL_ID=$(az container show \
  --name myContainerInstance \
  --resource-group myResourceGroup \
  --query identity.principalId \
  --output tsv)
```

### For Azure App Service

```bash
# Enable system-assigned managed identity
az webapp identity assign \
  --name myAppService \
  --resource-group myResourceGroup

# Get the principal ID
PRINCIPAL_ID=$(az webapp identity show \
  --name myAppService \
  --resource-group myResourceGroup \
  --query principalId \
  --output tsv)
```

### For User-Assigned Managed Identity

```bash
# Create a user-assigned managed identity
az identity create \
  --name myTerraformIdentity \
  --resource-group myResourceGroup

# Get the principal ID and client ID
PRINCIPAL_ID=$(az identity show \
  --name myTerraformIdentity \
  --resource-group myResourceGroup \
  --query principalId \
  --output tsv)

CLIENT_ID=$(az identity show \
  --name myTerraformIdentity \
  --resource-group myResourceGroup \
  --query clientId \
  --output tsv)

# Assign the identity to your resource (example: VM)
az vm identity assign \
  --name myVM \
  --resource-group myResourceGroup \
  --identities myTerraformIdentity
```

## Step 2: Grant API Permissions

The managed identity needs the AdminCenter.ReadWrite.All application permission:

```bash
# Business Central Admin Center API Application ID
BC_ADMIN_API="996def3d-b36c-4153-8607-a6fd3c01b89f"

# AdminCenter.ReadWrite.All permission ID (Application permission)
PERMISSION_ID="2e3cf0a5-be71-42b6-8b82-6f50da52005d"

# Get the Business Central Admin Center service principal
BC_SP_ID=$(az ad sp show --id $BC_ADMIN_API --query id --output tsv)

# Assign the app role to the managed identity
az rest --method POST \
  --uri "https://graph.microsoft.com/v1.0/servicePrincipals/$PRINCIPAL_ID/appRoleAssignments" \
  --headers "Content-Type=application/json" \
  --body "{
    \"principalId\": \"$PRINCIPAL_ID\",
    \"resourceId\": \"$BC_SP_ID\",
    \"appRoleId\": \"$PERMISSION_ID\"
  }"
```

> **Note**: This requires Azure AD Global Administrator or Privileged Role Administrator privileges.

## Step 3: Configure Business Central Admin Center Access

Two actions are required in the Business Central Admin Center before the provider can make any API calls.

### Add Managed Identity to Authorized Entra Apps

> **Important:** This step is required before running any Terraform commands. The provider will fail with an authorization error if the managed identity has not been added here first.

1. Navigate to the [Business Central Admin Center](https://businesscentral.dynamics.com/admin)
2. Go to **Settings** > **Authorized Microsoft Entra Apps**
3. Click **New**
4. Enter the Client ID of your managed identity (`$CLIENT_ID`) and click **OK**

> **Note:** This step requires Business Central Admin Center administrator privileges and cannot be performed through the provider itself.

### Add Managed Identity to AdminAgents Group

For delegated admin access across tenants, also add the managed identity to the AdminAgents group:

1. Navigate to the [Business Central Admin Center](https://businesscentral.dynamics.com/admin)
2. Go to **Settings** > **Admin Center API**
3. Click **Add** under AdminAgents
4. Search for the managed identity name (e.g., "myVM", "myTerraformIdentity")
5. Select the identity and click **Add**

## Step 4: Configure the Provider

### For System-Assigned Managed Identity

On the Azure resource, create your Terraform configuration:

```terraform
terraform {
  required_providers {
    bcadmincenter = {
      source  = "axiansinfoma/bcadmincenter"
      version = "~> 1.0"
    }
  }
}

provider "bcadmincenter" {
  use_msi = true  # Enable Managed Identity authentication
  
  # Optionally specify tenant ID if managing multiple tenants
  # tenant_id = "00000000-0000-0000-0000-000000000000"
}
```

### For User-Assigned Managed Identity

If using a user-assigned managed identity, specify the client ID:

```terraform
provider "bcadmincenter" {
  use_msi           = true
  msi_client_id     = "00000000-0000-0000-0000-000000000000"  # User-assigned identity client ID
  
  # Optionally specify tenant ID
  # tenant_id = "00000000-0000-0000-0000-000000000000"
}
```

Alternatively, use environment variables:

```bash
export AZURE_CLIENT_ID="00000000-0000-0000-0000-000000000000"  # For user-assigned identity
```

## Step 5: Test the Configuration

Create a simple test configuration:

```terraform
data "bcadmincenter_environments" "all" {
  application_family = "BusinessCentral"
}

output "environment_count" {
  value = length(data.bcadmincenter_environments.all.environments)
}
```

Run Terraform from within the Azure resource:

```bash
terraform init
terraform plan
```

## Use Cases

### Continuous Deployment from Azure Container Instance

```yaml
# Docker Compose example
version: '3.8'
services:
  terraform:
    image: hashicorp/terraform:latest
    volumes:
      - ./terraform:/workspace
    working_dir: /workspace
    environment:
      - AZURE_USE_MSI=true
    command: 
      - apply
      - -auto-approve
```

### Scheduled Terraform Runs on Azure VM

```bash
#!/bin/bash
# /usr/local/bin/terraform-apply.sh

cd /opt/terraform/bc-admin-center

# Managed Identity is automatically used
terraform init
terraform plan -out=tfplan
terraform apply tfplan
```

Create a cron job:
```bash
# Run Terraform every day at 2 AM
0 2 * * * /usr/local/bin/terraform-apply.sh >> /var/log/terraform.log 2>&1
```

### Azure App Service with Terraform

Deploy Terraform configurations through an Azure App Service:

```bash
# In your App Service deployment script
cd /home/site/wwwroot/terraform
export AZURE_USE_MSI=true
terraform init
terraform apply -auto-approve
```

## Architecture Patterns

### Centralized Terraform Runner

```
┌─────────────────────────────────────┐
│   Azure Container Instance          │
│   ┌─────────────────────────────┐  │
│   │  Terraform Runner            │  │
│   │  - Managed Identity enabled  │  │
│   │  - Scheduled runs            │  │
│   └─────────────────────────────┘  │
└─────────────────────────────────────┘
              │
              │ Authenticates with Managed Identity
              ▼
┌─────────────────────────────────────┐
│  Business Central Admin Center API  │
└─────────────────────────────────────┘
```

## Advantages of Managed Identity

1. **No credential management**: No secrets to store or rotate
2. **Automatic rotation**: Azure handles token lifecycle
3. **Secure**: Credentials never leave Azure platform
4. **Simple**: Minimal configuration required
5. **Auditable**: All actions tied to a specific identity

## Limitations

1. **Azure-only**: Only works on Azure compute resources
2. **Network dependency**: Requires access to Azure Instance Metadata Service (IMDS)
3. **Setup required**: Requires initial configuration of managed identity

## Troubleshooting

### "Failed to obtain token from Managed Identity"

Check that:
- Managed identity is enabled on the resource
- The resource has network access to IMDS endpoint (169.254.169.254)
- You're running on a supported Azure compute service

### Permission Denied Errors

Verify:
- The managed identity has the AdminCenter.ReadWrite.All permission
- The identity is added to the AdminAgents group
- API permissions have been granted admin consent

### User-Assigned Identity Not Found

Ensure:
- The correct client ID is specified
- The identity is assigned to the Azure resource
- The identity exists in the correct tenant

### Testing Managed Identity Locally

You can test managed identity authentication locally using Azure CLI:

```bash
# Simulate managed identity token acquisition
az account get-access-token \
  --resource 996def3d-b36c-4153-8607-a6fd3c01b89f \
  --query accessToken \
  --output tsv
```

## Best Practices

1. **Use system-assigned when possible**: Simpler lifecycle management
2. **Use user-assigned for shared identity**: When multiple resources need the same permissions
3. **Implement least privilege**: Only grant necessary permissions
4. **Monitor usage**: Enable diagnostic logs for your Azure resources
5. **Document identity assignments**: Keep track of which resources use which identities

## Security Considerations

- Managed identities cannot be used outside of Azure
- Protect access to the Azure resource (it has admin privileges)
- Use Azure RBAC to control who can modify managed identity assignments
- Enable logging and monitoring for the Azure resource

## Next Steps

- Learn about [workload identity for GitHub Actions](./workload-identity-github.md)
- Set up [workload identity for Azure DevOps](./workload-identity-azure-devops.md)
- Configure [service principal authentication](./service-principal-authentication.md) for hybrid scenarios
- Explore [Azure CLI authentication](./azure-cli-authentication.md) for development

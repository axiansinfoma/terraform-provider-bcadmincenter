---
page_title: "Guide: Authenticating using a Service Principal and Client Secret"
subcategory: "Authentication"
description: |-
  This guide demonstrates how to authenticate the Business Central Admin Center provider using a service principal with a client secret. This method is suitable for automated scenarios where interactive login is not possible.
---

# Authenticating with Service Principal and Client Secret

This tutorial demonstrates how to authenticate the Business Central Admin Center provider using a service principal with a client secret. This method is suitable for automated scenarios where interactive login is not possible.

## Prerequisites

- Azure CLI installed and configured
- Appropriate permissions to create Azure AD applications and service principals
- Access to Business Central Admin Center as an admin

## Step 1: Create an Azure AD Application

First, create an Azure AD application that will represent your Terraform automation:

```bash
# Create the application
APP_NAME="Terraform-BC-Admin-Center"
APP_ID=$(az ad app create \
  --display-name "$APP_NAME" \
  --query appId \
  --output tsv)

echo "Application ID: $APP_ID"
```

## Step 2: Create a Service Principal

Create a service principal for the application:

```bash
az ad sp create --id $APP_ID
```

## Step 3: Grant API Permissions

Grant the required permissions to access the Business Central Admin Center API:

```bash
# Business Central Admin Center API Application ID
BC_ADMIN_API="996def3d-b36c-4153-8607-a6fd3c01b89f"

# AdminCenter.ReadWrite.All permission ID
PERMISSION_ID="2e3cf0a5-be71-42b6-8b82-6f50da52005d"

# Add the API permission
az ad app permission add \
  --id $APP_ID \
  --api $BC_ADMIN_API \
  --api-permissions ${PERMISSION_ID}=Role

# Grant admin consent (requires Global Administrator or Privileged Role Administrator)
az ad app permission admin-consent --id $APP_ID
```

## Step 4: Create a Client Secret

Create a client secret for authentication:

```bash
# Create a secret that expires in 1 year
SECRET_OUTPUT=$(az ad app credential reset \
  --id $APP_ID \
  --append \
  --years 1 \
  --query password \
  --output tsv)

echo "Client Secret: $SECRET_OUTPUT"
```

> **Warning**: Store this secret securely! It will not be shown again. Consider using Azure Key Vault or a secure secrets management system.

## Step 5: Get Your Tenant ID

Retrieve your Azure AD tenant ID:

```bash
TENANT_ID=$(az account show --query tenantId --output tsv)
echo "Tenant ID: $TENANT_ID"
```

## Step 6: Configure Business Central Admin Center Access

Two actions are required in the Business Central Admin Center before the provider can make any API calls.

### Add Service Principal to Authorized Entra Apps

~> **Important:** This step is required before running any Terraform commands. The provider will fail with an authorization error if the service principal has not been added here first.

1. Navigate to the [Business Central Admin Center](https://businesscentral.dynamics.com/admin)
2. Go to **Settings** > **Authorized Microsoft Entra Apps**
3. Click **New**
4. Enter the Application (client) ID of your service principal (`$APP_ID`) and click **OK**

> **Note:** This step requires Business Central Admin Center administrator privileges and cannot be performed through the Terraform provider itself. You must complete it manually before running `terraform init` or `terraform plan`.

### Add Service Principal to AdminAgents Group

For delegated admin access across tenants, also add the service principal to the AdminAgents group:

1. Navigate to the [Business Central Admin Center](https://businesscentral.dynamics.com/admin)
2. Go to **Settings** > **Admin Center API**
3. Click **Add** under AdminAgents
4. Search for your application name ("Terraform-BC-Admin-Center")
5. Select the service principal and click **Add**

## Step 7: Configure the Provider

Create a Terraform configuration file with the provider settings:

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
  client_id     = "00000000-0000-0000-0000-000000000000"  # Replace with your Application ID
  client_secret = "your-client-secret"                     # Replace with your client secret
  tenant_id     = "00000000-0000-0000-0000-000000000000"  # Replace with your Tenant ID
}
```

## Step 8: Use Environment Variables (Recommended)

For better security, use environment variables instead of hardcoding credentials:

```bash
export ARM_CLIENT_ID="$APP_ID"
export ARM_CLIENT_SECRET="$SECRET_OUTPUT"
export ARM_TENANT_ID="$TENANT_ID"
```

Then use a simplified provider configuration:

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
  # Authentication will use environment variables
}
```

## Step 9: Test the Configuration

Create a simple test configuration to verify authentication:

```terraform
data "bcadmincenter_environments" "all" {
  application_family = "BusinessCentral"
}

output "environment_count" {
  value = length(data.bcadmincenter_environments.all.environments)
}
```

Run Terraform commands:

```bash
terraform init
terraform plan
```

If authentication is successful, you should see the number of environments in your output.

## Security Best Practices

1. **Never commit secrets to version control**: Use `.gitignore` to exclude files containing secrets
2. **Rotate secrets regularly**: Set up a schedule to rotate client secrets
3. **Use least privilege**: Only grant the minimum required permissions
4. **Store secrets securely**: Use Azure Key Vault, HashiCorp Vault, or your CI/CD system's secret management
5. **Monitor access**: Enable audit logging for your service principal

## Troubleshooting

### Permission Denied Errors

If you receive permission errors, verify:

- The API permissions were granted admin consent
- The service principal was added to the AdminAgents group
- You're using the correct tenant ID

### Authentication Failures

Check that:

- The client ID, client secret, and tenant ID are correct
- The client secret hasn't expired
- There are no network connectivity issues

## Next Steps

- Learn about [Azure CLI authentication for local development](./azure-cli-authentication.md)
- Set up [workload identity for GitHub Actions](./workload-identity-github.md)
- Configure [workload identity for Azure DevOps](./workload-identity-azure-devops.md)
- Explore [managed identity authentication](./managed-identity-authentication.md)

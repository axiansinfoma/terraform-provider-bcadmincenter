# Authenticating with Azure CLI (Local Development)

This tutorial demonstrates how to authenticate the Business Central Admin Center provider using Azure CLI credentials. This method is ideal for local development and interactive scenarios.

## Prerequisites

- [Azure CLI](https://learn.microsoft.com/en-us/cli/azure/install-azure-cli) installed
- Access to Business Central Admin Center as an admin
- Your user account added to the AdminAgents group

## Step 1: Install Azure CLI

If you haven't already installed Azure CLI, follow the installation instructions for your operating system:

**Windows:**
```powershell
winget install Microsoft.AzureCLI
```

**macOS:**
```bash
brew install azure-cli
```

**Linux (Ubuntu/Debian):**
```bash
curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash
```

## Step 2: Login to Azure

Authenticate with your Azure account:

```bash
az login
```

This will open a browser window for interactive authentication. Sign in with your organizational account that has access to Business Central Admin Center.

### Verify Your Login

Check your current account and tenant:

```bash
az account show
```

If you have access to multiple tenants, ensure you're logged in to the correct one:

```bash
# List all available tenants
az account list --query "[].{Name:name, TenantId:tenantId}" --output table

# Switch to a specific tenant if needed
az account set --subscription "<subscription-id-or-name>"
```

## Step 3: Add Your User to AdminAgents Group

Your user account must be added to the AdminAgents group in Business Central Admin Center:

1. Navigate to the [Business Central Admin Center](https://businesscentral.dynamics.com/admin)
2. Go to **Settings** > **Admin Center API**
3. Click **Add** under AdminAgents
4. Search for your user account
5. Select your account and click **Add**

> **Note**: This requires Business Central Admin Center administrator privileges. If you don't have access, ask your administrator to add your account.

## Step 4: Configure the Provider

Create a Terraform configuration that uses Azure CLI authentication:

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
  # When no credentials are specified, the provider will automatically
  # attempt to use Azure CLI authentication
  use_cli = true
}
```

Alternatively, you can omit all authentication parameters, and the provider will automatically fall back to Azure CLI:

```terraform
provider "bcadmincenter" {
  # Authentication will automatically use Azure CLI
}
```

## Step 5: Test the Configuration

Create a simple test to verify authentication:

```terraform
data "bcadmincenter_environments" "all" {
  application_family = "BusinessCentral"
}

output "environment_names" {
  value = [for env in data.bcadmincenter_environments.all.environments : env.name]
}
```

Run Terraform commands:

```bash
terraform init
terraform plan
```

If authentication is successful, you should see your Business Central environments listed.

## Step 6: Working with Multiple Tenants

If you manage multiple Business Central tenants, you can specify the tenant ID explicitly:

```terraform
provider "bcadmincenter" {
  tenant_id = "00000000-0000-0000-0000-000000000000"  # Specific tenant
  use_cli   = true
}
```

You can also switch tenants using Azure CLI:

```bash
az login --tenant <tenant-id>
```

## Authentication Flow

When using Azure CLI authentication, the provider:

1. Locates the Azure CLI configuration (`~/.azure` directory)
2. Reads the cached access token for your account
3. Refreshes the token if needed
4. Uses the token to authenticate with Business Central Admin Center API

## Advantages of Azure CLI Authentication

- **Quick setup**: No need to create service principals
- **Interactive**: Works with MFA and Conditional Access policies
- **Convenient**: Uses your existing Azure credentials
- **Secure**: Tokens are cached securely by Azure CLI

## Limitations

- **Not suitable for automation**: Requires interactive login
- **Session expiration**: Tokens expire and require re-authentication
- **Single user**: Only works with the logged-in user's permissions
- **Not recommended for CI/CD**: Use service principals or workload identity instead

## Switching Between Authentication Methods

You can easily switch between authentication methods by changing the provider configuration:

```terraform
# Development (Azure CLI)
provider "bcadmincenter" {
  use_cli = true
}

# Production (Service Principal)
# provider "bcadmincenter" {
#   client_id     = var.client_id
#   client_secret = var.client_secret
#   tenant_id     = var.tenant_id
# }
```

## Troubleshooting

### "No Azure CLI Installation Found"

Ensure Azure CLI is installed and available in your PATH:

```bash
az --version
```

### "Please run 'az login' to setup account"

You need to authenticate first:

```bash
az login
```

### "No subscriptions found"

Your account might not have access to any Azure subscriptions. Verify with:

```bash
az account list
```

### Permission Denied Errors

Verify:
- You're logged in to the correct tenant
- Your user account is in the AdminAgents group
- You have the necessary permissions in Business Central Admin Center

### Token Expired

Re-authenticate with Azure CLI:

```bash
az login
```

## Best Practices

1. **Use for development only**: Azure CLI authentication is best for local development
2. **Keep Azure CLI updated**: Regularly update to get security fixes
   ```bash
   az upgrade
   ```
3. **Clear credentials when done**: Logout when switching contexts
   ```bash
   az logout
   ```
4. **Use service principals for automation**: Don't use Azure CLI auth in CI/CD pipelines

## Next Steps

- Set up [service principal authentication](./service-principal-authentication.md) for automation
- Configure [workload identity for GitHub Actions](./workload-identity-github.md)
- Learn about [managed identity authentication](./managed-identity-authentication.md)
- Explore [workload identity for Azure DevOps](./workload-identity-azure-devops.md)

---
page_title: "Guide: Authenticating using Workload Identity in GitHub Actions"
subcategory: "Authentication"
description: |-
  This guide demonstrates how to authenticate the Business Central Admin Center provider in GitHub Actions using Azure AD Workload Identity Federation with OIDC. This modern, secure approach eliminates the need to store long-lived secrets in GitHub.
---

# Authenticating with Workload Identity in GitHub Actions

This tutorial demonstrates how to authenticate the Business Central Admin Center provider in GitHub Actions using Azure AD Workload Identity Federation with OIDC. This modern, secure approach eliminates the need to store long-lived secrets in GitHub.

## Prerequisites

- An Azure AD tenant with appropriate permissions
- A GitHub repository
- Permissions to create Azure AD applications and federated credentials
- Access to Business Central Admin Center as an admin

## What is Workload Identity Federation?

Workload Identity Federation allows GitHub Actions to authenticate to Azure AD using OpenID Connect (OIDC) tokens instead of storing secrets. GitHub generates short-lived tokens that Azure AD trusts, providing secure, credential-free authentication.

### Benefits

- **No secrets to manage**: No client secrets stored in GitHub
- **Automatic rotation**: Tokens are short-lived and automatically renewed
- **More secure**: Eliminates secret sprawl and exposure risks
- **Auditable**: Clear identity mapping between GitHub and Azure AD

## Step 1: Create an Azure AD Application

Create an application in Azure AD for your GitHub Actions workflow:

```bash
# Set variables
APP_NAME="GitHub-Terraform-BC-Admin"
REPO_OWNER="your-github-org"
REPO_NAME="your-repo-name"

# Create the application
APP_ID=$(az ad app create \
  --display-name "$APP_NAME" \
  --query appId \
  --output tsv)

echo "Application ID: $APP_ID"

# Create service principal
az ad sp create --id $APP_ID

# Get the Object ID of the application (needed for federated credentials)
APP_OBJECT_ID=$(az ad app show --id $APP_ID --query id --output tsv)
```

## Step 2: Configure Federated Credentials

Add federated identity credentials to trust GitHub Actions OIDC tokens:

### For Production Deployments (main branch)

```bash
# Create federated credential for main branch
az ad app federated-credential create \
  --id $APP_OBJECT_ID \
  --parameters '{
    "name": "GitHubActionsMain",
    "issuer": "https://token.actions.githubusercontent.com",
    "subject": "repo:'"$REPO_OWNER"'/'"$REPO_NAME"':ref:refs/heads/main",
    "description": "GitHub Actions deployments from main branch",
    "audiences": [
      "api://AzureADTokenExchange"
    ]
  }'
```

### For Pull Request Validation

```bash
# Create federated credential for pull requests
az ad app federated-credential create \
  --id $APP_OBJECT_ID \
  --parameters '{
    "name": "GitHubActionsPullRequests",
    "issuer": "https://token.actions.githubusercontent.com",
    "subject": "repo:'"$REPO_OWNER"'/'"$REPO_NAME"':pull_request",
    "description": "GitHub Actions for pull request validation",
    "audiences": [
      "api://AzureADTokenExchange"
    ]
  }'
```

### For Specific Environments

```bash
# Create federated credential for GitHub Environment (e.g., "production")
az ad app federated-credential create \
  --id $APP_OBJECT_ID \
  --parameters '{
    "name": "GitHubActionsProduction",
    "issuer": "https://token.actions.githubusercontent.com",
    "subject": "repo:'"$REPO_OWNER"'/'"$REPO_NAME"':environment:production",
    "description": "GitHub Actions for production environment",
    "audiences": [
      "api://AzureADTokenExchange"
    ]
  }'
```

## Step 3: Grant API Permissions

Grant the required permissions to access Business Central Admin Center:

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

# Grant admin consent
az ad app permission admin-consent --id $APP_ID
```

## Step 4: Configure Business Central Admin Center Access

Two actions are required in the Business Central Admin Center before the provider can make any API calls.

### Add Service Principal to Authorized Entra Apps

~> **Important:** This step is required before running any Terraform commands. The provider will fail with an authorization error if the service principal has not been added here first.

1. Navigate to the [Business Central Admin Center](https://businesscentral.dynamics.com/admin)
2. Go to **Settings** > **Authorized Microsoft Entra Apps**
3. Click **New**
4. Enter the Application (client) ID of your service principal (`$APP_ID`) and click **OK**

> **Note:** This step requires Business Central Admin Center administrator privileges and cannot be performed through the provider itself.

### Add Service Principal to AdminAgents Group

For delegated admin access across tenants, also add the service principal to the AdminAgents group:

1. Navigate to the [Business Central Admin Center](https://businesscentral.dynamics.com/admin)
2. Go to **Settings** > **Admin Center API**
3. Click **Add** under AdminAgents
4. Search for your application name ("GitHub-Terraform-BC-Admin")
5. Select the service principal and click **Add**

## Step 5: Get Your Tenant ID

```bash
TENANT_ID=$(az account show --query tenantId --output tsv)
echo "Tenant ID: $TENANT_ID"
```

## Step 6: Configure GitHub Secrets

Add the following secrets to your GitHub repository (Settings → Secrets and variables → Actions):

- `AZURE_CLIENT_ID`: The Application (client) ID from Step 1
- `AZURE_TENANT_ID`: Your Azure AD tenant ID

> **Note**: You do NOT need to store a client secret with workload identity!

## Step 7: Create GitHub Actions Workflow

Create `.github/workflows/terraform.yml`:

```yaml
name: Terraform Business Central Admin Center

on:
  push:
    branches:
      - main
  pull_request:
    branches:
      - main

permissions:
  id-token: write  # Required for OIDC token
  contents: read   # Required to checkout code

jobs:
  terraform:
    name: Terraform Plan and Apply
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Azure Login with OIDC
        uses: azure/login@v2
        with:
          client-id: ${{ secrets.AZURE_CLIENT_ID }}
          tenant-id: ${{ secrets.AZURE_TENANT_ID }}
          subscription-id: ${{ secrets.AZURE_SUBSCRIPTION_ID }}  # Optional
          
      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: 1.9.0

      - name: Terraform Init
        run: terraform init
        working-directory: ./terraform

      - name: Terraform Format Check
        run: terraform fmt -check
        working-directory: ./terraform

      - name: Terraform Validate
        run: terraform validate
        working-directory: ./terraform

      - name: Terraform Plan
        run: terraform plan -out=tfplan
        working-directory: ./terraform
        env:
          AZURE_CLIENT_ID: ${{ secrets.AZURE_CLIENT_ID }}
          AZURE_TENANT_ID: ${{ secrets.AZURE_TENANT_ID }}
          # Workload identity uses OIDC tokens - no secret needed!

      - name: Terraform Apply
        if: github.ref == 'refs/heads/main' && github.event_name == 'push'
        run: terraform apply -auto-approve tfplan
        working-directory: ./terraform
        env:
          AZURE_CLIENT_ID: ${{ secrets.AZURE_CLIENT_ID }}
          AZURE_TENANT_ID: ${{ secrets.AZURE_TENANT_ID }}
```

## Step 8: Configure Terraform Provider

In your Terraform configuration (`terraform/main.tf`):

```terraform
terraform {
  required_version = ">= 1.0"
  
  required_providers {
    bcadmincenter = {
      source  = "axiansinfoma/bcadmincenter"
      version = "~> 1.0"
    }
  }
}

provider "bcadmincenter" {
  # Authentication will use environment variables and OIDC tokens
  # No client_secret needed!
  
  # These can be set via environment variables:
  # AZURE_CLIENT_ID
  # AZURE_TENANT_ID
  
  # The provider will automatically use workload identity when:
  # 1. Running in GitHub Actions with OIDC token available
  # 2. AZURE_CLIENT_ID and AZURE_TENANT_ID are set
  # 3. No client_secret is provided
}
```

## Step 9: Test the Workflow

1. Commit and push your changes to a branch
2. Create a pull request
3. Check the Actions tab to see the workflow run
4. Verify that Terraform plan executes successfully

## Advanced Configuration

### Using GitHub Environments

Configure deployment protection rules with environments:

```yaml
jobs:
  terraform-production:
    name: Deploy to Production
    runs-on: ubuntu-latest
    environment: production  # GitHub Environment with protection rules
    
    steps:
      - name: Azure Login with OIDC
        uses: azure/login@v2
        with:
          client-id: ${{ secrets.AZURE_CLIENT_ID }}
          tenant-id: ${{ secrets.AZURE_TENANT_ID }}
      
      # ... rest of the steps
```

### Matrix Builds for Multiple Tenants

Deploy to multiple Business Central tenants:

```yaml
jobs:
  terraform:
    name: Deploy to ${{ matrix.tenant.name }}
    runs-on: ubuntu-latest
    
    strategy:
      matrix:
        tenant:
          - name: production
            tenant_id: ${{ secrets.PROD_TENANT_ID }}
            client_id: ${{ secrets.PROD_CLIENT_ID }}
          - name: staging
            tenant_id: ${{ secrets.STAGING_TENANT_ID }}
            client_id: ${{ secrets.STAGING_CLIENT_ID }}
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Azure Login
        uses: azure/login@v2
        with:
          client-id: ${{ matrix.tenant.client_id }}
          tenant-id: ${{ matrix.tenant.tenant_id }}
      
      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v3
      
      - name: Terraform Apply
        run: terraform apply -auto-approve
        working-directory: ./terraform/${{ matrix.tenant.name }}
        env:
          AZURE_CLIENT_ID: ${{ matrix.tenant.client_id }}
          AZURE_TENANT_ID: ${{ matrix.tenant.tenant_id }}
```

### Terraform Backend Configuration

Use Azure Storage for remote state:

```yaml
      - name: Terraform Init
        run: |
          terraform init \
            -backend-config="storage_account_name=${{ secrets.TFSTATE_STORAGE_ACCOUNT }}" \
            -backend-config="container_name=tfstate" \
            -backend-config="key=bc-admin-center.tfstate" \
            -backend-config="use_azuread_auth=true"
        working-directory: ./terraform
```

With backend configuration in `backend.tf`:

```terraform
terraform {
  backend "azurerm" {
    resource_group_name  = "terraform-state-rg"
    storage_account_name = "tfstate"  # Override with -backend-config
    container_name       = "tfstate"
    key                  = "bc-admin-center.tfstate"
    use_azuread_auth     = true  # Use workload identity for state storage
  }
}
```

## Troubleshooting

### "OIDC token verification failed"

Check that:
- The federated credential subject matches your repository and branch
- The credential has the correct issuer: `https://token.actions.githubusercontent.com`
- The audience is `api://AzureADTokenExchange`

### "id-token permission required"

Ensure the workflow has the correct permissions:

```yaml
permissions:
  id-token: write
  contents: read
```

### "Token exchange failed"

Verify:
- The GitHub Actions workflow has `id-token: write` permission
- The Azure AD application has the federated credential configured
- The subject pattern matches your workflow context

### Debugging OIDC Token

Add a debug step to inspect the token:

```yaml
      - name: Debug OIDC Token
        run: |
          echo "GitHub OIDC Token:"
          curl -H "Authorization: bearer $ACTIONS_ID_TOKEN_REQUEST_TOKEN" \
            "$ACTIONS_ID_TOKEN_REQUEST_URL&audience=api://AzureADTokenExchange" | \
            jq -R 'split(".") | .[1] | @base64d | fromjson'
```

## Security Best Practices

1. **Use GitHub Environments**: Implement deployment protection rules
2. **Limit federated credentials**: Create specific credentials for each branch/environment
3. **Review permissions regularly**: Audit what the service principal can access
4. **Enable branch protection**: Require reviews before merging to main
5. **Use least privilege**: Only grant necessary permissions to the application
6. **Monitor workflow runs**: Enable notifications for failed workflows

## Subject Claim Patterns

Different subject patterns for federated credentials:

| Scenario | Subject Pattern |
|----------|----------------|
| Specific branch | `repo:ORG/REPO:ref:refs/heads/BRANCH` |
| Any branch | `repo:ORG/REPO:ref:refs/heads/*` |
| Pull requests | `repo:ORG/REPO:pull_request` |
| Specific tag | `repo:ORG/REPO:ref:refs/tags/TAG` |
| GitHub Environment | `repo:ORG/REPO:environment:ENV_NAME` |

## Comparison with Client Secrets

| Feature | Workload Identity | Client Secret |
|---------|------------------|---------------|
| Security | High (no stored secrets) | Medium (secrets in GitHub) |
| Rotation | Automatic | Manual |
| Expiration | Short-lived tokens | Secret expires (up to 2 years) |
| Setup Complexity | Medium | Low |
| GitHub Actions Support | Native OIDC | Requires secrets |

## Next Steps

- Learn about [workload identity for Azure DevOps](./workload-identity-azure-devops.md)
- Explore [service principal authentication](./service-principal-authentication.md) for other scenarios
- Set up [managed identity authentication](./managed-identity-authentication.md) for Azure-hosted runners
- Configure [Azure CLI authentication](./azure-cli-authentication.md) for local development

## Additional Resources

- [GitHub Actions OIDC Documentation](https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/about-security-hardening-with-openid-connect)
- [Azure Workload Identity Federation](https://learn.microsoft.com/en-us/azure/active-directory/develop/workload-identity-federation)
- [Configuring OpenID Connect in Azure](https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/configuring-openid-connect-in-azure)

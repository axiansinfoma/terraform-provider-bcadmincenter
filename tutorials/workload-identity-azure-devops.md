# Authenticating with Workload Identity in Azure DevOps

This tutorial demonstrates how to authenticate the Business Central Admin Center provider in Azure DevOps using Azure AD Workload Identity Federation with a service connection. This secure approach eliminates the need to store client secrets in Azure DevOps.

## Prerequisites

- An Azure AD tenant with appropriate permissions
- An Azure DevOps organization and project
- Permissions to create Azure AD applications and service connections
- Access to Business Central Admin Center as an admin

## What is Workload Identity Federation with Azure DevOps?

Workload Identity Federation allows Azure DevOps pipelines to authenticate to Azure AD using federated credentials instead of storing client secrets. Azure DevOps generates short-lived tokens that Azure AD trusts, providing secure authentication without managing secrets.

### Benefits

- **No secrets in Azure DevOps**: Client secrets are not stored in variable groups
- **Automatic token rotation**: Tokens are short-lived and automatically renewed
- **Enhanced security**: Eliminates secret sprawl and reduces exposure risks
- **Native integration**: Works seamlessly with Azure DevOps service connections

## Step 1: Create an Azure AD Application

Create an application for your Azure DevOps pipeline:

```bash
# Set variables
APP_NAME="AzureDevOps-Terraform-BC-Admin"
ORG_NAME="your-azdo-org"
PROJECT_NAME="your-project-name"

# Create the application
APP_ID=$(az ad app create \
  --display-name "$APP_NAME" \
  --query appId \
  --output tsv)

echo "Application ID: $APP_ID"

# Create service principal
PRINCIPAL_ID=$(az ad sp create --id $APP_ID --query id --output tsv)

# Get the Object ID of the application (needed for federated credentials)
APP_OBJECT_ID=$(az ad app show --id $APP_ID --query id --output tsv)

echo "Application Object ID: $APP_OBJECT_ID"
```

## Step 2: Configure Federated Credentials for Azure DevOps

Azure DevOps uses a different subject claim format than GitHub Actions. Configure federated credentials for your Azure DevOps organization and project:

### For Specific Service Connection

```bash
# Set your Azure DevOps organization details
AZDO_ORG_NAME="your-org-name"
AZDO_PROJECT_NAME="your-project-name"
SERVICE_CONNECTION_NAME="bc-admin-center-workload-identity"

# Create federated credential for the service connection
az ad app federated-credential create \
  --id $APP_OBJECT_ID \
  --parameters '{
    "name": "AzureDevOpsServiceConnection",
    "issuer": "https://vstoken.dev.azure.com/'"$AZDO_ORG_NAME"'",
    "subject": "sc://'"$AZDO_ORG_NAME"'/'"$AZDO_PROJECT_NAME"'/'"$SERVICE_CONNECTION_NAME"'",
    "description": "Azure DevOps service connection for BC Admin Center",
    "audiences": [
      "api://AzureADTokenExchange"
    ]
  }'
```

### Alternative: For Any Service Connection in Project

```bash
# Allow any service connection in the project
az ad app federated-credential create \
  --id $APP_OBJECT_ID \
  --parameters '{
    "name": "AzureDevOpsProject",
    "issuer": "https://vstoken.dev.azure.com/'"$AZDO_ORG_NAME"'",
    "subject": "sc://'"$AZDO_ORG_NAME"'/'"$AZDO_PROJECT_NAME"'/*",
    "description": "Azure DevOps service connections in project",
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

## Step 4: Add Service Principal to AdminAgents Group

Add the service principal to the AdminAgents group:

1. Navigate to the [Business Central Admin Center](https://businesscentral.dynamics.com/admin)
2. Go to **Settings** > **Admin Center API**
3. Click **Add** under AdminAgents
4. Search for your application name ("AzureDevOps-Terraform-BC-Admin")
5. Select the service principal and click **Add**

## Step 5: Get Required IDs

```bash
TENANT_ID=$(az account show --query tenantId --output tsv)
SUBSCRIPTION_ID=$(az account show --query id --output tsv)

echo "Tenant ID: $TENANT_ID"
echo "Subscription ID: $SUBSCRIPTION_ID"
echo "Application ID: $APP_ID"
```

## Step 6: Create Azure DevOps Service Connection

### Option A: Using Azure DevOps Portal

1. Navigate to your Azure DevOps project
2. Go to **Project Settings** → **Service connections**
3. Click **New service connection**
4. Select **Azure Resource Manager**
5. Choose **Workload Identity federation (automatic)**
6. Fill in the details:
   - **Service connection name**: `bc-admin-center-workload-identity`
   - **Subscription**: Select your Azure subscription
   - **Service principal**: Use existing or create new
   - **Application (client) ID**: Enter the `$APP_ID` from Step 1
   - **Tenant ID**: Enter the `$TENANT_ID` from Step 5
7. **Important**: After creating the service connection, note its name. If you used a different name than specified in Step 2, you must recreate the federated credential with the correct service connection name.

### Option B: Using Azure CLI and REST API

```bash
# This requires the Azure DevOps extension
az extension add --name azure-devops

# Set your organization and project
az devops configure --defaults organization=https://dev.azure.com/$AZDO_ORG_NAME project=$AZDO_PROJECT_NAME

# Note: Creating workload identity service connections via CLI requires additional REST API calls
# It's recommended to use the portal for initial setup
```

## Step 7: Create Azure DevOps Pipeline

Create `azure-pipelines.yml` in your repository:

```yaml
trigger:
  branches:
    include:
      - main
  paths:
    include:
      - terraform/*

pr:
  branches:
    include:
      - main

pool:
  vmImage: 'ubuntu-latest'

variables:
  - name: terraformWorkingDirectory
    value: '$(System.DefaultWorkingDirectory)/terraform'

stages:
  - stage: Validate
    displayName: 'Terraform Validate'
    jobs:
      - job: Validate
        displayName: 'Validate Terraform'
        steps:
          - task: TerraformInstaller@1
            displayName: 'Install Terraform'
            inputs:
              terraformVersion: '1.9.0'

          - task: AzureCLI@2
            displayName: 'Terraform Format Check'
            inputs:
              azureSubscription: 'bc-admin-center-workload-identity'
              scriptType: 'bash'
              scriptLocation: 'inlineScript'
              inlineScript: |
                cd $(terraformWorkingDirectory)
                terraform fmt -check -recursive
              addSpnToEnvironment: true
              useGlobalConfig: true

          - task: AzureCLI@2
            displayName: 'Terraform Init'
            inputs:
              azureSubscription: 'bc-admin-center-workload-identity'
              scriptType: 'bash'
              scriptLocation: 'inlineScript'
              inlineScript: |
                cd $(terraformWorkingDirectory)
                terraform init
              addSpnToEnvironment: true
              useGlobalConfig: true

          - task: AzureCLI@2
            displayName: 'Terraform Validate'
            inputs:
              azureSubscription: 'bc-admin-center-workload-identity'
              scriptType: 'bash'
              scriptLocation: 'inlineScript'
              inlineScript: |
                cd $(terraformWorkingDirectory)
                terraform validate
              addSpnToEnvironment: true
              useGlobalConfig: true

  - stage: Plan
    displayName: 'Terraform Plan'
    dependsOn: Validate
    jobs:
      - job: Plan
        displayName: 'Plan Terraform Changes'
        steps:
          - task: TerraformInstaller@1
            displayName: 'Install Terraform'
            inputs:
              terraformVersion: '1.9.0'

          - task: AzureCLI@2
            displayName: 'Terraform Plan'
            inputs:
              azureSubscription: 'bc-admin-center-workload-identity'
              scriptType: 'bash'
              scriptLocation: 'inlineScript'
              inlineScript: |
                cd $(terraformWorkingDirectory)
                
                # Export Azure credentials for Terraform provider
                export AZURE_CLIENT_ID=$servicePrincipalId
                export AZURE_TENANT_ID=$tenantId
                export AZURE_FEDERATED_TOKEN_FILE=$AZURE_FEDERATED_TOKEN_FILE
                
                terraform init
                terraform plan -out=tfplan
              addSpnToEnvironment: true
              useGlobalConfig: true
              
          - task: PublishPipelineArtifact@1
            displayName: 'Publish Terraform Plan'
            inputs:
              targetPath: '$(terraformWorkingDirectory)/tfplan'
              artifact: 'terraform-plan'
              publishLocation: 'pipeline'

  - stage: Apply
    displayName: 'Terraform Apply'
    dependsOn: Plan
    condition: and(succeeded(), eq(variables['Build.SourceBranch'], 'refs/heads/main'))
    jobs:
      - deployment: Apply
        displayName: 'Apply Terraform Changes'
        environment: 'production'  # Requires approval
        strategy:
          runOnce:
            deploy:
              steps:
                - checkout: self

                - task: TerraformInstaller@1
                  displayName: 'Install Terraform'
                  inputs:
                    terraformVersion: '1.9.0'

                - task: DownloadPipelineArtifact@2
                  displayName: 'Download Terraform Plan'
                  inputs:
                    artifact: 'terraform-plan'
                    path: '$(terraformWorkingDirectory)'

                - task: AzureCLI@2
                  displayName: 'Terraform Apply'
                  inputs:
                    azureSubscription: 'bc-admin-center-workload-identity'
                    scriptType: 'bash'
                    scriptLocation: 'inlineScript'
                    inlineScript: |
                      cd $(terraformWorkingDirectory)
                      
                      # Export Azure credentials for Terraform provider
                      export AZURE_CLIENT_ID=$servicePrincipalId
                      export AZURE_TENANT_ID=$tenantId
                      export AZURE_FEDERATED_TOKEN_FILE=$AZURE_FEDERATED_TOKEN_FILE
                      
                      terraform init
                      terraform apply -auto-approve tfplan
                    addSpnToEnvironment: true
                    useGlobalConfig: true
```

## Step 8: Configure Terraform Provider

In your Terraform configuration:

```terraform
terraform {
  required_version = ">= 1.0"
  
  required_providers {
    bcadmincenter = {
      source  = "vllni/bcadmincenter"
      version = "~> 1.0"
    }
  }
}

provider "bcadmincenter" {
  # Authentication will use environment variables and federated token
  # No client_secret needed!
  
  # These are automatically set by AzureCLI@2 task with addSpnToEnvironment: true
  # AZURE_CLIENT_ID
  # AZURE_TENANT_ID
  # AZURE_FEDERATED_TOKEN_FILE
  
  # The provider will automatically detect workload identity when:
  # 1. AZURE_CLIENT_ID and AZURE_TENANT_ID are set
  # 2. AZURE_FEDERATED_TOKEN_FILE is present
  # 3. No client_secret is provided
}
```

## Step 9: Configure Remote State (Optional)

Use Azure Storage for Terraform state with workload identity:

```yaml
- task: AzureCLI@2
  displayName: 'Terraform Init with Backend'
  inputs:
    azureSubscription: 'bc-admin-center-workload-identity'
    scriptType: 'bash'
    scriptLocation: 'inlineScript'
    inlineScript: |
      cd $(terraformWorkingDirectory)
      
      terraform init \
        -backend-config="storage_account_name=$(TF_STATE_STORAGE_ACCOUNT)" \
        -backend-config="container_name=tfstate" \
        -backend-config="key=bc-admin-center.tfstate" \
        -backend-config="use_azuread_auth=true"
    addSpnToEnvironment: true
```

Backend configuration in `backend.tf`:

```terraform
terraform {
  backend "azurerm" {
    resource_group_name  = "terraform-state-rg"
    storage_account_name = "tfstate"  # Override with pipeline variable
    container_name       = "tfstate"
    key                  = "bc-admin-center.tfstate"
    use_azuread_auth     = true  # Use workload identity for state access
  }
}
```

## Advanced Configurations

### Multi-Environment Deployment

Deploy to different Business Central tenants based on environment:

```yaml
stages:
  - stage: DeployDev
    displayName: 'Deploy to Development'
    variables:
      - group: 'bc-admin-dev'
    jobs:
      - deployment: DeployDev
        environment: 'development'
        pool:
          vmImage: 'ubuntu-latest'
        strategy:
          runOnce:
            deploy:
              steps:
                - task: AzureCLI@2
                  inputs:
                    azureSubscription: 'bc-admin-dev-workload-identity'
                    scriptType: 'bash'
                    scriptLocation: 'inlineScript'
                    inlineScript: |
                      export AZURE_CLIENT_ID=$servicePrincipalId
                      export AZURE_TENANT_ID=$tenantId
                      export AZURE_FEDERATED_TOKEN_FILE=$AZURE_FEDERATED_TOKEN_FILE
                      
                      cd terraform/environments/dev
                      terraform init
                      terraform apply -auto-approve
                    addSpnToEnvironment: true

  - stage: DeployProd
    displayName: 'Deploy to Production'
    dependsOn: DeployDev
    condition: succeeded()
    variables:
      - group: 'bc-admin-prod'
    jobs:
      - deployment: DeployProd
        environment: 'production'
        pool:
          vmImage: 'ubuntu-latest'
        strategy:
          runOnce:
            deploy:
              steps:
                - task: AzureCLI@2
                  inputs:
                    azureSubscription: 'bc-admin-prod-workload-identity'
                    scriptType: 'bash'
                    scriptLocation: 'inlineScript'
                    inlineScript: |
                      export AZURE_CLIENT_ID=$servicePrincipalId
                      export AZURE_TENANT_ID=$tenantId
                      export AZURE_FEDERATED_TOKEN_FILE=$AZURE_FEDERATED_TOKEN_FILE
                      
                      cd terraform/environments/prod
                      terraform init
                      terraform apply -auto-approve
                    addSpnToEnvironment: true
```

### Using Terraform Task Extension

Alternatively, use the [Terraform extension for Azure Pipelines](https://marketplace.visualstudio.com/items?itemName=ms-devlabs.custom-terraform-tasks):

```yaml
- task: TerraformTaskV4@4
  displayName: 'Terraform Init'
  inputs:
    provider: 'azurerm'
    command: 'init'
    workingDirectory: '$(terraformWorkingDirectory)'
    backendServiceArm: 'bc-admin-center-workload-identity'
    backendAzureRmResourceGroupName: 'terraform-state-rg'
    backendAzureRmStorageAccountName: '$(TF_STATE_STORAGE_ACCOUNT)'
    backendAzureRmContainerName: 'tfstate'
    backendAzureRmKey: 'bc-admin-center.tfstate'

- task: TerraformTaskV4@4
  displayName: 'Terraform Plan'
  inputs:
    provider: 'azurerm'
    command: 'plan'
    workingDirectory: '$(terraformWorkingDirectory)'
    environmentServiceNameAzureRM: 'bc-admin-center-workload-identity'
    commandOptions: '-out=tfplan'
```

## Troubleshooting

### "Failed to get federated token"

Check that:
- The service connection is configured correctly with workload identity
- The federated credential subject matches the service connection name exactly
- `addSpnToEnvironment: true` is set in the AzureCLI task

### "Service connection not found"

Verify:
- The service connection name in the pipeline matches the one created
- The service connection is shared with the pipeline
- You have permissions to use the service connection

### Federated Credential Subject Mismatch

The subject must exactly match this pattern:
```
sc://ORGANIZATION_NAME/PROJECT_NAME/SERVICE_CONNECTION_NAME
```

Verify each component:
```bash
# Check your organization name (as it appears in the URL)
# https://dev.azure.com/{ORGANIZATION_NAME}

# Check project name (case-sensitive)
# https://dev.azure.com/{ORG}/{PROJECT_NAME}

# Check service connection name (case-sensitive)
# Project Settings -> Service connections -> Connection name
```

### Debugging Federated Token

Add a debug step to check the token:

```yaml
- task: AzureCLI@2
  displayName: 'Debug Federated Token'
  inputs:
    azureSubscription: 'bc-admin-center-workload-identity'
    scriptType: 'bash'
    scriptLocation: 'inlineScript'
    inlineScript: |
      echo "Client ID: $servicePrincipalId"
      echo "Tenant ID: $tenantId"
      echo "Token file: $AZURE_FEDERATED_TOKEN_FILE"
      
      if [ -f "$AZURE_FEDERATED_TOKEN_FILE" ]; then
        echo "Token file exists"
        # Decode the JWT token (don't do this in production!)
        cat $AZURE_FEDERATED_TOKEN_FILE | cut -d. -f2 | base64 -d 2>/dev/null | jq .
      else
        echo "Token file not found"
      fi
    addSpnToEnvironment: true
```

## Security Best Practices

1. **Use Azure DevOps Environments**: Configure approval gates for production deployments
2. **Limit service connection access**: Only grant access to specific pipelines
3. **Enable audit logging**: Monitor all pipeline runs and changes
4. **Use variable groups**: Store non-sensitive configuration in variable groups
5. **Implement branch policies**: Require pull requests and reviews
6. **Restrict federated credentials**: Use specific service connection names, not wildcards

## Subject Claim Patterns for Azure DevOps

| Scenario | Subject Pattern |
|----------|----------------|
| Specific service connection | `sc://ORG/PROJECT/CONNECTION_NAME` |
| Any connection in project | `sc://ORG/PROJECT/*` |

> **Note**: Unlike GitHub Actions, Azure DevOps does not support branch-specific or environment-specific federated credentials in the subject claim. Use Azure DevOps environments and service connection permissions for access control.

## Comparison: Workload Identity vs Client Secret

| Feature | Workload Identity | Client Secret |
|---------|------------------|---------------|
| Security | High (no stored secrets) | Medium (secrets in variable groups) |
| Rotation | Automatic | Manual |
| Setup Complexity | Medium | Low |
| Azure DevOps Support | Native (service connections) | Requires variable groups |
| Token Lifetime | Short (minutes) | Long (up to 2 years) |

## Next Steps

- Learn about [workload identity for GitHub Actions](./workload-identity-github.md)
- Explore [service principal authentication](./service-principal-authentication.md) with client secrets
- Set up [managed identity authentication](./managed-identity-authentication.md) for self-hosted agents
- Configure [Azure CLI authentication](./azure-cli-authentication.md) for local testing

## Additional Resources

- [Azure DevOps Workload Identity Federation](https://learn.microsoft.com/en-us/azure/devops/pipelines/release/configure-workload-identity)
- [Azure Resource Manager service connections](https://learn.microsoft.com/en-us/azure/devops/pipelines/library/connect-to-azure)
- [Azure Workload Identity Federation](https://learn.microsoft.com/en-us/azure/active-directory/develop/workload-identity-federation)
- [Terraform in Azure Pipelines](https://learn.microsoft.com/en-us/azure/developer/terraform/get-started-azure-devops)

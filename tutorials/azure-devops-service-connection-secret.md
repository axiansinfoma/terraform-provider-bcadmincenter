# Authenticating with an Azure DevOps Service Connection (Client Secret)

This guide shows how to run Terraform in Azure DevOps using an Azure Resource Manager service connection that authenticates with a service principal client secret.

## When to use this approach

Use this approach when workload identity federation is not yet available in your Azure DevOps environment.

## Prerequisites

- Azure DevOps project with permission to create service connections.
- Azure AD application (service principal) with `AdminCenter.ReadWrite.All` permission for Business Central Admin Center API.
- Service principal added to **Authorized Microsoft Entra Apps** in Business Central Admin Center (**Settings** > **Authorized Microsoft Entra Apps**). This is required before the provider can make any API calls.
- Optionally, service principal added to the **AdminAgents** group in Business Central Admin Center for delegated admin access.

## Step 1: Create a service principal secret

```bash
az ad app credential reset \
  --id <APP_REGISTRATION_CLIENT_ID> \
  --append \
  --display-name "azure-devops-bcadmincenter"
```

Save the generated secret value securely.

## Step 2: Create an Azure DevOps service connection (secret-based)

In Azure DevOps:

1. Go to **Project settings** → **Service connections**.
2. Select **New service connection**.
3. Choose **Azure Resource Manager**.
4. Select **Service principal (manual)**.
5. Enter tenant ID, client ID, and client secret.
6. Name the connection (example: `bc-admin-sp-secret`).

## Step 3: Use the service connection in your pipeline

```yaml
trigger:
  - main

pool:
  vmImage: ubuntu-latest

steps:
  - task: TerraformInstaller@1
    inputs:
      terraformVersion: '1.9.0'

  - task: AzureCLI@2
    inputs:
      azureSubscription: 'bc-admin-sp-secret'
      scriptType: 'bash'
      scriptLocation: 'inlineScript'
      addSpnToEnvironment: true
      inlineScript: |
        export AZURE_CLIENT_ID="$servicePrincipalId"
        export AZURE_CLIENT_SECRET="$servicePrincipalKey"
        export AZURE_TENANT_ID="$tenantId"

        terraform init
        terraform plan
```

## Step 4: Configure the provider

```hcl
provider "bcadmincenter" {
  # Uses AZURE_CLIENT_ID, AZURE_CLIENT_SECRET, AZURE_TENANT_ID
}
```

## Security recommendations

- Store client secrets only in the service connection, not in source control.
- Rotate service principal secrets regularly.
- Prefer workload identity service connections when possible.

## Related guides

- [Azure DevOps service connection with workload identity](./workload-identity-azure-devops.md)
- [Service principal authentication](./service-principal-authentication.md)

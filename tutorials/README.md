# Business Central Admin Center Provider - Tutorials

This directory contains comprehensive tutorials for setting up and using the Business Central Admin Center Terraform Provider.

## Available Tutorials

### End-to-End Use-Case Tutorials

These tutorials guide you through common real-world scenarios from start to finish:

1. **[Provisioning a Business Central Environment](./full-environment-tutorial.md)**
   - Best for: First-time users, teams setting up a new tenant
   - Difficulty: Easy
   - Topics: Data source discovery, environment creation, settings, support contact, import

2. **[Multi-Tenant Management](./multi-tenant-management.md)**
   - Best for: Partners or ISVs managing Business Central across multiple customer tenants
   - Difficulty: Medium
   - Topics: `manageable_tenants` data source, `for_each` iteration, provider aliases, import workflow

### Authentication Tutorials

These tutorials guide you through setting up authentication for different scenarios:

3. **[Service Principal with Client Secret](./service-principal-authentication.md)**
   - Best for: Automated scenarios, CI/CD pipelines (traditional approach)
   - Difficulty: Easy
   - Setup time: ~15 minutes
   - Security level: Medium (requires secret management)

4. **[Azure CLI Authentication](./azure-cli-authentication.md)**
   - Best for: Local development, interactive scenarios
   - Difficulty: Easy
   - Setup time: ~5 minutes
   - Security level: High (no stored credentials)

5. **[Managed Identity](./managed-identity-authentication.md)**
   - Best for: Azure-hosted Terraform runners (VMs, Container Instances, App Service)
   - Difficulty: Medium
   - Setup time: ~20 minutes
   - Security level: High (no credential management)

6. **[Workload Identity for GitHub Actions](./workload-identity-github.md)**
   - Best for: GitHub Actions workflows
   - Difficulty: Medium
   - Setup time: ~25 minutes
   - Security level: Very High (OIDC-based, no secrets)

7. **[Service Connection with Workload Identity (Azure DevOps)](./workload-identity-azure-devops.md)**
    - Best for: Azure DevOps pipelines
    - Difficulty: Medium
    - Setup time: ~25 minutes
    - Security level: Very High (federated credentials, no secrets)

8. **[Service Connection with Client Secret (Azure DevOps)](./azure-devops-service-connection-secret.md)**
   - Best for: Azure DevOps pipelines that require secret-based authentication
   - Difficulty: Easy
   - Setup time: ~15 minutes
   - Security level: Medium (requires secret management)

## Choosing the Right Authentication Method

### Decision Tree

```
Are you developing locally?
├─ Yes → Use Azure CLI Authentication
└─ No → Are you running in CI/CD?
    ├─ Yes → Which CI/CD platform?
    │   ├─ GitHub Actions → Use Workload Identity for GitHub Actions
    │   ├─ Azure DevOps → Use Service Connection with Workload Identity (preferred) or Client Secret
    │   └─ Other → Use Service Principal with Client Secret
    └─ No → Are you running on Azure compute?
        ├─ Yes → Use Managed Identity
        └─ No → Use Service Principal with Client Secret
```

### Comparison Table

| Method | Use Case | Pros | Cons | Setup Complexity |
|--------|----------|------|------|-----------------|
| **Azure CLI** | Local development | Quick setup, no secrets | Not for automation | Low |
| **Service Principal** | Generic automation | Works anywhere, simple | Requires secret management | Low |
| **Managed Identity** | Azure VMs/Containers | No secrets, automatic | Azure-only | Medium |
| **Workload Identity (GitHub)** | GitHub Actions | No secrets, secure | GitHub-only | Medium |
| **Workload Identity (Azure DevOps)** | Azure Pipelines | No secrets, secure | Azure DevOps-only | Medium |
| **Service Connection (Azure DevOps + Secret)** | Azure Pipelines | Simple setup | Requires secret rotation | Low |

## Common Prerequisites

All authentication methods require:

1. **Azure AD tenant access**
   - Permissions to create applications (or an existing application)
   - Permissions to grant admin consent for API permissions

2. **Business Central Admin Center access**
   - Admin role in Business Central Admin Center
   - Ability to add members to the AdminAgents group

3. **Required API Permissions**
   - `AdminCenter.ReadWrite.All` on the Business Central Admin Center API
   - Application ID: `996def3d-b36c-4153-8607-a6fd3c01b89f`

## Quick Start

### For Local Development

```bash
# 1. Install Azure CLI
curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash

# 2. Login
az login

# 3. Configure Terraform
cat > main.tf <<EOF
terraform {
  required_providers {
    bcadmincenter = {
      source  = "axiansinfoma/bcadmincenter"
      version = "~> 1.0"
    }
  }
}

provider "bcadmincenter" {
  use_cli = true
}
EOF

# 4. Test
terraform init
terraform plan
```

See the [Azure CLI Authentication tutorial](./azure-cli-authentication.md) for complete instructions.

### For GitHub Actions

```yaml
# .github/workflows/terraform.yml
name: Terraform
on: [push]

permissions:
  id-token: write
  contents: read

jobs:
  terraform:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - uses: azure/login@v2
        with:
          client-id: ${{ secrets.AZURE_CLIENT_ID }}
          tenant-id: ${{ secrets.AZURE_TENANT_ID }}
      
      - uses: hashicorp/setup-terraform@v3
      
      - run: terraform init
      - run: terraform plan
```

See the [Workload Identity for GitHub Actions tutorial](./workload-identity-github.md) for complete setup.

### For Azure DevOps

```yaml
# azure-pipelines.yml
trigger:
  - main

pool:
  vmImage: 'ubuntu-latest'

steps:
  - task: TerraformInstaller@1
    inputs:
      terraformVersion: '1.9.0'

  - task: AzureCLI@2
    inputs:
      azureSubscription: 'bc-admin-workload-identity'
      scriptType: 'bash'
      scriptLocation: 'inlineScript'
      inlineScript: |
        export AZURE_CLIENT_ID=$servicePrincipalId
        export AZURE_TENANT_ID=$tenantId
        terraform init
        terraform plan
      addSpnToEnvironment: true
```

See either:
- [Service Connection with Workload Identity (Azure DevOps)](./workload-identity-azure-devops.md)
- [Service Connection with Client Secret (Azure DevOps)](./azure-devops-service-connection-secret.md)

## Security Best Practices

Regardless of the authentication method you choose:

1. **Use least privilege**: Only grant the minimum required permissions
2. **Rotate credentials**: Regularly rotate client secrets (if using them)
3. **Enable audit logging**: Monitor access to Business Central environments
4. **Use remote state**: Store Terraform state securely in Azure Storage
5. **Protect state files**: State files may contain sensitive information
6. **Review permissions regularly**: Audit who has access to your service principals

## Migrating Between Authentication Methods

You can easily switch authentication methods by changing your provider configuration:

```terraform
# Development (Azure CLI)
provider "bcadmincenter" {
  use_cli = true
}

# Staging (Service Principal)
# provider "bcadmincenter" {
#   client_id     = var.client_id
#   client_secret = var.client_secret
#   tenant_id     = var.tenant_id
# }

# Production (Workload Identity - auto-detected in CI/CD)
# provider "bcadmincenter" {
#   # Credentials from environment variables
# }
```

## Troubleshooting

### Common Issues

1. **"Permission denied" errors**
   - Verify API permissions are granted and admin consent is given
   - Check that the service principal is in the AdminAgents group
   - Ensure you're using the correct tenant ID

2. **"Application not found"**
   - Verify the client ID is correct
   - Ensure the application exists in the correct Azure AD tenant
   - Check that the service principal was created

3. **"Invalid client secret"**
   - Check that the client secret hasn't expired
   - Verify you're using the correct secret value
   - Ensure there are no extra spaces or characters

### Getting Help

- Check the specific tutorial for troubleshooting steps
- Review the [provider documentation](../docs/index.md)
- Open an issue on [GitHub](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/issues)

## Additional Resources

### Microsoft Documentation

- [Business Central Admin Center API](https://learn.microsoft.com/en-us/dynamics365/business-central/dev-itpro/administration/administration-center-api)
- [Azure AD App Registration](https://learn.microsoft.com/en-us/azure/active-directory/develop/quickstart-register-app)
- [Workload Identity Federation](https://learn.microsoft.com/en-us/azure/active-directory/develop/workload-identity-federation)
- [Managed Identities](https://learn.microsoft.com/en-us/azure/active-directory/managed-identities-azure-resources/overview)

### Terraform Documentation

- [Terraform Provider Development](https://developer.hashicorp.com/terraform/plugin)
- [Terraform Backend Configuration](https://developer.hashicorp.com/terraform/language/settings/backends/configuration)
- [Terraform State](https://developer.hashicorp.com/terraform/language/state)

### CI/CD Platform Documentation

- [GitHub Actions with OIDC](https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/about-security-hardening-with-openid-connect)
- [Azure DevOps Service Connections](https://learn.microsoft.com/en-us/azure/devops/pipelines/library/service-endpoints)
- [Azure DevOps Workload Identity](https://learn.microsoft.com/en-us/azure/devops/pipelines/release/configure-workload-identity)

## Contributing

If you find issues with these tutorials or have suggestions for improvements, please open an issue or pull request on our [GitHub repository](https://github.com/axiansinfoma/terraform-provider-bcadmincenter).

## License

These tutorials are part of the Business Central Admin Center Terraform Provider and are licensed under the [Mozilla Public License 2.0](../LICENSE).

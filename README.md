# Terraform Provider for Business Central Admin Center

This Terraform provider enables Infrastructure as Code (IaC) management of Microsoft Dynamics 365 Business Central environments through the [Business Central Admin Center API](https://learn.microsoft.com/en-us/dynamics365/business-central/dev-itpro/administration/administration-center-api).

## ⚠️ Important Warnings

**This provider manages critical production infrastructure and requires administrator privileges.**

- **Destructive Operations**: This provider will permanently delete environments when Terraform determines it's necessary (e.g., when changing immutable attributes). Always carefully review `terraform plan` output before applying changes.
- **No Undo**: Environment deletions are permanent and cannot be reversed. Ensure you have proper backups before making changes.
- **Development Status**: This provider is in active development and has not been extensively tested in production environments. Use at your own risk.
- **No Warranty**: The authors and contributors are not responsible for any data loss, service interruption, or other issues that may occur from using this provider.
- **Version Updates Not Supported**: This provider **cannot schedule or apply version updates** to environments or apps. Environment version updates (`application_version`) and app updates must be managed through the [Business Central Admin Center portal](https://businesscentral.dynamics.com/?page=1801) or other automation tools.

**Best Practices**:
- Always run `terraform plan` and carefully review changes before `terraform apply`
- Test in non-production environments first
- Use version control for your Terraform configurations
- Implement proper backup strategies for critical environments
- Consider using `-target` flag to limit changes to specific resources when needed

## Features

- Manage Business Central production and sandbox environments
- Configure environment settings and access controls
- Configure administrative notifications
- Monitor environment quotas

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.13
- [Go](https://golang.org/doc/install) >= 1.24 (for development)
- Azure AD application with **AdminCenter.ReadWrite.All** permissions
- Membership in the **AdminAgents** group for delegated admin access

## Using the Provider

### Authentication

The provider supports multiple authentication methods via the Azure SDK:

1. **Service Principal with Client Secret**
2. **Service Principal with Workload Identity Federation** (recommended for CI/CD - authenticates via OIDC against an Azure AD app registration with federated credentials)
3. **Service Principal with Certificate**
4. **Managed Identity** (for Azure-hosted environments)
5. **Azure CLI Authentication** (for local development)
6. **Device Code Flow** (for interactive scenarios)

### Configuration

#### Using Service Principal with Client Secret

```hcl
terraform {
  required_providers {
    bcadmincenter = {
      source  = "vllni/bcadmincenter"
      version = "~> 1.0"
    }
  }
}

provider "bcadmincenter" {
  client_id     = "00000000-0000-0000-0000-000000000000"
  client_secret = "your-client-secret"
  tenant_id     = "00000000-0000-0000-0000-000000000000"
  environment   = "public" # optional: public, usgovernment, china
}
```

#### Using Environment Variables

The provider follows Azure SDK conventions and supports these environment variables:

- `AZURE_CLIENT_ID` - The Client ID (Application ID)
- `AZURE_CLIENT_SECRET` - The Client Secret
- `AZURE_TENANT_ID` - The Tenant ID
- `AZURE_ENVIRONMENT` - The Azure environment (public, usgovernment, china)

#### Using Azure Workload Identity (Recommended for CI/CD)

Azure Workload Identity allows authentication via OIDC against an Azure AD app registration that has a federated credential. This is the recommended approach for CI/CD pipelines (GitHub Actions, Azure DevOps) as it eliminates the need to manage long-lived secrets.

For Azure Workload Identity in Kubernetes environments or CI/CD:

```hcl
provider "bcadmincenter" {
  # Azure Workload Identity uses these environment variables:
  # AZURE_CLIENT_ID
  # AZURE_TENANT_ID
  # AZURE_FEDERATED_TOKEN_FILE - Path to the federated token file
  # AZURE_AUTHORITY_HOST - Azure Active Directory authority host
}
```

The provider will automatically detect and use workload identity credentials when available.

### Example Usage

```hcl
# Create a sandbox environment
resource "bc_environment" "sandbox" {
  name               = "my-sandbox"
  application_family = "BusinessCentral"
  type               = "Sandbox"
  country_code       = "US"
  ring_name          = "Production"
  azure_region       = "westus2"
}

# The application_version is read-only and assigned by the API
output "sandbox_version" {
  value = bc_environment.sandbox.application_version
}

# Configure environment settings
resource "bc_environment_settings" "sandbox_settings" {
  environment_name = bc_environment.sandbox.name
  
  # Add settings configuration here
}
```

## Building The Provider

1. Clone the repository:
```shell
git clone https://github.com/vllni/terraform-provider-bcadmincenter
cd terraform-provider-bcadmincenter
```

2. Build the provider:
```shell
go install
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules):

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Developing the Provider

### Prerequisites

- Go 1.24 or later
- Terraform 1.0 or later
- Access to Business Central Admin Center API
- Azure AD application registration with appropriate permissions

### Local Development

To compile the provider locally:

```shell
go build -o terraform-provider-bcadmincenter
```

To install locally for testing:

```shell
mkdir -p ~/.terraform.d/plugins/local/vllni/bcadmincenter/1.0.0/linux_amd64
cp terraform-provider-bcadmincenter ~/.terraform.d/plugins/local/vllni/bcadmincenter/1.0.0/linux_amd64/
```

Then use it in your Terraform configuration:

```hcl
terraform {
  required_providers {
    bcadmincenter = {
      source  = "local/vllni/bcadmincenter"
      version = "1.0.0"
    }
  }
}
```

### Testing

Generate or update documentation:
```shell
make generate
```

Run acceptance tests (note: creates real resources):
```shell
make testacc
```

### Documentation Development

This provider uses [terraform-plugin-docs](https://github.com/hashicorp/terraform-plugin-docs) for documentation generation.

```shell
# Generate documentation from templates
make docs

# Validate documentation compliance
make validate-docs

# Check if docs are up-to-date
make docs-check

# Format example files
terraform fmt -recursive examples/
```

**Important**: Documentation in the `docs/` directory is automatically generated during the release pipeline. For documentation changes, edit templates in `templates/` and example files in `examples/`. The documentation is generated from these sources during release.

See the [Documentation Quick Reference](docs/QUICK-REFERENCE.md) for more details.

## Documentation

Published documentation for the latest release is available on the [Terraform Registry](https://registry.terraform.io/providers/vllni/bcadmincenter/latest/docs).

**Documentation Development**:
- Edit templates in `templates/` directory
- Update example files in `examples/` directory
- Documentation is generated during release pipeline
- [Template Guide](templates/README.md) - How to write documentation templates

## Contributing

Contributions are welcome! Please see our contributing guidelines.

## License

Mozilla Public License 2.0 - see [LICENSE](LICENSE) for details.

## Support

For issues and questions:
- [GitHub Issues](https://github.com/vllni/terraform-provider-bcadmincenter/issues)
- [Business Central Admin Center API Documentation](https://learn.microsoft.com/en-us/dynamics365/business-central/dev-itpro/administration/administration-center-api)

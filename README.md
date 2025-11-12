# Terraform Provider for Business Central Admin Center

This Terraform provider enables Infrastructure as Code (IaC) management of Microsoft Dynamics 365 Business Central environments through the [Business Central Admin Center API](https://learn.microsoft.com/en-us/dynamics365/business-central/dev-itpro/administration/administration-center-api).

## Features

- Manage Business Central production and sandbox environments
- Configure environment settings and access controls
- Manage application installations and updates
- Configure administrative notifications
- Query environment operations and quotas

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.24 (for development)
- Azure AD application with **AdminCenter.ReadWrite.All** permissions
- Membership in the **AdminAgents** group for delegated admin access

## Using the Provider

### Authentication

The provider supports multiple authentication methods via the Azure SDK:

1. **Service Principal with Client Secret**
2. **Service Principal with Workload Identity Credential** (recommended for CI/CD)
3. **Service Principal with Certificate**
4. **Managed Identity** (for Azure-hosted environments)
5. **Azure CLI Authentication** (for local development)
6. **Device Code Flow** (for interactive scenarios)

### Configuration

#### Using Service Principal with Client Secret

```hcl
terraform {
  required_providers {
    bc_admin_center = {
      source  = "vllni/bc-admin-center"
      version = "~> 1.0"
    }
  }
}

provider "bc_admin_center" {
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

For Azure Workload Identity in Kubernetes environments:

```hcl
provider "bc_admin_center" {
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
  name                = "my-sandbox"
  application_family  = "BusinessCentral"
  type               = "Sandbox"
  country_code       = "US"
  ring_name          = "Production"
  application_version = "24.0"
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
git clone https://github.com/vllni/terraform-provider-bc-admin-center
cd terraform-provider-bc-admin-center
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
go build -o terraform-provider-bc-admin-center
```

To install locally for testing:

```shell
mkdir -p ~/.terraform.d/plugins/local/vllni/bc-admin-center/1.0.0/linux_amd64
cp terraform-provider-bc-admin-center ~/.terraform.d/plugins/local/vllni/bc-admin-center/1.0.0/linux_amd64/
```

Then use it in your Terraform configuration:

```hcl
terraform {
  required_providers {
    bc_admin_center = {
      source  = "local/vllni/bc-admin-center"
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

## Documentation

See the [docs](./docs) directory for detailed documentation on:
- Resources
- Data Sources
- Configuration options

## Contributing

Contributions are welcome! Please see our contributing guidelines.

## License

Mozilla Public License 2.0 - see [LICENSE](LICENSE) for details.

## Support

For issues and questions:
- [GitHub Issues](https://github.com/vllni/terraform-provider-bc-admin-center/issues)
- [Business Central Admin Center API Documentation](https://learn.microsoft.com/en-us/dynamics365/business-central/dev-itpro/administration/administration-center-api)

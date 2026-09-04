# Terraform Provider for Business Central Admin Center

[![Tests](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/actions/workflows/test.yml/badge.svg)](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/actions/workflows/test.yml)
[![Terraform Registry](https://img.shields.io/badge/registry-axiansinfoma%2Fbcadmincenter-844FBA?logo=terraform)](https://registry.terraform.io/providers/axiansinfoma/bcadmincenter/latest)
[![License: MPL 2.0](https://img.shields.io/badge/license-MPL--2.0-blue)](LICENSE)

Manage Microsoft Dynamics 365 Business Central environments as Infrastructure as
Code, through the [Business Central Admin Center
API](https://learn.microsoft.com/en-us/dynamics365/business-central/dev-itpro/administration/administration-center-api).

- 📖 **Documentation**: [Terraform Registry](https://registry.terraform.io/providers/axiansinfoma/bcadmincenter/latest/docs)
- 🎓 **Guides**: [Authentication and workflow tutorials](tutorials/README.md)
- 🐛 **Issues**: [GitHub Issues](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/issues)
- 💬 **Questions**: [GitHub Discussions](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/discussions)
- 🔒 **Security**: [Security Policy](.github/SECURITY.md)

## ⚠️ Important Warnings

**This provider manages critical production infrastructure and requires
administrator privileges.**

- **Destructive Operations**: This provider will permanently delete environments
  when Terraform determines it's necessary (e.g., when changing immutable
  attributes). Always carefully review `terraform plan` output before applying
  changes.
- **No Undo**: Environment deletions are permanent and cannot be reversed.
  Ensure you have proper backups before making changes.
- **Development Status**: This provider is pre-1.0 and in active development. It
  has not been extensively tested in production environments, and minor releases
  may contain breaking changes — see the [changelog](CHANGELOG.md). Use at your
  own risk.
- **No Warranty**: The authors and contributors are not responsible for any data
  loss, service interruption, or other issues that may occur from using this
  provider.

**Best Practices**:

- Always run `terraform plan` and carefully review changes before
  `terraform apply`
- Test in non-production environments first
- Pin the provider version with `~>` and read the changelog before upgrading
- Use version control for your Terraform configurations
- Implement proper backup strategies for critical environments
- Consider using `-target` to limit changes to specific resources when needed
- Treat Terraform state and plan files as secret material and keep them in an
  encrypted, access-controlled backend

## What It Manages

**Resources**

| Resource | Purpose |
| --- | --- |
| `bcadmincenter_environment` | Production and sandbox environments, including their settings and update window |
| `bcadmincenter_environment_update_schedule` | Scheduled version upgrades for an environment |
| `bcadmincenter_environment_app` | Install, update, and uninstall lifecycle of an app in an environment |
| `bcadmincenter_environment_support_contact` | The support contact shown to environment users |
| `bcadmincenter_per_tenant_extension` | Per-tenant extension (PTE) upload and lifecycle |
| `bcadmincenter_authorized_entra_app` | Microsoft Entra apps authorized to call the Admin Center API for a tenant |
| `bcadmincenter_notification_recipient` | Recipients of administrative notifications |

**Data sources**

| Data source | Purpose |
| --- | --- |
| `bcadmincenter_environment` / `bcadmincenter_environments` | Look up one environment, or all environments in a tenant |
| `bcadmincenter_environment_updates` | Available and scheduled updates for an environment |
| `bcadmincenter_available_applications` / `bcadmincenter_application_family` | Application versions available for new environments |
| `bcadmincenter_authorized_entra_apps` / `bcadmincenter_manageable_tenants` | Authorized Entra apps, and the tenants the caller can manage |
| `bcadmincenter_scheduled_pte_operations` | Pending per-tenant extension operations |
| `bcadmincenter_notification_settings` | Configured notification recipients and settings |
| `bcadmincenter_quotas` | Environment quota usage for the tenant |
| `bcadmincenter_timezones` | Time zone identifiers accepted by the API |

See the [registry
documentation](https://registry.terraform.io/providers/axiansinfoma/bcadmincenter/latest/docs)
for the full schema of each.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.13
  (tested against 1.13.x and 1.14.x)
- An Azure AD application with **AdminCenter.ReadWrite.All** on the *Dynamics
  365 Business Central administration center* API
  (`996def3d-b36c-4153-8607-a6fd3c01b89f`)
- The service principal added to the **Authorized Microsoft Entra Apps** list in
  the Business Central Admin Center
- Membership in the **AdminAgents** group for delegated admin access
- [Go](https://golang.org/doc/install) — only for development; use the version
  pinned in [`go.mod`](go.mod)

## Using the Provider

### Authentication

The provider authenticates through the Azure SDK for Go and follows the same
conventions as the AzureRM provider. Supported methods:

1. **Service Principal with Client Secret**
2. **Workload Identity Federation / OIDC** — recommended for CI/CD; authenticates
   against an Azure AD app registration with a federated credential, so there is
   no long-lived secret to manage
3. **Service Principal with Certificate**
4. **Managed Identity** — for Azure-hosted runners
5. **Azure CLI** — for local development
6. **Device Code Flow** — for interactive scenarios

Step-by-step guides for each are in [`tutorials/`](tutorials/README.md):
[service principal](tutorials/service-principal-authentication.md) ·
[Azure CLI](tutorials/azure-cli-authentication.md) ·
[managed identity](tutorials/managed-identity-authentication.md) ·
[workload identity on GitHub Actions](tutorials/workload-identity-github.md) ·
[workload identity on Azure DevOps](tutorials/workload-identity-azure-devops.md) ·
[Azure DevOps service connection with a secret](tutorials/azure-devops-service-connection-secret.md) ·
[multi-tenant management](tutorials/multi-tenant-management.md)

### Configuration

```hcl
terraform {
  required_providers {
    bcadmincenter = {
      source  = "axiansinfoma/bcadmincenter"
      version = "~> 0.2"
    }
  }
}

provider "bcadmincenter" {
  client_id     = "00000000-0000-0000-0000-000000000000"
  client_secret = "your-client-secret" # prefer a variable or the environment
  tenant_id     = "00000000-0000-0000-0000-000000000000"
}
```

#### Environment variables

Leave the provider block empty and let it pick credentials up from the
environment. The provider uses the `ARM_*` names, and accepts the `AZURE_*`
equivalents for backward compatibility:

```bash
export ARM_CLIENT_ID="00000000-0000-0000-0000-000000000000"
export ARM_CLIENT_SECRET="your-client-secret"
export ARM_TENANT_ID="00000000-0000-0000-0000-000000000000"
```

#### Workload identity (recommended for CI/CD)

```hcl
provider "bcadmincenter" {
  use_oidc = true

  # Everything else is discovered from the platform's environment:
  #   ARM_CLIENT_ID / ARM_TENANT_ID
  #   AZURE_FEDERATED_TOKEN_FILE          (Kubernetes projected volume)
  #   ACTIONS_ID_TOKEN_REQUEST_URL/TOKEN  (GitHub Actions)
  #   SYSTEM_OIDCREQUESTURI               (Azure DevOps)
}
```

The provider detects workload identity credentials automatically and re-reads
the federated token on every Azure AD token refresh, so platform-rotated and
short-lived tokens keep working through long Terraform runs.

### Example Usage

```hcl
# Create a sandbox environment with inline settings
resource "bcadmincenter_environment" "sandbox" {
  name               = "my-sandbox"
  application_family = "BusinessCentral"
  type               = "Sandbox"
  country_code       = "US"
  ring_name          = "PROD"
  azure_region       = "westus2"

  # Optional: configure environment settings inline. The three update-window
  # attributes must be set together.
  settings {
    update_window_start_time = "22:00"
    update_window_end_time   = "06:00"
    update_window_timezone   = "Pacific Standard Time"
  }

  timeouts {
    create = "90m"
    delete = "60m"
  }
}

# The application_version is read-only unless you set it to request an upgrade
output "sandbox_version" {
  value = bcadmincenter_environment.sandbox.application_version
}
```

More runnable examples live in [`examples/`](examples/README.md), and
[`dev/`](dev/README.md) has configurations covering every resource and data
source.

## Development

Contributions are welcome. **[CONTRIBUTING.md](.github/CONTRIBUTING.md) is the
full guide** — setup, testing layers, how to add a resource, documentation
generation, and the pull request process. What follows is the short version.

### Building

```shell
git clone https://github.com/axiansinfoma/terraform-provider-bcadmincenter
cd terraform-provider-bcadmincenter
go mod download
make build
```

A [dev container](.devcontainer/devcontainer.json) is provided with Go,
Terraform, the Azure CLI, and the pre-commit hooks preinstalled.

### Running your build

[`scripts/setup-local-testing.sh`](scripts/setup-local-testing.sh) builds the
provider and sets up a `dev_overrides` block in `~/.terraformrc` pointing at
your working tree, so you can `terraform plan` against your local binary
without `terraform init`.

### Testing

```shell
make test        # unit tests
make testmock    # acceptance tests against the local mock HTTP server (offline)
make lint        # golangci-lint
make testacc     # ⚠️ acceptance tests against a REAL tenant — creates and destroys
                 #    real environments, takes up to 2h, may incur cost
```

Configuration-level tests using `mock_provider` live in [`tests/`](tests/README.md):

```shell
cd tests && terraform init && terraform test
```

### Documentation

`docs/` is **generated** — don't edit it by hand. Edit `templates/` (see the
[template guide](templates/README.md)), the examples in `examples/`, and the
schema descriptions in `internal/`, then:

```shell
make docs           # regenerate docs/
make validate-docs  # verify docs/ is in sync and examples are formatted
```

Commit the regenerated `docs/` with your change: CI regenerates the docs and
fails if the committed output differs, and the pre-commit hook checks the same
thing locally.

### Adding dependencies

```shell
go get github.com/author/dependency
go mod tidy
```

Commit the resulting `go.mod` and `go.sum`.

## Contributing

- [Contribution guidelines](.github/CONTRIBUTING.md) — how to set up, test, and
  submit a change
- [Code of Conduct](.github/CODE_OF_CONDUCT.md)
- [Security Policy](.github/SECURITY.md) — please report vulnerabilities
  privately, not as an issue

Bug reports and feature requests belong in
[Issues](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/issues);
configuration and usage questions belong in
[Discussions](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/discussions).

## License

Mozilla Public License 2.0 — see [LICENSE](LICENSE) for details.

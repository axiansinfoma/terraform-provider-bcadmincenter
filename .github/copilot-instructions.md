# Terraform Provider for Business Central Admin Center

## Project Overview

This repository contains a Terraform provider for managing Microsoft Dynamics 365 Business Central environments through the [Business Central Admin Center API](https://learn.microsoft.com/en-us/dynamics365/business-central/dev-itpro/administration/administration-center-api). The provider enables Infrastructure as Code (IaC) for Business Central tenant administration tasks.

## Business Central Admin Center API

The provider is built on the Business Central Administration Center API (`https://api.businesscentral.dynamics.com`) which enables administrators to programmatically:

- Manage production and sandbox environments
- Configure administrative notifications
- Monitor telemetry and environment operations
- Manage applications and extensions
- Configure environment settings and access controls
- Handle environment lifecycle operations (create, copy, delete, restore)

### Key API Resources Covered

Based on the [Microsoft Learn documentation](https://learn.microsoft.com/en-us/dynamics365/business-central/dev-itpro/administration/administration-center-api), the provider should implement resources for:

#### Core Resources

**Environments (`/admin/v2.24/applications/{applicationFamily}/environments`)**
- `resource "bc_environment"` - Create, update, and delete Business Central environments
- `data "bc_environment"` - Get information about a specific environment
- `data "bc_environments"` - List all environments for a tenant

**Applications (`/admin/v2.24/applications`)**
- `resource "bc_application"` - Manage application installations and updates
- `data "bc_application"` - Get application details
- `data "bc_applications"` - List available applications and versions

**Environment Settings (`/admin/v2.24/applications/{applicationFamily}/environments/{environmentName}/settings`)**
- `resource "bc_environment_settings"` - Configure environment-specific settings
- `resource "bc_environment_access"` - Manage environment access controls
- `resource "bc_environment_telemetry"` - Configure Application Insights telemetry

**Notifications (`/admin/v2.24/settings/notification`)**
- `resource "bc_notification_settings"` - Configure tenant notification recipients
- `data "bc_notification_settings"` - Get current notification configuration

#### Operations Resources

**Environment Operations (`/admin/v2.24/environments/{environmentName}/operations`)**
- `data "bc_environment_operations"` - Query environment operation history
- `resource "bc_environment_operation"` - Trigger specific environment operations

**App Management (`/admin/v2.24/applications/{applicationFamily}/environments/{environmentName}/apps`)**
- `resource "bc_environment_app"` - Install, update, and uninstall apps
- `data "bc_environment_app"` - Get app installation details
- `data "bc_environment_apps"` - List installed apps

#### Administrative Resources

**Available Applications (`/admin/v2.24/applications/{applicationFamily}/Countries/{countryCode}/Rings/{ringName}`)**
- `data "bc_available_applications"` - List available application families, countries, and rings
- `data "bc_application_versions"` - Get available versions for a specific ring

**Quotas (`/admin/v2.24/environments/quotas`)**
- `data "bc_environment_quotas"` - Get environment quotas and limits

## Authentication Strategy

The provider will use the same authentication approach as the AzureRM and AzureAD providers, leveraging the Azure SDK for Go's authentication libraries:

### Dependencies
```go
github.com/Azure/azure-sdk-for-go/sdk/azidentity
github.com/Azure/azure-sdk-for-go/sdk/azcore
```

### Supported Authentication Methods
1. **Service Principal with Client Secret**
2. **Service Principal with Workload Identity Credential** (recommended for CI/CD)
3. **Service Principal with Certificate**
4. **Managed Identity** (for Azure-hosted environments)
5. **Azure CLI Authentication** (for local development)
6. **Device Code Flow** (for interactive scenarios)

### Required Permissions
- **AdminCenter.ReadWrite.All** on the "Dynamics 365 Business Central administration center" API
- The application must be added to the **AdminAgents** group for delegated admin access
- Additional environment-level permissions may be required for app management operations

### Authentication Configuration
The provider will support the same authentication patterns as azurerm:

```hcl
provider "bc_admin_center" {
  # Authentication via Service Principal
  client_id       = "00000000-0000-0000-0000-000000000000"
  client_secret   = "client-secret"
  tenant_id       = "00000000-0000-0000-0000-000000000000"
  
  # Optional: Override default endpoints
  environment = "public" # public, usgovernment, china
}
```

Or using environment variables (following Azure conventions):
- `AZURE_CLIENT_ID`
- `AZURE_CLIENT_SECRET`
- `AZURE_TENANT_ID`
- `AZURE_ENVIRONMENT`

### Azure Workload Identity Support
For Azure Workload Identity (Kubernetes environments), the provider will also support:
- `AZURE_FEDERATED_TOKEN_FILE` - Path to the federated token file
- `AZURE_AUTHORITY_HOST` - Azure Active Directory authority host
- `AZURE_CLIENT_ASSERTION` - Client assertion for federated identity credentials

## Terraform SDK Implementation

### Framework Choice
Follow the [Terraform Plugin SDK v2](https://developer.hashicorp.com/terraform/plugin/sdkv2) guidelines for implementation.

### Project Structure
```
provider-bc-admin-center/
├── .github/
│   ├── workflows/          # CI/CD workflows
│   └── instructions.md     # This file
├── docs/                   # Provider documentation
├── examples/               # Usage examples
├── internal/
│   ├── provider/          # Main provider implementation
│   ├── client/            # Business Central Admin Center API client
│   ├── services/          # Service-specific implementations
│   │   ├── environments/
│   │   ├── applications/
│   │   ├── notifications/
│   │   └── settings/
│   └── utils/             # Shared utilities
├── tests/                 # Integration and unit tests
├── go.mod
├── go.sum
├── main.go               # Provider entry point
└── README.md
```

### Key Implementation Guidelines

#### 1. Provider Configuration
```go
func Provider() *schema.Provider {
    return &schema.Provider{
        Schema: map[string]*schema.Schema{
            "client_id": {
                Type:        schema.TypeString,
                Optional:    true,
                DefaultFunc: schema.EnvDefaultFunc("AZURE_CLIENT_ID", ""),
            },
            "client_secret": {
                Type:        schema.TypeString,
                Optional:    true,
                Sensitive:   true,
                DefaultFunc: schema.EnvDefaultFunc("AZURE_CLIENT_SECRET", ""),
            },
            "tenant_id": {
                Type:        schema.TypeString,
                Optional:    true,
                DefaultFunc: schema.EnvDefaultFunc("AZURE_TENANT_ID", ""),
            },
        },
        ResourcesMap: map[string]*schema.Resource{
            "bc_environment":              resourceEnvironment(),
            "bc_environment_settings":     resourceEnvironmentSettings(),
            "bc_environment_app":          resourceEnvironmentApp(),
            "bc_notification_settings":    resourceNotificationSettings(),
        },
        DataSourcesMap: map[string]*schema.Resource{
            "bc_environment":              dataSourceEnvironment(),
            "bc_environments":             dataSourceEnvironments(),
            "bc_available_applications":   dataSourceAvailableApplications(),
            "bc_environment_quotas":       dataSourceEnvironmentQuotas(),
        },
        ConfigureContextFunc: providerConfigure,
    }
}
```

#### 2. Client Implementation
```go
type Client struct {
    credential   azcore.TokenCredential
    httpClient   *http.Client
    baseURL      string
    tenantID     string
}

func (c *Client) authenticatedRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
    token, err := c.credential.GetToken(ctx, policy.TokenRequestOptions{
        Scopes: []string{"996def3d-b36c-4153-8607-a6fd3c01b89f/.default"}, // Business Central resource ID
    })
    if err != nil {
        return nil, err
    }
    
    // Build and execute request with Bearer token
}
```

#### 3. Error Handling
Implement comprehensive error handling for Business Central Admin Center API error responses:

```go
type AdminCenterError struct {
    Code         string                 `json:"code"`
    Message      string                 `json:"message"`
    Target       string                 `json:"target,omitempty"`
    ClientError  []AdminCenterError     `json:"clientError,omitempty"`
}
```

#### 4. Resource Schema Patterns

**Environment Resource Example:**
```go
func resourceEnvironment() *schema.Resource {
    return &schema.Resource{
        CreateContext: resourceEnvironmentCreate,
        ReadContext:   resourceEnvironmentRead,
        UpdateContext: resourceEnvironmentUpdate,
        DeleteContext: resourceEnvironmentDelete,
        Importer: &schema.ResourceImporter{
            StateContext: schema.ImportStatePassthroughContext,
        },
        Schema: map[string]*schema.Schema{
            "name": {
                Type:         schema.TypeString,
                Required:     true,
                ForceNew:     true,
                ValidateFunc: validation.StringLenBetween(1, 30),
            },
            "application_family": {
                Type:     schema.TypeString,
                Required: true,
                ForceNew: true,
                ValidateFunc: validation.StringInSlice([]string{
                    "BusinessCentral",
                }, false),
            },
            "type": {
                Type:     schema.TypeString,
                Required: true,
                ForceNew: true,
                ValidateFunc: validation.StringInSlice([]string{
                    "Production",
                    "Sandbox",
                }, false),
            },
            "country_code": {
                Type:     schema.TypeString,
                Required: true,
                ForceNew: true,
            },
            "ring_name": {
                Type:     schema.TypeString,
                Required: true,
                ForceNew: true,
            },
            "application_version": {
                Type:     schema.TypeString,
                Required: true,
                ForceNew: true,
            },
            "status": {
                Type:     schema.TypeString,
                Computed: true,
            },
            "web_client_login_url": {
                Type:     schema.TypeString,
                Computed: true,
            },
        },
    }
}
```

#### 5. Async Operations Handling

Many Business Central Admin Center operations are asynchronous. Implement proper polling:

```go
func waitForOperation(ctx context.Context, client *Client, operationID string, timeout time.Duration) error {
    stateConf := &resource.StateChangeConf{
        Pending:    []string{"Running", "Queued"},
        Target:     []string{"Succeeded"},
        Refresh:    operationStateRefreshFunc(ctx, client, operationID),
        Timeout:    timeout,
        MinTimeout: 10 * time.Second,
        Delay:      30 * time.Second,
    }

    _, err := stateConf.WaitForStateContext(ctx)
    return err
}
```

### Testing Strategy

#### 1. Unit Tests
- Test individual functions and utilities
- Mock API responses for consistent testing
- Validate schema and validation functions

#### 2. Integration Tests
- Test against real Business Central Admin Center API
- Use separate test tenant/environments
- Implement cleanup procedures

#### 3. Acceptance Tests
Following Terraform provider conventions:

```go
func TestAccEnvironment_basic(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:          func() { testAccPreCheck(t) },
        ProviderFactories: testAccProviderFactories,
        CheckDestroy:      testAccCheckEnvironmentDestroy,
        Steps: []resource.TestStep{
            {
                Config: testAccEnvironment_basic(),
                Check: resource.ComposeTestCheckFunc(
                    testAccCheckEnvironmentExists("bc_environment.test"),
                    resource.TestCheckResourceAttr("bc_environment.test", "name", "test-env"),
                ),
            },
        },
    })
}
```

## Documentation Requirements

### 1. Provider Documentation
- Overview and getting started guide
- Authentication configuration
- Complete resource and data source reference

### 2. Resource Documentation
For each resource and data source:
- Description and use cases
- Complete schema documentation
- Usage examples
- Import instructions

### 3. Examples
- Basic environment setup
- Complete tenant configuration
- Multi-environment scenarios
- Integration with existing Terraform configurations

## Development Workflow

### 1. Prerequisites
- Go 1.21+
- Terraform 1.0+
- Access to Business Central Admin Center API
- Appropriate Azure AD application registration

### 2. Local Development
```bash
# Clone and build
git clone https://github.com/your-org/terraform-provider-bc-admin-center
cd terraform-provider-bc-admin-center
go build -o terraform-provider-bc-admin-center

# Install locally for testing
mkdir -p ~/.terraform.d/plugins/local/provider/bc-admin-center/1.0.0/linux_amd64
cp terraform-provider-bc-admin-center ~/.terraform.d/plugins/local/provider/bc-admin-center/1.0.0/linux_amd64/
```

### 3. Testing Configuration
```hcl
terraform {
  required_providers {
    bc_admin_center = {
      source = "local/provider/bc-admin-center"
      version = "1.0.0"
    }
  }
}
```

## Resource Design Principles

### 1. Stateful vs Non-Stateful Resources
Only implement resources for entities that can be managed statefully. Avoid creating resources for:

**Non-Stateful Items (Data Sources Only):**
- `bc_available_applications` - Application catalogs and available versions (read-only)
- `bc_application_versions` - Available versions for rings (read-only)  
- `bc_environment_quotas` - Quota limits and usage (read-only)
- `bc_environment_operations` - Historical operations log (read-only)

**Stateful Resources (Manageable):**
- `bc_environment` - Environment lifecycle (create/delete/manage)
- `bc_environment_settings` - Configurable environment properties
- `bc_environment_app` - App installations (install/uninstall/update)
- `bc_notification_settings` - Notification recipients and preferences

### 2. Resource Lifecycle Considerations
- **App Versions**: Do not create resources for specific app versions as they are immutable releases
- **Operations History**: Operations are events, not manageable state - use data sources only  
- **Quotas**: System-defined limits should be queried, not managed
- **Available Applications**: Catalogs are Microsoft-managed, not tenant-managed

### 3. State Management Guidelines
- Resources must have clear create/read/update/delete operations
- Avoid resources that represent point-in-time data or system-generated information
- Focus on tenant-configurable and environment-manageable entities

## Compliance and Best Practices

### 1. Security
- Never log sensitive information (tokens, secrets)
- Implement proper credential handling
- Use secure defaults for all configurations

### 2. API Best Practices
- Implement proper rate limiting and retry logic
- Handle API versioning appropriately
- Respect API quotas and limits

### 3. Terraform Best Practices
- Follow Terraform provider development guidelines
- Implement proper state management
- Provide clear error messages and documentation

### 4. Code Quality
- Comprehensive test coverage (>80%)
- Consistent code formatting (gofmt)
- Proper error handling and logging
- Clear and maintainable code structure

### 5. Testing
- Write unit tests for all functions
- Implement integration tests against real API
- Use acceptance tests for end-to-end validation

## Release Strategy

### 1. Versioning
Follow semantic versioning (semver) for releases

### 2. Release Pipeline
- Automated testing on multiple Terraform versions
- Automated builds for multiple platforms
- Signed releases and checksums
- Terraform Registry publication

### 3. Compatibility Matrix
Maintain compatibility with:
- Terraform 1.13+
- Business Central Admin Center API versions
- Go 1.25+

This provider will enable teams to manage their Business Central environments as code, providing consistent, repeatable, and version-controlled infrastructure management for Business Central tenants.
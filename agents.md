# Terraform Provider for Business Central Admin Center

## ⚠️ CRITICAL REQUIREMENTS

**TESTING IS MANDATORY**: Every new resource, data source, or service method MUST have corresponding tests before it is considered complete. See the [Testing Strategy](#testing-strategy) section for detailed requirements.

**DOCUMENTATION IS REQUIRED**: All resources and data sources must have complete documentation templates and examples. Do NOT create separate markdown files to summarize work - only update existing templates or generated docs.

**RUN `make generate` AFTER CHANGES**: After making code, schema, examples, templates, or documentation changes, you MUST run `make generate` before considering work complete or opening a pull request. This keeps generated documentation and related artifacts up to date.

**CONVENTIONAL COMMITS REQUIRED**: All commit messages and pull request titles MUST follow the [Conventional Commits](https://www.conventionalcommits.org/) specification. This ensures consistent, meaningful commit history and enables automated changelog generation. See the [Conventional Commits Guidelines](#conventional-commits-guidelines) section for details.

**CHANGELOG MUST BE UPDATED**: Every change — new resources, bug fixes, breaking changes, deprecations, or notable improvements — MUST be reflected in `CHANGELOG.md` before the work is considered complete. See the [Changelog Guidelines](#changelog-guidelines) section for details.

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
provider "bcadmincenter" {
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
provider-bcadmincenter/
├── .github/
│   ├── workflows/          # CI/CD workflows
│   └── instructions.md     # This file
├── docs/                   # Provider documentation
├── examples/               # Usage examples
├── internal/
│   ├── constants/         # Shared constants (ProviderNamespace, API version, etc.)
│   ├── provider/          # Main provider implementation
│   ├── client/            # Business Central Admin Center API client
│   ├── services/          # Service-specific implementations
│   │   ├── environments/
│   │   │   ├── resourceid.go      # Environment-specific resource ID functions
│   │   │   ├── resourceid_test.go
│   │   │   └── ...
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

## Shared Constants Package

The `internal/constants` package provides centralized constants used across the provider:

```go
package constants

// ProviderNamespace for Business Central Admin Center resources
const ProviderNamespace = "Microsoft.Dynamics365.BusinessCentral"

// DefaultBaseURL is the default Business Central Admin Center API endpoint
const DefaultBaseURL = "https://api.businesscentral.dynamics.com"

// BusinessCentralResourceID is the Azure AD resource ID for Business Central
const BusinessCentralResourceID = "996def3d-b36c-4153-8607-a6fd3c01b89f"

// DefaultAPIVersion is the default API version to use
const DefaultAPIVersion = "v2.24"
```

**When to use constants:**
- Use `constants.ProviderNamespace` in resource ID functions
- Use `constants.DefaultAPIVersion` in tests and client initialization
- Use `constants.DefaultBaseURL` when configuring clients
- Use `constants.BusinessCentralResourceID` for authentication scopes

**When NOT to use constants:**
- Business logic values (e.g., environment types, application families)
- User-provided configuration values
- Dynamic or computed values

### Key Implementation Guidelines

**REMINDER: All implementations require corresponding tests. See the Testing Strategy section below for requirements.**

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

#### 6. Resource ID Format

All resources in this provider use an ARM-like resource ID format to support multi-tenant scenarios. Resource IDs follow this structure:

```
/tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/{resourcePath}
```

**Examples:**

- **Notification Recipient**: `/tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/notificationRecipients/{recipientId}`
- **Environment**: `/tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/applications/{applicationFamily}/environments/{environmentName}`
- **Environment Settings**: `/tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/applications/{applicationFamily}/environments/{environmentName}/settings`
- **Environment Support Contact**: `/tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/applications/{applicationFamily}/environments/{environmentName}/supportContact`

**Implementation Guidelines:**

1. **Decentralized Resource IDs**: Each service package manages its own resource ID functions in a local `resourceid.go` file within the service directory (e.g., `internal/services/environments/resourceid.go`).
   
2. **Shared Constants**: Common constants are centralized in `internal/constants/constants.go`:
   - `ProviderNamespace` - The provider namespace for all resources
   - `DefaultAPIVersion` - The default API version
   - `DefaultBaseURL` - The base API endpoint
   - `BusinessCentralResourceID` - Azure AD resource ID for authentication

3. **Builder Functions**: Each service implements its own builder function to create resource IDs:
   ```go
   import "github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/constants"
   
   // In service's resourceid.go
   func BuildEnvironmentID(tenantID, applicationFamily, environmentName string) string {
       return fmt.Sprintf("/tenants/%s/providers/%s/applications/%s/environments/%s",
           tenantID, constants.ProviderNamespace, applicationFamily, environmentName)
   }
   ```

4. **Parser Functions**: Each service implements its own parser function:
   ```go
   // In service's resourceid.go
   func ParseEnvironmentID(id string) (string, string, string, error) {
       parts := strings.Split(strings.TrimPrefix(id, "/"), "/")
       
       if len(parts) != 8 {
           return "", "", "", fmt.Errorf("invalid environment ID format...")
       }
       
       // Validation logic using constants.ProviderNamespace
       
       return parts[1], parts[5], parts[7], nil
   }
   ```

5. **Multi-Tenant Support**: All resources support an optional `aad_tenant_id` attribute that:
   - Defaults to the provider's configured tenant ID if not specified
   - Allows managing resources in different tenants when explicitly set
   - Is included in the resource ID for proper multi-tenant isolation

6. **Testing Resource IDs**: When adding new resource types:
   - Add `resourceid.go` and `resourceid_test.go` in the service package
   - Include tests for: valid IDs, invalid formats, wrong providers, missing parts, and round-trip conversions
   - Import `internal/constants` package for shared constants

**Example Resource Implementation:**
```go
func (r *EnvironmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    // ... create resource via API ...
    
    // Build ARM-like resource ID using local function
    data.ID = types.StringValue(BuildEnvironmentID(
        data.AADTenantID.ValueString(),
        data.ApplicationFamily.ValueString(),
        data.Name.ValueString(),
    ))
    
    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EnvironmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    // Parse ARM-like resource ID using local function
    tenantID, appFamily, envName, err := ParseEnvironmentID(req.ID)
    if err != nil {
        resp.Diagnostics.AddError("Invalid Import ID", err.Error())
        return
    }
    
    // Set parsed values in state
    resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("aad_tenant_id"), tenantID)...)
    resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("application_family"), appFamily)...)
    resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), envName)...)
}
```

### Testing Strategy

**CRITICAL: Always create tests when implementing new resources or data sources.**

Every resource and data source MUST have comprehensive test coverage before being considered complete. Follow this testing checklist:

#### Required Test Files

When creating a new resource or data source, you MUST create the following test files:

1. **Service Tests** (`service_test.go` in the service package)
   - Test all service methods with mock HTTP responses
   - Test success scenarios
   - Test error scenarios (API errors, network errors, invalid responses)
   - Test edge cases (empty responses, malformed data)

2. **Data Source/Resource Tests** (`data_source_test.go` or `resource_test.go`)
   - Test Metadata() method returns correct type name
   - Test Schema() method defines all required and optional attributes
   - Test Configure() method handles provider data correctly
   - Test model structs can be created and populated

3. **Provider Registration Tests** (update `provider_test.go`)
   - Update DataSources() or Resources() test to expect the new count

#### 1. Unit Tests

**Service Layer Tests:**
```go
func TestService_MethodName(t *testing.T) {
    tests := []struct {
        name           string
        responseBody   interface{}
        responseStatus int
        wantErr        bool
        // Additional expectations
    }{
        {
            name: "successful response",
            responseBody: ExpectedResponse{...},
            responseStatus: http.StatusOK,
            wantErr: false,
        },
        {
            name: "not found error",
            responseBody: map[string]string{"error": "not found"},
            responseStatus: http.StatusNotFound,
            wantErr: true,
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Create mock server
            server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                w.WriteHeader(tt.responseStatus)
                json.NewEncoder(w).Encode(tt.responseBody)
            }))
            defer server.Close()

            // Create client with mock server
            mockCred := &mockTokenCredential{token: "test-token"}
            c := &client.Client{}
            c.SetCredential(mockCred)
            c.SetBaseURL(server.URL)
            c.SetAPIVersion(constants.DefaultAPIVersion)
            c.SetHTTPClient(&http.Client{})

            // Test the method
            svc := NewService(c)
            result, err := svc.MethodName(context.Background(), args...)

            // Assert results
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

**Data Source/Resource Tests:**
```go
func TestDataSourceName_Metadata(t *testing.T) {
    d := NewDataSource()
    req := datasource.MetadataRequest{ProviderTypeName: "bcadmincenter"}
    resp := &datasource.MetadataResponse{}
    
    d.Metadata(context.Background(), req, resp)
    
    expected := "bcadmincenter_resource_name"
    if resp.TypeName != expected {
        t.Errorf("TypeName = %v, want %v", resp.TypeName, expected)
    }
}

func TestDataSourceName_Schema(t *testing.T) {
    d := NewDataSource()
    req := datasource.SchemaRequest{}
    resp := &datasource.SchemaResponse{}
    
    d.Schema(context.Background(), req, resp)
    
    if resp.Diagnostics.HasError() {
        t.Fatalf("Schema() errors: %v", resp.Diagnostics)
    }
    
    // Verify required attributes exist
    if _, ok := resp.Schema.Attributes["required_attr"]; !ok {
        t.Error("Schema missing required_attr")
    }
}
```

- Test individual functions and utilities
- Mock API responses for consistent testing
- Validate schema and validation functions
- Test model struct creation and population

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

### 1. Provider Documentation Structure

The provider uses [terraform-plugin-docs](https://github.com/hashicorp/terraform-plugin-docs) (`tfplugindocs`) to generate documentation that complies with the [Terraform Registry documentation requirements](https://developer.hashicorp.com/terraform/registry/providers/docs).

#### Directory Structure
```
provider-bcadmincenter/
├── docs/                          # Generated documentation (do not edit manually)
│   ├── index.md                  # Provider overview and configuration
│   ├── resources/                # Resource documentation
│   │   └── environment.md
│   └── data-sources/             # Data source documentation
│       └── available_applications.md
├── templates/                     # Documentation templates (edit these)
│   ├── index.md.tmpl             # Provider documentation template
│   ├── resources/                # Resource documentation templates
│   │   └── environment.md.tmpl
│   └── data-sources/             # Data source documentation templates
│       └── available_applications.md.tmpl
└── examples/                      # Example Terraform configurations
    ├── provider/
    │   └── provider.tf           # Provider configuration examples
    ├── resources/
    │   └── bc_admin_center_environment/
    │       └── resource.tf       # Resource usage examples
    └── data-sources/
        └── bc_admin_center_available_applications/
            └── data-source.tf    # Data source usage examples
```

#### Documentation Generation Workflow

1. **Edit Templates**: Modify files in `templates/` directory
   - Use `{{.SchemaMarkdown}}` placeholder for schema documentation
   - Use `{{tffile "path/to/example.tf"}}` to include example files
   - Follow Terraform Registry markdown conventions

2. **Create Examples**: Add example Terraform configurations in `examples/`
   - Each resource/data source should have a dedicated subdirectory
   - Examples should be complete, working configurations
   - Include provider configuration when needed for context

3. **Generate Documentation**: Run the documentation generator
   ```bash
    make generate
   ```
   This will:
   - Extract schema from provider code
   - Process template files
   - Include example files
   - Generate final markdown in `docs/`

4. **Review Generated Docs**: Check `docs/` directory
   - Ensure schema is correctly rendered
   - Verify examples are properly included
   - Check for broken links or formatting issues

### 2. Documentation Template Guidelines

#### Provider Template (index.md.tmpl)

Must include:
- **Clear description** of provider purpose and capabilities
- **Authentication methods** with complete setup instructions
- **Required permissions** and how to configure them
- **Multiple usage examples** covering different authentication scenarios
- **Environment variables** reference table
- **Schema documentation** using `{{ .SchemaMarkdown }}` placeholder
- **Links to additional resources**

Best practices:
- Use callouts for warnings (`~>`) and notes (`->`)
- Provide step-by-step setup instructions
- Include CLI commands for common setup tasks
- Document all supported authentication methods
- Link to official Microsoft documentation

**IMPORTANT**: Do NOT create separate markdown files to summarize work, document changes, or log completed steps. All documentation should be updates to existing template files or generated docs only.

#### Resource Templates (resources/*.md.tmpl)

Must include:
- **Clear description** of what the resource manages
- **Important warnings** about destructive operations or limitations
- **Multiple usage examples** showing common patterns
- **Import instructions** with exact format and examples
- **Timeouts documentation** if supported
- **Attribute reference** (auto-generated from schema)
- **Best practices** section
- **Common issues** and troubleshooting

Template structure:
```markdown
---
page_title: "{{.Type}} {{.Name}} - {{.ProviderName}}"
subcategory: ""
description: |-
{{ .Description | plainmarkdown | trimspace | prefixlines "  " }}
---

# {{.Type}} ({{.Name}})

{{ .Description | trimspace }}

[Additional context about the resource]

## Important Notes

~> **Warning:** [Critical warnings about the resource]

## Example Usage

### Basic Example

{{tffile "examples/resources/[resource_name]/resource.tf"}}

### Advanced Example

[Inline example or additional tffile reference]

{{ .SchemaMarkdown | trimspace }}

## Import

[Import instructions with examples]

## Best Practices

[Usage recommendations]

## Common Issues

[Troubleshooting guidance]
```

#### Data Source Templates (data-sources/*.md.tmpl)

Must include:
- **Clear description** of what data is retrieved
- **Usage examples** showing common query patterns
- **Attribute reference** (auto-generated)
- **Use cases** demonstrating practical applications
- **Integration examples** with resources

### 3. Example File Requirements

All example files must:
- **Include copyright headers**
- **Be complete, working configurations**
- **Use realistic but safe values** (no real credentials)
- **Include comments** explaining non-obvious configurations
- **Follow Terraform style guidelines**
- **Be formatted** with `terraform fmt`

Example file template:
```terraform
# Copyright (c) 2025 Axians Infoma GmbH
# SPDX-License-Identifier: MPL-2.0

# [Brief description of what this example demonstrates]

terraform {
  required_providers {
    bcadmincenter = {
      source = "axiansinfoma/bcadmincenter"
    }
  }
}

provider "bcadmincenter" {
  # Configuration
}

# [Resource or data source usage]
```

### 4. Documentation Validation

Before submitting documentation:

1. **Generate docs**: `make generate`
2. **Review output**: Check for warnings or errors
3. **Validate examples**: Run `terraform fmt -recursive examples/`
4. **Check links**: Ensure all links are valid
5. **Test imports**: Verify import examples are accurate
6. **Spell check**: Review for typos and grammar

### 5. Continuous Documentation Updates

When adding new resources or data sources:

1. Create template file in `templates/resources/` or `templates/data-sources/`
2. Add example file in `examples/resources/` or `examples/data-sources/`
3. Run `make generate`
4. Commit both templates and generated docs
5. Update main README.md with new capabilities

### 6. Documentation Best Practices

- **Be concise but complete**: Provide necessary detail without overwhelming users
- **Use consistent terminology**: Match Terraform and Business Central terminology
- **Include error scenarios**: Document common errors and solutions
- **Show real-world patterns**: Examples should reflect actual use cases
- **Link to Microsoft docs**: Reference official BC documentation for detailed API behavior
- **Keep examples updated**: Ensure examples work with current provider version
- **Use semantic line breaks**: Break lines at sentence boundaries in templates for better diffs

### 7. Required Documentation Sections

Every resource and data source must document:

- [ ] Description and purpose
- [ ] At least one basic example
- [ ] Complete schema (auto-generated)
- [ ] Import instructions (resources only)
- [ ] Timeouts (if applicable)
- [ ] Important warnings or limitations
- [ ] Related resources and data sources

Provider documentation must include:

- [ ] Feature overview
- [ ] All authentication methods
- [ ] Permission requirements
- [ ] Environment variables reference
- [ ] Example configurations
- [ ] Links to additional resources

## Changelog Guidelines

`CHANGELOG.md` MUST be kept up to date as part of every change. Follow these rules:

### When to Update

Update `CHANGELOG.md` whenever you:
- Add a new resource or data source
- Fix a bug (even a minor one)
- Introduce a breaking change or deprecation
- Make a notable improvement, refactor, or performance change
- Add or change documentation in a user-visible way
- Bump dependencies that affect behaviour

### Format

The changelog follows [HashiCorp's provider changelog conventions](https://developer.hashicorp.com/terraform/plugin/best-practices/versioning):

```markdown
## X.Y.Z (Unreleased)

BREAKING CHANGES:
* resource/bc_example: attribute `foo` has been removed ([#123](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/issues/123))

FEATURES:
* **New Resource:** `bc_new_resource` ([#456](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/456))
* **New Data Source:** `bc_new_data_source` ([#456](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/456))

ENHANCEMENTS:
* resource/bc_environment: add `ignore_update_window` attribute ([#789](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/789))

BUG FIXES:
* resource/bc_environment: fix drift caused by version casing ([#101](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/101))
```

### Categories (in order)

| Category | Use for |
|---|---|
| `BREAKING CHANGES:` | Removals, renames, or behaviour changes requiring user action |
| `FEATURES:` | New resources, data sources, or provider-level capabilities |
| `ENHANCEMENTS:` | New attributes, improved behaviour, better error messages |
| `BUG FIXES:` | Corrections to existing behaviour |
| `NOTES:` | Deprecations, upgrade guidance, dependency bumps |

### Rules

- Every entry MUST reference a PR or issue number as a Markdown link.
- Entries go under the **topmost** `(Unreleased)` version block. Create one if it does not exist.
- Keep entries short: one line per change.
- Use the `resource/<name>:` or `data-source/<name>:` prefix for resource-scoped changes.
- Do NOT edit previously released version blocks.

## Development Workflow

### 1. Conventional Commits Guidelines

All commit messages and pull request titles MUST follow the [Conventional Commits](https://www.conventionalcommits.org/) specification.

#### Commit Message Format

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

#### Commit Types

- **feat**: A new feature (triggers minor version bump)
- **fix**: A bug fix (triggers patch version bump)
- **docs**: Documentation only changes
- **style**: Changes that don't affect code meaning (formatting, missing semicolons, etc.)
- **refactor**: Code change that neither fixes a bug nor adds a feature
- **perf**: Code change that improves performance
- **test**: Adding missing tests or correcting existing tests
- **build**: Changes that affect the build system or external dependencies
- **ci**: Changes to CI configuration files and scripts
- **chore**: Other changes that don't modify src or test files
- **revert**: Reverts a previous commit

#### Breaking Changes

Breaking changes MUST be indicated by:
1. A `!` after the type/scope: `feat!: remove support for X`
2. A `BREAKING CHANGE:` footer in the commit body

#### Examples

```bash
# Feature addition
feat(environments): add support for environment tags

# Bug fix
fix(auth): handle token expiration correctly

# Documentation
docs: update authentication examples in README

# Breaking change (method 1)
feat(client)!: change API client interface

# Breaking change (method 2)
feat(client): change API client interface

BREAKING CHANGE: The client now requires explicit configuration
```

#### Pull Request Titles

Pull request titles MUST follow the same format as commit messages:

```
feat: add environment scheduling resource
fix: correct validation in notification settings
docs: improve provider configuration examples
```

### 2. Prerequisites
- Go 1.21+
- Terraform 1.0+
- Access to Business Central Admin Center API
- Appropriate Azure AD application registration

### 3. Local Development
```bash
# Clone and build
git clone https://github.com/your-org/terraform-provider-bcadmincenter
cd terraform-provider-bcadmincenter
go build -o terraform-provider-bcadmincenter

# Install locally for testing
mkdir -p ~/.terraform.d/plugins/local/provider/bcadmincenter/1.0.0/linux_amd64
cp terraform-provider-bcadmincenter ~/.terraform.d/plugins/local/provider/bcadmincenter/1.0.0/linux_amd64/
```

### 4. Testing Configuration
```hcl
terraform {
  required_providers {
    bcadmincenter = {
      source = "local/provider/bcadmincenter"
      version = "1.0.0"
    }
  }
}
```

### 5. Running Tests

**CRITICAL: Always run tests before committing code.**

```bash
# Run all tests
go test ./... -v

# Run tests for a specific package
go test ./internal/services/available_applications/... -v

# Run tests with coverage
go test ./... -cover

# Run tests with coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

**Test Checklist Before Commit:**
- [ ] All unit tests pass
- [ ] New code has test coverage
- [ ] Service tests include success and error scenarios
- [ ] Data source/resource tests verify Metadata, Schema, and Configure
- [ ] Provider tests updated with new resource/data source counts
- [ ] No test files are missing for new implementations
- [ ] Run `make generate` and include generated documentation/artifact updates

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

### 5. Linting Guidelines

**CRITICAL: Always respect standardized boilerplate formats.**

The project uses golangci-lint v2.x with multiple enabled linters. When fixing linting issues:

**DO NOT modify standardized boilerplate:**
- ✅ Copyright headers: `// Copyright (c) 2025 Axians Infoma GmbH` (no period)
- ✅ SPDX identifiers: `// SPDX-License-Identifier: MPL-2.0` (no period)
- ✅ Build tags: `// +build` or `//go:build` directives (exact format required)
- ✅ Generated code markers: `// Code generated ... DO NOT EDIT` (exact format)
- ✅ Linter directives: `// nolint:lintername` (exact format)
- ✅ Go directives: `//go:generate`, `//go:embed` (exact format)

**Configuration limitations with golangci-lint v2.x:**
- The `godot` linter (comment period checker) is disabled because v2.x doesn't support excluding patterns
- Copyright and SPDX headers must not have periods (per SPDX specification)
- Regular code comments should still end with periods for consistency and readability

**When linter rules conflict with standardized formats:**
- Disable the linter rather than modifying boilerplate
- Document the decision in `.golangci.yml`
- NEVER modify standardized boilerplate to satisfy linter rules

**Running linting:**
```bash
# Run all linters
golangci-lint run ./...

# Verify configuration
golangci-lint config verify

# Auto-fix safe issues (when available)
golangci-lint run --fix ./...
```

### 6. Testing
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
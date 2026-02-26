# Support Contact Service

This package implements the Business Central Admin Center API support contact management functionality.

## Overview

The support contact service allows partners and administrators to configure support contact information that is displayed to users in the **Help and Support** page within Business Central environments. This provides a way to direct users to the appropriate support channels.

## API Endpoints

The service interacts with the following Business Central Admin Center API endpoint:

- `GET /admin/v2.24/support/applications/{applicationFamily}/environments/{environmentName}/supportcontact` - Retrieves support contact information
- `PUT /admin/v2.24/support/applications/{applicationFamily}/environments/{environmentName}/supportcontact` - Sets support contact information

**Note:** There is no DELETE endpoint. Support contacts can only be updated, not removed via the API.

## Service Methods

### Get(ctx, applicationFamily, environmentName)

Retrieves the current support contact information for an environment.

**Parameters:**
- `ctx` - Context for the request
- `applicationFamily` - Application family (e.g., "BusinessCentral")
- `environmentName` - Name of the environment

**Returns:**
- `*SupportContact` - The support contact information (nil if not configured)
- `error` - Error if the request fails

**Special Handling:**
- Returns `nil, nil` when no support contact is configured (404 response)
- Returns error for actual failures (network errors, permission issues, etc.)

### Set(ctx, applicationFamily, environmentName, contact)

Sets or updates the support contact information for an environment.

**Parameters:**
- `ctx` - Context for the request
- `applicationFamily` - Application family (e.g., "BusinessCentral")
- `environmentName` - Name of the environment
- `contact` - SupportContact struct with name, email, and URL

**Returns:**
- `*SupportContact` - The updated support contact information
- `error` - Error if the request fails

## Data Models

### SupportContact

Represents support contact information for an environment.

```go
type SupportContact struct {
    Name  string `json:"name"`
    Email string `json:"email"`
    URL   string `json:"url"`
}
```

**Fields:**
- `Name` - The name of the support contact (displayed to users)
- `Email` - The email address for support inquiries
- `URL` - A URL for additional support information (support portal, knowledge base, etc.)

## Terraform Resources

This service implements:

- **Resource:** `bcadmincenter_support_contact` - Manages support contact configuration

The resource follows standard Terraform CRUD patterns:
- **Create:** Sets initial support contact information
- **Read:** Retrieves current support contact information
- **Update:** Updates support contact information
- **Delete:** Removes from Terraform state with warning (API has no delete endpoint)

## Error Handling

The service uses consistent error handling patterns:

### 404 Not Found

A 404 response has special meaning:
- Indicates no support contact is configured
- Service returns `nil, nil` to indicate absence
- Resource treats this as "not exists" state

### Other Errors

All other errors are propagated to the caller:
- Authentication errors (401, 403)
- Environment not found
- Validation errors (invalid email format, etc.)
- Network errors

**Implementation Detail:** The client wrapper returns errors for status codes >= 400 before the response can be inspected. Therefore, 404 detection is done by parsing the error message for "404" string.

## Testing

The service includes comprehensive tests:

### Service Tests (`service_test.go`)
- Success scenarios (Get and Set)
- Not found scenarios (404 handling)
- Error scenarios (environment not found, validation errors)

### Resource Tests (`resource_test.go`)
- Metadata validation
- Schema validation
- Configure method validation

All tests use `httptest.NewServer` to mock API responses without requiring real API access.

## Usage Examples

### Basic Support Contact

```hcl
resource "bcadmincenter_support_contact" "production" {
  application_family = "BusinessCentral"
  environment_name   = "Production"
  
  name  = "IT Support Team"
  email = "support@example.com"
  url   = "https://support.example.com"
}
```

### Multiple Environments

```hcl
variable "environments" {
  type = map(object({
    name  = string
    email = string
    url   = string
  }))
}

resource "bcadmincenter_support_contact" "contacts" {
  for_each = var.environments
  
  application_family = "BusinessCentral"
  environment_name   = each.key
  
  name  = each.value.name
  email = each.value.email
  url   = each.value.url
}
```

## Best Practices

1. **Use Dedicated Email**: Use dedicated support email addresses rather than personal addresses
2. **Keep URLs Current**: Ensure support portal URLs are actively maintained
3. **Environment-Specific Contacts**: Consider different contacts for production vs. sandbox
4. **Monitor Contacts**: Ensure contact email addresses are actively monitored
5. **Test Before Production**: Verify contact information in sandbox environments first

## API Limitations

- No DELETE endpoint - contacts cannot be removed via API
- Contact information is environment-specific
- Email validation is performed by the API
- Changes may take time to propagate to user sessions

## Dependencies

- `github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/client` - HTTP client wrapper
- `github.com/hashicorp/terraform-plugin-framework` - Terraform plugin framework

## Related Services

- `environments` - Environment management (support contacts are per-environment)
- `environment_settings` - Other environment configuration options

## References

- [Business Central Admin Center API Documentation](https://learn.microsoft.com/en-us/dynamics365/business-central/dev-itpro/administration/administration-center-api)
- [Support Contact API Reference](https://learn.microsoft.com/en-us/dynamics365/business-central/dev-itpro/administration/administration-center-api_support_contact)

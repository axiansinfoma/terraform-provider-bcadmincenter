# Notification Recipients Service

This service manages notification recipients for Business Central tenants through the Admin Center API.

## API Endpoints

- `GET /admin/v2.24/settings/notification/recipients` - List all notification recipients
- `PUT /admin/v2.24/settings/notification/recipients` - Create a new notification recipient
- `DELETE /admin/v2.24/settings/notification/recipients/{id}` - Delete a notification recipient

## Features

- List all notification recipients for a tenant
- Get a specific notification recipient by ID
- Create new notification recipients
- Delete notification recipients

## Notes

- Up to 100 notification recipients can be configured per tenant
- The API does not support updating existing recipients (email and name are immutable)
- Recipients receive notifications from `no-reply-dynamics365@microsoft.com`
- Notification events include update availability, successful updates, failures, and extension validations

## API Response Example

```json
{
  "value": [
    {
      "id": "00000000-0000-0000-0000-000000000001",
      "email": "admin@example.com",
      "name": "Administrator"
    }
  ]
}
```

## Error Codes

- `invalidInput` - Invalid email or name (empty/null)
- `requestBodyRequired` - Request body must be provided
- `tenantNotFound` - Tenant information not found

## Related Documentation

- [Business Central Admin Center API - Notifications](https://learn.microsoft.com/en-us/dynamics365/business-central/dev-itpro/administration/administration-center-api_notifications)
- [Managing Tenant Notifications](https://learn.microsoft.com/en-us/dynamics365/business-central/dev-itpro/administration/tenant-admin-center-notifications)

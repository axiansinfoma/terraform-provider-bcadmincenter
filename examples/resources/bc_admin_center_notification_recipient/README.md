# Notification Recipient Resource Example

This example demonstrates how to configure notification recipients for a Business Central tenant.

## Basic Usage

The example creates a notification recipient who will receive email notifications about environment lifecycle events such as:
- Update availability
- Successful updates
- Update failures
- Extension validations

## Important Notes

- Up to 100 notification recipients can be configured per tenant
- Email addresses must be valid
- Notifications are sent from `no-reply-dynamics365@microsoft.com`
- Ensure emails are not filtered to spam folders
- Both email and name are required when creating a recipient
- Updating email or name requires replacing the resource (delete and recreate)

## Related Resources

- Consider using a distribution list email if you need more than 100 recipients
- Microsoft 365 Message Center also posts environment lifecycle events

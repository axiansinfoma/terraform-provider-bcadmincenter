# Support Contact Resource Example

This example demonstrates how to configure support contact information for a Business Central environment.

## What This Example Shows

- Configuring basic support contact information
- Setting environment-specific support contacts
- Integrating with environment resources

## Prerequisites

- Business Central environment already exists
- Service principal with appropriate permissions (AdminCenter.ReadWrite.All)
- Environment name and application family information

## Usage

1. Update the provider configuration with your credentials
2. Modify the environment name to match your target environment
3. Set the contact information (name, email, URL)
4. Run terraform apply

## Notes

- The support contact information is displayed to users in the Help and Support page
- Each environment can have its own support contact configuration
- The API does not support deleting contacts - they must be updated or manually removed
- Contact information should be kept current and actively monitored

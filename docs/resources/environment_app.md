---
page_title: "Resource bcadmincenter_environment_app - bcadmincenter"
subcategory: "Environment"
description: |-
  Manages the install/update/uninstall lifecycle for an app in a Business Central environment.
  This resource installs a Business Central app and manages its version. Install, update and uninstall are asynchronous operations that can take several minutes to complete.
---

# Resource (bcadmincenter_environment_app)

Manages the install/update/uninstall lifecycle for an app in a Business Central environment.

This resource installs a Business Central app and manages its version. Install, update and uninstall are asynchronous operations that can take several minutes to complete.

Manages the install/update/uninstall lifecycle of a Business Central app in an environment. This resource:

- Installs the specified app (optionally at a pinned version) via the Admin Center API.
- Monitors the asynchronous install/update/uninstall operation until completion.
- Supports in-place version upgrades by changing the `target_version` attribute.
- Automatically removes from state if the app is no longer installed.

## Important Notes

~> **Warning:** App install, update, and uninstall are asynchronous operations that can take several minutes to complete.

~> **Warning:** Downgrading `target_version` to a lower value is blocked at plan time. Only version upgrades are supported via this resource.

-> **Note:** When the app reaches a terminal failure state (`"installFailed"` or `"updateFailed"`), Terraform will automatically plan a replacement (destroy + recreate) on the next `terraform plan`, allowing the environment to converge without manual state manipulation.

-> **Note:** All identity attributes (`application_family`, `environment_name`, `app_id`) are `ForceNew`. Changing any of them destroys the existing app installation and installs the new app.

## Deferred Updates and the Environment Update Window

When `use_environment_update_window = true` (the default), the Business Central API does not execute the update immediately. Instead, it schedules the operation to run during the environment's configured [update window](https://learn.microsoft.com/en-us/dynamics365/business-central/dev-itpro/administration/tenant-admin-center-update-management#set-the-update-window). The operation appears in the `"scheduled"` state until the window opens.

Because the operation has not yet completed, Terraform cannot verify the final installed version at apply time. The resource handles this transparently:

- `pending_target_version` — set to the version requested during a deferred apply. On subsequent `terraform plan` / `terraform refresh` runs, this value suppresses false drift: Terraform will not report the installed version as changed until the scheduled operation completes and `pending_target_version` is cleared.
- `pending_operation_id` — the ID of the scheduled BC operation, used internally if you later change `use_environment_update_window` and the resource needs to cancel the existing scheduled operation before resubmitting.

**Typical workflow with update windows:**

1. Run `terraform apply` — BC accepts the update and schedules it. State records `pending_target_version`.
2. Wait for the environment's update window to pass and BC to execute the update.
3. Run `terraform apply` again (or `terraform refresh`) — BC reports the installed version now matches `target_version`, `pending_target_version` is cleared, and state converges.

**Changing `use_environment_update_window` while an update is pending:**

If you toggle `use_environment_update_window` while `pending_target_version` is set, the resource will automatically cancel the existing scheduled operation and resubmit the update with the new setting. The cancellation requires the scheduled operation ID from the BC App Operations API (`/apps/{appId}/operations`). If cancellation is rejected by BC (for example, because the operation has already started running), a descriptive error is returned and you should wait for the update to complete before applying again.

## Example Usage

### Install a Specific App Version

```terraform
# Copyright (c) 2025 Axians Infoma GmbH
# SPDX-License-Identifier: MPL-2.0

# This example demonstrates how to install a Business Central app into an environment.
# The app version is pinned to a specific version; omitting the version attribute installs
# the latest available version. Changing the version to a higher value on a subsequent
# apply triggers an in-place update without recreating the resource.

terraform {
  required_providers {
    bcadmincenter = {
      source = "axiansinfoma/bcadmincenter"
    }
  }
}

provider "bcadmincenter" {
  # Authentication via Service Principal (or use environment variables)
  # client_id     = var.client_id
  # client_secret = var.client_secret
  # tenant_id     = var.tenant_id
}

resource "bcadmincenter_environment" "sandbox" {
  name               = "my-sandbox"
  application_family = "BusinessCentral"
  type               = "Sandbox"
  country_code       = "US"
  ring_name          = "PROD"
}

resource "bcadmincenter_environment_app" "contoso_app" {
  application_family = bcadmincenter_environment.sandbox.application_family
  environment_name   = bcadmincenter_environment.sandbox.name

  app_id         = "00000000-0000-0000-0000-000000000000"
  target_version = "1.0.0.0" # Omit to install the latest available version.

  install_or_update_needed_dependencies = true
  allow_preview_version                 = false
}

output "app_status" {
  description = "The current install status of the app."
  value       = bcadmincenter_environment_app.contoso_app.status
}

output "app_version" {
  description = "The installed version of the app."
  value       = bcadmincenter_environment_app.contoso_app.target_version
}
```

### Install Latest Version

```terraform
resource "bcadmincenter_environment_app" "my_app" {
  application_family = "BusinessCentral"
  environment_name   = "my-sandbox"
  app_id             = "00000000-0000-0000-0000-000000000000"
  # target_version omitted — installs the latest available version
}
```

### In-Place Version Upgrade

Changing `target_version` from `"1.0.0.0"` to `"1.1.0.0"` on a subsequent apply triggers an in-place update without recreating the resource:

```terraform
resource "bcadmincenter_environment_app" "my_app" {
  application_family = "BusinessCentral"
  environment_name   = "my-sandbox"
  app_id             = "00000000-0000-0000-0000-000000000000"
  target_version     = "1.1.0.0"  # changed from "1.0.0.0" → in-place update
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `app_id` (String) The app GUID to install. Changing this forces a new resource to be created.
- `application_family` (String) The application family for the environment (e.g. `"BusinessCentral"`). Changing this forces a new resource to be created.
- `environment_name` (String) The name of the target environment. Changing this forces a new resource to be created.

### Optional

- `aad_tenant_id` (String) The Azure AD tenant ID. If not specified, defaults to the provider's configured tenant ID.
- `accept_isv_eula` (Boolean) When `true`, accepts the ISV End User License Agreement (EULA) for the app. Required for some ISV apps. Defaults to `false`. Changing this forces a new resource to be created.
- `allow_preview_version` (Boolean) When `true`, allows installing preview versions of the app. Defaults to `false`.
- `install_or_update_needed_dependencies` (Boolean) When `true`, automatically installs or updates app dependencies. Defaults to `true`.
- `language_id` (String) The language identifier for the app installation (e.g. `"en-US"`). If not specified, the default language is used. Changing this forces a new resource to be created.
- `target_version` (String) The target app version to install or update to (e.g. `"1.2.3.4"`). Omit or leave null to install the latest available version. Changing this to a higher version schedules an in-place update. Downgrading is blocked at plan time.
- `timeouts` (Attributes) Timeout configuration for the resource operations. (see [below for nested schema](#nestedatt--timeouts))
- `use_environment_update_window` (Boolean) When `true` (default), update and uninstall operations respect the environment's configured update window. Set to `false` to bypass the update window and apply the operation immediately.

### Read-Only

- `id` (String) The ARM-like resource ID (format: /tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/applications/{applicationFamily}/environments/{environmentName}/apps/{appId})
- `name` (String) The display name of the app (read from the API).
- `pending_operation_id` (String) The BC operation ID of a currently scheduled (deferred) update. Non-empty when an update has been deferred to the environment's update window. Used internally to cancel and reschedule the operation when `use_environment_update_window` changes.
- `pending_target_version` (String) The target version of a currently scheduled or running update. Non-empty when an update has been deferred to the environment's update window. While non-empty, `target_version` is suppressed to this value so no drift is reported.
- `published_as` (String) How the app is published (e.g. `"Global"`).
- `publisher` (String) The publisher of the app (read from the API).
- `status` (String) The current install status of the app (e.g. `"installed"`, `"installFailed"`, `"updateFailed"`). When the status is `"installFailed"` or `"updateFailed"`, the resource will be replaced on the next apply.

<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String) Timeout for create operations. Defaults to 60 minutes.
- `delete` (String) Timeout for delete operations. Defaults to 60 minutes.

## Import

Environment app resources can be imported using the ARM-like resource ID:

```shell
terraform import bcadmincenter_environment_app.example /tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/applications/{applicationFamily}/environments/{environmentName}/apps/{appId}
```

For example:

```shell
terraform import bcadmincenter_environment_app.contoso_app \
  /tenants/00000000-0000-0000-0000-000000000001/providers/Microsoft.Dynamics365.BusinessCentral/applications/BusinessCentral/environments/my-sandbox/apps/00000000-0000-0000-0000-000000000002
```

## Best Practices

- Pin `target_version` to a specific value to ensure deterministic deployments.
- Omit `target_version` to always track the latest available version.
- Use `install_or_update_needed_dependencies = true` (default) to avoid dependency conflicts.
- Use `allow_preview_version = false` (default) for production environments.
- Leave `use_environment_update_window = true` (default) to respect the environment's configured update window. Set it to `false` only for urgent updates where you want immediate execution.
- The `pending_target_version` and `pending_operation_id` attributes are managed automatically — do not set them manually.
- If an app enters `"installFailed"` or `"updateFailed"` status, the next `terraform plan` will propose a replacement — run `terraform apply` to recover.

## Related Resources

- [`bcadmincenter_environment`](../resources/environment.md) — Manages the Business Central environment that hosts the app.

---
page_title: "Resource bcadmincenter_per_tenant_extension - bcadmincenter"
subcategory: "Extension Management"
description: |-
  Manages the full lifecycle of a Per-Tenant Extension (PTE) in a Business Central environment.
  This resource uploads a .app extension package, installs or schedules it, updates it when the package changes, and uninstalls it on destroy. All operations use the Business Central Admin Center API PTE endpoints, which require API version 2.29 or later.
  ~> Note: Exactly one of file_path or file_content must be set.
---

# Resource (bcadmincenter_per_tenant_extension)

Manages the full lifecycle of a Per-Tenant Extension (PTE) in a Business Central environment.

This resource uploads a `.app` extension package, installs or schedules it, updates it when the package changes, and uninstalls it on destroy. All operations use the **Business Central Admin Center API** PTE endpoints, which require **API version 2.29 or later**.

~> **Note:** Exactly one of `file_path` or `file_content` must be set.

Manages the full lifecycle of a Per-Tenant Extension (PTE) in a Business Central environment using the **Business Central Admin Center API**. This resource:

- Uploads a `.app` extension package to `apps/pteInstall` and installs it, or schedules it for a later deployment window.
- Polls the app operation until it reaches a terminal state (immediate deployments only).
- Updates the extension when `file_sha256` changes — uploading a new package installs or schedules the new version, no prior uninstall required.
- Cancels any still-scheduled versions and uninstalls the extension on destroy.

~> **Warning:** This resource requires **Admin Center API version 2.29 or later**, which introduced the PTE endpoints. Set `api_version` on the provider only if you need to pin a different version; the provider defaults to a version that supports these endpoints.

~> **Warning:** Extension install and update operations are asynchronous and can take several minutes. The resource polls until the operation reaches a terminal state. Deployments scheduled for a future window (`deployment_schedule` other than `"Immediate"`) are **not** waited on — the apply returns as soon as the API accepts the upload, and `pending_target_version` records what will be deployed.

~> **Warning:** Setting `delete_data = true` permanently deletes all data associated with the extension on destroy. This is irreversible.

-> **Note:** Exactly one of `file_path` or `file_content` must be set. Setting both or neither produces a plan-time error.

-> **Note:** The `.app` package cannot exceed 50 MB, the limit enforced by the `pteInstall` endpoint.

## Migrating from the Automation API implementation

Before Admin Center API 2.29 this resource drove PTEs through the BC Automation API. That path has been replaced, which changes the configuration surface:

| Removed attribute | Replacement |
|---|---|
| `company_id` | No longer needed — the Admin Center API is environment-scoped, not company-scoped. |
| `schedule` | `deployment_schedule`. The old values `"Current version"`, `"Next minor version"`, and `"Next major version"` are still accepted and mapped to `"Immediate"`, `"NextMinorUpdate"`, and `"NextMajorUpdate"`. |
| `schema_sync_mode` | `sync_mode`. The old value `"Force Sync"` is still accepted and mapped to `"ForceSync"`. |
| `unpublish_on_delete` | No equivalent — the Admin Center API does not expose an unpublish operation. Old versions are cleaned up by the platform. |
| `package_id` | `app_id` is the stable identity; there is no per-upload package ID in this API. |

The new `accept_isv_eula` attribute is **required** and must be set to `true`.

## Required Permissions

### Azure AD (Entra) Application Permissions

Only the permission the provider already requires for all other operations is needed:

| API | Permission name | Type |
|---|---|---|
| Dynamics 365 Business Central (`996def3d-b36c-4153-8607-a6fd3c01b89f`) | `AdminCenter.ReadWrite.All` | Application |

-> **Note:** The `Automation.ReadWrite.All` permission and the `D365 AUTOMATION` / `EXT. MGT. - ADMIN` Business Central permission sets that the previous Automation API implementation required are no longer needed.

## Example Usage

### Install from a Local .app File

```terraform
# Copyright (c) 2025 Axians Infoma GmbH
# SPDX-License-Identifier: MPL-2.0

# Example A: install a PTE from a local .app file, immediately
resource "bcadmincenter_per_tenant_extension" "my_pte" {
  environment_name   = "MyProdEnvironment"
  application_family = "BusinessCentral"

  # Exactly one of file_path or file_content must be set.
  file_path   = "./extensions/MyExtension_1.0.0.0.app"
  file_sha256 = filesha256("./extensions/MyExtension_1.0.0.0.app")

  deployment_schedule = "Immediate"
  sync_mode           = "Add"
  language_id         = "en-US"

  # Required: accepts the publisher's EULA and the associated Marketplace terms.
  accept_isv_eula = true

  install_or_update_needed_dependencies = true

  delete_data          = false
  uninstall_dependents = false
}

# Example B: install a PTE from base64-encoded content (e.g. an Azure Storage blob)
data "azurerm_storage_blob" "pte_package" {
  name                   = "MyExtension_1.0.0.0.app"
  storage_account_name   = "mystorageaccount"
  storage_container_name = "bc-extensions"
}

resource "bcadmincenter_per_tenant_extension" "my_pte_from_blob" {
  environment_name   = "MySandboxEnvironment"
  application_family = "BusinessCentral"

  # file_content accepts base64-encoded .app bytes (e.g. from a storage blob data source).
  # file_sha256 must be the SHA-256 hash of the decoded content.
  # Azure Storage blobs expose content_md5 (MD5), not SHA-256.
  # Compute the SHA-256 hash separately and pass it here.
  file_content = data.azurerm_storage_blob.pte_package.content
  file_sha256  = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" # replace with actual SHA-256

  # The API requires an uploaded file name ending in .app. Set it explicitly when the
  # bytes do not come from file_path.
  file_name = "MyExtension_1.0.0.0.app"

  accept_isv_eula = true
}

# Example C: stage an update to deploy during the environment's update window
resource "bcadmincenter_per_tenant_extension" "my_pte_scheduled" {
  environment_name   = "MyDevEnvironment"
  application_family = "BusinessCentral"

  file_path   = "./extensions/MyExtension_2.0.0.0.app"
  file_sha256 = filesha256("./extensions/MyExtension_2.0.0.0.app")

  # "UpdateWindow" defers the install to the environment's update window.
  # "NextMinorUpdate" and "NextMajorUpdate" are also valid, but only for an
  # extension that is already installed.
  deployment_schedule = "UpdateWindow"

  accept_isv_eula = true

  # Cancels any still-scheduled versions of this extension on destroy so they
  # cannot reinstall it after the uninstall. Defaults to true.
  cancel_scheduled_on_destroy = true

  timeouts = {
    create = "90m"
    update = "90m"
    delete = "30m"
  }
}
```

### Update an Existing Extension

Changing `file_sha256` (and the accompanying `file_path`/`file_content`) on a subsequent apply
uploads and installs the new package version. No prior uninstall step is required.

```terraform
resource "bcadmincenter_per_tenant_extension" "my_pte" {
  environment_name   = "MyProdEnvironment"
  application_family = "BusinessCentral"

  file_path   = "./extensions/MyExtension_2.0.0.0.app"              # updated path
  file_sha256 = filesha256("./extensions/MyExtension_2.0.0.0.app")  # updated hash

  accept_isv_eula = true
}
```

### Defer an Update to the Next Minor Release

`"NextMinorUpdate"` and `"NextMajorUpdate"` stage the package so it deploys alongside the
platform update. They are only valid for an extension that is already installed.

```terraform
resource "bcadmincenter_per_tenant_extension" "my_pte" {
  environment_name   = "MyProdEnvironment"
  application_family = "BusinessCentral"

  file_path   = "./extensions/MyExtension_2.0.0.0.app"
  file_sha256 = filesha256("./extensions/MyExtension_2.0.0.0.app")

  deployment_schedule = "NextMinorUpdate"
  accept_isv_eula     = true
}

output "pending_version" {
  value = bcadmincenter_per_tenant_extension.my_pte.pending_target_version
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `accept_isv_eula` (Boolean) Must be set to `true` for the installation to proceed. Setting it to `true` accepts the publisher's end-user license terms and the associated Microsoft Marketplace terms, exactly as when installing the extension through the Business Central admin center.
- `application_family` (String) The application family of the environment (e.g. `"BusinessCentral"`). Changing this forces a new resource to be created.
- `environment_name` (String) The name of the target environment. Changing this forces a new resource to be created.
- `file_sha256` (String) SHA-256 hash of the `.app` file content. Drives change detection — changing this value triggers an update.

### Optional

- `aad_tenant_id` (String) The Azure AD tenant ID. If not specified, defaults to the provider's configured tenant ID. Changing this forces a new resource to be created.
- `cancel_scheduled_on_destroy` (Boolean) When `true` (default), any versions of this extension still scheduled for a future deployment window are cancelled on destroy before the uninstall runs. Uninstalling alone does not remove scheduled versions, so they would otherwise reinstall the extension later.
- `delete_data` (Boolean) When `true`, the uninstall performed on destroy also deletes the extension's data and syncs it clean (irreversible). Defaults to `false`.
- `deployment_schedule` (String) When the uploaded package is installed. One of `"Immediate"` (default), `"UpdateWindow"`, `"NextMinorUpdate"`, or `"NextMajorUpdate"`. Must be `"Immediate"` or `"UpdateWindow"` for a PTE that is not yet installed; updates to an installed PTE may use any value. The pre-2.29 names `"Current version"`, `"Next minor version"`, and `"Next major version"` are accepted and mapped to their modern equivalents.
- `file_content` (String, Sensitive) Base64-encoded `.app` file bytes. Mutually exclusive with `file_path`. Enables passing content directly from a data source (e.g. `azurerm_storage_blob`). When set, `file_name` should also be set because the API requires an uploaded file name ending in `.app`.
- `file_name` (String) File name submitted with the upload. Must end in `.app`. Defaults to the base name of `file_path`, or `extension.app` when `file_content` is used.
- `file_path` (String) Local path to the `.app` file. Mutually exclusive with `file_content`.
- `install_or_update_needed_dependencies` (Boolean) When `true`, dependencies of the uploaded extension are installed or updated automatically. When `false` (default), the install fails and returns the missing dependencies as error details.
- `language_id` (String) Microsoft Language Code ID applied during install (e.g. `"en-US"`). Defaults to `"en-US"`.
- `sync_mode` (String) Schema synchronisation mode applied during install. One of `"Add"` (default) or `"ForceSync"`. The pre-2.29 name `"Force Sync"` is accepted and mapped to `"ForceSync"`.
- `timeouts` (Attributes) Timeout configuration for the resource operations. Values are Go duration strings (e.g. `"90m"`). (see [below for nested schema](#nestedatt--timeouts))
- `uninstall_dependents` (Boolean) When `true`, apps that depend on this extension are uninstalled along with it on destroy. When `false` (default), the uninstall fails and returns the dependent apps as error details.
- `uninstall_in_update_window` (Boolean) When `true`, the uninstall performed on destroy runs only inside the environment's update window and is reported as `scheduled` until then. Defaults to `false`, which uninstalls immediately.

### Read-Only

- `app_id` (String) Stable extension identity read from the uploaded `.app` package. Remains constant across version updates.
- `app_type` (String) App type reported by the API. `"tenant"` for per-tenant extensions.
- `display_name` (String) Display name of the extension.
- `id` (String) ARM-like resource ID (format: `/tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/applications/{applicationFamily}/environments/{environmentName}/perTenantExtensions/{appId}`).
- `last_operation_id` (String) ID of the most recent install or update operation triggered for this extension.
- `pending_schedule_kind` (String) Schedule kind of a pending install or update (e.g. `"UpdateWindow"`, `"NextMinorUpdate"`). Empty when nothing is pending.
- `pending_target_version` (String) Target version of an install or update that is still scheduled for a future deployment window. Empty when nothing is pending.
- `publisher` (String) Publisher of the extension.
- `state` (String) Install state reported by the API (e.g. `"installed"`, `"updatePending"`, `"updating"`).
- `version` (String) Currently installed version. Empty while the first install is still scheduled for a future deployment window.

<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String) Timeout for create operations. Defaults to 60 minutes.
- `delete` (String) Timeout for delete operations. Defaults to 60 minutes.
- `update` (String) Timeout for update operations. Defaults to 60 minutes.

## Import

Per-tenant extension resources can be imported using the ARM-like resource ID:

```shell
terraform import bcadmincenter_per_tenant_extension.example \
  /tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/applications/{applicationFamily}/environments/{environmentName}/perTenantExtensions/{appId}
```

For example:

```shell
terraform import bcadmincenter_per_tenant_extension.my_pte \
  /tenants/00000000-0000-0000-0000-000000000001/providers/Microsoft.Dynamics365.BusinessCentral/applications/BusinessCentral/environments/MyProdEnvironment/perTenantExtensions/d0e4c7e2-1234-5678-abcd-ef0123456789
```

After import, you must configure `file_path` or `file_content`, `file_sha256`, and `accept_isv_eula` manually in your Terraform configuration — these are write-only inputs that are not stored in or read back from BC state.

## Best Practices

- Use `file_sha256 = filesha256(...)` to drive change detection automatically from the local file.
- Leave `sync_mode = "Add"` (default) unless you explicitly need `"ForceSync"`, which can cause data loss.
- Leave `cancel_scheduled_on_destroy = true` (default). Uninstalling a PTE does **not** remove versions already scheduled for a future window, so without this a destroyed extension can reinstall itself later.
- Set `install_or_update_needed_dependencies = true` when the extension has dependencies you also manage; otherwise the install fails and reports the missing dependencies.
- Use the [`bcadmincenter_scheduled_pte_operations`](../data-sources/scheduled_pte_operations.md) data source to audit what is still queued for an environment.

## Related Resources

- [`bcadmincenter_environment`](../resources/environment.md) — Manages the Business Central environment that hosts the extension.
- [`bcadmincenter_environment_app`](../resources/environment_app.md) — Manages marketplace (global) apps, which use the same app management API but a different install flow.
- [`bcadmincenter_scheduled_pte_operations`](../data-sources/scheduled_pte_operations.md) — Lists PTE installs and updates scheduled for a future deployment window.

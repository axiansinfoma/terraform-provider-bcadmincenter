---
page_title: "Resource bcadmincenter_per_tenant_extension - bcadmincenter"
subcategory: "Extension Management"
description: |-
  Manages the full lifecycle of a Per-Tenant Extension (PTE) in a Business Central environment.
  This resource uploads a .app extension package, installs it, updates it when the package changes, and uninstalls it on destroy. All operations use the Business Central Automation API.
  ~> Note: Exactly one of file_path or file_content must be set.
---

# Resource (bcadmincenter_per_tenant_extension)

Manages the full lifecycle of a Per-Tenant Extension (PTE) in a Business Central environment.

This resource uploads a `.app` extension package, installs it, updates it when the package changes, and uninstalls it on destroy. All operations use the **Business Central Automation API**.

~> **Note:** Exactly one of `file_path` or `file_content` must be set.

Manages the full lifecycle of a Per-Tenant Extension (PTE) in a Business Central environment using the **BC Automation API**. This resource:

- Uploads a `.app` extension package and installs it (3-step sequence).
- Polls `extensionDeploymentStatus` until the deployment completes.
- Updates the extension when `file_sha256` changes (upload new package, no prior uninstall required).
- Optionally unpublishes the old package version after an update.
- Uninstalls (and optionally deletes data, and optionally unpublishes) on destroy.

~> **Warning:** Extension install and update operations are asynchronous and can take several minutes. The resource polls until the operation reaches a terminal state.

~> **Warning:** Setting `delete_data = true` permanently deletes all data associated with the extension on destroy. This is irreversible.

-> **Note:** Exactly one of `file_path` or `file_content` must be set. Setting both or neither produces a plan-time error.

## Required Permissions

### Azure AD (Entra) Application Permissions

In addition to `AdminCenter.ReadWrite.All` required by the provider for all other operations, the Azure AD app registration must also be granted:

| API | Permission name | Type |
|---|---|---|
| Dynamics 365 Business Central (`996def3d-b36c-4153-8607-a6fd3c01b89f`) | `Automation.ReadWrite.All` | Application |

### Business Central Permission Sets

The service principal (automation user) in Business Central must be assigned **both** of the following permission sets in each environment where PTEs are managed:

| Permission Set ID | Purpose |
|---|---|
| `D365 AUTOMATION` | Grants access to Automation API endpoints including `extensionUpload` and `extensions`. |
| `EXT. MGT. - ADMIN` | Grants rights to install, uninstall, and manage extensions. |

These are assigned per-environment inside Business Central:
**Settings → Users → select the automation user → Permission Sets**.

## Example Usage

### Install from a Local .app File

```terraform
# Copyright (c) 2025 Axians Infoma GmbH
# SPDX-License-Identifier: MPL-2.0

# Example A: install a PTE from a local .app file
resource "bcadmincenter_per_tenant_extension" "my_pte" {
  environment_name   = "MyProdEnvironment"
  application_family = "BusinessCentral"

  # Exactly one of file_path or file_content must be set.
  file_path   = "./extensions/MyExtension_1.0.0.0.app"
  file_sha256 = filesha256("./extensions/MyExtension_1.0.0.0.app")

  schedule         = "Current version"
  schema_sync_mode = "Add"

  delete_data         = false
  unpublish_on_delete = false
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

  schedule         = "Next minor version"
  schema_sync_mode = "Add"

  unpublish_on_delete = true
}

# Example C: install from a local file with an explicit company_id
resource "bcadmincenter_per_tenant_extension" "my_pte_explicit_company" {
  environment_name   = "MyDevEnvironment"
  application_family = "BusinessCentral"

  # company_id is optional; when omitted the first company in the environment is used.
  company_id = "00000000-0000-0000-0000-000000000001"

  file_path   = "./extensions/MyExtension_1.0.0.0.app"
  file_sha256 = filesha256("./extensions/MyExtension_1.0.0.0.app")
}
```

### Install from Base64 Content (e.g. Azure Storage Blob)

```terraform
data "azurerm_storage_blob" "pte_package" {
  name                   = "MyExtension_1.0.0.0.app"
  storage_account_name   = "mystorageaccount"
  storage_container_name = "bc-extensions"
}

resource "bcadmincenter_per_tenant_extension" "my_pte" {
  environment_name   = "MySandboxEnvironment"
  application_family = "BusinessCentral"

  # file_content accepts base64-encoded .app bytes.
  # file_sha256 must be the SHA-256 hash of the decoded content.
  # Azure Storage blobs expose content_md5 (MD5), not SHA-256.
  # Compute the SHA-256 hash separately (e.g. sha256sum of the downloaded file)
  # and pass it as a string literal.
  file_content = data.azurerm_storage_blob.pte_package.content
  file_sha256  = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" # replace with actual SHA-256

  schedule         = "Next minor version"
  schema_sync_mode = "Add"
}
```

### Update an Existing Extension

Changing `file_sha256` (and optionally `file_path`/`file_content`) on a subsequent apply
uploads and installs the new package version. No prior uninstall step is required.

```terraform
resource "bcadmincenter_per_tenant_extension" "my_pte" {
  environment_name   = "MyProdEnvironment"
  application_family = "BusinessCentral"

  file_path   = "./extensions/MyExtension_2.0.0.0.app"   # updated path
  file_sha256 = filesha256("./extensions/MyExtension_2.0.0.0.app")  # updated hash

  unpublish_on_delete = true  # also unpublishes the old package version after update
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `application_family` (String) The application family of the environment (e.g. `"BusinessCentral"`). Changing this forces a new resource to be created.
- `environment_name` (String) The name of the target environment. Changing this forces a new resource to be created.
- `file_sha256` (String) SHA-256 hash of the `.app` file content. Drives change detection — changing this value triggers an update.

### Optional

- `aad_tenant_id` (String) The Azure AD tenant ID. If not specified, defaults to the provider's configured tenant ID.
- `company_id` (String) The ID of the Business Central company used for Automation API calls. When not set the provider resolves it automatically by using the first company in the environment. PTEs are published globally across all companies so the choice of company is only an implementation detail.
- `delete_data` (Boolean) When `true`, calls `Microsoft.NAV.uninstallAndDeleteExtensionData` on destroy (irreversible). Defaults to `false`.
- `file_content` (String, Sensitive) Base64-encoded `.app` file bytes. Mutually exclusive with `file_path`. Enables passing content directly from a data source (e.g. `azurerm_storage_blob`).
- `file_path` (String) Local path to the `.app` file. Mutually exclusive with `file_content`.
- `schedule` (String) Installation schedule. One of `"Current version"` (default), `"Next minor version"`, or `"Next major version"`.
- `schema_sync_mode` (String) Schema synchronisation mode. One of `"Add"` (default) or `"Force Sync"`.
- `unpublish_on_delete` (Boolean) When `true`, calls `Microsoft.NAV.unpublish` after uninstall/update. Requires BC v25.4 or later. The call is silently skipped on older BC versions. Defaults to `false`.

### Read-Only

- `app_id` (String) Stable extension identity (`id` field) that remains constant across version updates.
- `display_name` (String) Display name of the extension.
- `id` (String) ARM-like resource ID (format: `/tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/applications/{applicationFamily}/environments/{environmentName}/perTenantExtensions/{appId}`).
- `package_id` (String) `packageId` of the currently installed upload. Changes with every new upload.
- `publisher` (String) Publisher of the extension.
- `version` (String) Installed version in `major.minor.build.revision` format.

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

After import, you must configure `file_path` or `file_content` and `file_sha256` manually in your Terraform configuration — these are write-only inputs that are not stored in or read back from BC state.

## Best Practices

- Use `file_sha256 = filesha256(...)` to drive change detection automatically from the local file.
- Set `unpublish_on_delete = true` to keep the extension list clean after updates and deletes (requires BC v25.4+; silently skipped on older versions).
- Leave `schema_sync_mode = "Add"` (default) unless you explicitly need `"Force Sync"`, which can cause data loss.
- Set `company_id` explicitly only if you have a specific requirement; the resource resolves it automatically from the first available company.

## Related Resources

- [`bcadmincenter_environment`](../resources/environment.md) — Manages the Business Central environment that hosts the extension.
- [`bcadmincenter_environment_app`](../resources/environment_app.md) — Manages marketplace apps via the Admin Center API (different API from PTEs).

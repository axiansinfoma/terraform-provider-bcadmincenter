## 0.1.6 (Unreleased)

BREAKING CHANGES:
* **resource/bcadmincenter_per_tenant_extension: reimplemented on the Business Central Admin Center API PTE endpoints introduced in API version 2.29.** The resource no longer uses the BC Automation API, so the `Automation.ReadWrite.All` Entra permission and the `D365 AUTOMATION` / `EXT. MGT. - ADMIN` Business Central permission sets are no longer required — `AdminCenter.ReadWrite.All` alone is sufficient. Configuration changes:
    * `accept_isv_eula` is new and **required**; it must be set to `true` for an install to proceed.
    * `schedule` was replaced by `deployment_schedule` (`"Immediate"`, `"UpdateWindow"`, `"NextMinorUpdate"`, `"NextMajorUpdate"`). The old values `"Current version"`, `"Next minor version"`, and `"Next major version"` are still accepted and mapped to their new equivalents.
    * `schema_sync_mode` was replaced by `sync_mode` (`"Add"`, `"ForceSync"`). The old value `"Force Sync"` is still accepted.
    * `company_id` was removed — the Admin Center API is environment-scoped rather than company-scoped.
    * `unpublish_on_delete` was removed — the Admin Center API exposes no unpublish operation.
    * `package_id` was removed — `app_id` is the stable identity in this API.

ENHANCEMENTS:
* provider: the default Admin Center API version is now `v2.29` (was `v2.27`), which is the version that introduced the per-tenant extension endpoints
* resource/bcadmincenter_per_tenant_extension: add `language_id`, `install_or_update_needed_dependencies`, `uninstall_dependents`, `uninstall_in_update_window`, and `file_name` attributes, plus a configurable `timeouts` block for create, update, and delete
* resource/bcadmincenter_per_tenant_extension: add `cancel_scheduled_on_destroy` (default `true`), which removes versions still scheduled for a future deployment window on destroy. Uninstalling a PTE does not clear scheduled versions, so without this a destroyed extension could reinstall itself later
* resource/bcadmincenter_per_tenant_extension: expose `state`, `app_type`, `last_operation_id`, `pending_target_version`, and `pending_schedule_kind`. Installs deferred to a future deployment window no longer block the apply and are tracked through the pending attributes
* client: support `multipart/form-data` requests and per-request HTTP timeouts, which the `.app` package upload requires
* resource/bcadmincenter_per_tenant_extension: destroy now waits until the extension has actually left the environment. The API reports the uninstall operation as `succeeded` while the app lingers in `uninstallPending` for another ~30-60 seconds, so a destroy immediately followed by an apply previously failed with "already installed"

FEATURES:
* **New Data Source:** `bcadmincenter_scheduled_pte_operations` — lists per-tenant extension installs and updates scheduled for a future deployment window on an environment

BUG FIXES:
* resource/bcadmincenter_environment: fix perpetual drift where the `settings` block was always shown as being added even after a successful apply ([#60](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/issues/60)). Two root causes were fixed: (1) `application_version` was missing `UseStateForUnknown()`, causing the plan to show it as `(known after apply)` when unset — this made `versionChanged = true` in the Update function and prevented the settings state from being saved; (2) `access_with_m365_licenses` inside the `settings` block was also missing `UseStateForUnknown()`, causing it to appear as `(known after apply)` on every plan and triggering unnecessary update cycles even when settings were unchanged.
## 0.1.3

BREAKING CHANGES:
* **Resource `bcadmincenter_environment_settings` has been removed.** Use the `settings` nested block on `bcadmincenter_environment` instead ([#42](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/issues/42))

ENHANCEMENTS:
* provider: `ARM_CLIENT_ID`, `ARM_CLIENT_SECRET`, `ARM_TENANT_ID`, and `ARM_ENVIRONMENT` environment variables are now supported, matching the `azurerm` provider convention; the existing `AZURE_*` variables remain supported for backward compatibility ([#57](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/57))
* resource/bcadmincenter_environment: add optional `settings` nested block to manage environment settings inline ([#42](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/issues/42))

FEATURES:
* **New Resource:** `bcadmincenter_authorized_entra_app` ([#1](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/1))
* **New Resource:** `bcadmincenter_environment` ([#1](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/1))
* **New Resource:** `bcadmincenter_environment_app` ([#11](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/issues/11))
* **New Resource:** `bcadmincenter_environment_support_contact` ([#1](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/1))
* **New Resource:** `bcadmincenter_environment_update_schedule` ([#35](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/issues/35))
* **New Resource:** `bcadmincenter_notification_recipient` ([#1](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/1))
* **New Resource:** `bcadmincenter_per_tenant_extension` — manage Per-Tenant Extension (PTE) lifecycle via the BC Automation API: upload `.app` package, install, update, uninstall, and optionally unpublish ([#4](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/issues/4))
* **New Data Source:** `bcadmincenter_application_family` ([#1](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/1))
* **New Data Source:** `bcadmincenter_authorized_entra_apps` ([#1](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/1))
* **New Data Source:** `bcadmincenter_available_applications` ([#1](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/1))
* **New Data Source:** `bcadmincenter_environment` ([#1](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/1))
* **New Data Source:** `bcadmincenter_environment_updates` ([#35](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/issues/35))
* **New Data Source:** `bcadmincenter_environments` ([#1](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/1))
* **New Data Source:** `bcadmincenter_manageable_tenants` ([#1](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/1))
* **New Data Source:** `bcadmincenter_notification_settings` ([#1](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/1))
* **New Data Source:** `bcadmincenter_quotas` ([#1](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/1))
* **New Data Source:** `bcadmincenter_timezones` ([#1](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/1))

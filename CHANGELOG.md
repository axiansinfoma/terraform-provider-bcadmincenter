## 0.2.1

BREAKING CHANGES:
* **provider: a static access token in `BCADMINCENTER_TEST_TOKEN` is now refused, and `base_url` must use `https`.** Both were testing affordances compiled into every release binary with nothing but a docstring restraining them: the token took precedence over every credential path and bypassed Azure AD entirely, while an unvalidated `base_url` received a live Azure AD access token on every request, because the `Authorization` header is attached before the destination is examined. Both are now gated behind the `bcadmincenter_testing` build tag, so the code paths are compiled out of released binaries ([#102](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/102))
* **resource/bcadmincenter_environment, resource/bcadmincenter_environment_app, resource/bcadmincenter_environment_support_contact, resource/bcadmincenter_environment_update_schedule: `aad_tenant_id` now forces replacement.** It selects which tenant the resource is read from and written to, so an in-place update talked to the old tenant while recording the new one. `bcadmincenter_per_tenant_extension` already behaved this way ([#103](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/103))
* **resource/bcadmincenter_environment: removing `app_update_cadence`, `partner_access_status`, or a configured update window from the `settings` block is now rejected at plan time.** The Admin Center API has no operation to unset them, so the removal previously produced no request at all — Terraform recorded the attribute as gone while the environment kept it. Set the attribute explicitly to the value you want (`"Default"`, `"Disabled"`) instead. The three update-window attributes must also now be configured together, because they are sent as a single payload in which absent fields are ignored ([#107](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/107))

ENHANCEMENTS:
* provider: built with Go 1.27.1 and an updated module graph, which resolves eight vulnerabilities reachable from provider code — seven in the standard library and one in `google.golang.org/grpc` (bumped to v1.83.2) ([#101](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/101))
* resource/bcadmincenter_environment, resource/bcadmincenter_environment_app: the documented `timeouts` block is now honoured. Both resources advertised `create`/`delete` and documented a 60 minute default, then hardcoded 60 minutes regardless of the configured value ([#103](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/103))
* provider: overriding `base_url`, or selecting an `environment` other than `public`, now raises a warning. The `environment` setting is currently accepted but not implemented — every request goes to the public cloud endpoint ([#105](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/105))
* all resources: ship an `import.sh` example, and correct the `examples/` directory names, which used a `bc_admin_center_*` prefix matching no actual resource type ([#105](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/105))

BUG FIXES:
* resource/bcadmincenter_environment: `azure_region` is no longer discarded on refresh. The API does not return it, and the provider overwrote the configured value with null on both create and read, so every apply failed with "Provider produced inconsistent result after apply" and every subsequent plan proposed replacing the environment ([#103](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/103))
* resource/bcadmincenter_authorized_entra_app: API calls now go to the tenant named by `aad_tenant_id`. The resource accepted the attribute, wrote it into state and into the resource ID, and then sent every request to the provider's own tenant, so Terraform reported success for an authorization that did not exist in the named tenant ([#103](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/103))
* resource/bcadmincenter_notification_recipient: a failed lookup no longer removes the resource from state. Read treated every error as "the recipient is gone", so a transient 503 or a client timeout dropped it from state and the next apply created a duplicate ([#103](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/103))
* provider: API status values and app IDs are compared case-insensitively, and an unrecognised operation status no longer aborts a healthy long-running operation. The live API returns statuses capitalised on some endpoints and lower-cased on others, so an install that had actually succeeded could fail with `unknown operation status: Succeeded`. Both spellings of "cancel(l)ed" are accepted ([#103](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/103))
* resource/bcadmincenter_environment, resource/bcadmincenter_environment_app, resource/bcadmincenter_per_tenant_extension, resource/bcadmincenter_environment_update_schedule: a failure after the resource already exists remotely no longer discards it. Terraform previously recorded nothing, so the environment or app finished provisioning untracked and the next apply failed as "already exists" ([#103](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/103), [#104](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/104))
* resource/bcadmincenter_environment, resource/bcadmincenter_environment_app, resource/bcadmincenter_per_tenant_extension, resource/bcadmincenter_environment_update_schedule: a resource deleted outside Terraform is now removed from state on refresh instead of raising an error that broke every subsequent plan, apply and destroy ([#103](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/103), [#104](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/104))
* resource/bcadmincenter_per_tenant_extension: `file_sha256` is now verified against the uploaded package. Nothing hashed or compared it, so rebuilding a `.app` in place without changing the hash silently uploaded nothing, and a package rebuilt between plan and apply was uploaded unchecked ([#104](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/104))
* resource/bcadmincenter_environment_app, resource/bcadmincenter_per_tenant_extension: `terraform import` no longer leaves attributes unset that the following plan reads as changes. For `bcadmincenter_environment_app` this proposed destroying and recreating — uninstalling and reinstalling — an app that was already installed ([#104](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/104))
* resource/bcadmincenter_environment: removing `security_group_id` or `app_insights_key` from the `settings` block now reaches the API. Removal produced no request, so state recorded the change while the environment kept the setting — and because these settings are not read back, no later plan could detect the difference ([#104](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/104), [#107](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/107))
* resource/bcadmincenter_environment: an update window written through the `settings` block no longer fails when the API answers `204 No Content`, and a failed read of one setting no longer discards a value that was applied successfully ([#104](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/104))
* data-source/bcadmincenter_timezones: `supports_daylight_savings` and `offset_from_utc` are populated again. Two structs decoded the same endpoint with mutually exclusive field names, so one of them produced `false` and `""` for every entry ([#104](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/104))
* provider: request paths are percent-escaped. Values from configuration were interpolated raw, so an environment name containing `?`, `#`, or `/` changed the structure of the request and could silently target a different environment ([#102](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/102))
* provider: API errors now carry their HTTP status. Error responses that do not use the documented `{code, message}` shape decoded to an empty struct and rendered as `": "`, which also defeated the 404 detection used to remove deleted resources from state ([#102](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/102))
* resource/bcadmincenter_notification_recipient, resource/bcadmincenter_environment_support_contact, resource/bcadmincenter_environment_update_schedule: values echoed back by the API no longer overwrite required or optional-but-unset attributes, which failed the apply with "inconsistent result after apply" after the change had already been written ([#104](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/104))

## 0.2.0

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

## 0.1.7

ENHANCEMENTS:
* provider: dependency and GitHub Actions updates only; no functional changes

## 0.1.6

BUG FIXES:
* resource/bcadmincenter_environment: fix perpetual drift where the `settings` block was always shown as being added even after a successful apply ([#60](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/issues/60)). Two root causes were fixed: (1) `application_version` was missing `UseStateForUnknown()`, causing the plan to show it as `(known after apply)` when unset — this made `versionChanged = true` in the Update function and prevented the settings state from being saved; (2) `access_with_m365_licenses` inside the `settings` block was also missing `UseStateForUnknown()`, causing it to appear as `(known after apply)` on every plan and triggering unnecessary update cycles even when settings were unchanged.

ENHANCEMENTS:
* provider: dependency and GitHub Actions updates

## 0.1.5

BUG FIXES:
* resource/bcadmincenter_environment: fix handling of the environment version ([#59](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/59))

## 0.1.4

BUG FIXES:
* resource/bcadmincenter_environment: use the correct application family while waiting for environment creation, and correct the accompanying documentation ([#58](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/58))

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

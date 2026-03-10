## 0.1.0 (Unreleased)

BREAKING CHANGES:
* **Resource `bcadmincenter_environment_settings` has been removed.** Use the `settings` nested block on `bcadmincenter_environment` instead ([#42](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/issues/42))

ENHANCEMENTS:
* resource/bcadmincenter_environment: add optional `settings` nested block to manage environment settings inline ([#42](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/issues/42))

FEATURES:
* **New Resource:** `bcadmincenter_authorized_entra_app` ([#1](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/1))
* **New Resource:** `bcadmincenter_environment` ([#1](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/1))
* **New Resource:** `bcadmincenter_environment_app` ([#11](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/issues/11))
* **New Resource:** `bcadmincenter_environment_support_contact` ([#1](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/1))
* **New Resource:** `bcadmincenter_environment_update_schedule` ([#35](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/issues/35))
* **New Resource:** `bcadmincenter_notification_recipient` ([#1](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/1))
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

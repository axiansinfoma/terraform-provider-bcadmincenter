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

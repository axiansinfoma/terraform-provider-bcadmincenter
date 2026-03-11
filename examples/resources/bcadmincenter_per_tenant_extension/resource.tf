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

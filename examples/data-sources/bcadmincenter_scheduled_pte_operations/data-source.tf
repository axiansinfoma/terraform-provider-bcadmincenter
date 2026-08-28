# Copyright (c) 2025 Axians Infoma GmbH
# SPDX-License-Identifier: MPL-2.0

# List every per-tenant extension install or update still scheduled for a future
# deployment window on an environment.
data "bcadmincenter_scheduled_pte_operations" "prod" {
  application_family = "BusinessCentral"
  environment_name   = "MyProdEnvironment"
}

output "scheduled_ptes" {
  value = [
    for op in data.bcadmincenter_scheduled_pte_operations.prod.operations :
    "${op.name} ${op.target_app_version} (${op.schedule_kind})"
  ]
}

# Narrow the result to a single extension by passing app_id.
data "bcadmincenter_scheduled_pte_operations" "my_pte" {
  application_family = "BusinessCentral"
  environment_name   = "MyProdEnvironment"
  app_id             = bcadmincenter_per_tenant_extension.my_pte.app_id
}

output "my_pte_pending_version" {
  value = try(data.bcadmincenter_scheduled_pte_operations.my_pte.operations[0].target_app_version, "none")
}

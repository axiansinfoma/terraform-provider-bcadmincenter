// Copyright (c) 2025 Michael Villani
// SPDX-License-Identifier: MPL-2.0

package quotas

// QuotasResponse represents the API response for environment quotas.
type QuotasResponse struct {
	ProductionEnvironmentsQuota     int `json:"productionEnvironmentsQuota"`
	ProductionEnvironmentsAllocated int `json:"productionEnvironmentsAllocated"`
	SandboxEnvironmentsQuota        int `json:"sandboxEnvironmentsQuota"`
	SandboxEnvironmentsAllocated    int `json:"sandboxEnvironmentsAllocated"`
	StorageQuotaGB                  int `json:"storageQuotaGB"`
	StorageAllocatedGB              int `json:"storageAllocatedGB"`
}

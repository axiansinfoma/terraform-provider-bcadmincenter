// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package pertenantextensions

import "strings"

// Deployment schedule values accepted by the `apps/pteInstall` endpoint
// (API version 2.29 and later). Send these exact values; the API echoes them back
// lower-cased, so compare with strings.EqualFold rather than by equality.
const (
	DeploymentScheduleImmediate       = "Immediate"
	DeploymentScheduleUpdateWindow    = "UpdateWindow"
	DeploymentScheduleNextMinorUpdate = "NextMinorUpdate"
	DeploymentScheduleNextMajorUpdate = "NextMajorUpdate"
)

// Schema synchronisation modes accepted by the `apps/pteInstall` endpoint.
const (
	SyncModeAdd       = "Add"
	SyncModeForceSync = "ForceSync"
)

// Defaults applied when the corresponding attribute is omitted.
const (
	DefaultDeploymentSchedule = DeploymentScheduleImmediate
	DefaultSyncMode           = SyncModeAdd
	DefaultLanguageID         = "en-US"
)

// MaxExtensionFileSize is the upload limit the pteInstall endpoint enforces on the
// .app package (50 MB).
const MaxExtensionFileSize = 50 * 1024 * 1024

// App operation status values. The API returns these capitalised on some endpoints and
// lower-cased on others, so always compare with strings.EqualFold.
const (
	OperationStatusScheduled = "scheduled"
	OperationStatusRunning   = "running"
	OperationStatusQueued    = "queued"
	OperationStatusSucceeded = "succeeded"
	OperationStatusFailed    = "failed"
	OperationStatusCanceled  = "canceled"
	// OperationStatusCancelled is the double-l spelling used by the environment and
	// environment_apps operations endpoints. Accept both rather than depending on which
	// one a given endpoint returns: matching only one spelling meant a cancelled
	// operation was never recognised, so the poller ran to the full timeout and then
	// reported a misleading "timed out" instead of "was canceled".
	OperationStatusCancelled = "cancelled"
	OperationStatusSkipped   = "skipped"
)

// Installed app states reported by the `apps` endpoint. The live API returns these in
// camelCase even though the documentation shows TitleCase, so compare with
// strings.EqualFold rather than by equality.
const (
	AppStateInstalled     = "installed"
	AppStateUpdatePending = "updatePending"
	AppStateUpdating      = "updating"
)

// AppTypeTenant is the `appType` the `apps` endpoint reports for per-tenant extensions.
const AppTypeTenant = "tenant"

// legacyDeploymentSchedules maps the Automation API schedule names this resource used
// before API version 2.29 onto their Admin Center API equivalents, so existing
// configurations keep working without edits.
var legacyDeploymentSchedules = map[string]string{
	"current version":    DeploymentScheduleImmediate,
	"next minor version": DeploymentScheduleNextMinorUpdate,
	"next major version": DeploymentScheduleNextMajorUpdate,
	"update window":      DeploymentScheduleUpdateWindow,
	"immediate":          DeploymentScheduleImmediate,
	"nextminorupdate":    DeploymentScheduleNextMinorUpdate,
	"nextmajorupdate":    DeploymentScheduleNextMajorUpdate,
	"updatewindow":       DeploymentScheduleUpdateWindow,
}

// legacySyncModes maps the Automation API schema sync mode names onto their Admin
// Center API equivalents.
var legacySyncModes = map[string]string{
	"add":        SyncModeAdd,
	"force sync": SyncModeForceSync,
	"forcesync":  SyncModeForceSync,
}

// NormalizeDeploymentSchedule converts a configured deployment schedule — either a
// canonical 2.29 value or one of the pre-2.29 Automation API names — into the value the
// pteInstall endpoint expects. Unrecognised values are returned unchanged so the API
// reports the error.
func NormalizeDeploymentSchedule(value string) string {
	if value == "" {
		return DefaultDeploymentSchedule
	}
	if normalized, ok := legacyDeploymentSchedules[strings.ToLower(strings.TrimSpace(value))]; ok {
		return normalized
	}
	return value
}

// NormalizeSyncMode converts a configured sync mode — either a canonical 2.29 value or
// the pre-2.29 Automation API name — into the value the pteInstall endpoint expects.
// Unrecognised values are returned unchanged so the API reports the error.
func NormalizeSyncMode(value string) string {
	if value == "" {
		return DefaultSyncMode
	}
	if normalized, ok := legacySyncModes[strings.ToLower(strings.TrimSpace(value))]; ok {
		return normalized
	}
	return value
}

// PteInstallRequest is the multipart/form-data payload for the
// `apps/pteInstall` endpoint.
type PteInstallRequest struct {
	// FileName is the name of the uploaded part. It must carry the .app extension.
	FileName string
	// Content is the raw .app package. It cannot exceed MaxExtensionFileSize.
	Content []byte
	// DeploymentSchedule determines when the package is installed. Must be
	// "Immediate" or "UpdateWindow" for a PTE that is not yet installed.
	DeploymentSchedule string
	// SyncMode is the schema synchronisation mode applied during install.
	SyncMode string
	// LanguageID is a Microsoft Language Code ID such as "en-US".
	LanguageID string
	// AcceptIsvEula must be true for the installation to proceed.
	AcceptIsvEula bool
	// InstallOrUpdateNeededDependencies installs or updates missing dependencies
	// instead of failing the operation.
	InstallOrUpdateNeededDependencies bool
}

// UninstallAppRequest is the request body for the `apps/{appId}/uninstall` endpoint.
type UninstallAppRequest struct {
	UseEnvironmentUpdateWindow bool `json:"useEnvironmentUpdateWindow"`
	UninstallDependents        bool `json:"uninstallDependents"`
	DeleteData                 bool `json:"deleteData"`
}

// RemoveScheduledPteVersionRequest is the request body for the
// `apps/{appId}/removeScheduledPteVersion` endpoint.
type RemoveScheduledPteVersionRequest struct {
	TargetVersion string `json:"targetVersion"`
	ScheduleKind  string `json:"scheduleKind"`
}

// AppOperation is an install, update, or uninstall operation returned by the app
// management endpoints.
//
// The version fields are duplicated because the API is inconsistent: `pteInstall`,
// `scheduledPteOperations`, and `removeScheduledPteVersion` return
// sourceAppVersion/targetAppVersion, while `apps/{appId}/operations` returns
// sourceVersion/targetVersion. Read them through SourceVersionValue and
// TargetVersionValue rather than directly.
type AppOperation struct {
	ID                   string `json:"id"`
	Type                 string `json:"type"`
	Status               string `json:"status"`
	CreatedOn            string `json:"createdOn,omitempty"`
	StartedOn            string `json:"startedOn,omitempty"`
	CompletedOn          string `json:"completedOn,omitempty"`
	ErrorMessage         string `json:"errorMessage,omitempty"`
	CreatedBy            string `json:"createdBy,omitempty"`
	CanceledBy           string `json:"canceledBy,omitempty"`
	CreatorPrincipalType string `json:"creatorPrincipalType,omitempty"`
	AppID                string `json:"appId,omitempty"`
	AADTenantID          string `json:"aadTenantId,omitempty"`
	ScheduleKind         string `json:"scheduleKind,omitempty"`

	SourceAppVersion string `json:"sourceAppVersion,omitempty"`
	TargetAppVersion string `json:"targetAppVersion,omitempty"`
	SourceVersion    string `json:"sourceVersion,omitempty"`
	TargetVersion    string `json:"targetVersion,omitempty"`
}

// SourceVersionValue returns the source version regardless of which field name the
// responding endpoint used.
func (o *AppOperation) SourceVersionValue() string {
	if o.SourceAppVersion != "" {
		return o.SourceAppVersion
	}
	return o.SourceVersion
}

// TargetVersionValue returns the target version regardless of which field name the
// responding endpoint used.
func (o *AppOperation) TargetVersionValue() string {
	if o.TargetAppVersion != "" {
		return o.TargetAppVersion
	}
	return o.TargetVersion
}

// IsScheduled reports whether the operation is waiting for a future deployment window.
func (o *AppOperation) IsScheduled() bool {
	return strings.EqualFold(o.Status, OperationStatusScheduled)
}

// AppOperationListResponse is the list wrapper returned by the app operations endpoint.
type AppOperationListResponse struct {
	Value []AppOperation `json:"value"`
}

// ScheduledPteOperationParameters is the snapshot of the install request a scheduled
// PTE operation was created with.
type ScheduledPteOperationParameters struct {
	AppID            string `json:"appId,omitempty"`
	TargetAppVersion string `json:"targetAppVersion,omitempty"`
	CountryCode      string `json:"countryCode,omitempty"`
	LanguageID       string `json:"languageId,omitempty"`
	Name             string `json:"name,omitempty"`
	Publisher        string `json:"publisher,omitempty"`
	ScheduleKind     string `json:"scheduleKind,omitempty"`
	TargetRelease    string `json:"targetRelease,omitempty"`
	SyncMode         string `json:"syncMode,omitempty"`
}

// ScheduledPteOperation is an entry from the `apps/scheduledPteOperations` endpoint.
type ScheduledPteOperation struct {
	AppOperation
	Parameters *ScheduledPteOperationParameters `json:"parameters,omitempty"`
}

// ScheduledPteOperationListResponse is the list wrapper returned by the
// `apps/scheduledPteOperations` endpoint.
type ScheduledPteOperationListResponse struct {
	Value []ScheduledPteOperation `json:"value"`
}

// App is an app installed on an environment, as returned by the `apps` endpoint.
//
// The published documentation names the identity field `appId`, but the live API returns
// it as `id` (matching the older environment app endpoints). Both are decoded and read
// through Identity so either shape resolves.
type App struct {
	ID                         string `json:"id"`
	AppID                      string `json:"appId"`
	Name                       string `json:"name"`
	Publisher                  string `json:"publisher"`
	Version                    string `json:"version"`
	State                      string `json:"state"`
	LastOperationID            string `json:"lastOperationId,omitempty"`
	LastUpdateAttemptResult    string `json:"lastUpdateAttemptResult,omitempty"`
	LastUninstallOperationID   string `json:"lastUninstallOperationId,omitempty"`
	LastUninstallAttemptResult string `json:"lastUninstallAttemptResult,omitempty"`
	AppType                    string `json:"appType,omitempty"`
	CanBeUninstalled           bool   `json:"canBeUninstalled,omitempty"`
}

// Identity returns the app's stable ID regardless of which field name the API used.
func (a *App) Identity() string {
	if a.ID != "" {
		return a.ID
	}
	return a.AppID
}

// AppListResponse is the list wrapper returned by the `apps` endpoint.
type AppListResponse struct {
	Value []App `json:"value"`
}

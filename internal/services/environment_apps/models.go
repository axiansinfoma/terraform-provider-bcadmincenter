// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package environmentapps

// App represents a Business Central app installed in an environment.
type App struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Publisher   string `json:"publisher"`
	Version     string `json:"version"`
	Status      string `json:"status"`
	PublishedAs string `json:"publishedAs"`
}

// AppListResponse represents the response when listing apps for an environment.
type AppListResponse struct {
	Value []App `json:"value"`
}

// InstallAppRequest represents the request body for installing an app.
type InstallAppRequest struct {
	TargetVersion                     string `json:"targetVersion,omitempty"`
	AllowPreviewVersion               bool   `json:"allowPreviewVersion"`
	InstallOrUpdateNeededDependencies bool   `json:"installOrUpdateNeededDependencies"`
	AcceptIsvEula                     bool   `json:"acceptIsvEula,omitempty"`
	LanguageID                        string `json:"languageId,omitempty"`
}

// UpdateAppRequest represents the request body for updating an app.
type UpdateAppRequest struct {
	TargetVersion                     string `json:"targetVersion,omitempty"`
	AllowPreviewVersion               bool   `json:"allowPreviewVersion"`
	InstallOrUpdateNeededDependencies bool   `json:"installOrUpdateNeededDependencies"`
	UseEnvironmentUpdateWindow        bool   `json:"useEnvironmentUpdateWindow"`
}

// UninstallAppRequest represents the request body for uninstalling an app.
type UninstallAppRequest struct {
	DoNotSaveData              bool `json:"doNotSaveData"`
	UninstallDependents        bool `json:"uninstallDependents"`
	UseEnvironmentUpdateWindow bool `json:"useEnvironmentUpdateWindow"`
}

// Operation represents an asynchronous operation returned by the app lifecycle API.
type Operation struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	Status       string `json:"status"`
	ErrorMessage string `json:"errorMessage,omitempty"`
}

// CancelUpdateRequest represents the request body for cancelling a scheduled app update.
type CancelUpdateRequest struct {
	ScheduledOperationID string `json:"ScheduledOperationId"`
}

// AppOperation represents a single operation entry returned by the app operations endpoint.
type AppOperation struct {
	ID            string `json:"id"`
	CreatedOn     string `json:"createdOn"`
	StartedOn     string `json:"startedOn,omitempty"`
	CompletedOn   string `json:"completedOn,omitempty"`
	Status        string `json:"status"`
	SourceVersion string `json:"sourceVersion,omitempty"`
	TargetVersion string `json:"targetVersion,omitempty"`
	Type          string `json:"type"`
	ErrorMessage  string `json:"errorMessage,omitempty"`
}

// AppOperationsResponse represents the paginated list returned by the app operations endpoint.
type AppOperationsResponse struct {
	Value []AppOperation `json:"value"`
}

// OperationStatus constants for app operations.
const (
	OperationStatusScheduled = "scheduled"
	OperationStatusQueued    = "queued"
	OperationStatusRunning   = "running"
	OperationStatusSucceeded = "succeeded"
	OperationStatusFailed    = "failed"
	OperationStatusCancelled = "cancelled"
)

// App status constants.
const (
	AppStatusInstalled     = "installed"
	AppStatusInstalling    = "installing"
	AppStatusUninstalling  = "uninstalling"
	AppStatusUpdatePending = "updatePending"
	AppStatusInstallFailed = "installFailed"
	AppStatusUpdateFailed  = "updateFailed"
)

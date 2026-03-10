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
	IgnoreUpdateWindow                bool   `json:"ignoreUpdateWindow,omitempty"`
}

// UninstallAppRequest represents the request body for uninstalling an app.
type UninstallAppRequest struct {
	DoNotSaveData       bool `json:"doNotSaveData"`
	UninstallDependents bool `json:"uninstallDependents"`
	IgnoreUpdateWindow  bool `json:"ignoreUpdateWindow,omitempty"`
}

// Operation represents an asynchronous operation returned by the app lifecycle API.
type Operation struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	Status       string `json:"status"`
	ErrorMessage string `json:"errorMessage,omitempty"`
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

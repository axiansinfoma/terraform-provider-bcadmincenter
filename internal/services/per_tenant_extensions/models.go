// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package pertenantextensions

// Company represents a Business Central company returned by the Automation API.
type Company struct {
	ID          string `json:"id"`
	SystemID    string `json:"systemId"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

// CompanyListResponse is the OData list response for companies.
type CompanyListResponse struct {
	Value []Company `json:"value"`
}

// ExtensionUploadRequest is the request body for creating an extension upload record.
type ExtensionUploadRequest struct {
	Schedule       string `json:"schedule"`
	SchemaSyncMode string `json:"schemaSyncMode"`
}

// ExtensionUpload is the response body from creating an extension upload record.
type ExtensionUpload struct {
	SystemID string `json:"systemId"`
	Schedule string `json:"schedule,omitempty"`
}

// ExtensionDeploymentStatus represents a single entry from the extensionDeploymentStatus collection.
type ExtensionDeploymentStatus struct {
	OperationID   string `json:"operationID"`
	Name          string `json:"name"`
	Publisher     string `json:"publisher"`
	OperationType string `json:"operationType"`
	Status        string `json:"status"`
	Schedule      string `json:"schedule"`
	AppVersion    string `json:"appVersion"`
	StartedOn     string `json:"startedOn"`
}

// ExtensionDeploymentStatusListResponse is the OData list response for deployment statuses.
type ExtensionDeploymentStatusListResponse struct {
	Value []ExtensionDeploymentStatus `json:"value"`
}

// Extension represents an installed extension from the extensions collection.
type Extension struct {
	PackageID       string `json:"packageId"`
	ID              string `json:"id"`
	DisplayName     string `json:"displayName"`
	Publisher       string `json:"publisher"`
	VersionMajor    int    `json:"versionMajor"`
	VersionMinor    int    `json:"versionMinor"`
	VersionBuild    int    `json:"versionBuild"`
	VersionRevision int    `json:"versionRevision"`
	IsInstalled     bool   `json:"isInstalled"`
	PublishedAs     string `json:"publishedAs"`
}

// ExtensionListResponse is the OData list response for extensions.
type ExtensionListResponse struct {
	Value []Extension `json:"value"`
}

// Deployment status terminal states.
const (
	DeploymentStatusCompleted = "Completed"
	DeploymentStatusFailed    = "Failed"
)

// Default schedule and schema sync mode values.
const (
	DefaultSchedule       = "Current version"
	DefaultSchemaSyncMode = "Add"
)

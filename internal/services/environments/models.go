// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package environments

import "time"

// Environment represents a Business Central environment from the Admin Center API.
type Environment struct {
	AADTenantID           string           `json:"aadTenantId"`
	ApplicationFamily     string           `json:"applicationFamily"`
	Type                  string           `json:"type"`
	Name                  string           `json:"name"`
	CountryCode           string           `json:"countryCode"`
	Status                string           `json:"status"`
	WebClientLoginURL     string           `json:"webClientLoginUrl"`
	WebServiceURL         string           `json:"webServiceUrl,omitempty"`
	AppInsightsKey        string           `json:"appInsightsKey,omitempty"`
	RingName              string           `json:"ringName,omitempty"`
	ApplicationVersion    string           `json:"applicationVersion,omitempty"`
	PlatformVersion       string           `json:"platformVersion,omitempty"`
	LocationOptions       []LocationOption `json:"locationOptions,omitempty"`
	SourceEnvironmentName string           `json:"sourceEnvironmentName,omitempty"`
	SourceEnvironmentType string           `json:"sourceEnvironmentType,omitempty"`
}

// LocationOption represents available Azure regions for the environment.
type LocationOption struct {
	Type   string `json:"type"`
	Code   string `json:"code"`
	Name   string `json:"name"`
	Region string `json:"region,omitempty"`
}

// CreateEnvironmentRequest represents the request body for creating a new environment.
type CreateEnvironmentRequest struct {
	EnvironmentType    string `json:"environmentType"`
	Name               string `json:"name"`
	CountryCode        string `json:"countryCode"`
	RingName           string `json:"ringName,omitempty"`
	ApplicationVersion string `json:"applicationVersion,omitempty"` // set only when specified by user
	AzureRegion        string `json:"azureRegion,omitempty"`
}

// UpdateEnvironmentRequest represents the request body for updating an environment.
type UpdateEnvironmentRequest struct {
	// Currently the BC Admin Center API has limited update capabilities.
	// Most changes require recreating the environment.
}

// EnvironmentListResponse represents the response when listing environments.
type EnvironmentListResponse struct {
	Value []Environment `json:"value"`
}

// Operation represents an asynchronous operation in the Admin Center API.
type Operation struct {
	ID                     string    `json:"id"`
	Type                   string    `json:"type"`
	AADTenantID            string    `json:"aadTenantId"`
	ApplicationFamily      string    `json:"applicationFamily"`
	ProductFamily          string    `json:"productFamily"` // API returns this instead of applicationFamily
	Status                 string    `json:"status"`
	ErrorMessage           string    `json:"errorMessage,omitempty"`
	CreatedOn              time.Time `json:"createdOn"`
	LastModified           time.Time `json:"lastModified"`
	SourceEnvironment      string    `json:"sourceEnvironment,omitempty"`
	DestinationEnvironment string    `json:"destinationEnvironment,omitempty"`
	EnvironmentName        string    `json:"environmentName,omitempty"` // Alternative field for environment name
	EnvironmentType        string    `json:"environmentType,omitempty"` // Type of the environment
}

// OperationStatus represents the possible states of an operation.
// Note: API returns lowercase status values.
const (
	OperationStatusQueued    = "queued"
	OperationStatusRunning   = "running"
	OperationStatusSucceeded = "succeeded"
	OperationStatusFailed    = "failed"
	OperationStatusCancelled = "cancelled"
)

// EnvironmentStatus represents the possible states of an environment.
const (
	EnvironmentStatusActive   = "Active"
	EnvironmentStatusCreating = "Creating"
	EnvironmentStatusDeleting = "Deleting"
	EnvironmentStatusFailed   = "Failed"
)

// EnvironmentType represents the type of Business Central environment.
const (
	EnvironmentTypeProduction = "Production"
	EnvironmentTypeSandbox    = "Sandbox"
)

// EnvironmentUpdatesResponse represents the response when listing available updates.
type EnvironmentUpdatesResponse struct {
	Value []EnvironmentUpdate `json:"value"`
}

// EnvironmentUpdate represents a single available update entry for an environment.
type EnvironmentUpdate struct {
	TargetVersion     string                 `json:"targetVersion"`
	Available         bool                   `json:"available"`
	Selected          bool                   `json:"selected"`
	UpdateStatus      string                 `json:"updateStatus,omitempty"`
	ScheduleDetails   *UpdateScheduleDetails `json:"scheduleDetails,omitempty"`
	TargetVersionType string                 `json:"targetVersionType,omitempty"`
}

// UpdateScheduleDetails holds scheduling information for an environment update.
type UpdateScheduleDetails struct {
	LatestSelectableDateTime string `json:"latestSelectableDateTime,omitempty"`
	SelectedDateTime         string `json:"selectedDateTime,omitempty"`
	IgnoreUpdateWindow       bool   `json:"ignoreUpdateWindow,omitempty"`
	RolloutStatus            string `json:"rolloutStatus,omitempty"`
}

// SelectUpdateRequest is used by SelectUpdateVersion and ScheduleUpdateVersion.
type SelectUpdateRequest struct {
	Selected        bool                   `json:"selected"`
	ScheduleDetails *UpdateScheduleDetails `json:"scheduleDetails,omitempty"`
}

// UpdateScheduleDetailsRequest is used by UpdateScheduleDetails (no "selected" field).
type UpdateScheduleDetailsRequest struct {
	ScheduleDetails *UpdateScheduleDetails `json:"scheduleDetails"`
}

// UpdateStatus constants for EnvironmentUpdate.updateStatus field.
const (
	UpdateStatusScheduled = "scheduled"
	UpdateStatusRunning   = "running"
	UpdateStatusFailed    = "failed"
)

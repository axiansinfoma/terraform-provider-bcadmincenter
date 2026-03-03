// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package environments

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/client"
	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/utils"
)

// Service handles environment-related operations for the Business Central Admin Center API.
type Service struct {
	client *client.Client
}

// NewService creates a new environment service.
func NewService(c *client.Client) *Service {
	return &Service{
		client: c,
	}
}

// List retrieves all environments for the specified application family.
func (s *Service) List(ctx context.Context, applicationFamily string) ([]Environment, error) {
	path := fmt.Sprintf("applications/%s/environments", applicationFamily)

	resp, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to list environments: %w", err)
	}
	defer resp.Body.Close()

	var envList EnvironmentListResponse
	if err := json.NewDecoder(resp.Body).Decode(&envList); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return envList.Value, nil
}

// Get retrieves a specific environment by name.
func (s *Service) Get(ctx context.Context, applicationFamily, environmentName string) (*Environment, error) {
	path := fmt.Sprintf("applications/%s/environments/%s", applicationFamily, environmentName)

	resp, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get environment: %w", err)
	}
	defer resp.Body.Close()

	var env Environment
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &env, nil
}

// Create creates a new Business Central environment.
func (s *Service) Create(ctx context.Context, applicationFamily string, req *CreateEnvironmentRequest) (*Operation, error) {
	// The API uses PUT with the environment name in the URL path.
	path := fmt.Sprintf("applications/%s/environments/%s", applicationFamily, req.Name)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := s.client.Put(ctx, path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create environment: %w", err)
	}
	defer resp.Body.Close()

	// The API returns a 202 Accepted with an operation in the response.
	if resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, utils.ReadResponseBody(resp.Body))
	}

	var operation Operation
	if err := json.NewDecoder(resp.Body).Decode(&operation); err != nil {
		return nil, fmt.Errorf("failed to decode operation response: %w", err)
	}

	return &operation, nil
}

// Delete deletes a Business Central environment.
func (s *Service) Delete(ctx context.Context, applicationFamily, environmentName string) (*Operation, error) {
	path := fmt.Sprintf("applications/%s/environments/%s", applicationFamily, environmentName)

	resp, err := s.client.Delete(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to delete environment: %w", err)
	}
	defer resp.Body.Close()

	// The API returns a 202 Accepted with an operation in the response.
	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusNoContent {
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, utils.ReadResponseBody(resp.Body))
	}

	// If 204 No Content, the environment was already deleted or doesn't exist.
	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	var operation Operation
	if err := json.NewDecoder(resp.Body).Decode(&operation); err != nil {
		return nil, fmt.Errorf("failed to decode operation response: %w", err)
	}

	return &operation, nil
}

// GetOperation retrieves the status of an operation.
// Uses the environment-specific operations endpoint.
func (s *Service) GetOperation(ctx context.Context, applicationFamily, environmentName, operationID string) (*Operation, error) {
	// GET /admin/{version}/applications/{applicationFamily}/environments/{environmentName}/operations/{id}.
	path := fmt.Sprintf("applications/%s/environments/%s/operations/%s", applicationFamily, environmentName, operationID)

	resp, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get operation: %w", err)
	}
	defer resp.Body.Close()

	var operation Operation
	if err := json.NewDecoder(resp.Body).Decode(&operation); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &operation, nil
}

// WaitForOperation polls an operation until it completes or times out.
func (s *Service) WaitForOperation(ctx context.Context, applicationFamily, environmentName, operationID string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// Check immediately first.
	operation, err := s.GetOperation(ctx, applicationFamily, environmentName, operationID)
	if err != nil {
		// For delete operations, if the environment is not found, consider it success.
		if isEnvironmentNotFoundError(err) {
			return nil
		}
		return fmt.Errorf("failed to check operation status: %w", err)
	}

	// Log initial operation status.
	fmt.Printf("[DEBUG] Initial operation status: %s (ID: %s)\n", operation.Status, operation.ID)

	if operation.Status == OperationStatusSucceeded {
		fmt.Printf("[DEBUG] Operation already succeeded\n")
		return nil
	}
	if operation.Status == OperationStatusFailed {
		return fmt.Errorf("operation failed: %s", operation.ErrorMessage)
	}
	if operation.Status == OperationStatusCancelled {
		return fmt.Errorf("operation was cancelled")
	}

	// Then poll at intervals.
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("operation timeout after %v", timeout)
		case <-ticker.C:
			operation, err := s.GetOperation(ctx, applicationFamily, environmentName, operationID)
			if err != nil {
				// For delete operations, if the environment is not found, consider it success.
				if isEnvironmentNotFoundError(err) {
					return nil
				}
				return fmt.Errorf("failed to check operation status: %w", err)
			}

			// Log polling status.
			fmt.Printf("[DEBUG] Polling operation status: %s (ID: %s)\n", operation.Status, operation.ID)

			switch operation.Status {
			case OperationStatusSucceeded:
				fmt.Printf("[DEBUG] Operation succeeded\n")
				return nil
			case OperationStatusFailed:
				return fmt.Errorf("operation failed: %s", operation.ErrorMessage)
			case OperationStatusCancelled:
				return fmt.Errorf("operation was cancelled")
			case OperationStatusQueued, OperationStatusRunning:
				// Continue polling.
				continue
			default:
				return fmt.Errorf("unknown operation status: %s", operation.Status)
			}
		}
	}
}

// isEnvironmentNotFoundError checks if an error is an EnvironmentNotFound error.
// This is useful for delete operations where the environment no longer exists.
func isEnvironmentNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	var apiErr *client.AdminCenterError
	if errors.As(err, &apiErr) {
		return apiErr.Code == "EnvironmentNotFound"
	}

	return strings.Contains(err.Error(), "EnvironmentNotFound")
}

// GetUpdates returns available and selected updates for an environment.
// GET /admin/{apiVersion}/applications/{applicationFamily}/environments/{environmentName}/updates
func (s *Service) GetUpdates(ctx context.Context, applicationFamily, environmentName string) ([]EnvironmentUpdate, error) {
	path := fmt.Sprintf("applications/%s/environments/%s/updates", applicationFamily, environmentName)

	resp, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get updates: %w", err)
	}
	defer resp.Body.Close()

	var updates EnvironmentUpdatesResponse
	if err := json.NewDecoder(resp.Body).Decode(&updates); err != nil {
		return nil, fmt.Errorf("failed to decode updates response: %w", err)
	}

	return updates.Value, nil
}

// patchUpdate is a shared helper that sends a PATCH request to the updates endpoint.
func (s *Service) patchUpdate(ctx context.Context, applicationFamily, environmentName, targetVersion string, body interface{}) error {
	path := fmt.Sprintf("applications/%s/environments/%s/updates/%s", applicationFamily, environmentName, targetVersion)

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := s.client.Patch(ctx, path, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to patch update: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, utils.ReadResponseBody(resp.Body))
	}

	return nil
}

// SelectUpdateVersion schedules an upgrade to the target version in the next update window.
// Used by the bcadmincenter_environment resource (application_version change).
// PATCH /admin/{apiVersion}/applications/{applicationFamily}/environments/{environmentName}/updates/{targetVersion}
// Body: {"selected": true, "scheduleDetails": {"ignoreUpdateWindow": <bool>}}
//
// The API rejects re-selection when the existing update entry holds a past selectedDateTime
// ("EntityValidationFailed: Update currently has selected date time in the past").
// To handle this, we first send a PATCH that explicitly nulls out selectedDateTime, then
// send the select request. The first step is best-effort — if it fails (e.g. the entry
// doesn't exist yet) we proceed with the selection.
func (s *Service) SelectUpdateVersion(ctx context.Context, applicationFamily, environmentName, targetVersion string, ignoreUpdateWindow bool) error {
	// Step 1: clear any stored selectedDateTime so the API doesn't reject the select
	// because of a past datetime from a prior schedule.
	clearReq := map[string]interface{}{
		"scheduleDetails": map[string]interface{}{
			"selectedDateTime":   nil,
			"ignoreUpdateWindow": ignoreUpdateWindow,
		},
	}
	if err := s.patchUpdate(ctx, applicationFamily, environmentName, targetVersion, clearReq); err != nil {
		// Best-effort: if clearing fails (e.g. no existing entry), proceed with selection.
		// Log the error for observability without blocking the upgrade.
		fmt.Printf("[WARN] SelectUpdateVersion: failed to clear selectedDateTime for %s/%s/%s: %v; proceeding with select\n",
			applicationFamily, environmentName, targetVersion, err)
	}

	// Step 2: select the version.
	req := SelectUpdateRequest{
		Selected: true,
		ScheduleDetails: &UpdateScheduleDetails{
			IgnoreUpdateWindow: ignoreUpdateWindow,
		},
	}
	return s.patchUpdate(ctx, applicationFamily, environmentName, targetVersion, req)
}

// ScheduleUpdateVersion schedules an upgrade with an explicit datetime.
// Used by the bcadmincenter_environment_update_schedule resource.
// PATCH /admin/{apiVersion}/applications/{applicationFamily}/environments/{environmentName}/updates/{targetVersion}
// Body: {"selected": true, "scheduleDetails": {"selectedDateTime": <datetime>, "ignoreUpdateWindow": <bool>}}
func (s *Service) ScheduleUpdateVersion(ctx context.Context, applicationFamily, environmentName, targetVersion, scheduledDateTime string, ignoreUpdateWindow bool) error {
	scheduleDetails := &UpdateScheduleDetails{
		IgnoreUpdateWindow: ignoreUpdateWindow,
	}
	if scheduledDateTime != "" {
		scheduleDetails.SelectedDateTime = scheduledDateTime
	}
	req := SelectUpdateRequest{
		Selected:        true,
		ScheduleDetails: scheduleDetails,
	}
	return s.patchUpdate(ctx, applicationFamily, environmentName, targetVersion, req)
}

// UpdateScheduleDetails updates scheduleDetails for an already-selected version without reselecting.
// Used when only scheduled_datetime or ignore_update_window changes on the update_schedule resource.
// PATCH /admin/{apiVersion}/applications/{applicationFamily}/environments/{environmentName}/updates/{targetVersion}
// Body: {"scheduleDetails": {"selectedDateTime": <datetime>, "ignoreUpdateWindow": <bool>}}
func (s *Service) UpdateScheduleDetails(ctx context.Context, applicationFamily, environmentName, targetVersion, scheduledDateTime string, ignoreUpdateWindow bool) error {
	scheduleDetails := &UpdateScheduleDetails{
		IgnoreUpdateWindow: ignoreUpdateWindow,
	}
	if scheduledDateTime != "" {
		scheduleDetails.SelectedDateTime = scheduledDateTime
	}
	req := UpdateScheduleDetailsRequest{
		ScheduleDetails: scheduleDetails,
	}
	return s.patchUpdate(ctx, applicationFamily, environmentName, targetVersion, req)
}

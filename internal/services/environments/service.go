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
	"github.com/hashicorp/terraform-plugin-log/tflog"
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
	path := client.BuildPath("applications", applicationFamily, "environments")

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
	path := client.BuildPath("applications", applicationFamily, "environments", environmentName)

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
	path := client.BuildPath("applications", applicationFamily, "environments", req.Name)

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
	path := client.BuildPath("applications", applicationFamily, "environments", environmentName)

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
	path := client.BuildPath("applications", applicationFamily, "environments", environmentName, "operations", operationID)

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

	tflog.Debug(ctx, "Initial operation status", map[string]interface{}{
		"status":       operation.Status,
		"operation_id": operation.ID,
	})

	if done, err := classifyOperation(operation); done {
		return err
	}

	// Then poll at intervals.
	for {
		select {
		case <-ctx.Done():
			return operationWaitError(ctx, timeout)
		case <-ticker.C:
			operation, err := s.GetOperation(ctx, applicationFamily, environmentName, operationID)
			if err != nil {
				// For delete operations, if the environment is not found, consider it success.
				if isEnvironmentNotFoundError(err) {
					return nil
				}
				return fmt.Errorf("failed to check operation status: %w", err)
			}

			tflog.Debug(ctx, "Polling operation status", map[string]interface{}{
				"status":       operation.Status,
				"operation_id": operation.ID,
			})

			if done, err := classifyOperation(operation); done {
				return err
			}
			// Queued, running, or a status this provider does not recognise: keep
			// polling until the deadline. Failing on an unrecognised status aborted
			// long-running operations that were progressing normally.
		}
	}
}

// classifyOperation reports whether an operation has reached a terminal state, and with
// what outcome. Statuses are matched case-insensitively and across both spellings of
// "cancelled", because the API is inconsistent about each.
func classifyOperation(operation *Operation) (done bool, err error) {
	switch {
	case utils.StatusIs(operation.Status, OperationStatusSucceeded):
		return true, nil
	case utils.StatusIs(operation.Status, OperationStatusFailed):
		return true, fmt.Errorf("operation failed: %s", operation.ErrorMessage)
	case utils.StatusIs(operation.Status, OperationStatusCancelled, OperationStatusCanceled):
		return true, fmt.Errorf("operation was cancelled")
	default:
		return false, nil
	}
}

// operationWaitError distinguishes a genuine timeout from the user interrupting the run,
// which otherwise reported a misleading "operation timeout after 1h0m0s" after a Ctrl-C.
func operationWaitError(ctx context.Context, timeout time.Duration) error {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return fmt.Errorf("operation timeout after %v", timeout)
	}
	return fmt.Errorf("stopped waiting for operation: %w", ctx.Err())
}

// isEnvironmentNotFoundError checks if an error is an EnvironmentNotFound error.
// This is useful for delete operations where the environment no longer exists.
func isEnvironmentNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	// HTTP 404 is authoritative. The Admin Center also returns EnvironmentNotFound on
	// some non-404 statuses, so both checks are needed.
	if client.IsNotFound(err) {
		return true
	}

	var apiErr *client.AdminCenterError
	if errors.As(err, &apiErr) {
		return strings.EqualFold(apiErr.Code, "EnvironmentNotFound")
	}

	return strings.Contains(err.Error(), "EnvironmentNotFound")
}

// GetUpdates returns available and selected updates for an environment.
// Calls GET /admin/{apiVersion}/applications/{applicationFamily}/environments/{environmentName}/updates.
func (s *Service) GetUpdates(ctx context.Context, applicationFamily, environmentName string) ([]EnvironmentUpdate, error) {
	path := client.BuildPath("applications", applicationFamily, "environments", environmentName, "updates")

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
	path := client.BuildPath("applications", applicationFamily, "environments", environmentName, "updates", targetVersion)

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
// To handle this, we first attempt a plain select (no selectedDateTime). If the API rejects
// it due to a stale past datetime, we retry supplying a fresh valid selectedDateTime — set to
// now+1h, capped to latestSelectableDateTime (fetched from the updates list). This preserves
// the natural "next update window" behaviour on first selection while recovering cleanly on
// re-selection after a previously scheduled upgrade that never ran.
// NOTE: the API does NOT support selected:false (deselect); that returns EntityValidationFailed.
func (s *Service) SelectUpdateVersion(ctx context.Context, applicationFamily, environmentName, targetVersion string, ignoreUpdateWindow bool) error {
	req := SelectUpdateRequest{
		Selected: true,
		ScheduleDetails: &UpdateScheduleDetails{
			IgnoreUpdateWindow: ignoreUpdateWindow,
		},
	}
	err := s.patchUpdate(ctx, applicationFamily, environmentName, targetVersion, req)
	if err == nil {
		return nil
	}

	// If the API rejected because of a stale past selectedDateTime, retry with a fresh datetime.
	if !isPastSelectedDateTimeError(err) {
		return err
	}

	tflog.Warn(ctx, "SelectUpdateVersion: stale past selectedDateTime detected; retrying with refreshed datetime", map[string]interface{}{
		"application_family": applicationFamily,
		"environment_name":   environmentName,
		"target_version":     targetVersion,
	})

	safeDateTime := time.Now().UTC().Add(1 * time.Hour).Format(time.RFC3339)
	// Cap to latestSelectableDateTime if the API provides one for this version.
	if updates, getErr := s.GetUpdates(ctx, applicationFamily, environmentName); getErr == nil {
		for _, u := range updates {
			if u.TargetVersion == targetVersion && u.ScheduleDetails != nil && u.ScheduleDetails.LatestSelectableDateTime != "" {
				if latest, parseErr := time.Parse(time.RFC3339, u.ScheduleDetails.LatestSelectableDateTime); parseErr == nil {
					candidate := time.Now().UTC().Add(1 * time.Hour)
					if candidate.After(latest) {
						safeDateTime = latest.Add(-1 * time.Minute).Format(time.RFC3339)
					}
				}
				break
			}
		}
	}

	req.ScheduleDetails.SelectedDateTime = safeDateTime
	return s.patchUpdate(ctx, applicationFamily, environmentName, targetVersion, req)
}

// isPastSelectedDateTimeError reports whether err is the "selected date time in the past" API error.
func isPastSelectedDateTimeError(err error) bool {
	if err == nil {
		return false
	}
	var apiErr *client.AdminCenterError
	if errors.As(err, &apiErr) {
		return apiErr.Code == "EntityValidationFailed" &&
			strings.Contains(apiErr.Message, "selected date time in the past")
	}
	return strings.Contains(err.Error(), "selected date time in the past")
}

// ScheduleUpdateVersion schedules an upgrade with an explicit datetime.
// Used by the bcadmincenter_environment_update_schedule resource.
// PATCH /admin/{apiVersion}/applications/{applicationFamily}/environments/{environmentName}/updates/{targetVersion}
// Body: {"selected": true, "scheduleDetails": {"selectedDateTime": <datetime>, "ignoreUpdateWindow": <bool>}}
//
// Same deselect-first strategy as SelectUpdateVersion to handle past-datetime state.
func (s *Service) ScheduleUpdateVersion(ctx context.Context, applicationFamily, environmentName, targetVersion, scheduledDateTime string, ignoreUpdateWindow bool) error {
	// Step 1: deselect to clear any past selectedDateTime state (best-effort).
	deselect := SelectUpdateRequest{Selected: false}
	if err := s.patchUpdate(ctx, applicationFamily, environmentName, targetVersion, deselect); err != nil {
		tflog.Warn(ctx, "ScheduleUpdateVersion: deselect failed; proceeding with select", map[string]interface{}{
			"application_family": applicationFamily,
			"environment_name":   environmentName,
			"target_version":     targetVersion,
			"error":              err.Error(),
		})
	}

	// Step 2: select the version with the desired schedule.
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
// Calls PATCH /admin/{apiVersion}/applications/{applicationFamily}/environments/{environmentName}/updates/{targetVersion}
// with body {"scheduleDetails": {"selectedDateTime": <datetime>, "ignoreUpdateWindow": <bool>}}.
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

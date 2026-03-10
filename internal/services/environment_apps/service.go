// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package environmentapps

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

// Service handles app lifecycle operations for the Business Central Admin Center API.
type Service struct {
	client *client.Client
}

// NewService creates a new environment app service.
func NewService(c *client.Client) *Service {
	return &Service{
		client: c,
	}
}

// IsAlreadyScheduledError reports whether err is an API response indicating that
// an update for the app has already been scheduled for the environment.  This
// happens when a previous deferred apply registered the update in the BC
// update queue but the state was not fully persisted, causing a second apply to
// attempt the same update.  The update is already in progress as desired, so
// callers should treat this as a deferred success rather than a hard error.
func IsAlreadyScheduledError(err error) bool {
	var apiErr *client.AdminCenterError
	if !errors.As(err, &apiErr) {
		return false
	}
	return apiErr.Code == "EntityValidationFailed" &&
		strings.Contains(apiErr.Message, "already been scheduled")
}

// IsCancelNotAllowedError reports whether err is a BC API error indicating that
// the scheduled update cannot be cancelled — typically because the operation has
// already transitioned to a running or terminal state.  Any AdminCenterError
// returned by the cancel endpoint is treated as "not cancellable", while
// lower-level errors (network, timeout) are not matched so they propagate normally.
func IsCancelNotAllowedError(err error) bool {
	var apiErr *client.AdminCenterError
	return errors.As(err, &apiErr)
}

// GetByID retrieves a specific installed app by its ID.
// Returns (nil, nil) when the app is not found (not installed).
func (s *Service) GetByID(ctx context.Context, applicationFamily, environmentName, appID string) (*App, error) {
	path := fmt.Sprintf("applications/%s/environments/%s/apps", applicationFamily, environmentName)

	resp, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to list apps: %w", err)
	}
	defer resp.Body.Close()

	var appList AppListResponse
	if err := json.NewDecoder(resp.Body).Decode(&appList); err != nil {
		return nil, fmt.Errorf("failed to decode app list response: %w", err)
	}

	for i := range appList.Value {
		if appList.Value[i].ID == appID {
			return &appList.Value[i], nil
		}
	}

	// App not found — not installed.
	return nil, nil
}

// Install installs an app into the environment.
// Returns the async operation to poll.
func (s *Service) Install(ctx context.Context, applicationFamily, environmentName, appID string, req *InstallAppRequest) (*Operation, error) {
	path := fmt.Sprintf("applications/%s/environments/%s/apps/%s/install", applicationFamily, environmentName, appID)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal install request: %w", err)
	}

	resp, err := s.client.Post(ctx, path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to install app: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, utils.ReadResponseBody(resp.Body))
	}

	var operation Operation
	if err := json.NewDecoder(resp.Body).Decode(&operation); err != nil {
		return nil, fmt.Errorf("failed to decode install operation response: %w", err)
	}

	return &operation, nil
}

// Update updates an installed app to a new version.
// Returns the async operation to poll.
func (s *Service) Update(ctx context.Context, applicationFamily, environmentName, appID string, req *UpdateAppRequest) (*Operation, error) {
	path := fmt.Sprintf("applications/%s/environments/%s/apps/%s/update", applicationFamily, environmentName, appID)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal update request: %w", err)
	}

	resp, err := s.client.Post(ctx, path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to update app: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, utils.ReadResponseBody(resp.Body))
	}

	var operation Operation
	if err := json.NewDecoder(resp.Body).Decode(&operation); err != nil {
		return nil, fmt.Errorf("failed to decode update operation response: %w", err)
	}

	return &operation, nil
}

// Uninstall uninstalls an app from the environment.
// Returns the async operation to poll.
func (s *Service) Uninstall(ctx context.Context, applicationFamily, environmentName, appID string, req *UninstallAppRequest) (*Operation, error) {
	path := fmt.Sprintf("applications/%s/environments/%s/apps/%s/uninstall", applicationFamily, environmentName, appID)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal uninstall request: %w", err)
	}

	resp, err := s.client.Post(ctx, path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to uninstall app: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, utils.ReadResponseBody(resp.Body))
	}

	var operation Operation
	if err := json.NewDecoder(resp.Body).Decode(&operation); err != nil {
		return nil, fmt.Errorf("failed to decode uninstall operation response: %w", err)
	}

	return &operation, nil
}

// GetScheduledUpdateOperationID looks up the ID of the most recently created
// "update" operation in "scheduled" state for the given app.  This is needed
// when calling CancelUpdate because the BC API requires the scheduled operation
// ID in the request body.  Returns an error if no scheduled update operation is
// found or the API call fails.
func (s *Service) GetScheduledUpdateOperationID(ctx context.Context, applicationFamily, environmentName, appID string) (string, error) {
	path := fmt.Sprintf("applications/%s/environments/%s/apps/%s/operations", applicationFamily, environmentName, appID)

	resp, err := s.client.Get(ctx, path)
	if err != nil {
		return "", fmt.Errorf("failed to list app operations: %w", err)
	}
	defer resp.Body.Close()

	var ops AppOperationsResponse
	if err := json.NewDecoder(resp.Body).Decode(&ops); err != nil {
		return "", fmt.Errorf("failed to decode app operations response: %w", err)
	}

	// BC returns operations in reverse-chronological order; the first scheduled
	// update operation is the one we want to cancel.
	for _, op := range ops.Value {
		if strings.EqualFold(op.Status, "scheduled") && strings.EqualFold(op.Type, "update") {
			return op.ID, nil
		}
	}

	return "", fmt.Errorf("no scheduled update operation found for app %s", appID)
}

// CancelUpdate cancels a scheduled (deferred) app update for the environment.
// The BC API requires the ScheduledOperationId in the request body; obtain it
// from the operation returned when scheduling the update or via
// GetScheduledUpdateOperationID.  The BC API rejects the request with an
// AdminCenterError if the operation is already running or in a non-cancellable
// state; callers should check with IsCancelNotAllowedError to distinguish that
// from transient errors.
func (s *Service) CancelUpdate(ctx context.Context, applicationFamily, environmentName, appID, scheduledOperationID string) error {
	path := fmt.Sprintf("applications/%s/environments/%s/apps/%s/update/cancel", applicationFamily, environmentName, appID)

	cancelReq := CancelUpdateRequest{ScheduledOperationID: scheduledOperationID}
	body, err := json.Marshal(cancelReq)
	if err != nil {
		return fmt.Errorf("failed to marshal cancel request: %w", err)
	}

	resp, err := s.client.Post(ctx, path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to cancel app update: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, utils.ReadResponseBody(resp.Body))
	}

	return nil
}

// getOperation retrieves the current status of an async operation.
func (s *Service) getOperation(ctx context.Context, applicationFamily, environmentName, operationID string) (*Operation, error) {
	path := fmt.Sprintf("applications/%s/environments/%s/operations/%s", applicationFamily, environmentName, operationID)

	resp, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get operation: %w", err)
	}
	defer resp.Body.Close()

	var operation Operation
	if err := json.NewDecoder(resp.Body).Decode(&operation); err != nil {
		return nil, fmt.Errorf("failed to decode operation response: %w", err)
	}

	return &operation, nil
}

// WaitForOperation polls an operation until it completes or times out.
// When skipIfScheduled is true the function returns immediately (without error)
// if the operation is in the "scheduled" state, meaning BC has accepted the
// request but deferred execution to the environment's update window.  This
// mirrors the environment resource behaviour: the pending state is recorded and
// the next Terraform refresh will observe the final result.
// The returned bool is true when the operation was deferred (scheduled) and not
// yet executed; callers should preserve the intended target version in state
// rather than overwriting it with the API's currently-installed version.
func (s *Service) WaitForOperation(ctx context.Context, applicationFamily, environmentName, operationID string, timeout time.Duration, skipIfScheduled bool) (deferred bool, err error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// Check immediately first.
	operation, err := s.getOperation(ctx, applicationFamily, environmentName, operationID)
	if err != nil {
		return false, fmt.Errorf("failed to check operation status: %w", err)
	}

	tflog.Debug(ctx, "Initial app operation status", map[string]interface{}{"status": operation.Status, "operation_id": operation.ID})

	switch operation.Status {
	case OperationStatusSucceeded:
		return false, nil
	case OperationStatusFailed:
		return false, fmt.Errorf("operation failed: %s", operation.ErrorMessage)
	case OperationStatusCancelled:
		return false, fmt.Errorf("operation was cancelled")
	case OperationStatusScheduled:
		if skipIfScheduled {
			// Operation has been deferred to the update window — do not block.
			return true, nil
		}
	}

	// Then poll at intervals.
	for {
		select {
		case <-ctx.Done():
			return false, fmt.Errorf("operation timeout after %v", timeout)
		case <-ticker.C:
			operation, err := s.getOperation(ctx, applicationFamily, environmentName, operationID)
			if err != nil {
				return false, fmt.Errorf("failed to check operation status: %w", err)
			}

			tflog.Debug(ctx, "Polling app operation status", map[string]interface{}{"status": operation.Status, "operation_id": operation.ID})

			switch operation.Status {
			case OperationStatusSucceeded:
				return false, nil
			case OperationStatusFailed:
				return false, fmt.Errorf("operation failed: %s", operation.ErrorMessage)
			case OperationStatusCancelled:
				return false, fmt.Errorf("operation was cancelled")
			case OperationStatusScheduled:
				if skipIfScheduled {
					// Transitioned to scheduled mid-poll — deferred to update window.
					return true, nil
				}
				continue
			case OperationStatusQueued, OperationStatusRunning:
				// Continue polling.
				continue
			default:
				return false, fmt.Errorf("unknown operation status: %s", operation.Status)
			}
		}
	}
}

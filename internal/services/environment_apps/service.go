// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package environmentapps

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/client"
	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/utils"
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

	if resp.StatusCode != http.StatusAccepted {
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

	if resp.StatusCode != http.StatusAccepted {
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

	if resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, utils.ReadResponseBody(resp.Body))
	}

	var operation Operation
	if err := json.NewDecoder(resp.Body).Decode(&operation); err != nil {
		return nil, fmt.Errorf("failed to decode uninstall operation response: %w", err)
	}

	return &operation, nil
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
func (s *Service) WaitForOperation(ctx context.Context, applicationFamily, environmentName, operationID string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// Check immediately first.
	operation, err := s.getOperation(ctx, applicationFamily, environmentName, operationID)
	if err != nil {
		return fmt.Errorf("failed to check operation status: %w", err)
	}

	fmt.Printf("[DEBUG] Initial app operation status: %s (ID: %s)\n", operation.Status, operation.ID)

	switch operation.Status {
	case OperationStatusSucceeded:
		return nil
	case OperationStatusFailed:
		return fmt.Errorf("operation failed: %s", operation.ErrorMessage)
	case OperationStatusCancelled:
		return fmt.Errorf("operation was cancelled")
	}

	// Then poll at intervals.
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("operation timeout after %v", timeout)
		case <-ticker.C:
			operation, err := s.getOperation(ctx, applicationFamily, environmentName, operationID)
			if err != nil {
				return fmt.Errorf("failed to check operation status: %w", err)
			}

			fmt.Printf("[DEBUG] Polling app operation status: %s (ID: %s)\n", operation.Status, operation.ID)

			switch operation.Status {
			case OperationStatusSucceeded:
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

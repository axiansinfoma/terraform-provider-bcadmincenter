// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package pertenantextensions

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/client"
	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/utils"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// uploadTimeout bounds the single HTTP request that streams the .app package. It is
// deliberately generous because the endpoint accepts packages up to 50 MB.
const uploadTimeout = 15 * time.Minute

// Service handles per-tenant extension lifecycle operations via the Business Central
// Admin Center API. The PTE endpoints it uses were introduced in API version 2.29.
type Service struct {
	client *client.Client
}

// NewService creates a new per-tenant extension service.
func NewService(c *client.Client) *Service {
	return &Service{client: c}
}

// appsPath builds the base apps path for an environment.
func appsPath(applicationFamily, environmentName string) string {
	return client.BuildPath("applications", applicationFamily, "environments", environmentName, "apps")
}

// appPath returns the path for a single app under an environment, with any further
// segments appended. Every segment is escaped by client.BuildPath.
func appPath(applicationFamily, environmentName, appID string, extra ...string) string {
	segments := append([]string{"applications", applicationFamily, "environments", environmentName, "apps", appID}, extra...)
	return client.BuildPath(segments...)
}

// IsNotFoundError reports whether err is an API error indicating the targeted app or
// scheduled version does not exist. Callers use it to treat "already gone" as success.
func IsNotFoundError(err error) bool {
	// HTTP 404 is authoritative. The Admin Center also returns not-found codes on some
	// non-404 statuses, so both checks are needed.
	if client.IsNotFound(err) {
		return true
	}

	var apiErr *client.AdminCenterError
	if !errors.As(err, &apiErr) {
		return false
	}
	return strings.EqualFold(apiErr.Code, "ResourceDoesNotExist") ||
		strings.EqualFold(apiErr.Code, "NotFound") ||
		strings.Contains(strings.ToLower(apiErr.Message), "not installed on environment")
}

// buildPteInstallForm encodes req as a multipart/form-data body and returns the body
// together with the boundary-carrying content type.
func buildPteInstallForm(req *PteInstallRequest) (*bytes.Buffer, string, error) {
	if req == nil {
		return nil, "", fmt.Errorf("pte install request cannot be nil")
	}
	if len(req.Content) == 0 {
		return nil, "", fmt.Errorf("extension package is empty")
	}
	if len(req.Content) > MaxExtensionFileSize {
		return nil, "", fmt.Errorf("extension package is %d bytes, which exceeds the %d byte limit enforced by the pteInstall endpoint",
			len(req.Content), MaxExtensionFileSize)
	}

	fileName := req.FileName
	if fileName == "" {
		fileName = "extension.app"
	}
	if !strings.EqualFold(filepath.Ext(fileName), ".app") {
		return nil, "", fmt.Errorf("extension package file name %q must have the .app extension", fileName)
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("extensionFile", fileName)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create extension file part: %w", err)
	}
	if _, err := part.Write(req.Content); err != nil {
		return nil, "", fmt.Errorf("failed to write extension file part: %w", err)
	}

	fields := map[string]string{
		"deploymentSchedule":                NormalizeDeploymentSchedule(req.DeploymentSchedule),
		"syncMode":                          NormalizeSyncMode(req.SyncMode),
		"acceptIsvEula":                     strconv.FormatBool(req.AcceptIsvEula),
		"installOrUpdateNeededDependencies": strconv.FormatBool(req.InstallOrUpdateNeededDependencies),
	}
	if req.LanguageID != "" {
		fields["languageId"] = req.LanguageID
	}

	for name, value := range fields {
		if err := writer.WriteField(name, value); err != nil {
			return nil, "", fmt.Errorf("failed to write %s field: %w", name, err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, "", fmt.Errorf("failed to finalize multipart body: %w", err)
	}

	return &body, writer.FormDataContentType(), nil
}

// UploadAndInstall uploads a .app package and schedules its install or update via the
// `apps/pteInstall` endpoint. The returned operation carries the app ID and target
// version read from the uploaded package, so callers do not need to look them up.
func (s *Service) UploadAndInstall(ctx context.Context, applicationFamily, environmentName string, req *PteInstallRequest) (*AppOperation, error) {
	body, contentType, err := buildPteInstallForm(req)
	if err != nil {
		return nil, err
	}

	path := appsPath(applicationFamily, environmentName) + "/pteInstall"

	resp, err := s.client.PostMultipart(ctx, path, body, contentType, uploadTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to upload per-tenant extension: %w", err)
	}
	defer resp.Body.Close()

	var operation AppOperation
	if err := json.NewDecoder(resp.Body).Decode(&operation); err != nil {
		return nil, fmt.Errorf("failed to decode pteInstall response: %w", err)
	}

	if operation.ID == "" {
		return nil, fmt.Errorf("pteInstall response did not contain an operation id")
	}

	tflog.Debug(ctx, "Uploaded per-tenant extension", map[string]interface{}{
		"operation_id":   operation.ID,
		"app_id":         operation.AppID,
		"target_version": operation.TargetVersionValue(),
		"status":         operation.Status,
		"schedule_kind":  operation.ScheduleKind,
	})

	return &operation, nil
}

// GetApp returns the installed app with the given app ID.
// Returns (nil, nil) when the app is not installed on the environment.
func (s *Service) GetApp(ctx context.Context, applicationFamily, environmentName, appID string) (*App, error) {
	resp, err := s.client.Get(ctx, appsPath(applicationFamily, environmentName))
	if err != nil {
		return nil, fmt.Errorf("failed to list apps: %w", err)
	}
	defer resp.Body.Close()

	var list AppListResponse
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, fmt.Errorf("failed to decode app list response: %w", err)
	}

	for i := range list.Value {
		if strings.EqualFold(list.Value[i].Identity(), appID) {
			return &list.Value[i], nil
		}
	}

	return nil, nil
}

// GetOperation returns a single app operation. The API returns either a bare operation
// object or a list wrapper depending on the endpoint version, so both shapes are
// accepted — but a list is always searched for the requested ID rather than taking the
// first entry, because the endpoint returns the environment's whole operation history
// (in no particular order) whenever it does not narrow to a single operation.
//
// operationID must be non-empty: the API answers a request with an empty operation ID by
// returning that full history, whose first entry is an unrelated, long-completed
// operation. Treating it as the requested one would report a stale success.
func (s *Service) GetOperation(ctx context.Context, applicationFamily, environmentName, appID, operationID string) (*AppOperation, error) {
	if operationID == "" {
		return nil, fmt.Errorf("an operation id is required to look up an app operation")
	}

	path := appPath(applicationFamily, environmentName, appID, "operations", operationID)

	resp, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get app operation: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read app operation response: %w", err)
	}

	var list AppOperationListResponse
	if err := json.Unmarshal(raw, &list); err == nil && len(list.Value) > 0 {
		for i := range list.Value {
			if strings.EqualFold(list.Value[i].ID, operationID) {
				return &list.Value[i], nil
			}
		}
		return nil, fmt.Errorf("app operation %q was not found among the %d operations returned for app %s",
			operationID, len(list.Value), appID)
	}

	var operation AppOperation
	if err := json.Unmarshal(raw, &operation); err != nil {
		return nil, fmt.Errorf("failed to decode app operation response: %w", err)
	}
	if operation.ID == "" {
		return nil, fmt.Errorf("app operation %q was not found", operationID)
	}

	return &operation, nil
}

// WaitForOperation polls an app operation until it reaches a terminal state or the
// timeout elapses.
//
// scheduledIsTerminal controls how the "scheduled" status is treated, and must reflect
// what the caller actually requested. The API uses "scheduled" for two different things:
// a deployment genuinely deferred to a future window (unbounded, so it must not be
// waited on), and a transient queued state that immediate operations — an uninstall in
// particular — pass through before they start running. Treating the latter as terminal
// reports success while the operation has not even begun.
func (s *Service) WaitForOperation(ctx context.Context, applicationFamily, environmentName, appID, operationID string, timeout time.Duration, scheduledIsTerminal bool) (*AppOperation, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		operation, err := s.GetOperation(ctx, applicationFamily, environmentName, appID, operationID)
		if err != nil {
			return nil, fmt.Errorf("failed to poll app operation status: %w", err)
		}

		tflog.Debug(ctx, "Per-tenant extension operation status", map[string]interface{}{
			"operation_id": operation.ID,
			"status":       operation.Status,
			"type":         operation.Type,
		})

		switch {
		case strings.EqualFold(operation.Status, OperationStatusSucceeded):
			return operation, nil
		case strings.EqualFold(operation.Status, OperationStatusScheduled):
			if scheduledIsTerminal {
				// Genuinely deferred to a future deployment window — do not block on it.
				return operation, nil
			}
			// Transient queued state for an immediate operation; keep polling.
		case strings.EqualFold(operation.Status, OperationStatusFailed):
			return nil, fmt.Errorf("per-tenant extension operation %s failed: %s", operation.ID, operation.ErrorMessage)
		case utils.StatusIs(operation.Status, OperationStatusCanceled, OperationStatusCancelled):
			return nil, fmt.Errorf("per-tenant extension operation %s was canceled", operation.ID)
		case strings.EqualFold(operation.Status, OperationStatusSkipped):
			return nil, fmt.Errorf("per-tenant extension operation %s was skipped", operation.ID)
		}

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timed out after %v waiting for per-tenant extension operation %s", timeout, operationID)
		case <-ticker.C:
			// Continue polling.
		}
	}
}

// Uninstall uninstalls an app from the environment and returns the operation to poll.
func (s *Service) Uninstall(ctx context.Context, applicationFamily, environmentName, appID string, req *UninstallAppRequest) (*AppOperation, error) {
	path := appPath(applicationFamily, environmentName, appID, "uninstall")

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal uninstall request: %w", err)
	}

	resp, err := s.client.Post(ctx, path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to uninstall per-tenant extension: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("unexpected status code %d from uninstall endpoint", resp.StatusCode)
	}

	var operation AppOperation
	if err := json.NewDecoder(resp.Body).Decode(&operation); err != nil {
		return nil, fmt.Errorf("failed to decode uninstall operation response: %w", err)
	}

	return &operation, nil
}

// WaitForAppRemoval polls the apps list until the app is no longer present.
//
// A successful uninstall operation does not mean the app is gone: the API reports the
// operation as "succeeded" while the app lingers in an "uninstallPending" state for
// roughly another half minute. Without this wait, a destroy followed by an apply fails
// because the extension is still considered installed.
func (s *Service) WaitForAppRemoval(ctx context.Context, applicationFamily, environmentName, appID string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		app, err := s.GetApp(ctx, applicationFamily, environmentName, appID)
		if err != nil {
			return fmt.Errorf("failed to poll app removal: %w", err)
		}
		if app == nil {
			return nil
		}

		tflog.Debug(ctx, "Waiting for per-tenant extension removal", map[string]interface{}{
			"app_id": appID,
			"state":  app.State,
		})

		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out after %v waiting for per-tenant extension %s to be removed (last state %q)",
				timeout, appID, app.State)
		case <-ticker.C:
			// Continue polling.
		}
	}
}

// ListScheduledPteOperations returns the PTE installs and updates that are scheduled for
// a future deployment window on the environment.
func (s *Service) ListScheduledPteOperations(ctx context.Context, applicationFamily, environmentName string) ([]ScheduledPteOperation, error) {
	path := appsPath(applicationFamily, environmentName) + "/scheduledPteOperations"

	resp, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to list scheduled per-tenant extension operations: %w", err)
	}
	defer resp.Body.Close()

	var list ScheduledPteOperationListResponse
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, fmt.Errorf("failed to decode scheduled per-tenant extension operations response: %w", err)
	}

	return list.Value, nil
}

// GetScheduledPteOperationsForApp returns the scheduled operations that target the given
// app ID.
func (s *Service) GetScheduledPteOperationsForApp(ctx context.Context, applicationFamily, environmentName, appID string) ([]ScheduledPteOperation, error) {
	all, err := s.ListScheduledPteOperations(ctx, applicationFamily, environmentName)
	if err != nil {
		return nil, err
	}

	matches := make([]ScheduledPteOperation, 0, len(all))
	for i := range all {
		if strings.EqualFold(all[i].AppID, appID) {
			matches = append(matches, all[i])
		}
	}

	return matches, nil
}

// RemoveScheduledPteVersion cancels one scheduled PTE install or update, permanently
// removing the staged .app package for that version. The scheduled operation is
// identified by app ID, target version, and schedule kind.
func (s *Service) RemoveScheduledPteVersion(ctx context.Context, applicationFamily, environmentName, appID string, req *RemoveScheduledPteVersionRequest) (*AppOperation, error) {
	if req == nil || req.TargetVersion == "" || req.ScheduleKind == "" {
		return nil, fmt.Errorf("targetVersion and scheduleKind are required to remove a scheduled per-tenant extension version")
	}

	path := appPath(applicationFamily, environmentName, appID, "removeScheduledPteVersion")

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal remove scheduled version request: %w", err)
	}

	resp, err := s.client.Post(ctx, path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to remove scheduled per-tenant extension version: %w", err)
	}
	defer resp.Body.Close()

	var operation AppOperation
	if err := json.NewDecoder(resp.Body).Decode(&operation); err != nil {
		return nil, fmt.Errorf("failed to decode remove scheduled version response: %w", err)
	}

	return &operation, nil
}

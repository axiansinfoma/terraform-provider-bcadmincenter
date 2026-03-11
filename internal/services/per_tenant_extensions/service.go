// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package pertenantextensions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/client"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Service handles per-tenant extension lifecycle operations via the BC Automation API.
type Service struct {
	client *client.Client
}

// NewService creates a new per-tenant extension service.
func NewService(c *client.Client) *Service {
	return &Service{client: c}
}

// GetFirstCompany fetches automation companies and returns the ID of the first one.
// BC PTEs are published globally across all companies so the choice of company is only
// an implementation detail for the Automation API endpoint.
func (s *Service) GetFirstCompany(ctx context.Context, environmentName string) (string, error) {
	resp, err := s.client.DoAutomationRequest(ctx, http.MethodGet, environmentName, "companies", nil, "", nil)
	if err != nil {
		return "", fmt.Errorf("failed to list automation companies: %w", err)
	}
	defer resp.Body.Close()

	var list CompanyListResponse
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return "", fmt.Errorf("failed to decode companies response: %w", err)
	}

	if len(list.Value) == 0 {
		return "", fmt.Errorf("no companies found in environment %q", environmentName)
	}

	return list.Value[0].ID, nil
}

// CreateExtensionUpload creates an extension upload record and returns the system ID.
func (s *Service) CreateExtensionUpload(ctx context.Context, environmentName, companyID string, req *ExtensionUploadRequest) (string, error) {
	path := fmt.Sprintf("companies(%s)/extensionUpload", companyID)

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal extension upload request: %w", err)
	}

	resp, err := s.client.DoAutomationRequest(ctx, http.MethodPost, environmentName, path, bytes.NewReader(body), "", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create extension upload: %w", err)
	}
	defer resp.Body.Close()

	var upload ExtensionUpload
	if err := json.NewDecoder(resp.Body).Decode(&upload); err != nil {
		return "", fmt.Errorf("failed to decode extension upload response: %w", err)
	}

	if upload.SystemID == "" {
		return "", fmt.Errorf("extension upload response missing systemId")
	}

	return upload.SystemID, nil
}

// UploadExtensionContent streams the raw .app file bytes to the upload record.
func (s *Service) UploadExtensionContent(ctx context.Context, environmentName, companyID, uploadID string, data []byte) error {
	path := fmt.Sprintf("companies(%s)/extensionUpload(%s)/extensionContent", companyID, uploadID)

	resp, err := s.client.DoAutomationRequest(
		ctx, http.MethodPatch, environmentName, path,
		bytes.NewReader(data),
		"application/octet-stream",
		map[string]string{"If-Match": "*"},
	)
	if err != nil {
		return fmt.Errorf("failed to upload extension content: %w", err)
	}
	resp.Body.Close()

	return nil
}

// TriggerInstall calls Microsoft.NAV.upload to trigger the installation.
func (s *Service) TriggerInstall(ctx context.Context, environmentName, companyID, uploadID string) error {
	path := fmt.Sprintf("companies(%s)/extensionUpload(%s)/Microsoft.NAV.upload", companyID, uploadID)

	resp, err := s.client.DoAutomationRequest(ctx, http.MethodPost, environmentName, path, nil, "", nil)
	if err != nil {
		return fmt.Errorf("failed to trigger extension install: %w", err)
	}
	resp.Body.Close()

	return nil
}

// WaitForDeployment polls extensionDeploymentStatus until the deployment reaches a terminal state.
// It matches by name (or publisher+name) because the API does not return an operation ID from the
// install trigger. Returns the final status entry on success, or an error on failure/timeout.
func (s *Service) WaitForDeployment(ctx context.Context, environmentName, companyID string, timeout time.Duration) (*ExtensionDeploymentStatus, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	// Poll immediately first, then at intervals.
	for {
		status, err := s.getLatestDeploymentStatus(ctx, environmentName, companyID)
		if err != nil {
			return nil, fmt.Errorf("failed to poll deployment status: %w", err)
		}

		if status != nil {
			tflog.Debug(ctx, "Extension deployment status", map[string]interface{}{
				"name":           status.Name,
				"status":         status.Status,
				"operation_type": status.OperationType,
			})

			switch status.Status {
			case DeploymentStatusCompleted:
				return status, nil
			case DeploymentStatusFailed:
				return nil, fmt.Errorf("extension deployment failed (operationType=%s, status=%s, name=%q)",
					status.OperationType, status.Status, status.Name)
			}
		}

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timed out waiting for extension deployment after %v", timeout)
		case <-ticker.C:
			// continue polling
		}
	}
}

// getLatestDeploymentStatus returns the most recently started non-terminal or terminal deployment status entry.
func (s *Service) getLatestDeploymentStatus(ctx context.Context, environmentName, companyID string) (*ExtensionDeploymentStatus, error) {
	path := fmt.Sprintf("companies(%s)/extensionDeploymentStatus", companyID)

	resp, err := s.client.DoAutomationRequest(ctx, http.MethodGet, environmentName, path, nil, "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment status: %w", err)
	}
	defer resp.Body.Close()

	var list ExtensionDeploymentStatusListResponse
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, fmt.Errorf("failed to decode deployment status response: %w", err)
	}

	if len(list.Value) == 0 {
		return nil, nil
	}

	// Return the first entry – BC returns entries in reverse-chronological order.
	return &list.Value[0], nil
}

// GetExtensionByPackageID looks up an extension from the extensions collection by packageId.
// Returns (nil, nil) when no matching extension is found.
func (s *Service) GetExtensionByPackageID(ctx context.Context, environmentName, companyID, packageID string) (*Extension, error) {
	path := fmt.Sprintf("companies(%s)/extensions", companyID)

	resp, err := s.client.DoAutomationRequest(ctx, http.MethodGet, environmentName, path, nil, "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list extensions: %w", err)
	}
	defer resp.Body.Close()

	var list ExtensionListResponse
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, fmt.Errorf("failed to decode extensions response: %w", err)
	}

	for i := range list.Value {
		if list.Value[i].PackageID == packageID {
			return &list.Value[i], nil
		}
	}

	return nil, nil
}

// GetExtensionByAppID looks up an installed extension by its stable appId (id field).
// Returns (nil, nil) when no matching extension is found.
func (s *Service) GetExtensionByAppID(ctx context.Context, environmentName, companyID, appID string) (*Extension, error) {
	path := fmt.Sprintf("companies(%s)/extensions", companyID)

	resp, err := s.client.DoAutomationRequest(ctx, http.MethodGet, environmentName, path, nil, "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list extensions: %w", err)
	}
	defer resp.Body.Close()

	var list ExtensionListResponse
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, fmt.Errorf("failed to decode extensions response: %w", err)
	}

	for i := range list.Value {
		if list.Value[i].ID == appID && list.Value[i].IsInstalled {
			return &list.Value[i], nil
		}
	}

	return nil, nil
}

// Uninstall uninstalls an extension by packageId.
// When deleteData is true it calls Microsoft.NAV.uninstallAndDeleteExtensionData; otherwise Microsoft.NAV.uninstall.
func (s *Service) Uninstall(ctx context.Context, environmentName, companyID, packageID string, deleteData bool) error {
	action := "Microsoft.NAV.uninstall"
	if deleteData {
		action = "Microsoft.NAV.uninstallAndDeleteExtensionData"
	}

	path := fmt.Sprintf("companies(%s)/extensions(%s)/%s", companyID, packageID, action)

	resp, err := s.client.DoAutomationRequest(ctx, http.MethodPost, environmentName, path, nil, "", nil)
	if err != nil {
		return fmt.Errorf("failed to uninstall extension: %w", err)
	}
	resp.Body.Close()

	return nil
}

// Unpublish calls Microsoft.NAV.unpublish on the extension identified by packageId.
// Gracefully ignores 404/405 responses (indicating the BC version does not support unpublish).
func (s *Service) Unpublish(ctx context.Context, environmentName, companyID, packageID string) error {
	path := fmt.Sprintf("companies(%s)/extensions(%s)/Microsoft.NAV.unpublish", companyID, packageID)

	resp, err := s.client.DoAutomationRequest(ctx, http.MethodPost, environmentName, path, nil, "", nil)
	if err != nil {
		// Gracefully skip if the endpoint does not exist (older BC versions).
		if _, ok := err.(*client.AdminCenterError); ok {
			tflog.Warn(ctx, "Microsoft.NAV.unpublish not supported on this BC version, skipping", map[string]interface{}{
				"package_id": packageID,
			})
			return nil
		}
		return fmt.Errorf("failed to unpublish extension: %w", err)
	}
	resp.Body.Close()

	return nil
}

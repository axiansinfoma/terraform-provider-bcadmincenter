// Copyright (c) 2025 Michael Villani
// SPDX-License-Identifier: MPL-2.0

package environmentsupportcontact

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/vllni/terraform-provider-bcadmincenter/internal/client"
)

// Service handles support contact operations for the Business Central Admin Center API.
type Service struct {
	client *client.Client
}

// NewService creates a new support contact service.
func NewService(c *client.Client) *Service {
	return &Service{
		client: c,
	}
}

// Get retrieves the support contact information for an environment.
func (s *Service) Get(ctx context.Context, applicationFamily, environmentName string) (*SupportContact, error) {
	path := fmt.Sprintf("support/applications/%s/environments/%s/supportcontact", applicationFamily, environmentName)

	resp, err := s.client.Get(ctx, path)
	if err != nil {
		// Check if it's a 404 error (no support contact configured)
		if isNotFoundError(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get support contact: %w", err)
	}
	defer resp.Body.Close()

	var contact SupportContact
	if err := json.NewDecoder(resp.Body).Decode(&contact); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &contact, nil
}

// isNotFoundError checks if an error is a 404 Not Found error.
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	var apiErr *client.AdminCenterError
	if errors.As(err, &apiErr) {
		return apiErr.Code == "ResourceNotFound" || apiErr.Code == "EnvironmentNotFound"
	}

	return strings.Contains(err.Error(), "404")
}

// Set updates the support contact information for an environment.
func (s *Service) Set(ctx context.Context, applicationFamily, environmentName string, contact *SupportContact) (*SupportContact, error) {
	path := fmt.Sprintf("support/applications/%s/environments/%s/supportcontact", applicationFamily, environmentName)

	body, err := json.Marshal(contact)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := s.client.Put(ctx, path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to set support contact: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, readResponseBody(resp.Body))
	}

	var updatedContact SupportContact
	if err := json.NewDecoder(resp.Body).Decode(&updatedContact); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &updatedContact, nil
}

func readResponseBody(body io.Reader) string {
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return fmt.Sprintf("failed to read response body: %v", err)
	}
	return string(bodyBytes)
}

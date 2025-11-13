// Copyright (c) 2025 Michael Villani
// SPDX-License-Identifier: MPL-2.0

package environmentsupportcontact

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/vllni/terraform-provider-bcadmincenter/internal/client"
)

// Service handles support contact operations for the Business Central Admin Center API
type Service struct {
	client *client.Client
}

// NewService creates a new support contact service
func NewService(c *client.Client) *Service {
	return &Service{
		client: c,
	}
}

// Get retrieves the support contact information for an environment
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

// isNotFoundError checks if an error is a 404 Not Found error
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	// Check if error message contains "404"
	errMsg := err.Error()
	return contains(errMsg, "404")
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || indexOfSubstring(s, substr) >= 0)
}

// indexOfSubstring returns the index of the first instance of substr in s, or -1 if substr is not present in s
func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// Set updates the support contact information for an environment
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
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var updatedContact SupportContact
	if err := json.NewDecoder(resp.Body).Decode(&updatedContact); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &updatedContact, nil
}

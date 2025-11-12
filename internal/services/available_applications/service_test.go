// Copyright (c) 2025 Michael Villani
// SPDX-License-Identifier: MPL-2.0

package available_applications

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/vllni/terraform-provider-bc-admin-center/internal/client"
)

// mockTokenCredential implements azcore.TokenCredential for testing
type mockTokenCredential struct {
	token string
	err   error
}

func (m *mockTokenCredential) GetToken(ctx context.Context, options policy.TokenRequestOptions) (azcore.AccessToken, error) {
	if m.err != nil {
		return azcore.AccessToken{}, m.err
	}
	return azcore.AccessToken{
		Token: m.token,
	}, nil
}

func TestService_GetAvailableApplications(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   interface{}
		responseStatus int
		wantErr        bool
		wantFamilies   int
		wantCountries  int // Expected countries for first family
	}{
		{
			name: "successful response with BusinessCentral",
			responseBody: AvailableApplicationsResponse{
				Value: []ApplicationFamily{
					{
						ApplicationFamily: "BusinessCentral",
						CountriesRingDetails: []CountryRingDetails{
							{
								CountryCode: "US",
								Rings: []Ring{
									{
										Name:           "PROD",
										ProductionRing: true,
										FriendlyName:   "Production",
									},
									{
										Name:           "PREVIEW",
										ProductionRing: false,
										FriendlyName:   "Preview",
									},
								},
							},
							{
								CountryCode: "GB",
								Rings: []Ring{
									{
										Name:           "PROD",
										ProductionRing: true,
										FriendlyName:   "Production",
									},
								},
							},
						},
					},
				},
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
			wantFamilies:   1,
			wantCountries:  2,
		},
		{
			name: "successful response with multiple families",
			responseBody: AvailableApplicationsResponse{
				Value: []ApplicationFamily{
					{
						ApplicationFamily: "BusinessCentral",
						CountriesRingDetails: []CountryRingDetails{
							{
								CountryCode: "US",
								Rings: []Ring{
									{
										Name:           "PROD",
										ProductionRing: true,
										FriendlyName:   "Production",
									},
								},
							},
						},
					},
					{
						ApplicationFamily: "FinancialManagement",
						CountriesRingDetails: []CountryRingDetails{
							{
								CountryCode: "DK",
								Rings: []Ring{
									{
										Name:           "PROD",
										ProductionRing: true,
										FriendlyName:   "Production",
									},
								},
							},
						},
					},
				},
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
			wantFamilies:   2,
			wantCountries:  1,
		},
		{
			name:           "empty response",
			responseBody:   AvailableApplicationsResponse{Value: []ApplicationFamily{}},
			responseStatus: http.StatusOK,
			wantErr:        false,
			wantFamilies:   0,
		},
		{
			name:           "server error",
			responseBody:   map[string]string{"error": "internal server error"},
			responseStatus: http.StatusInternalServerError,
			wantErr:        true,
		},
		{
			name:           "unauthorized",
			responseBody:   map[string]string{"error": "unauthorized"},
			responseStatus: http.StatusUnauthorized,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify the request path
				expectedPath := "/admin/v2.24/applications/"
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}

				// Verify authorization header
				authHeader := r.Header.Get("Authorization")
				if authHeader != "Bearer test-token" {
					t.Errorf("Expected Authorization header 'Bearer test-token', got '%s'", authHeader)
				}

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			// Create a client with the test server
			mockCred := &mockTokenCredential{token: "test-token"}
			c := &client.Client{}
			c.SetCredential(mockCred)
			c.SetBaseURL(server.URL)
			c.SetAPIVersion("v2.24")
			c.SetHTTPClient(&http.Client{})

			// Create service
			svc := NewService(c)

			// Call the method
			result, err := svc.GetAvailableApplications(context.Background())

			// Check error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAvailableApplications() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Verify results
			if len(result.Value) != tt.wantFamilies {
				t.Errorf("GetAvailableApplications() returned %d families, want %d", len(result.Value), tt.wantFamilies)
			}

			if tt.wantFamilies > 0 && tt.wantCountries > 0 {
				if len(result.Value[0].CountriesRingDetails) != tt.wantCountries {
					t.Errorf("First family has %d countries, want %d", len(result.Value[0].CountriesRingDetails), tt.wantCountries)
				}
			}
		})
	}
}

func TestService_GetApplicationFamily(t *testing.T) {
	tests := []struct {
		name           string
		familyName     string
		responseBody   interface{}
		responseStatus int
		wantErr        bool
		wantCountries  int
		errContains    string
	}{
		{
			name:       "successful retrieval of BusinessCentral",
			familyName: "BusinessCentral",
			responseBody: AvailableApplicationsResponse{
				Value: []ApplicationFamily{
					{
						ApplicationFamily: "BusinessCentral",
						CountriesRingDetails: []CountryRingDetails{
							{
								CountryCode: "US",
								Rings: []Ring{
									{
										Name:           "PROD",
										ProductionRing: true,
										FriendlyName:   "Production",
									},
								},
							},
							{
								CountryCode: "GB",
								Rings: []Ring{
									{
										Name:           "PROD",
										ProductionRing: true,
										FriendlyName:   "Production",
									},
								},
							},
						},
					},
				},
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
			wantCountries:  2,
		},
		{
			name:       "application family not found",
			familyName: "NonExistent",
			responseBody: AvailableApplicationsResponse{
				Value: []ApplicationFamily{
					{
						ApplicationFamily: "BusinessCentral",
						CountriesRingDetails: []CountryRingDetails{
							{
								CountryCode: "US",
								Rings:       []Ring{{Name: "PROD", ProductionRing: true, FriendlyName: "Production"}},
							},
						},
					},
				},
			},
			responseStatus: http.StatusOK,
			wantErr:        true,
			errContains:    "not found",
		},
		{
			name:           "server error",
			familyName:     "BusinessCentral",
			responseBody:   map[string]string{"error": "internal server error"},
			responseStatus: http.StatusInternalServerError,
			wantErr:        true,
		},
		{
			name:       "case sensitive family name",
			familyName: "businesscentral", // lowercase
			responseBody: AvailableApplicationsResponse{
				Value: []ApplicationFamily{
					{
						ApplicationFamily: "BusinessCentral", // PascalCase
						CountriesRingDetails: []CountryRingDetails{
							{
								CountryCode: "US",
								Rings:       []Ring{{Name: "PROD", ProductionRing: true, FriendlyName: "Production"}},
							},
						},
					},
				},
			},
			responseStatus: http.StatusOK,
			wantErr:        true,
			errContains:    "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			// Create a client with the test server
			mockCred := &mockTokenCredential{token: "test-token"}
			c := &client.Client{}
			c.SetCredential(mockCred)
			c.SetBaseURL(server.URL)
			c.SetAPIVersion("v2.24")
			c.SetHTTPClient(&http.Client{})

			// Create service
			svc := NewService(c)

			// Call the method
			result, err := svc.GetApplicationFamily(context.Background(), tt.familyName)

			// Check error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("GetApplicationFamily() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.errContains != "" && err != nil {
					if !contains(err.Error(), tt.errContains) {
						t.Errorf("GetApplicationFamily() error = %v, want error containing %v", err, tt.errContains)
					}
				}
				return
			}

			// Verify results
			if result.ApplicationFamily != tt.familyName {
				t.Errorf("GetApplicationFamily() returned family %s, want %s", result.ApplicationFamily, tt.familyName)
			}

			if len(result.CountriesRingDetails) != tt.wantCountries {
				t.Errorf("GetApplicationFamily() returned %d countries, want %d", len(result.CountriesRingDetails), tt.wantCountries)
			}
		})
	}
}

func TestNewService(t *testing.T) {
	mockCred := &mockTokenCredential{token: "test-token"}
	c := &client.Client{}
	c.SetCredential(mockCred)
	c.SetHTTPClient(&http.Client{})

	svc := NewService(c)

	if svc == nil {
		t.Fatal("NewService() returned nil")
	}

	if svc.client != c {
		t.Error("NewService() did not set client correctly")
	}
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > 0 && len(substr) > 0 && stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Copyright (c) 2025 Michael Villani
// SPDX-License-Identifier: MPL-2.0

package quotas

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/vllni/terraform-provider-bcadmincenter/internal/client"
)

// mockTokenCredential implements azcore.TokenCredential for testing
type mockTokenCredential struct {
	token string
}

func (m *mockTokenCredential) GetToken(ctx context.Context, options policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{
		Token: m.token,
	}, nil
}

func TestService_GetQuotas(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   interface{}
		responseStatus int
		wantErr        bool
		validateResult func(*testing.T, *QuotasResponse)
	}{
		{
			name: "successful response",
			responseBody: QuotasResponse{
				ProductionEnvironmentsQuota:    3,
				ProductionEnvironmentsAllocated: 1,
				SandboxEnvironmentsQuota:        10,
				SandboxEnvironmentsAllocated:    5,
				StorageQuotaGB:                  100,
				StorageAllocatedGB:              25,
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
			validateResult: func(t *testing.T, result *QuotasResponse) {
				if result.ProductionEnvironmentsQuota != 3 {
					t.Errorf("expected 3 production environments quota, got %d", result.ProductionEnvironmentsQuota)
				}
				if result.SandboxEnvironmentsQuota != 10 {
					t.Errorf("expected 10 sandbox environments quota, got %d", result.SandboxEnvironmentsQuota)
				}
			},
		},
		{
			name:           "not found error",
			responseBody:   map[string]string{"error": "not found"},
			responseStatus: http.StatusNotFound,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			// Create client with mock server
			mockCred := &mockTokenCredential{token: "test-token"}
			c := &client.Client{}
			c.SetCredential(mockCred)
			c.SetBaseURL(server.URL)
			c.SetAPIVersion("v2.24")
			c.SetHTTPClient(&http.Client{})

			// Test the method
			svc := NewService(c)
			result, err := svc.GetQuotas(context.Background())

			// Assert results
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validateResult != nil {
				tt.validateResult(t, result)
			}
		})
	}
}

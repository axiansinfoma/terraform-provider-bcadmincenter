// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package provider_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/constants"
	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/provider"
)

// testAccProtoV6ProviderFactories creates a provider factory for acceptance testing.
// The provider is served in-process using the ProtoV6 protocol.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"bcadmincenter": providerserver.NewProtocol6WithError(provider.New("test")()),
}

// testAccProviderConfig returns the HCL provider configuration block for acceptance tests.
// It uses the mock server URL and a static test token to bypass Azure AD authentication.
func testAccProviderConfig(serverURL string) string {
	return fmt.Sprintf(`
provider "bcadmincenter" {
  tenant_id = "00000000-0000-0000-0000-000000000001"
  base_url  = %q
}
`, serverURL)
}

// TestAccQuotasDataSource verifies the quotas data source works end-to-end
// with a mock Business Central Admin Center API.
func TestAccQuotasDataSource(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test in short mode.")
	}

	server := newMockServer(t)
	defer server.Close()

	t.Setenv("BCADMINCENTER_TEST_TOKEN", "test-token-for-acceptance-tests")

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(server.URL) + `
data "bcadmincenter_quotas" "test" {}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.bcadmincenter_quotas.test", "production_environments_quota", "3"),
					resource.TestCheckResourceAttr("data.bcadmincenter_quotas.test", "sandbox_environments_quota", "3"),
					resource.TestCheckResourceAttrSet("data.bcadmincenter_quotas.test", "id"),
				),
			},
		},
	})
}

// TestAccTimezonesDataSource verifies the timezones data source works end-to-end
// with a mock Business Central Admin Center API.
func TestAccTimezonesDataSource(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test in short mode.")
	}

	server := newMockServer(t)
	defer server.Close()

	t.Setenv("BCADMINCENTER_TEST_TOKEN", "test-token-for-acceptance-tests")

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(server.URL) + `
data "bcadmincenter_timezones" "test" {}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.bcadmincenter_timezones.test", "id"),
					resource.TestCheckResourceAttr("data.bcadmincenter_timezones.test", "timezones.#", "2"),
					resource.TestCheckResourceAttr("data.bcadmincenter_timezones.test", "timezones.0.id", "UTC"),
				),
			},
		},
	})
}

// TestAccEnvironmentsDataSource verifies the environments data source works end-to-end
// with a mock Business Central Admin Center API.
func TestAccEnvironmentsDataSource(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test in short mode.")
	}

	server := newMockServer(t)
	defer server.Close()

	t.Setenv("BCADMINCENTER_TEST_TOKEN", "test-token-for-acceptance-tests")

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(server.URL) + `
data "bcadmincenter_environments" "test" {
  application_family = "BusinessCentral"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.bcadmincenter_environments.test", "environments.#", "1"),
					resource.TestCheckResourceAttr("data.bcadmincenter_environments.test", "environments.0.name", "production"),
					resource.TestCheckResourceAttr("data.bcadmincenter_environments.test", "environments.0.type", "Production"),
					resource.TestCheckResourceAttr("data.bcadmincenter_environments.test", "environments.0.status", "Active"),
				),
			},
		},
	})
}

// TestAccNotificationRecipientResource verifies the notification recipient resource works
// end-to-end with a mock Business Central Admin Center API.
func TestAccNotificationRecipientResource(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test in short mode.")
	}

	server := newMockServer(t)
	defer server.Close()

	t.Setenv("BCADMINCENTER_TEST_TOKEN", "test-token-for-acceptance-tests")

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(server.URL) + `
resource "bcadmincenter_notification_recipient" "test" {
  email = "test@example.com"
  name  = "Test User"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("bcadmincenter_notification_recipient.test", "email", "test@example.com"),
					resource.TestCheckResourceAttr("bcadmincenter_notification_recipient.test", "name", "Test User"),
					resource.TestCheckResourceAttrSet("bcadmincenter_notification_recipient.test", "id"),
				),
			},
		},
	})
}

// newMockServer creates a mock HTTP server that simulates the Business Central Admin Center API.
func newMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	apiBase := "/admin/" + constants.DefaultAPIVersion + "/"

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		t.Logf("Mock server received: %s %s", r.Method, r.URL.Path)

		switch {
		case r.URL.Path == apiBase+"environments/quotas":
			handleQuotas(w)
		case r.URL.Path == apiBase+"applications/settings/timezones":
			handleTimezones(w)
		case r.URL.Path == apiBase+"applications/BusinessCentral/environments":
			handleEnvironments(w, r)
		case r.URL.Path == apiBase+"settings/notification/recipients":
			handleNotificationRecipients(w, r)
		case strings.HasPrefix(r.URL.Path, apiBase+"settings/notification/recipients/"):
			// Individual recipient operations (DELETE by ID)
			handleNotificationRecipients(w, r)
		case r.URL.Path == apiBase+"settings/notification":
			handleNotificationSettings(w)
		default:
			t.Logf("Mock server: unhandled path %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	}))
}

func handleQuotas(w http.ResponseWriter) {
	resp := map[string]interface{}{
		"productionEnvironmentsQuota":     3,
		"productionEnvironmentsAllocated": 1,
		"sandboxEnvironmentsQuota":        3,
		"sandboxEnvironmentsAllocated":    2,
		"storageQuotaGB":                  80,
		"storageAllocatedGB":              20,
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func handleTimezones(w http.ResponseWriter) {
	resp := map[string]interface{}{
		"value": []map[string]interface{}{
			{
				"id":                      "UTC",
				"displayName":             "(UTC) Coordinated Universal Time",
				"supportsDaylightSavings": false,
				"offsetFromUTC":           "+00:00",
			},
			{
				"id":                      "W. Europe Standard Time",
				"displayName":             "(UTC+01:00) Amsterdam, Berlin, Bern, Rome, Stockholm, Vienna",
				"supportsDaylightSavings": true,
				"offsetFromUTC":           "+01:00",
			},
		},
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func handleEnvironments(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		resp := map[string]interface{}{
			"value": []map[string]interface{}{
				{
					"name":               "production",
					"type":               "Production",
					"applicationFamily":  "BusinessCentral",
					"countryCode":        "US",
					"status":             "Active",
					"webClientLoginUrl":  "https://businesscentral.dynamics.com/00000000-0000-0000-0000-000000000001/production",
					"aadTenantId":        "00000000-0000-0000-0000-000000000001",
					"ringName":           "PROD",
					"applicationVersion": "25.0.0.0",
				},
			},
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func handleNotificationRecipients(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		resp := map[string]interface{}{
			"value": []map[string]interface{}{
				{
					"id":    "recipient-test-id",
					"email": "test@example.com",
					"name":  "Test User",
				},
			},
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	case http.MethodPut:
		resp := map[string]interface{}{
			"id":    "recipient-test-id",
			"email": "test@example.com",
			"name":  "Test User",
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	case http.MethodDelete:
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleNotificationSettings(w http.ResponseWriter) {
	resp := map[string]interface{}{
		"aadTenantId": "00000000-0000-0000-0000-000000000001",
		"recipients": []map[string]interface{}{
			{
				"id":    "recipient-test-id",
				"email": "test@example.com",
				"name":  "Test User",
			},
		},
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

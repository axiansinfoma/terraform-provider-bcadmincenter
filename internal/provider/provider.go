// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/client"
	authorizedentraapps "github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/services/authorized_entra_apps"
	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/services/available_applications"
	environmentapps "github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/services/environment_apps"
	environmentsupportcontact "github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/services/environment_support_contact"
	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/services/environments"
	notificationrecipients "github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/services/notification_recipients"
	pertenantextensions "github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/services/per_tenant_extensions"
	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/services/quotas"
	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/services/timezones"
)

// Ensure BCAdminCenterProvider satisfies various provider interfaces.
var _ provider.Provider = &BCAdminCenterProvider{}
var _ provider.ProviderWithFunctions = &BCAdminCenterProvider{}
var _ provider.ProviderWithEphemeralResources = &BCAdminCenterProvider{}

// BCAdminCenterProvider defines the provider implementation.
type BCAdminCenterProvider struct {
	// version is set to the provider version on release, "dev" when the.
	// provider is built and ran locally, and "test" when running acceptance.
	// testing.
	version string
}

// BCAdminCenterProviderModel describes the provider data model.
type BCAdminCenterProviderModel struct {
	ClientID                       types.String `tfsdk:"client_id"`
	ClientSecret                   types.String `tfsdk:"client_secret"`
	TenantID                       types.String `tfsdk:"tenant_id"`
	Environment                    types.String `tfsdk:"environment"`
	AuxiliaryTenantIDs             types.List   `tfsdk:"auxiliary_tenant_ids"`
	BaseURL                        types.String `tfsdk:"base_url"`
	UseOIDC                        types.Bool   `tfsdk:"use_oidc"`
	OIDCToken                      types.String `tfsdk:"oidc_token"`
	OIDCTokenFilePath              types.String `tfsdk:"oidc_token_file_path"`
	OIDCRequestURL                 types.String `tfsdk:"oidc_request_url"`
	OIDCRequestToken               types.String `tfsdk:"oidc_request_token"`
	ADOPipelineServiceConnectionID types.String `tfsdk:"ado_pipeline_service_connection_id"`
}

func (p *BCAdminCenterProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "bcadmincenter"
	resp.Version = p.version
}

func (p *BCAdminCenterProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Provider for managing Microsoft Dynamics 365 Business Central environments through the Business Central Admin Center API.",
		Attributes: map[string]schema.Attribute{
			"client_id": schema.StringAttribute{
				MarkdownDescription: "The Client ID (Application ID) for Azure AD authentication. Can also be set via the `ARM_CLIENT_ID` environment variable (or `AZURE_CLIENT_ID` for backward compatibility).",
				Optional:            true,
			},
			"client_secret": schema.StringAttribute{
				MarkdownDescription: "The Client Secret for Azure AD authentication. Can also be set via the `ARM_CLIENT_SECRET` environment variable (or `AZURE_CLIENT_SECRET` for backward compatibility).",
				Optional:            true,
				Sensitive:           true,
			},
			"tenant_id": schema.StringAttribute{
				MarkdownDescription: "The Tenant ID for Azure AD authentication. Can also be set via the `ARM_TENANT_ID` environment variable (or `AZURE_TENANT_ID` for backward compatibility).",
				Optional:            true,
			},
			"environment": schema.StringAttribute{
				MarkdownDescription: "The Azure environment to use (public, usgovernment, china). Defaults to 'public'. Can also be set via the `ARM_ENVIRONMENT` environment variable (or `AZURE_ENVIRONMENT` for backward compatibility).",
				Optional:            true,
			},
			"auxiliary_tenant_ids": schema.ListAttribute{
				MarkdownDescription: "List of auxiliary tenant IDs for multi-tenant scenarios.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"base_url": schema.StringAttribute{
				MarkdownDescription: "Override the base URL for the Business Central Admin Center API. Can also be set via BCADMINCENTER_BASE_URL environment variable. Primarily used for testing.",
				Optional:            true,
			},
			"use_oidc": schema.BoolAttribute{
				MarkdownDescription: "Force the use of OIDC / Workload Identity (federated credential) authentication. When true, the provider uses `WorkloadIdentityCredential` from the Azure SDK, which reads the federated token from the file specified by `oidc_token_file_path` (or `AZURE_FEDERATED_TOKEN_FILE`). Can also be set via `ARM_USE_OIDC` or `AZURE_USE_OIDC` environment variable.",
				Optional:            true,
			},
			"oidc_token": schema.StringAttribute{
				MarkdownDescription: "A JWT bearer token to use as the OIDC client assertion. Useful when the token is provided directly by the CI/CD platform. Can also be set via `ARM_OIDC_TOKEN` or `AZURE_OIDC_TOKEN` environment variable. Setting this implies `use_oidc = true`.",
				Optional:            true,
				Sensitive:           true,
			},
			"oidc_token_file_path": schema.StringAttribute{
				MarkdownDescription: "Path to a file containing the OIDC / federated token. The file is re-read on every Azure AD token refresh so platform-rotated tokens (e.g. Kubernetes projected volumes) are picked up automatically. Can also be set via `ARM_OIDC_TOKEN_FILE_PATH` or `AZURE_FEDERATED_TOKEN_FILE` environment variable. Used when `use_oidc = true`.",
				Optional:            true,
			},
			"oidc_request_url": schema.StringAttribute{
				MarkdownDescription: "The URL of the OIDC token endpoint. In GitHub Actions this is set automatically via `ACTIONS_ID_TOKEN_REQUEST_URL`; in Azure DevOps via `SYSTEM_OIDCREQUESTURI`. A fresh JWT is fetched from this endpoint on every Azure AD token refresh, so short-lived tokens are automatically renewed during long Terraform runs. Can also be set via `ARM_OIDC_REQUEST_URL`, `ACTIONS_ID_TOKEN_REQUEST_URL`, or `SYSTEM_OIDCREQUESTURI` environment variable.",
				Optional:            true,
			},
			"oidc_request_token": schema.StringAttribute{
				MarkdownDescription: "The bearer token used to authenticate requests to `oidc_request_url`. In GitHub Actions this is set automatically via `ACTIONS_ID_TOKEN_REQUEST_TOKEN`; in Azure DevOps via `SYSTEM_ACCESSTOKEN`. Can also be set via `ARM_OIDC_REQUEST_TOKEN`, `ACTIONS_ID_TOKEN_REQUEST_TOKEN`, or `SYSTEM_ACCESSTOKEN` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"ado_pipeline_service_connection_id": schema.StringAttribute{
				MarkdownDescription: "The Azure DevOps service connection ID used when authenticating via Azure DevOps Pipeline OIDC (`SYSTEM_OIDCREQUESTURI`). When set, the provider uses the ADO OIDC endpoint protocol (POST with `serviceConnectionId` and `api-version` query parameters) instead of the GitHub Actions endpoint. Can also be set via `ARM_ADO_PIPELINE_SERVICE_CONNECTION_ID` or `ARM_OIDC_AZURE_SERVICE_CONNECTION_ID` environment variable.",
				Optional:            true,
			},
		},
	}
}

func (p *BCAdminCenterProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data BCAdminCenterProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get configuration values from provider config or environment variables.
	// ARM_ prefixed variables take precedence (matching the azurerm provider convention);
	// AZURE_ prefixed variables are supported for backward compatibility.
	clientID := getConfigValue(data.ClientID, "ARM_CLIENT_ID", "AZURE_CLIENT_ID")
	clientSecret := getConfigValue(data.ClientSecret, "ARM_CLIENT_SECRET", "AZURE_CLIENT_SECRET")
	tenantID := getConfigValue(data.TenantID, "ARM_TENANT_ID", "AZURE_TENANT_ID")
	environment := getConfigValue(data.Environment, "ARM_ENVIRONMENT", "AZURE_ENVIRONMENT")
	baseURL := getConfigValue(data.BaseURL, "BCADMINCENTER_BASE_URL")
	useOIDC := getConfigBoolValue(data.UseOIDC, "ARM_USE_OIDC", "AZURE_USE_OIDC")
	oidcToken := getConfigValue(data.OIDCToken, "ARM_OIDC_TOKEN", "AZURE_OIDC_TOKEN")
	oidcTokenFilePath := getConfigValue(data.OIDCTokenFilePath, "ARM_OIDC_TOKEN_FILE_PATH", "AZURE_FEDERATED_TOKEN_FILE")
	oidcRequestURL := getConfigValue(data.OIDCRequestURL, "ARM_OIDC_REQUEST_URL", "ACTIONS_ID_TOKEN_REQUEST_URL", "SYSTEM_OIDCREQUESTURI")
	oidcRequestToken := getConfigValue(data.OIDCRequestToken, "ARM_OIDC_REQUEST_TOKEN", "ACTIONS_ID_TOKEN_REQUEST_TOKEN", "SYSTEM_ACCESSTOKEN")
	adoPipelineServiceConnectionID := getConfigValue(data.ADOPipelineServiceConnectionID, "ARM_ADO_PIPELINE_SERVICE_CONNECTION_ID", "ARM_OIDC_AZURE_SERVICE_CONNECTION_ID")
	// accessToken allows bypassing Azure AD authentication for testing purposes.
	accessToken := os.Getenv("BCADMINCENTER_TEST_TOKEN")

	// Validate required configuration.
	if tenantID == "" {
		resp.Diagnostics.AddError(
			"Missing Tenant ID",
			"Tenant ID must be provided either through the provider configuration or the ARM_TENANT_ID (or AZURE_TENANT_ID) environment variable",
		)
		return
	}

	// Set default environment if not specified.
	if environment == "" {
		environment = "public"
	}

	tflog.Debug(ctx, "Configuring Business Central Admin Center client", map[string]interface{}{
		"tenant_id":   tenantID,
		"environment": environment,
	})

	// Create the client.
	config := &client.Config{
		ClientID:                       clientID,
		ClientSecret:                   clientSecret,
		TenantID:                       tenantID,
		Environment:                    environment,
		BaseURL:                        baseURL,
		AccessToken:                    accessToken,
		UseOIDC:                        useOIDC,
		OIDCToken:                      oidcToken,
		OIDCTokenFilePath:              oidcTokenFilePath,
		OIDCRequestURL:                 oidcRequestURL,
		OIDCRequestToken:               oidcRequestToken,
		ADOPipelineServiceConnectionID: adoPipelineServiceConnectionID,
	}

	bcClient, err := client.NewClient(ctx, config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create Business Central Admin Center client",
			"Error: "+err.Error(),
		)
		return
	}

	// Make the client available to data sources and resources.
	resp.DataSourceData = bcClient
	resp.ResourceData = bcClient
}

// getConfigValue returns the config value if set, otherwise returns the first non-empty environment variable value.
func getConfigValue(configValue types.String, envVars ...string) string {
	if !configValue.IsNull() && configValue.ValueString() != "" {
		return configValue.ValueString()
	}
	for _, envVar := range envVars {
		if v := os.Getenv(envVar); v != "" {
			return v
		}
	}
	return ""
}

// getConfigBoolValue returns the config bool value if set, otherwise parses the first matching environment variable.
// An unset or empty environment variable is treated as false.
// Accepted truthy values (case-insensitive): "true", "1", "yes", "on".
func getConfigBoolValue(configValue types.Bool, envVars ...string) bool {
	if !configValue.IsNull() {
		return configValue.ValueBool()
	}
	for _, envVar := range envVars {
		if v := os.Getenv(envVar); v != "" {
			switch strings.ToLower(v) {
			case "true", "1", "yes", "on":
				return true
			}
			return false
		}
	}
	return false
}

func (p *BCAdminCenterProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		authorizedentraapps.NewAuthorizedEntraAppResource,
		environments.NewEnvironmentResource,
		environments.NewUpdateScheduleResource,
		environmentapps.NewEnvironmentAppResource,
		environmentsupportcontact.NewEnvironmentSupportContactResource,
		notificationrecipients.NewNotificationRecipientResource,
		pertenantextensions.NewPerTenantExtensionResource,
	}
}

func (p *BCAdminCenterProvider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{
		// TODO: Add ephemeral resources here as they are implemented.
	}
}

func (p *BCAdminCenterProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		authorizedentraapps.NewAuthorizedEntraAppsDataSource,
		authorizedentraapps.NewManageableTenantsDataSource,
		available_applications.NewAvailableApplicationsDataSource,
		available_applications.NewApplicationFamilyDataSource,
		environments.NewEnvironmentDataSource,
		environments.NewEnvironmentsDataSource,
		environments.NewEnvironmentUpdatesDataSource,
		notificationrecipients.NewNotificationSettingsDataSource,
		quotas.NewQuotasDataSource,
		timezones.NewTimeZonesDataSource,
	}
}

func (p *BCAdminCenterProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		// TODO: Add functions here as they are implemented.
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &BCAdminCenterProvider{
			version: version,
		}
	}
}

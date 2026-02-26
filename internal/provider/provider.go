// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

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
	environmentsettings "github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/services/environment_settings"
	environmentsupportcontact "github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/services/environment_support_contact"
	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/services/environments"
	notificationrecipients "github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/services/notification_recipients"
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
	ClientID           types.String `tfsdk:"client_id"`
	ClientSecret       types.String `tfsdk:"client_secret"`
	TenantID           types.String `tfsdk:"tenant_id"`
	Environment        types.String `tfsdk:"environment"`
	AuxiliaryTenantIDs types.List   `tfsdk:"auxiliary_tenant_ids"`
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
				MarkdownDescription: "The Client ID (Application ID) for Azure AD authentication. Can also be set via AZURE_CLIENT_ID environment variable.",
				Optional:            true,
			},
			"client_secret": schema.StringAttribute{
				MarkdownDescription: "The Client Secret for Azure AD authentication. Can also be set via AZURE_CLIENT_SECRET environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"tenant_id": schema.StringAttribute{
				MarkdownDescription: "The Tenant ID for Azure AD authentication. Can also be set via AZURE_TENANT_ID environment variable.",
				Optional:            true,
			},
			"environment": schema.StringAttribute{
				MarkdownDescription: "The Azure environment to use (public, usgovernment, china). Defaults to 'public'. Can also be set via AZURE_ENVIRONMENT environment variable.",
				Optional:            true,
			},
			"auxiliary_tenant_ids": schema.ListAttribute{
				MarkdownDescription: "List of auxiliary tenant IDs for multi-tenant scenarios.",
				Optional:            true,
				ElementType:         types.StringType,
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
	clientID := getConfigValue(data.ClientID, "AZURE_CLIENT_ID")
	clientSecret := getConfigValue(data.ClientSecret, "AZURE_CLIENT_SECRET")
	tenantID := getConfigValue(data.TenantID, "AZURE_TENANT_ID")
	environment := getConfigValue(data.Environment, "AZURE_ENVIRONMENT")

	// Validate required configuration.
	if tenantID == "" {
		resp.Diagnostics.AddError(
			"Missing Tenant ID",
			"Tenant ID must be provided either through the provider configuration or AZURE_TENANT_ID environment variable",
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
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TenantID:     tenantID,
		Environment:  environment,
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

// getConfigValue returns the config value if set, otherwise returns the environment variable value.
func getConfigValue(configValue types.String, envVar string) string {
	if !configValue.IsNull() && configValue.ValueString() != "" {
		return configValue.ValueString()
	}
	return os.Getenv(envVar)
}

func (p *BCAdminCenterProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		authorizedentraapps.NewAuthorizedEntraAppResource,
		environments.NewEnvironmentResource,
		environmentsettings.NewEnvironmentSettingsResource,
		environmentsupportcontact.NewEnvironmentSupportContactResource,
		notificationrecipients.NewNotificationRecipientResource,
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

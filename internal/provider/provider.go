// Copyright (c) 2025 Michael Villani
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure BCAdminCenterProvider satisfies various provider interfaces.
var _ provider.Provider = &BCAdminCenterProvider{}
var _ provider.ProviderWithFunctions = &BCAdminCenterProvider{}
var _ provider.ProviderWithEphemeralResources = &BCAdminCenterProvider{}

// BCAdminCenterProvider defines the provider implementation.
type BCAdminCenterProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// BCAdminCenterProviderModel describes the provider data model.
type BCAdminCenterProviderModel struct {
	ClientID       types.String `tfsdk:"client_id"`
	ClientSecret   types.String `tfsdk:"client_secret"`
	TenantID       types.String `tfsdk:"tenant_id"`
	Environment    types.String `tfsdk:"environment"`
	AuxiliaryTenantIDs types.List `tfsdk:"auxiliary_tenant_ids"`
}

func (p *BCAdminCenterProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "bc_admin_center"
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

	// TODO: Implement Azure authentication and client initialization
	// This will be implemented when we create the client package
	
	// Example client configuration for data sources and resources
	client := http.DefaultClient
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *BCAdminCenterProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		// TODO: Add resources here as they are implemented
	}
}

func (p *BCAdminCenterProvider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{
		// TODO: Add ephemeral resources here as they are implemented
	}
}

func (p *BCAdminCenterProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// TODO: Add data sources here as they are implemented
	}
}

func (p *BCAdminCenterProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		// TODO: Add functions here as they are implemented
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &BCAdminCenterProvider{
			version: version,
		}
	}
}

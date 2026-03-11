// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package pertenantextensions

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &PerTenantExtensionResource{}
	_ resource.ResourceWithConfigure   = &PerTenantExtensionResource{}
	_ resource.ResourceWithImportState = &PerTenantExtensionResource{}
)

// NewPerTenantExtensionResource is a helper function to simplify the provider implementation.
func NewPerTenantExtensionResource() resource.Resource {
	return &PerTenantExtensionResource{}
}

// PerTenantExtensionResource is the resource implementation.
type PerTenantExtensionResource struct {
	client *client.Client
}

// PerTenantExtensionResourceModel describes the resource data model.
type PerTenantExtensionResourceModel struct {
	ID                types.String `tfsdk:"id"`
	AADTenantID       types.String `tfsdk:"aad_tenant_id"`
	CompanyID         types.String `tfsdk:"company_id"`
	EnvironmentName   types.String `tfsdk:"environment_name"`
	ApplicationFamily types.String `tfsdk:"application_family"`
	FilePath          types.String `tfsdk:"file_path"`
	FileContent       types.String `tfsdk:"file_content"`
	FileSHA256        types.String `tfsdk:"file_sha256"`
	Schedule          types.String `tfsdk:"schedule"`
	SchemaSyncMode    types.String `tfsdk:"schema_sync_mode"`
	DeleteData        types.Bool   `tfsdk:"delete_data"`
	UnpublishOnDelete types.Bool   `tfsdk:"unpublish_on_delete"`
	PackageID         types.String `tfsdk:"package_id"`
	AppID             types.String `tfsdk:"app_id"`
	DisplayName       types.String `tfsdk:"display_name"`
	Publisher         types.String `tfsdk:"publisher"`
	Version           types.String `tfsdk:"version"`
}

// Metadata returns the resource type name.
func (r *PerTenantExtensionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_per_tenant_extension"
}

// Schema defines the schema for the resource.
func (r *PerTenantExtensionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the full lifecycle of a Per-Tenant Extension (PTE) in a Business Central environment.\n\n" +
			"This resource uploads a `.app` extension package, installs it, updates it when the package changes, " +
			"and uninstalls it on destroy. All operations use the **Business Central Automation API**.\n\n" +
			"~> **Note:** Exactly one of `file_path` or `file_content` must be set.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ARM-like resource ID (format: `/tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/applications/{applicationFamily}/environments/{environmentName}/perTenantExtensions/{appId}`).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"aad_tenant_id": schema.StringAttribute{
				MarkdownDescription: "The Azure AD tenant ID. If not specified, defaults to the provider's configured tenant ID.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"company_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the Business Central company used for Automation API calls. " +
					"When not set the provider resolves it automatically by using the first company in the environment. " +
					"PTEs are published globally across all companies so the choice of company is only an implementation detail.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"environment_name": schema.StringAttribute{
				MarkdownDescription: "The name of the target environment. Changing this forces a new resource to be created.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"application_family": schema.StringAttribute{
				MarkdownDescription: "The application family of the environment (e.g. `\"BusinessCentral\"`). Changing this forces a new resource to be created.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"file_path": schema.StringAttribute{
				MarkdownDescription: "Local path to the `.app` file. Mutually exclusive with `file_content`.",
				Optional:            true,
			},
			"file_content": schema.StringAttribute{
				MarkdownDescription: "Base64-encoded `.app` file bytes. Mutually exclusive with `file_path`. " +
					"Enables passing content directly from a data source (e.g. `azurerm_storage_blob`).",
				Optional:  true,
				Sensitive: true,
			},
			"file_sha256": schema.StringAttribute{
				MarkdownDescription: "SHA-256 hash of the `.app` file content. Drives change detection — changing this value triggers an update.",
				Required:            true,
			},
			"schedule": schema.StringAttribute{
				MarkdownDescription: "Installation schedule. One of `\"Current version\"` (default), `\"Next minor version\"`, or `\"Next major version\"`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(DefaultSchedule),
				Validators: []validator.String{
					stringvalidator.OneOf("Current version", "Next minor version", "Next major version"),
				},
			},
			"schema_sync_mode": schema.StringAttribute{
				MarkdownDescription: "Schema synchronisation mode. One of `\"Add\"` (default) or `\"Force Sync\"`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(DefaultSchemaSyncMode),
				Validators: []validator.String{
					stringvalidator.OneOf("Add", "Force Sync"),
				},
			},
			"delete_data": schema.BoolAttribute{
				MarkdownDescription: "When `true`, calls `Microsoft.NAV.uninstallAndDeleteExtensionData` on destroy (irreversible). Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"unpublish_on_delete": schema.BoolAttribute{
				MarkdownDescription: "When `true`, calls `Microsoft.NAV.unpublish` after uninstall/update. Requires BC v25.4 or later. The call is silently skipped on older BC versions. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"package_id": schema.StringAttribute{
				MarkdownDescription: "`packageId` of the currently installed upload. Changes with every new upload.",
				Computed:            true,
			},
			"app_id": schema.StringAttribute{
				MarkdownDescription: "Stable extension identity (`id` field) that remains constant across version updates.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"display_name": schema.StringAttribute{
				MarkdownDescription: "Display name of the extension.",
				Computed:            true,
			},
			"publisher": schema.StringAttribute{
				MarkdownDescription: "Publisher of the extension.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"version": schema.StringAttribute{
				MarkdownDescription: "Installed version in `major.minor.build.revision` format.",
				Computed:            true,
			},
		},
	}
}

// Configure stores the provider client on the resource.
func (r *PerTenantExtensionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = c
}

// resolveFileBytes returns the raw .app bytes from either file_path or file_content.
func resolveFileBytes(data *PerTenantExtensionResourceModel) ([]byte, error) {
	if !data.FilePath.IsNull() && !data.FilePath.IsUnknown() && data.FilePath.ValueString() != "" {
		return os.ReadFile(data.FilePath.ValueString())
	}

	if !data.FileContent.IsNull() && !data.FileContent.IsUnknown() && data.FileContent.ValueString() != "" {
		raw, err := base64.StdEncoding.DecodeString(data.FileContent.ValueString())
		if err != nil {
			return nil, fmt.Errorf("failed to base64-decode file_content: %w", err)
		}
		return raw, nil
	}

	return nil, fmt.Errorf("exactly one of file_path or file_content must be set")
}

// validateFileInputs checks that exactly one of file_path / file_content is provided.
func validateFileInputs(data *PerTenantExtensionResourceModel) error {
	hasFilePath := !data.FilePath.IsNull() && !data.FilePath.IsUnknown() && data.FilePath.ValueString() != ""
	hasFileContent := !data.FileContent.IsNull() && !data.FileContent.IsUnknown() && data.FileContent.ValueString() != ""

	if hasFilePath && hasFileContent {
		return fmt.Errorf("exactly one of file_path or file_content must be set, but both are set")
	}

	if !hasFilePath && !hasFileContent {
		return fmt.Errorf("exactly one of file_path or file_content must be set, but neither is set")
	}

	return nil
}

// uploadAndInstall performs the 3-step PTE upload/install sequence and waits for completion.
// Returns the packageId of the uploaded extension.
func (r *PerTenantExtensionResource) uploadAndInstall(ctx context.Context, data *PerTenantExtensionResourceModel, svc *Service) (string, error) {
	fileBytes, err := resolveFileBytes(data)
	if err != nil {
		return "", fmt.Errorf("failed to read extension file: %w", err)
	}

	// Step 1: Create upload record.
	uploadReq := &ExtensionUploadRequest{
		Schedule:       data.Schedule.ValueString(),
		SchemaSyncMode: data.SchemaSyncMode.ValueString(),
	}

	uploadID, err := svc.CreateExtensionUpload(ctx, data.EnvironmentName.ValueString(), data.CompanyID.ValueString(), uploadReq)
	if err != nil {
		return "", fmt.Errorf("failed to create extension upload record: %w", err)
	}

	tflog.Debug(ctx, "Created extension upload record", map[string]interface{}{"upload_id": uploadID})

	// Step 2: Stream the .app file.
	if err := svc.UploadExtensionContent(ctx, data.EnvironmentName.ValueString(), data.CompanyID.ValueString(), uploadID, fileBytes); err != nil {
		return "", fmt.Errorf("failed to upload extension content: %w", err)
	}

	tflog.Debug(ctx, "Uploaded extension content")

	// Step 3: Trigger install.
	if err := svc.TriggerInstall(ctx, data.EnvironmentName.ValueString(), data.CompanyID.ValueString(), uploadID); err != nil {
		return "", fmt.Errorf("failed to trigger extension install: %w", err)
	}

	tflog.Debug(ctx, "Triggered extension install")

	// Step 4: Poll for completion.
	if _, err := svc.WaitForDeployment(ctx, data.EnvironmentName.ValueString(), data.CompanyID.ValueString(), 30*time.Minute); err != nil {
		return "", err
	}

	return uploadID, nil
}

// populateComputedFields reads extension details from BC and populates computed state fields.
func (r *PerTenantExtensionResource) populateComputedFields(ctx context.Context, data *PerTenantExtensionResourceModel, svc *Service, packageID string) error {
	ext, err := svc.GetExtensionByPackageID(ctx, data.EnvironmentName.ValueString(), data.CompanyID.ValueString(), packageID)
	if err != nil {
		return fmt.Errorf("failed to read extension details: %w", err)
	}

	if ext == nil {
		tflog.Warn(ctx, "Extension not found by packageId after install", map[string]interface{}{"package_id": packageID})
		return nil
	}

	data.PackageID = types.StringValue(ext.PackageID)
	data.AppID = types.StringValue(ext.ID)
	data.DisplayName = types.StringValue(ext.DisplayName)
	data.Publisher = types.StringValue(ext.Publisher)
	data.Version = types.StringValue(fmt.Sprintf("%d.%d.%d.%d",
		ext.VersionMajor, ext.VersionMinor, ext.VersionBuild, ext.VersionRevision))

	return nil
}

// Create creates the PTE resource.
func (r *PerTenantExtensionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PerTenantExtensionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := validateFileInputs(&data); err != nil {
		resp.Diagnostics.AddError("Invalid Configuration", err.Error())
		return
	}

	// Use provider tenant ID if not explicitly set.
	if data.AADTenantID.IsNull() || data.AADTenantID.ValueString() == "" {
		data.AADTenantID = types.StringValue(r.client.GetTenantID())
	}

	c := r.client.ForTenant(data.AADTenantID.ValueString())
	svc := NewService(c)

	// Resolve company ID if not provided.
	if data.CompanyID.IsNull() || data.CompanyID.ValueString() == "" {
		companyID, err := svc.GetFirstCompany(ctx, data.EnvironmentName.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to resolve company ID", err.Error())
			return
		}
		data.CompanyID = types.StringValue(companyID)
	}

	uploadID, err := r.uploadAndInstall(ctx, &data, svc)
	if err != nil {
		resp.Diagnostics.AddError("Failed to install per-tenant extension", err.Error())
		return
	}

	if err := r.populateComputedFields(ctx, &data, svc, uploadID); err != nil {
		resp.Diagnostics.AddError("Failed to read extension details after install", err.Error())
		return
	}

	data.ID = types.StringValue(BuildPerTenantExtensionID(
		data.AADTenantID.ValueString(),
		data.ApplicationFamily.ValueString(),
		data.EnvironmentName.ValueString(),
		data.AppID.ValueString(),
	))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the state from BC.
func (r *PerTenantExtensionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PerTenantExtensionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.client.ForTenant(data.AADTenantID.ValueString())
	svc := NewService(c)

	// If company_id is not set, try to resolve it.
	if data.CompanyID.IsNull() || data.CompanyID.ValueString() == "" {
		companyID, err := svc.GetFirstCompany(ctx, data.EnvironmentName.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to resolve company ID", err.Error())
			return
		}
		data.CompanyID = types.StringValue(companyID)
	}

	// Try to find the extension by stable appId first.
	if !data.AppID.IsNull() && data.AppID.ValueString() != "" {
		ext, err := svc.GetExtensionByAppID(ctx, data.EnvironmentName.ValueString(), data.CompanyID.ValueString(), data.AppID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to read per-tenant extension", err.Error())
			return
		}

		if ext == nil {
			// Extension no longer installed – remove from state.
			resp.State.RemoveResource(ctx)
			return
		}

		data.PackageID = types.StringValue(ext.PackageID)
		data.AppID = types.StringValue(ext.ID)
		data.DisplayName = types.StringValue(ext.DisplayName)
		data.Publisher = types.StringValue(ext.Publisher)
		data.Version = types.StringValue(fmt.Sprintf("%d.%d.%d.%d",
			ext.VersionMajor, ext.VersionMinor, ext.VersionBuild, ext.VersionRevision))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update uploads a new version of the PTE.
func (r *PerTenantExtensionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PerTenantExtensionResourceModel
	var state PerTenantExtensionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := validateFileInputs(&data); err != nil {
		resp.Diagnostics.AddError("Invalid Configuration", err.Error())
		return
	}

	// Preserve tenant/company from state.
	data.AADTenantID = state.AADTenantID
	data.CompanyID = state.CompanyID

	c := r.client.ForTenant(data.AADTenantID.ValueString())
	svc := NewService(c)

	oldPackageID := state.PackageID.ValueString()

	uploadID, err := r.uploadAndInstall(ctx, &data, svc)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update per-tenant extension", err.Error())
		return
	}

	if err := r.populateComputedFields(ctx, &data, svc, uploadID); err != nil {
		resp.Diagnostics.AddError("Failed to read extension details after update", err.Error())
		return
	}

	// Optionally unpublish the old package version.
	if data.UnpublishOnDelete.ValueBool() && oldPackageID != "" && oldPackageID != data.PackageID.ValueString() {
		if err := svc.Unpublish(ctx, data.EnvironmentName.ValueString(), data.CompanyID.ValueString(), oldPackageID); err != nil {
			tflog.Warn(ctx, "Failed to unpublish old extension package", map[string]interface{}{
				"old_package_id": oldPackageID,
				"error":          err.Error(),
			})
		}
	}

	data.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete uninstalls (and optionally unpublishes) the PTE.
func (r *PerTenantExtensionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PerTenantExtensionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.PackageID.IsNull() || data.PackageID.ValueString() == "" {
		// Nothing to uninstall.
		return
	}

	c := r.client.ForTenant(data.AADTenantID.ValueString())
	svc := NewService(c)

	if err := svc.Uninstall(ctx, data.EnvironmentName.ValueString(), data.CompanyID.ValueString(), data.PackageID.ValueString(), data.DeleteData.ValueBool()); err != nil {
		resp.Diagnostics.AddError("Failed to uninstall per-tenant extension", err.Error())
		return
	}

	if data.UnpublishOnDelete.ValueBool() {
		if err := svc.Unpublish(ctx, data.EnvironmentName.ValueString(), data.CompanyID.ValueString(), data.PackageID.ValueString()); err != nil {
			tflog.Warn(ctx, "Failed to unpublish extension after uninstall", map[string]interface{}{
				"package_id": data.PackageID.ValueString(),
				"error":      err.Error(),
			})
		}
	}
}

// ImportState imports a per-tenant extension by its ARM-like resource ID.
func (r *PerTenantExtensionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tenantID, appFamily, envName, appID, err := ParsePerTenantExtensionID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("aad_tenant_id"), tenantID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("application_family"), appFamily)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("environment_name"), envName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("app_id"), appID)...)
}

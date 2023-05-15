// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	oxideSDK "github.com/oxidecomputer/oxide.go/oxide"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = (*imageResource)(nil)
	_ resource.ResourceWithConfigure = (*imageResource)(nil)
)

// NewImageResource is a helper function to simplify the provider implementation.
func NewImageResource() resource.Resource {
	return &imageResource{}
}

// imageResource is the resource implementation.
type imageResource struct {
	client *oxideSDK.Client
}

type imageResourceModel struct {
	BlockSize        types.Int64    `tfsdk:"block_size"`
	Description      types.String   `tfsdk:"description"`
	Digest           types.Object   `tfsdk:"digest"`
	ID               types.String   `tfsdk:"id"`
	Name             types.String   `tfsdk:"name"`
	OS               types.String   `tfsdk:"os"`
	ProjectID        types.String   `tfsdk:"project_id"`
	Size             types.Int64    `tfsdk:"size"`
	SourceSnapshotID types.String   `tfsdk:"source_snapshot_id"`
	SourceURL        types.String   `tfsdk:"source_url"`
	TimeCreated      types.String   `tfsdk:"time_created"`
	TimeModified     types.String   `tfsdk:"time_modified"`
	Version          types.String   `tfsdk:"version"`
	Timeouts         timeouts.Value `tfsdk:"timeouts"`
}

type imageResourceDigestModel struct {
	Type  types.String `tfsdk:"type"`
	Value types.String `tfsdk:"value"`
}

// Metadata returns the resource type name.
func (r *imageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "oxide_image"
}

// Configure adds the provider configured client to the data source.
func (r *imageResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*oxideSDK.Client)
}

func (r *imageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Schema defines the schema for the resource.
func (r *imageResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the image.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Required:    true,
				Description: "Description for the image.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"os": schema.StringAttribute{
				Required:    true,
				Description: "OS image distribution. Example: alpine",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"version": schema.StringAttribute{
				Required:    true,
				Description: "OS image version. Example: 3.16.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"block_size": schema.Int64Attribute{
				Required:    true,
				Description: "Size of blocks in bytes.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"project_id": schema.StringAttribute{
				Optional:    true,
				Description: "ID of the project that will contain the image.",
				PlanModifiers: []planmodifier.String{
					ProjectIDImagePlanModifier(),
				},
			},
			"source_snapshot_id": schema.StringAttribute{
				Optional:    true,
				Description: "Snapshot ID of the image source if applicable.",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.Expressions{
						path.MatchRoot("source_url"),
					}...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source_url": schema.StringAttribute{
				Optional:    true,
				Description: "URL source of this image, if applicable.",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.Expressions{
						path.MatchRoot("source_snapshot_id"),
					}...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				// TODO: Restore once updates and deletes are enabled
				// Update: true,
				// Delete: true,
			}),
			"digest": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Hash of the image contents, if applicable.",
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Description: "Digest type.",
						Computed:    true,
					},
					"value": schema.StringAttribute{
						Description: "Digest type value.",
						Computed:    true,
					},
				},
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier of the image.",
			},
			"size": schema.Int64Attribute{
				Computed:    true,
				Description: "Total size in bytes.",
			},
			"time_created": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this image was created.",
			},
			"time_modified": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this image was last modified.",
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *imageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan imageResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := plan.Timeouts.Create(ctx, defaultTimeout())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	params := oxideSDK.ImageCreateParams{
		Project: oxideSDK.NameOrId(plan.ProjectID.ValueString()),
		Body: &oxideSDK.ImageCreate{
			Description: plan.Description.ValueString(),
			Name:        oxideSDK.Name(plan.Name.ValueString()),
			BlockSize:   oxideSDK.BlockSize(plan.BlockSize.ValueInt64()),
			Os:          plan.OS.ValueString(),
			Version:     plan.Version.ValueString(),
		},
	}

	is := oxideSDK.ImageSource{}
	if !plan.SourceSnapshotID.IsNull() {
		is.Id = plan.SourceSnapshotID.ValueString()
		is.Type = oxideSDK.ImageSourceTypeSnapshot
	} else if !plan.SourceURL.IsNull() {
		is.Id = plan.SourceURL.ValueString()
		is.Type = oxideSDK.ImageSourceTypeUrl
		// TODO: Remove before releasing, for testing purposes only
		if plan.SourceURL.Equal(types.StringValue("you_can_boot_anything_as_long_as_its_alpine")) {
			is.Type = oxideSDK.ImageSourceTypeYouCanBootAnythingAsLongAsItsAlpine
		}
	} else {
		resp.Diagnostics.AddError(
			"Error creating image",
			"One of `source_url` or `source_snapshot_id` must be set",
		)
		return
	}
	params.Body.Source = is

	image, err := r.client.ImageCreate(params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating image",
			"API error: "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("created image with ID: %v", image.Id), map[string]any{"success": true})

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(image.Id)
	plan.Size = types.Int64Value(int64(image.Size))
	plan.TimeCreated = types.StringValue(image.TimeCreated.String())
	plan.TimeModified = types.StringValue(image.TimeCreated.String())
	plan.Version = types.StringValue(image.Version)

	// Parse imageResourceDigestModel into types.Object
	dm := imageResourceDigestModel{
		Type:  types.StringValue(string(image.Digest.Type)),
		Value: types.StringValue(image.Digest.Value),
	}
	attributeTypes := map[string]attr.Type{
		"type":  types.StringType,
		"value": types.StringType,
	}
	digest, diags := types.ObjectValueFrom(ctx, attributeTypes, dm)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Digest = digest

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *imageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state imageResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	readTimeout, diags := state.Timeouts.Read(ctx, defaultTimeout())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()

	image, err := r.client.ImageView(oxideSDK.ImageViewParams{
		Image: oxideSDK.NameOrId(state.ID.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read image:",
			"API error: "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("read image with ID: %v", image.Id), map[string]any{"success": true})

	state.BlockSize = types.Int64Value(int64(image.BlockSize))
	state.Description = types.StringValue(image.Description)
	state.ID = types.StringValue(image.Id)
	state.Name = types.StringValue(string(image.Name))
	state.OS = types.StringValue(image.Os)
	state.Size = types.Int64Value(int64(image.Size))
	state.TimeCreated = types.StringValue(image.TimeCreated.String())
	state.TimeModified = types.StringValue(image.TimeCreated.String())
	state.Version = types.StringValue(image.Version)

	// Only set ProjectID and SourceURL if they exist to avoid unintentional drift
	if image.ProjectId != "" {
		state.ProjectID = types.StringValue(image.ProjectId)
	}
	if image.Url != "" {
		state.SourceURL = types.StringValue(image.Url)
	}

	// Parse imageResourceDigestModel into types.Object
	dm := imageResourceDigestModel{
		Type:  types.StringValue(string(image.Digest.Type)),
		Value: types.StringValue(image.Digest.Value),
	}
	attributeTypes := map[string]attr.Type{
		"type":  types.StringType,
		"value": types.StringType,
	}
	digest, diags := types.ObjectValueFrom(ctx, attributeTypes, dm)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Digest = digest

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *imageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan imageResourceModel
	var state imageResourceModel

	// Read Terraform plan data into the plan model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read Terraform prior state data into the state model to retrieve ID
	// which is a computed attribute, so it won't show up in the plan.
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout, diags := plan.Timeouts.Update(ctx, defaultTimeout())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	var image *oxideSDK.Image
	var err error
	if plan.ProjectID.IsNull() {
		image, err = r.client.ImagePromote(oxideSDK.ImagePromoteParams{
			Image: oxideSDK.NameOrId(state.ID.ValueString()),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating image",
				"API error: "+err.Error(),
			)
			return
		}
		tflog.Trace(ctx, fmt.Sprintf("promoted image with ID: %v to silo visibility", image.Id), map[string]any{"success": true})
	} else {
		image, err = r.client.ImageDemote(oxideSDK.ImageDemoteParams{
			Image:   oxideSDK.NameOrId(state.ID.ValueString()),
			Project: oxideSDK.NameOrId(plan.ProjectID.ValueString()),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating image",
				"API error: "+err.Error(),
			)
			return
		}
		tflog.Trace(ctx, fmt.Sprintf("demoted image with ID: %v to single project visibility", image.Id), map[string]any{"success": true})
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(image.Id)
	plan.Size = types.Int64Value(int64(image.Size))
	plan.TimeCreated = types.StringValue(image.TimeCreated.String())
	plan.TimeModified = types.StringValue(image.TimeModified.String())
	plan.Version = types.StringValue(image.Version)

	// Parse imageResourceDigestModel into types.Object
	dm := imageResourceDigestModel{
		Type:  types.StringValue(string(image.Digest.Type)),
		Value: types.StringValue(image.Digest.Value),
	}
	attributeTypes := map[string]attr.Type{
		"type":  types.StringType,
		"value": types.StringType,
	}
	digest, diags := types.ObjectValueFrom(ctx, attributeTypes, dm)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Digest = digest

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *imageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddError(
		"Error delete image",
		"the oxide API currently does not support deleting images")

	// TODO: Uncomment once image delete is enabled in the API
	//
	//	var state imageResourceModel
	//
	//	// Read Terraform prior state data into the model
	//	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	//	if resp.Diagnostics.HasError() {
	//		return
	//	}
	//
	//	if err := r.client.ImageDelete(oxideSDK.ImageDeleteParams{
	//		Image: oxideSDK.NameOrId(state.ID.ValueString()),
	//	}); err != nil {
	//
	//		resp.Diagnostics.AddError(
	//			"Unable to read image:",
	//			"API error: "+err.Error(),
	//		)
	//		return
	//	}
}

// ProjectIDImagePlanModifier is a helper function to simplify the provider implementation.
func ProjectIDImagePlanModifier() planmodifier.String {
	return &projectIDPlanModifier{}
}

type projectIDPlanModifier struct{}

func (d *projectIDPlanModifier) Description(ctx context.Context) string {
	return "Ensures that `project_id` can only be modified from UUID -> null and null -> UUID."
}

func (d *projectIDPlanModifier) MarkdownDescription(ctx context.Context) string {
	return d.Description(ctx)
}

func (d *projectIDPlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	var planProjectID types.String

	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("project_id"), &planProjectID)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var stateProjectID types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("project_id"), &stateProjectID)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Do nothing if there are no changes
	if stateProjectID.Equal(planProjectID) {
		return
	}

	// Only allow the following changes for ProjectID: UUID -> null and null -> UUID.
	// We cannot move images from one project to another directly yet
	if !stateProjectID.IsNull() && !planProjectID.IsNull() {
		resp.Diagnostics.AddError(
			"Unable to update image:",
			fmt.Sprintf("Please change the value for `project_id`,"+
				" '%v' is not a valid value. The only allowed updates for `project_id`"+
				" are: UUID -> null and null -> UUID", planProjectID.ValueString()),
		)
		return
	}
}

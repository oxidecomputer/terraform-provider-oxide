// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/oxidecomputer/oxide.go/oxide"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = (*diskResource)(nil)
	_ resource.ResourceWithConfigure = (*diskResource)(nil)
)

// NewDiskResource is a helper function to simplify the provider implementation.
func NewDiskResource() resource.Resource {
	return &diskResource{}
}

// diskResource is the resource implementation.
type diskResource struct {
	client *oxide.Client
}

type diskResourceModel struct {
	BlockSize        types.Int64    `tfsdk:"block_size"`
	Description      types.String   `tfsdk:"description"`
	DevicePath       types.String   `tfsdk:"device_path"`
	ID               types.String   `tfsdk:"id"`
	SourceImageID    types.String   `tfsdk:"source_image_id"`
	Name             types.String   `tfsdk:"name"`
	ProjectID        types.String   `tfsdk:"project_id"`
	Size             types.Int64    `tfsdk:"size"`
	SourceSnapshotID types.String   `tfsdk:"source_snapshot_id"`
	TimeCreated      types.String   `tfsdk:"time_created"`
	TimeModified     types.String   `tfsdk:"time_modified"`
	Timeouts         timeouts.Value `tfsdk:"timeouts"`
}

// Metadata returns the resource type name.
func (r *diskResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "oxide_disk"
}

// Configure adds the provider configured client to the data source.
func (r *diskResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*oxide.Client)
}

// ImportState imports an existing disk resource into Terraform state.
func (r *diskResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Schema defines the schema for the resource.
func (r *diskResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: replaceBackticks(`
This resource manages disks.

To create a blank disk it's necessary to set ''block_size''. Otherwise, one of ''source_image_id'' or ''source_snapshot_id'' must be set; ''block_size'' will be automatically calculated.

!> Disks cannot be deleted while attached to instances. Please detach or delete associated instances before attempting to delete.

-> This resource currently only provides create, read and delete actions. An update requires a resource replacement
`),
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the project that will contain the disk.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the disk.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"size": schema.Int64Attribute{
				Required:    true,
				Description: "Size of the disk in bytes.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Required:    true,
				Description: "Description for the disk.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source_image_id": schema.StringAttribute{
				Optional:    true,
				Description: "Image ID of the disk source if applicable.",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.Expressions{
						path.MatchRoot("block_size"),
					}...),
					stringvalidator.ConflictsWith(path.Expressions{
						path.MatchRoot("source_snapshot_id"),
					}...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source_snapshot_id": schema.StringAttribute{
				Optional:    true,
				Description: "Snapshot ID of the disk source if applicable.",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.Expressions{
						path.MatchRoot("block_size"),
					}...),
					stringvalidator.ConflictsWith(path.Expressions{
						path.MatchRoot("source_image_id"),
					}...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"block_size": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Size of blocks in bytes.",
				Validators: []validator.Int64{
					int64validator.ConflictsWith(path.Expressions{
						path.MatchRoot("source_image_id"),
					}...),
					int64validator.ConflictsWith(path.Expressions{
						path.MatchRoot("source_snapshot_id"),
					}...),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				// TODO: Restore once updates are enabled
				// Update: true,
				Delete: true,
			}),
			"device_path": schema.StringAttribute{
				Computed:    true,
				Description: "Path of the disk.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier of the disk.",
			},
			"time_created": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this disk was created.",
			},
			"time_modified": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this disk was last modified.",
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *diskResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan diskResourceModel

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

	params := oxide.DiskCreateParams{
		Project: oxide.NameOrId(plan.ProjectID.ValueString()),
		Body: &oxide.DiskCreate{
			Description: plan.Description.ValueString(),
			Name:        oxide.Name(plan.Name.ValueString()),
			Size:        oxide.ByteCount(plan.Size.ValueInt64()),
			DiskBackend: oxide.DiskBackend{
				// As of r18, disk type must be specified as
				// "distributed" (the only option in prior
				// releases) or "local". For now, always set
				// disk type to distributed. We'll support
				// local disks as well once r18 is available
				// for testing.
				Type: oxide.DiskBackendTypeDistributed,
			},
		},
	}

	ds := oxide.DiskSource{}
	if !plan.SourceImageID.IsNull() {
		ds.ImageId = plan.SourceImageID.ValueString()
		ds.Type = oxide.DiskSourceTypeImage
	} else if !plan.SourceSnapshotID.IsNull() {
		ds.SnapshotId = plan.SourceSnapshotID.ValueString()
		ds.Type = oxide.DiskSourceTypeSnapshot
	} else if !plan.BlockSize.IsNull() {
		ds.BlockSize = oxide.BlockSize(plan.BlockSize.ValueInt64())
		ds.Type = oxide.DiskSourceTypeBlank
	}
	params.Body.DiskBackend.DiskSource = ds

	disk, err := r.client.DiskCreate(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating disk",
			"API error: "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("created disk with ID: %v", disk.Id), map[string]any{"success": true})

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(disk.Id)
	plan.DevicePath = types.StringValue(disk.DevicePath)
	plan.BlockSize = types.Int64Value(int64(disk.BlockSize))
	plan.TimeCreated = types.StringValue(disk.TimeCreated.String())
	plan.TimeModified = types.StringValue(disk.TimeModified.String())

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *diskResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state diskResourceModel

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

	params := oxide.DiskViewParams{
		Disk: oxide.NameOrId(state.ID.ValueString()),
	}
	disk, err := r.client.DiskView(ctx, params)
	if err != nil {
		if is404(err) {
			// Remove resource from state during a refresh
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to read disk:",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("read disk with ID: %v", disk.Id), map[string]any{"success": true})

	state.BlockSize = types.Int64Value(int64(disk.BlockSize))
	state.Description = types.StringValue(disk.Description)
	state.DevicePath = types.StringValue(disk.DevicePath)
	state.ID = types.StringValue(disk.Id)
	state.Name = types.StringValue(string(disk.Name))
	state.ProjectID = types.StringValue(disk.ProjectId)
	state.Size = types.Int64Value(int64(disk.Size))
	state.TimeCreated = types.StringValue(disk.TimeCreated.String())
	state.TimeModified = types.StringValue(disk.TimeModified.String())

	// Only set SourceImageID and SourceSnapshotID if they've been set to avoid unintentional drift
	if disk.ImageId != "" {
		state.SourceImageID = types.StringValue(disk.ImageId)
	}
	if disk.SnapshotId != "" {
		state.SourceSnapshotID = types.StringValue(disk.SnapshotId)
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *diskResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Error updating disk",
		"the oxide API currently does not support updating disks")
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *diskResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state diskResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := state.Timeouts.Delete(ctx, defaultTimeout())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	params := oxide.DiskDeleteParams{
		Disk: oxide.NameOrId(state.ID.ValueString()),
	}
	if err := r.client.DiskDelete(ctx, params); err != nil {
		if !is404(err) {
			resp.Diagnostics.AddError(
				"Unable to delete disk:",
				"API error: "+err.Error(),
			)
			return
		}
	}

	tflog.Trace(ctx, fmt.Sprintf("deleted disk with ID: %v", state.ID.ValueString()), map[string]any{"success": true})
}

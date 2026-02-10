// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/oxidecomputer/oxide.go/oxide"
)

var (
	_ datasource.DataSource              = (*diskDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*diskDataSource)(nil)
)

// NewDiskDataSource is a helper function to simplify the provider implementation.
func NewDiskDataSource() datasource.DataSource {
	return &diskDataSource{}
}

// diskDataSource is the data source implementation.
type diskDataSource struct {
	client *oxide.Client
}

// diskDataSourceModel are the attributes that are supported on this data source.
type diskDataSourceModel struct {
	ID           types.String              `tfsdk:"id"`
	Name         types.String              `tfsdk:"name"`
	Description  types.String              `tfsdk:"description"`
	BlockSize    types.Int64               `tfsdk:"block_size"`
	DevicePath   types.String              `tfsdk:"device_path"`
	ProjectName  types.String              `tfsdk:"project_name"`
	ProjectID    types.String              `tfsdk:"project_id"`
	Size         types.Int64               `tfsdk:"size"`
	State        *diskDataSourceModelState `tfsdk:"state"`
	ImageID      types.String              `tfsdk:"image_id"`
	SnapshotID   types.String              `tfsdk:"snapshot_id"`
	TimeCreated  timetypes.RFC3339         `tfsdk:"time_created"`
	TimeModified timetypes.RFC3339         `tfsdk:"time_modified"`
	Timeouts     timeouts.Value            `tfsdk:"timeouts"`
}

// diskDataSourceModelState are the attributes for the disk state.
type diskDataSourceModelState struct {
	State    types.String `tfsdk:"state"`
	Instance types.String `tfsdk:"instance"`
}

// Metadata sets the resource type name.
func (d *diskDataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = "oxide_disk"
}

// Configure adds the provider configured client to the data source.
func (d *diskDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	_ *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*oxide.Client)
}

// Schema defines the schema for the data source.
func (d *diskDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
Retrieve information about a specified disk.
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier of the disk.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the disk.",
			},
			"project_name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the project that contains the disk.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description for the disk.",
			},
			"block_size": schema.Int64Attribute{
				Computed:    true,
				Description: "Size of blocks in bytes.",
			},
			"device_path": schema.StringAttribute{
				Computed:    true,
				Description: "Path of the disk.",
			},
			"project_id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the project that contains the disk.",
			},
			"size": schema.Int64Attribute{
				Computed:    true,
				Description: "Size of the disk in bytes.",
			},
			"image_id": schema.StringAttribute{
				Computed:    true,
				Description: "Image ID of the disk source if applicable.",
			},
			"snapshot_id": schema.StringAttribute{
				Computed:    true,
				Description: "Snapshot ID of the disk source if applicable.",
			},
			"state": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "State of the disk.",
				Attributes: map[string]schema.Attribute{
					"state": schema.StringAttribute{
						Computed:    true,
						Description: "The state of the disk (e.g., detached, attached, creating, etc.).",
					},
					"instance": schema.StringAttribute{
						Computed:    true,
						Description: "ID of the instance the disk is attached to, if any.",
					},
				},
			},
			"time_created": schema.StringAttribute{
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
				Description: "Timestamp of when this disk was created.",
			},
			"time_modified": schema.StringAttribute{
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
				Description: "Timestamp of when this disk was last modified.",
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *diskDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state diskDataSourceModel

	// Read Terraform configuration data into the model.
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
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
		Disk:    oxide.NameOrId(state.Name.ValueString()),
		Project: oxide.NameOrId(state.ProjectName.ValueString()),
	}
	disk, err := d.client.DiskView(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read disk:",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("read disk with ID: %v", disk.Id), map[string]any{"success": true})

	state.ID = types.StringValue(disk.Id)
	state.Name = types.StringValue(string(disk.Name))
	state.Description = types.StringValue(disk.Description)
	state.BlockSize = types.Int64Value(int64(disk.BlockSize))
	state.DevicePath = types.StringValue(disk.DevicePath)
	state.ProjectID = types.StringValue(disk.ProjectId)
	state.Size = types.Int64Value(int64(disk.Size))
	state.TimeCreated = timetypes.NewRFC3339TimeValue(disk.TimeCreated.UTC())
	state.TimeModified = timetypes.NewRFC3339TimeValue(disk.TimeModified.UTC())

	// Only set ImageID and SnapshotID if they are not empty
	if disk.ImageId != "" {
		state.ImageID = types.StringValue(disk.ImageId)
	}
	if disk.SnapshotId != "" {
		state.SnapshotID = types.StringValue(disk.SnapshotId)
	}

	// Set disk state
	state.State = &diskDataSourceModelState{
		State: types.StringValue(string(disk.State.State)),
	}
	if disk.State.Instance != "" {
		state.State.Instance = types.StringValue(disk.State.Instance)
	}

	// Save retrieved state into Terraform state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
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
	ID               types.String   `tfsdk:"id"`
	Name             types.String   `tfsdk:"name"`
	Description      types.String   `tfsdk:"description"`
	BlockSize        types.Int64    `tfsdk:"block_size"`
	DevicePath       types.String   `tfsdk:"device_path"`
	ProjectID        types.String   `tfsdk:"project_id"`
	Size             types.Int64    `tfsdk:"size"`
	SourceImageID    types.String   `tfsdk:"source_image_id"`
	SourceSnapshotID types.String   `tfsdk:"source_snapshot_id"`
	TimeCreated      types.String   `tfsdk:"time_created"`
	TimeModified     types.String   `tfsdk:"time_modified"`
	Timeouts         timeouts.Value `tfsdk:"timeouts"`
}

// Metadata sets the resource type name.
func (d *diskDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "oxide_disk"
}

// Configure adds the provider configured client to the data source.
func (d *diskDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*oxide.Client)
}

// Schema defines the schema for the data source.
func (d *diskDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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
			"source_image_id": schema.StringAttribute{
				Computed:    true,
				Description: "Image ID of the disk source if applicable.",
			},
			"source_snapshot_id": schema.StringAttribute{
				Computed:    true,
				Description: "Snapshot ID of the disk source if applicable.",
			},
			"time_created": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this disk was created.",
			},
			"time_modified": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this disk was last modified.",
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *diskDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
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
		Disk: oxide.NameOrId(state.Name.ValueString()),
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
	state.TimeCreated = types.StringValue(disk.TimeCreated.String())
	state.TimeModified = types.StringValue(disk.TimeModified.String())

	// Only set SourceImageID and SourceSnapshotID if they are not empty
	if disk.ImageId != "" {
		state.SourceImageID = types.StringValue(disk.ImageId)
	}
	if disk.SnapshotId != "" {
		state.SourceSnapshotID = types.StringValue(disk.SnapshotId)
	}

	// Save retrieved state into Terraform state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

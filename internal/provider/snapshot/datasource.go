// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package snapshot

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/oxidecomputer/oxide.go/oxide"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/shared"
)

var (
	_ datasource.DataSource              = (*DataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*DataSource)(nil)
)

// NewDataSource is a helper function to simplify the provider implementation.
func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

// DataSource is the data source implementation.
type DataSource struct {
	client *oxide.Client
}

// DataSourceModel are the attributes that are supported on this data source.
type DataSourceModel struct {
	Description  types.String   `tfsdk:"description"`
	DiskID       types.String   `tfsdk:"disk_id"`
	ID           types.String   `tfsdk:"id"`
	Name         types.String   `tfsdk:"name"`
	ProjectID    types.String   `tfsdk:"project_id"`
	ProjectName  types.String   `tfsdk:"project_name"`
	Size         types.Int64    `tfsdk:"size"`
	State        types.String   `tfsdk:"state"`
	TimeCreated  types.String   `tfsdk:"time_created"`
	TimeModified types.String   `tfsdk:"time_modified"`
	Timeouts     timeouts.Value `tfsdk:"timeouts"`
}

// Metadata sets the data source type name.
func (d *DataSource) Metadata(
	_ context.Context,
	_ datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = "oxide_snapshot"
}

// Configure adds the provider configured client to the data source.
func (d *DataSource) Configure(
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
func (d *DataSource) Schema(
	ctx context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
Retrieve information about a specified snapshot.
`,
		Attributes: map[string]schema.Attribute{
			"project_name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the project that contains the snapshot.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the snapshot.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description for the snapshot.",
			},
			"disk_id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the disk used to create the snapshot.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier of the snapshot.",
			},
			"project_id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the project that contains the snapshot.",
			},
			"size": schema.Int64Attribute{
				Computed:    true,
				Description: "Size of the snapshot in bytes.",
			},
			"state": schema.StringAttribute{
				Computed:    true,
				Description: "State of the snapshot.",
			},
			"time_created": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this snapshot was created.",
			},
			"time_modified": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this snapshot was last modified.",
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *DataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state DataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	readTimeout, diags := state.Timeouts.Read(ctx, shared.DefaultTimeout())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()

	snapshot, err := d.client.SnapshotView(ctx, oxide.SnapshotViewParams{
		Snapshot: oxide.NameOrId(state.Name.ValueString()),
		Project:  oxide.NameOrId(state.ProjectName.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read snapshot:",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(
		ctx,
		fmt.Sprintf("read snapshot with ID: %v", snapshot.Id),
		map[string]any{"success": true},
	)

	state.Description = types.StringValue(snapshot.Description)
	state.DiskID = types.StringValue(snapshot.DiskId)
	state.ID = types.StringValue(snapshot.Id)
	state.Name = types.StringValue(string(snapshot.Name))
	state.ProjectID = types.StringValue(snapshot.ProjectId)
	state.Size = types.Int64Value(int64(snapshot.Size))
	state.State = types.StringValue(string(snapshot.State))
	state.TimeCreated = types.StringValue(snapshot.TimeCreated.String())
	state.TimeModified = types.StringValue(snapshot.TimeModified.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

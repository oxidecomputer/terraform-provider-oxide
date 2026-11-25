// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package snapshots

import (
	"context"
	"fmt"

	"github.com/google/uuid"
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

type DataSource struct {
	client *oxide.Client
}

type DataSourceModel struct {
	ID        types.String              `tfsdk:"id"`
	Project   types.String              `tfsdk:"project"`
	Snapshots []SnapshotDataSourceModel `tfsdk:"snapshots"`
	Timeouts  timeouts.Value            `tfsdk:"timeouts"`
}

type SnapshotDataSourceModel struct {
	Description  types.String `tfsdk:"description"`
	DiskID       types.String `tfsdk:"disk_id"`
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	ProjectID    types.String `tfsdk:"project_id"`
	Size         types.Int64  `tfsdk:"size"`
	State        types.String `tfsdk:"state"`
	TimeCreated  types.String `tfsdk:"time_created"`
	TimeModified types.String `tfsdk:"time_modified"`
}

// NewDataSource initialises a snapshots data source.
func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

func (d *DataSource) Metadata(
	_ context.Context,
	_ datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = "oxide_snapshots"
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

func (d *DataSource) Schema(
	ctx context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieve all snapshots in a project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"project": schema.StringAttribute{
				Required:    true,
				Description: "Name or ID of the project containing the snapshots.",
			},
			"snapshots": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"description": schema.StringAttribute{
							Computed:    true,
							Description: "Description for the snapshot.",
						},
						"disk_id": schema.StringAttribute{
							Computed:    true,
							Description: "ID of the disk from which the snapshot was created.",
						},
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "Unique, immutable, system-controlled identifier of the snapshot.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Name of the snapshot.",
						},
						"project_id": schema.StringAttribute{
							Computed:    true,
							Description: "ID of the project containing the snapshot.",
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
					},
				},
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

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

	params := oxide.SnapshotListParams{
		Project: oxide.NameOrId(state.Project.ValueString()),
		SortBy:  oxide.NameOrIdSortModeIdAscending,
	}
	snapshots, err := d.client.SnapshotListAllPages(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read snapshots list:",
			err.Error(),
		)
		return
	}

	tflog.Trace(
		ctx,
		fmt.Sprintf("read all snapshots from project: %v", state.Project.ValueString()),
		map[string]any{"success": true},
	)

	state.ID = types.StringValue(uuid.New().String())
	state.Snapshots = make([]SnapshotDataSourceModel, len(snapshots))
	for i, snapshot := range snapshots {
		state.Snapshots[i] = SnapshotDataSourceModel{
			Description:  types.StringValue(snapshot.Description),
			DiskID:       types.StringValue(snapshot.DiskId),
			ID:           types.StringValue(snapshot.Id),
			Name:         types.StringValue(string(snapshot.Name)),
			ProjectID:    types.StringValue(snapshot.ProjectId),
			Size:         types.Int64Value(int64(snapshot.Size)),
			State:        types.StringValue(string(snapshot.State)),
			TimeCreated:  types.StringValue(snapshot.TimeCreated.String()),
			TimeModified: types.StringValue(snapshot.TimeModified.String()),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

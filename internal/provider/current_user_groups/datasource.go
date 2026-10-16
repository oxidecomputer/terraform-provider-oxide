// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package currentusergroups

import (
	"context"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/oxidecomputer/oxide.go/oxide"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/shared"
)

var _ datasource.DataSource = (*DataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*DataSource)(nil)

type DataSource struct {
	client *oxide.Client
}

type DataSourceModel struct {
	Groups   []GroupDataSourceModel `tfsdk:"groups"`
	ID       types.String           `tfsdk:"id"`
	Timeouts timeouts.Value         `tfsdk:"timeouts"`
}

type GroupDataSourceModel struct {
	DisplayName  types.String `tfsdk:"display_name"`
	ID           types.String `tfsdk:"id"`
	SiloID       types.String `tfsdk:"silo_id"`
	TimeCreated  types.String `tfsdk:"time_created"`
	TimeModified types.String `tfsdk:"time_modified"`
}

// NewDataSource constructs the data source.
func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

// Metadata returns the data source type name.
func (d *DataSource) Metadata(
	_ context.Context,
	_ datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = "oxide_current_user_groups"
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
		MarkdownDescription: "Retrieve the groups to which the currently authenticated user belongs.",
		Attributes: map[string]schema.Attribute{
			"groups": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"display_name": schema.StringAttribute{
							Computed:    true,
							Description: "Human-readable name that identifies the group.",
						},
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "Unique, immutable, system-controlled identifier of the group.",
						},
						"silo_id": schema.StringAttribute{
							Computed:    true,
							Description: "Unique, immutable, system-controlled identifier of the silo to which the group belongs.",
						},
						"time_created": schema.StringAttribute{
							Computed:    true,
							Description: "Timestamp of when the group was created.",
						},
						"time_modified": schema.StringAttribute{
							Computed:    true,
							Description: "Timestamp of when the group was last modified.",
						},
					},
				},
			},
			"id": schema.StringAttribute{
				Computed: true,
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

	groups, err := d.client.CurrentUserGroupsAllPages(ctx, oxide.CurrentUserGroupsParams{
		SortBy: oxide.IdSortModeIdAscending,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read current user groups:",
			err.Error(),
		)
		return
	}

	tflog.Trace(ctx, "read all current user groups", map[string]any{"success": true})

	state.ID = types.StringValue(uuid.New().String())
	state.Groups = make([]GroupDataSourceModel, 0, len(groups))
	for _, group := range groups {
		state.Groups = append(state.Groups, GroupDataSourceModel{
			DisplayName:  types.StringValue(group.DisplayName),
			ID:           types.StringValue(group.Id),
			SiloID:       types.StringValue(group.SiloId),
			TimeCreated:  types.StringValue(group.TimeCreated.String()),
			TimeModified: types.StringValue(group.TimeModified.String()),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

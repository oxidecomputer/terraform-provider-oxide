// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/oxidecomputer/oxide.go/oxide"
)

var _ datasource.DataSource = (*systemIpPoolsDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*systemIpPoolsDataSource)(nil)

type systemIpPoolsDataSource struct {
	client *oxide.Client
}

type systemIpPoolsDataSourceModel struct {
	ID       types.String        `tfsdk:"id"`
	Timeouts timeouts.Value      `tfsdk:"timeouts"`
	IpPools  []systemIpPoolModel `tfsdk:"ip_pools"`
}

type systemIpPoolModel struct {
	Description  types.String `tfsdk:"description"`
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	TimeCreated  types.String `tfsdk:"time_created"`
	TimeModified types.String `tfsdk:"time_modified"`
}

// NewSystemIpPoolsDataSource initialises a system_ip_pools data source.
func NewSystemIpPoolsDataSource() datasource.DataSource {
	return &systemIpPoolsDataSource{}
}

func (d *systemIpPoolsDataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = "oxide_system_ip_pools"
}

// Configure adds the provider configured client to the data source.
func (d *systemIpPoolsDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	_ *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*oxide.Client)
}

func (d *systemIpPoolsDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
Retrieve all configured IP pools for the Oxide system.
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"timeouts": timeouts.Attributes(ctx),
			"ip_pools": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Name of the IP pool.",
						},
						"description": schema.StringAttribute{
							Computed:    true,
							Description: "Description for the IP pool.",
						},
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "Unique, immutable, system-controlled identifier of the IP pool.",
						},
						"time_created": schema.StringAttribute{
							Computed:    true,
							Description: "Timestamp of when this IP pool was created.",
						},
						"time_modified": schema.StringAttribute{
							Computed:    true,
							Description: "Timestamp of when this IP pool was last modified.",
						},
					},
				},
			},
		},
	}
}

func (d *systemIpPoolsDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state systemIpPoolsDataSourceModel

	// Read Terraform configuration data into the model
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

	params := oxide.IpPoolListParams{
		SortBy: oxide.NameOrIdSortModeIdAscending,
	}
	ipPools, err := d.client.IpPoolListAllPages(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read system IP pools list:",
			err.Error(),
		)
		return
	}

	tflog.Trace(ctx, "read all pools from system")

	// Set a unique ID for the data source payload
	state.ID = types.StringValue(uuid.New().String())

	for _, ipPool := range ipPools {
		poolState := systemIpPoolModel{
			Description:  types.StringValue(ipPool.Description),
			ID:           types.StringValue(ipPool.Id),
			Name:         types.StringValue(string(ipPool.Name)),
			TimeCreated:  types.StringValue(ipPool.TimeCreated.String()),
			TimeModified: types.StringValue(ipPool.TimeCreated.String()),
		}

		state.IpPools = append(state.IpPools, poolState)
	}

	// Save state into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

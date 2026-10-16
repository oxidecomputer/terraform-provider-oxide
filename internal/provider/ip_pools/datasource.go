// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ippools

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
	ID       types.String            `tfsdk:"id"`
	IPPools  []IPPoolDataSourceModel `tfsdk:"ip_pools"`
	Timeouts timeouts.Value          `tfsdk:"timeouts"`
}

type IPPoolDataSourceModel struct {
	Description  types.String `tfsdk:"description"`
	ID           types.String `tfsdk:"id"`
	IPVersion    types.String `tfsdk:"ip_version"`
	IsDefault    types.Bool   `tfsdk:"is_default"`
	Name         types.String `tfsdk:"name"`
	PoolType     types.String `tfsdk:"pool_type"`
	TimeCreated  types.String `tfsdk:"time_created"`
	TimeModified types.String `tfsdk:"time_modified"`
}

// NewDataSource initialises an ip_pools data source.
func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

func (d *DataSource) Metadata(
	_ context.Context,
	_ datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = "oxide_ip_pools"
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
		MarkdownDescription: "Retrieve all IP pools available to the current silo.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"ip_pools": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"description": schema.StringAttribute{
							Computed:    true,
							Description: "Description for the IP pool.",
						},
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "Unique, immutable, system-controlled identifier of the IP pool.",
						},
						"ip_version": schema.StringAttribute{
							Computed:    true,
							Description: "IP version for the pool.",
						},
						"is_default": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the pool is the default for the current silo.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Name of the IP pool.",
						},
						"pool_type": schema.StringAttribute{
							Computed:    true,
							Description: "Pool type for the IP pool.",
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

	ipPools, err := d.client.IpPoolListAllPages(ctx, oxide.IpPoolListParams{
		SortBy: oxide.NameOrIdSortModeIdAscending,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read IP pools list:",
			err.Error(),
		)
		return
	}

	tflog.Trace(ctx, "read all IP pools available to the current silo")

	state.ID = types.StringValue(uuid.New().String())
	state.IPPools = make([]IPPoolDataSourceModel, len(ipPools))
	for i, ipPool := range ipPools {
		state.IPPools[i] = IPPoolDataSourceModel{
			Description:  types.StringValue(ipPool.Description),
			ID:           types.StringValue(ipPool.Id),
			IPVersion:    types.StringValue(string(ipPool.IpVersion)),
			IsDefault:    types.BoolPointerValue(ipPool.IsDefault),
			Name:         types.StringValue(string(ipPool.Name)),
			PoolType:     types.StringValue(string(ipPool.PoolType)),
			TimeCreated:  types.StringValue(ipPool.TimeCreated.String()),
			TimeModified: types.StringValue(ipPool.TimeModified.String()),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

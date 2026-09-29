// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package systemsubnetpools

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

const (
	dataSourceTypeName           = "oxide_system_subnet_pools"
	deprecatedDataSourceTypeName = "oxide_subnet_pools"
	deprecationMessage           = "Use oxide_system_subnet_pools instead. " +
		"The oxide_subnet_pools data source will be removed in a future release."
)

type DataSource struct {
	client             *oxide.Client
	typeName           string
	deprecationMessage string
}

type DataSourceModel struct {
	ID          types.String                `tfsdk:"id"`
	SubnetPools []SubnetPoolDataSourceModel `tfsdk:"subnet_pools"`
	Timeouts    timeouts.Value              `tfsdk:"timeouts"`
}

type SubnetPoolDataSourceModel struct {
	Description  types.String `tfsdk:"description"`
	ID           types.String `tfsdk:"id"`
	IpVersion    types.String `tfsdk:"ip_version"`
	Name         types.String `tfsdk:"name"`
	TimeCreated  types.String `tfsdk:"time_created"`
	TimeModified types.String `tfsdk:"time_modified"`
}

// NewDataSource initialises a system_subnet_pools data source.
func NewDataSource() datasource.DataSource {
	return &DataSource{typeName: dataSourceTypeName}
}

// NewDeprecatedDataSource initialises the deprecated subnet_pools data source.
func NewDeprecatedDataSource() datasource.DataSource {
	return &DataSource{
		typeName:           deprecatedDataSourceTypeName,
		deprecationMessage: deprecationMessage,
	}
}

func (d *DataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = d.typeName
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
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	markdownDescription := "Retrieve all configured subnet pools for the Oxide system."
	if d.deprecationMessage != "" {
		markdownDescription = "-> **Deprecated:** Use the `oxide_system_subnet_pools` data source instead.\n\n" +
			markdownDescription
	}

	resp.Schema = schema.Schema{
		DeprecationMessage:  d.deprecationMessage,
		MarkdownDescription: markdownDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"subnet_pools": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"description": schema.StringAttribute{
							Computed:    true,
							Description: "Description for the subnet pool.",
						},
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "Unique, immutable, system-controlled identifier of the subnet pool.",
						},
						"ip_version": schema.StringAttribute{
							Computed:    true,
							Description: "The IP version for this pool (v4 or v6).",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Name of the subnet pool.",
						},
						"time_created": schema.StringAttribute{
							Computed:    true,
							Description: "Timestamp of when this subnet pool was created.",
						},
						"time_modified": schema.StringAttribute{
							Computed:    true,
							Description: "Timestamp of when this subnet pool was last modified.",
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

	params := oxide.SystemSubnetPoolListParams{
		SortBy: oxide.NameOrIdSortModeIdAscending,
	}
	subnetPools, err := d.client.SystemSubnetPoolListAllPages(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read system subnet pools list:",
			err.Error(),
		)
		return
	}

	tflog.Trace(ctx, "read all system subnet pools")

	state.ID = types.StringValue(uuid.New().String())
	state.SubnetPools = make([]SubnetPoolDataSourceModel, len(subnetPools))
	for i, subnetPool := range subnetPools {
		state.SubnetPools[i] = SubnetPoolDataSourceModel{
			Description:  types.StringValue(subnetPool.Description),
			ID:           types.StringValue(subnetPool.Id),
			IpVersion:    types.StringValue(string(subnetPool.IpVersion)),
			Name:         types.StringValue(string(subnetPool.Name)),
			TimeCreated:  types.StringValue(subnetPool.TimeCreated.String()),
			TimeModified: types.StringValue(subnetPool.TimeModified.String()),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

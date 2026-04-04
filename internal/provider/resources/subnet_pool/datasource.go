// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package subnet_pool

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

// DataSource is the data source implementation.
type DataSource struct {
	client *oxide.Client
}

// DataSourceModel are the attributes that are supported on this
// data source.
type DataSourceModel struct {
	Description  types.String            `tfsdk:"description"`
	ID           types.String            `tfsdk:"id"`
	IpVersion    types.String            `tfsdk:"ip_version"`
	Name         types.String            `tfsdk:"name"`
	Members      []DataSourceMemberModel `tfsdk:"members"`
	Timeouts     timeouts.Value          `tfsdk:"timeouts"`
	TimeCreated  types.String            `tfsdk:"time_created"`
	TimeModified types.String            `tfsdk:"time_modified"`
}

// DataSourceMemberModel represents a subnet pool member in the data
// source.
type DataSourceMemberModel struct {
	Subnet          types.String `tfsdk:"subnet"`
	MinPrefixLength types.Int64  `tfsdk:"min_prefix_length"`
	MaxPrefixLength types.Int64  `tfsdk:"max_prefix_length"`
}

// NewDataSource initialises a subnet_pool datasource.
func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

// Metadata returns the data source type name.
func (d *DataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = "oxide_subnet_pool"
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
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieve information about a" +
			" specified subnet pool.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the subnet pool.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description for the subnet pool.",
			},
			"id": schema.StringAttribute{
				Computed: true,
				Description: "Unique, immutable," +
					" system-controlled identifier of" +
					" the subnet pool.",
			},
			"ip_version": schema.StringAttribute{
				Computed: true,
				Description: "The IP version for this pool" +
					" (v4 or v6).",
			},
			"members": schema.ListNestedAttribute{
				Computed: true,
				Description: "List of subnet members in the" +
					" pool. Members are managed via the" +
					" `oxide_subnet_pool_member` resource.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"subnet": schema.StringAttribute{
							Computed:    true,
							Description: "The subnet CIDR.",
						},
						"min_prefix_length": schema.Int64Attribute{
							Computed: true,
							Description: "Minimum prefix" +
								" length for allocations" +
								" from this subnet.",
						},
						"max_prefix_length": schema.Int64Attribute{
							Computed: true,
							Description: "Maximum prefix" +
								" length for allocations" +
								" from this subnet.",
						},
					},
				},
			},
			"time_created": schema.StringAttribute{
				Computed: true,
				Description: "Timestamp of when this subnet" +
					" pool was created.",
			},
			"time_modified": schema.StringAttribute{
				Computed: true,
				Description: "Timestamp of when this subnet" +
					" pool was last modified.",
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

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	readTimeout, diags := state.Timeouts.Read(
		ctx, shared.DefaultTimeout(),
	)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()

	params := oxide.SystemSubnetPoolViewParams{
		Pool: oxide.NameOrId(state.Name.ValueString()),
	}
	pool, err := d.client.SystemSubnetPoolView(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read subnet pool:",
			err.Error(),
		)
		return
	}

	tflog.Trace(
		ctx,
		fmt.Sprintf(
			"read subnet pool with ID: %v", pool.Id,
		),
		map[string]any{"success": true},
	)

	// Map response body to model
	state.Description = types.StringValue(pool.Description)
	state.ID = types.StringValue(pool.Id)
	state.IpVersion = types.StringValue(string(pool.IpVersion))
	state.Name = types.StringValue(string(pool.Name))
	state.TimeCreated = types.StringValue(
		pool.TimeCreated.String(),
	)
	state.TimeModified = types.StringValue(
		pool.TimeModified.String(),
	)

	// Read members
	members, err := d.client.SystemSubnetPoolMemberListAllPages(
		ctx,
		oxide.SystemSubnetPoolMemberListParams{
			Pool: oxide.NameOrId(pool.Id),
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read subnet pool members:",
			err.Error(),
		)
		return
	}

	tflog.Trace(
		ctx,
		fmt.Sprintf(
			"read %d subnet pool members from pool with ID: %v",
			len(members), pool.Id,
		),
		map[string]any{"success": true},
	)

	state.Members = make(
		[]DataSourceMemberModel, len(members),
	)
	for i, member := range members {
		state.Members[i] = DataSourceMemberModel{
			Subnet: types.StringValue(
				member.Subnet.String(),
			),
			MinPrefixLength: types.Int64Value(
				int64(*member.MinPrefixLength),
			),
			MaxPrefixLength: types.Int64Value(
				int64(*member.MaxPrefixLength),
			),
		}
	}

	// Save state into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

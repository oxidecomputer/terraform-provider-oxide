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
	_ datasource.DataSource              = (*vpcSubnetDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*vpcSubnetDataSource)(nil)
)

// NewVPCSubnetDataSource initialises an images datasource
func NewVPCSubnetDataSource() datasource.DataSource {
	return &vpcSubnetDataSource{}
}

type vpcSubnetDataSource struct {
	client *oxide.Client
}

type vpcSubnetDataSourceModel struct {
	Description  types.String      `tfsdk:"description"`
	ID           types.String      `tfsdk:"id"`
	IPV4Block    types.String      `tfsdk:"ipv4_block"`
	IPV6Block    types.String      `tfsdk:"ipv6_block"`
	Name         types.String      `tfsdk:"name"`
	ProjectName  types.String      `tfsdk:"project_name"`
	VPCID        types.String      `tfsdk:"vpc_id"`
	VPCName      types.String      `tfsdk:"vpc_name"`
	TimeCreated  timetypes.RFC3339 `tfsdk:"time_created"`
	TimeModified timetypes.RFC3339 `tfsdk:"time_modified"`
	Timeouts     timeouts.Value    `tfsdk:"timeouts"`
}

func (d *vpcSubnetDataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = "oxide_vpc_subnet"
}

// Configure adds the provider configured client to the data source.
func (d *vpcSubnetDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	_ *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*oxide.Client)
}

func (d *vpcSubnetDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
Retrieve information about a specified VPC subnet.
`,
		Attributes: map[string]schema.Attribute{
			"project_name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the project that contains the subnet.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the subnet.",
			},
			"vpc_name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the VPC that contains the subnet.",
			},
			"vpc_id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the VPC that contains the subnet.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description for the VPC subnet.",
			},
			"ipv4_block": schema.StringAttribute{
				Computed: true,
				Description: "IPv4 address range for this VPC subnet. " +
					"It must be allocated from an RFC 1918 private address range, " +
					"and must not overlap with any other existing subnet in the VPC.",
			},
			"ipv6_block": schema.StringAttribute{
				Computed: true,
				Description: "IPv6 address range for this VPC subnet. " +
					"It must be allocated from the RFC 4193 Unique Local Address range, " +
					"with the prefix equal to the parent VPC's prefix. ",
			},
			"timeouts": timeouts.Attributes(ctx),
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier of the VPC subnet.",
			},
			"time_created": schema.StringAttribute{
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
				Description: "Timestamp of when this VPC subnet was created.",
			},
			"time_modified": schema.StringAttribute{
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
				Description: "Timestamp of when this VPC subnet was last modified.",
			},
		},
	}
}

func (d *vpcSubnetDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state vpcSubnetDataSourceModel

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

	params := oxide.VpcSubnetViewParams{
		Subnet:  oxide.NameOrId(state.Name.ValueString()),
		Vpc:     oxide.NameOrId(state.VPCName.ValueString()),
		Project: oxide.NameOrId(state.ProjectName.ValueString()),
	}
	subnet, err := d.client.VpcSubnetView(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read VPC subnet:",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(
		ctx,
		fmt.Sprintf("read VPC subnet with ID: %v", subnet.Id),
		map[string]any{"success": true},
	)

	state.Description = types.StringValue(subnet.Description)
	state.ID = types.StringValue(subnet.Id)
	state.IPV4Block = types.StringValue(string(subnet.Ipv4Block))
	state.IPV6Block = types.StringValue(string(subnet.Ipv6Block))
	state.Name = types.StringValue(string(subnet.Name))
	state.VPCID = types.StringValue(subnet.VpcId)
	state.TimeCreated = timetypes.NewRFC3339TimeValue(subnet.TimeCreated.UTC())
	state.TimeModified = timetypes.NewRFC3339TimeValue(subnet.TimeModified.UTC())

	// Save state into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

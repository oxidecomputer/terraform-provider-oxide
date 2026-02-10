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
	_ datasource.DataSource              = (*vpcInternetGatewayDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*vpcInternetGatewayDataSource)(nil)
)

// NewVPCInternetGatewayDataSource initialises an images datasource
func NewVPCInternetGatewayDataSource() datasource.DataSource {
	return &vpcInternetGatewayDataSource{}
}

type vpcInternetGatewayDataSource struct {
	client *oxide.Client
}

type vpcInternetGatewayDataSourceModel struct {
	Description  types.String      `tfsdk:"description"`
	ID           types.String      `tfsdk:"id"`
	Name         types.String      `tfsdk:"name"`
	ProjectName  types.String      `tfsdk:"project_name"`
	VPCID        types.String      `tfsdk:"vpc_id"`
	VPCName      types.String      `tfsdk:"vpc_name"`
	TimeCreated  timetypes.RFC3339 `tfsdk:"time_created"`
	TimeModified timetypes.RFC3339 `tfsdk:"time_modified"`
	Timeouts     timeouts.Value    `tfsdk:"timeouts"`
}

func (d *vpcInternetGatewayDataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = "oxide_vpc_internet_gateway"
}

// Configure adds the provider configured client to the data source.
func (d *vpcInternetGatewayDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	_ *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*oxide.Client)
}

func (d *vpcInternetGatewayDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
Retrieve information about a specified VPC internet gateway.
`,
		Attributes: map[string]schema.Attribute{
			"project_name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the project that contains the VPC internet gateway.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the VPC internet gateway.",
			},
			"vpc_name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the VPC that contains the VPC internet gateway.",
			},
			"vpc_id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the VPC that contains the VPC internet gateway.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description for the VPC internet gateway.",
			},
			"timeouts": timeouts.Attributes(ctx),
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier of the VPC.",
			},
			"time_created": schema.StringAttribute{
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
				Description: "Timestamp of when this VPC internet gateway was created.",
			},
			"time_modified": schema.StringAttribute{
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
				Description: "Timestamp of when this VPC internet gateway was last modified.",
			},
		},
	}
}

func (d *vpcInternetGatewayDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state vpcInternetGatewayDataSourceModel

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

	params := oxide.InternetGatewayViewParams{
		Gateway: oxide.NameOrId(state.Name.ValueString()),
		Vpc:     oxide.NameOrId(state.VPCName.ValueString()),
		Project: oxide.NameOrId(state.ProjectName.ValueString()),
	}
	vpcInternetGateway, err := d.client.InternetGatewayView(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read VPC internet gateway:",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(
		ctx,
		fmt.Sprintf("read VPC internet gateway with ID: %v", vpcInternetGateway.Id),
		map[string]any{"success": true},
	)

	state.Description = types.StringValue(vpcInternetGateway.Description)
	state.ID = types.StringValue(vpcInternetGateway.Id)
	state.Name = types.StringValue(string(vpcInternetGateway.Name))
	state.VPCID = types.StringValue(vpcInternetGateway.VpcId)
	state.TimeCreated = timetypes.NewRFC3339TimeValue(vpcInternetGateway.TimeCreated.UTC())
	state.TimeModified = timetypes.NewRFC3339TimeValue(vpcInternetGateway.TimeModified.UTC())

	// Save state into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

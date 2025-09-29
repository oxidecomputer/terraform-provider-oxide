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
	_ datasource.DataSource              = (*vpcDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*vpcDataSource)(nil)
)

// NewVPCDataSource initialises an images datasource
func NewVPCDataSource() datasource.DataSource {
	return &vpcDataSource{}
}

type vpcDataSource struct {
	client *oxide.Client
}

type vpcDataSourceModel struct {
	Description    types.String   `tfsdk:"description"`
	DNSName        types.String   `tfsdk:"dns_name"`
	ID             types.String   `tfsdk:"id"`
	IPV6Prefix     types.String   `tfsdk:"ipv6_prefix"`
	Name           types.String   `tfsdk:"name"`
	ProjectID      types.String   `tfsdk:"project_id"`
	ProjectName    types.String   `tfsdk:"project_name"`
	SystemRouterID types.String   `tfsdk:"system_router_id"`
	TimeCreated    types.String   `tfsdk:"time_created"`
	TimeModified   types.String   `tfsdk:"time_modified"`
	Timeouts       timeouts.Value `tfsdk:"timeouts"`
}

func (d *vpcDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "oxide_vpc"
}

// Configure adds the provider configured client to the data source.
func (d *vpcDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*oxide.Client)
}

func (d *vpcDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
Retrieve information about a specified VPC.
`,
		Attributes: map[string]schema.Attribute{
			"project_name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the project that contains the VPC.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the VPC.",
			},
			"project_id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the project that contains the VPC.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description for the VPC.",
			},
			"dns_name": schema.StringAttribute{
				Computed:    true,
				Description: "DNS name of the VPC.",
			},
			"ipv6_prefix": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "All IPv6 subnets created from this VPC must be taken from this range, which should be a unique local address in the range `fd00::/48`. The default VPC Subnet will have the first `/64` range from this prefix.",
			},
			"timeouts": timeouts.Attributes(ctx),
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier of the VPC.",
			},
			"system_router_id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier of the system router.",
			},
			"time_created": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this VPC was created.",
			},
			"time_modified": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this VPC was last modified.",
			},
		},
	}
}

func (d *vpcDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state vpcDataSourceModel

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

	params := oxide.VpcViewParams{
		Vpc:     oxide.NameOrId(state.Name.ValueString()),
		Project: oxide.NameOrId(state.ProjectName.ValueString()),
	}
	vpc, err := d.client.VpcView(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read VPC:",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("read VPC with ID: %v", vpc.Id), map[string]any{"success": true})

	state.Description = types.StringValue(vpc.Description)
	state.DNSName = types.StringValue(string(vpc.DnsName))
	state.ID = types.StringValue(vpc.Id)
	state.IPV6Prefix = types.StringValue(string(vpc.Ipv6Prefix))
	state.Name = types.StringValue(string(vpc.Name))
	state.ProjectID = types.StringValue(vpc.ProjectId)
	state.SystemRouterID = types.StringValue(vpc.SystemRouterId)
	state.TimeCreated = types.StringValue(vpc.TimeCreated.String())
	state.TimeModified = types.StringValue(vpc.TimeModified.String())

	// Save state into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

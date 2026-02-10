// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/oxidecomputer/oxide.go/oxide"
)

var (
	_ datasource.DataSource              = (*vpcRouterRouteDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*vpcRouterRouteDataSource)(nil)
)

// NewVPCRouterRouteDataSource initialises a VPC router route datasource
func NewVPCRouterRouteDataSource() datasource.DataSource {
	return &vpcRouterRouteDataSource{}
}

type vpcRouterRouteDataSource struct {
	client *oxide.Client
}

type vpcRouterRouteDataSourceModel struct {
	Description   types.String      `tfsdk:"description"`
	Destination   types.Object      `tfsdk:"destination"`
	ID            types.String      `tfsdk:"id"`
	Kind          types.String      `tfsdk:"kind"`
	Name          types.String      `tfsdk:"name"`
	Target        types.Object      `tfsdk:"target"`
	ProjectName   types.String      `tfsdk:"project_name"`
	VPCName       types.String      `tfsdk:"vpc_name"`
	VPCRouterName types.String      `tfsdk:"vpc_router_name"`
	VPCRouterID   types.String      `tfsdk:"vpc_router_id"`
	TimeCreated   timetypes.RFC3339 `tfsdk:"time_created"`
	TimeModified  timetypes.RFC3339 `tfsdk:"time_modified"`
	Timeouts      timeouts.Value    `tfsdk:"timeouts"`
}

type vpcRouterRouteDestinationDataSourceModel struct {
	Type  types.String `tfsdk:"type"`
	Value types.String `tfsdk:"value"`
}

type vpcRouterRouteTargetDataSourceModel struct {
	Type  types.String `tfsdk:"type"`
	Value types.String `tfsdk:"value"`
}

func (d *vpcRouterRouteDataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = "oxide_vpc_router_route"
}

// Configure adds the provider configured client to the data source.
func (d *vpcRouterRouteDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	_ *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*oxide.Client)
}

func (d *vpcRouterRouteDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
Retrieve information about a specified VPC router route.
`,
		Attributes: map[string]schema.Attribute{
			"project_name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the project that contains the VPC router route.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the VPC router route.",
			},
			"vpc_name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the VPC that contains the VPC router route.",
			},
			"vpc_router_name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the VPC router that contains the VPC router route.",
			},
			"vpc_router_id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the VPC router that contains the VPC router route.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description for the VPC router route.",
			},
			"destination": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Selects which traffic this routing rule will apply to",
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: "Route destination type. Possible values: `vpc`, `subnet`, `ip`, `ip_net`.",
						Computed:            true,
					},
					"value": schema.StringAttribute{
						MarkdownDescription: replaceBackticks(`
Depending on the type, it will be one of the following:
  - ''vpc'': Name of the VPC.
  - ''subnet'': Name of the VPC subnet.
  - ''ip'': IP address.
  - ''ip_net'': IPv4 or IPv6 subnet.
 `),
						Computed: true,
					},
				},
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier of the VPC router route.",
			},
			"kind": schema.StringAttribute{
				Computed:    true,
				Description: "Whether the VPC router is custom or system created.",
			},
			"target": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Location that matched packets should be forwarded to",
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: "Route destination type. Possible values: `vpc`, `subnet`, `instance`, `ip`, `internet_gateway`, `drop`.",
						Computed:            true,
					},
					"value": schema.StringAttribute{
						MarkdownDescription: replaceBackticks(`
Depending on the type, it will be one of the following:
  - ''vpc'': Name of the VPC.
  - ''subnet'': Name of the VPC subnet.
  - ''instance'': Name of the instance.
  - ''ip'': IP address.
  - ''internet_gateway'': Name of the internet gateway.
`),
						Computed: true,
					},
				},
			},
			"timeouts": timeouts.Attributes(ctx),
			"time_created": schema.StringAttribute{
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
				Description: "Timestamp of when this VPC router route was created.",
			},
			"time_modified": schema.StringAttribute{
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
				Description: "Timestamp of when this VPC router route was last modified.",
			},
		},
	}
}

func (d *vpcRouterRouteDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state vpcRouterRouteDataSourceModel

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

	params := oxide.VpcRouterRouteViewParams{
		Route:   oxide.NameOrId(state.Name.ValueString()),
		Project: oxide.NameOrId(state.ProjectName.ValueString()),
		Router:  oxide.NameOrId(state.VPCRouterName.ValueString()),
		Vpc:     oxide.NameOrId(state.VPCName.ValueString()),
	}
	route, err := d.client.VpcRouterRouteView(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read VPC router route:",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(
		ctx,
		fmt.Sprintf("read VPC router route with ID: %v", route.Id),
		map[string]any{"success": true},
	)

	// Parse vpcRouterRouteDestinationDataSourceModel into types.Object
	dm := vpcRouterRouteDestinationDataSourceModel{
		Type:  types.StringValue(string(route.Destination.Type())),
		Value: types.StringValue(route.Destination.String()),
	}
	attributeTypes := map[string]attr.Type{
		"type":  types.StringType,
		"value": types.StringType,
	}
	destination, diags := types.ObjectValueFrom(ctx, attributeTypes, dm)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Destination = destination

	// Parse vpcRouterRouteTargetDataSourceModel into types.Object
	tm := vpcRouterRouteTargetDataSourceModel{
		Type: types.StringValue(string(route.Target.Type())),
	}

	// When the target type is set to "drop" the value will be empty
	if targetValue := route.Target.String(); targetValue != "" {
		tm.Value = types.StringValue(targetValue)
	}

	targetAttributeTypes := map[string]attr.Type{
		"type":  types.StringType,
		"value": types.StringType,
	}
	target, diags := types.ObjectValueFrom(ctx, targetAttributeTypes, tm)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Target = target

	state.Description = types.StringValue(route.Description)
	state.ID = types.StringValue(route.Id)
	state.Kind = types.StringValue(string(route.Kind))
	state.Name = types.StringValue(string(route.Name))
	state.VPCRouterID = types.StringValue(string(route.VpcRouterId))
	state.TimeCreated = timetypes.NewRFC3339TimeValue(route.TimeCreated.UTC())
	state.TimeModified = timetypes.NewRFC3339TimeValue(route.TimeModified.UTC())

	// Save state into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

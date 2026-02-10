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
	_ datasource.DataSource              = (*vpcRouterDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*vpcRouterDataSource)(nil)
)

// NewVPCRouterDataSource initialises an images datasource
func NewVPCRouterDataSource() datasource.DataSource {
	return &vpcRouterDataSource{}
}

type vpcRouterDataSource struct {
	client *oxide.Client
}

type vpcRouterDataSourceModel struct {
	Description  types.String      `tfsdk:"description"`
	ID           types.String      `tfsdk:"id"`
	Kind         types.String      `tfsdk:"kind"`
	Name         types.String      `tfsdk:"name"`
	ProjectName  types.String      `tfsdk:"project_name"`
	VPCID        types.String      `tfsdk:"vpc_id"`
	VPCName      types.String      `tfsdk:"vpc_name"`
	TimeCreated  timetypes.RFC3339 `tfsdk:"time_created"`
	TimeModified timetypes.RFC3339 `tfsdk:"time_modified"`
	Timeouts     timeouts.Value    `tfsdk:"timeouts"`
}

func (d *vpcRouterDataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = "oxide_vpc_router"
}

// Configure adds the provider configured client to the data source.
func (d *vpcRouterDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	_ *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*oxide.Client)
}

func (d *vpcRouterDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
Retrieve information about a specified VPC router.
`,
		Attributes: map[string]schema.Attribute{
			"project_name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the project that contains the VPC router.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the VPC router.",
			},
			"vpc_name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the VPC that contains the VPC router.",
			},
			"vpc_id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the VPC that contains the VPC router.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description for the VPC router.",
			},
			"timeouts": timeouts.Attributes(ctx),
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier of the VPC router.",
			},
			"kind": schema.StringAttribute{
				Computed:    true,
				Description: "Whether the VPC router is custom or system created.",
			},
			"time_created": schema.StringAttribute{
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
				Description: "Timestamp of when this VPC router was created.",
			},
			"time_modified": schema.StringAttribute{
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
				Description: "Timestamp of when this VPC router was last modified.",
			},
		},
	}
}

func (d *vpcRouterDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state vpcRouterDataSourceModel

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

	params := oxide.VpcRouterViewParams{
		Router:  oxide.NameOrId(state.Name.ValueString()),
		Vpc:     oxide.NameOrId(state.VPCName.ValueString()),
		Project: oxide.NameOrId(state.ProjectName.ValueString()),
	}
	router, err := d.client.VpcRouterView(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read VPC router:",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(
		ctx,
		fmt.Sprintf("read VPC router with ID: %v", router.Id),
		map[string]any{"success": true},
	)

	state.Description = types.StringValue(router.Description)
	state.ID = types.StringValue(router.Id)
	state.Kind = types.StringValue(string(router.Kind))
	state.Name = types.StringValue(string(router.Name))
	state.VPCID = types.StringValue(router.VpcId)
	state.TimeCreated = timetypes.NewRFC3339TimeValue(router.TimeCreated.UTC())
	state.TimeModified = timetypes.NewRFC3339TimeValue(router.TimeModified.UTC())

	// Save state into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package vpcs

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

var (
	_ datasource.DataSource              = (*DataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*DataSource)(nil)
)

// NewDataSource initialises a VPCs data source.
func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

type DataSource struct {
	client *oxide.Client
}

type DataSourceModel struct {
	ID       types.String         `tfsdk:"id"`
	Project  types.String         `tfsdk:"project"`
	Timeouts timeouts.Value       `tfsdk:"timeouts"`
	VPCs     []VPCDataSourceModel `tfsdk:"vpcs"`
}

type VPCDataSourceModel struct {
	Description    types.String `tfsdk:"description"`
	DNSName        types.String `tfsdk:"dns_name"`
	ID             types.String `tfsdk:"id"`
	IPV6Prefix     types.String `tfsdk:"ipv6_prefix"`
	Name           types.String `tfsdk:"name"`
	ProjectID      types.String `tfsdk:"project_id"`
	SystemRouterID types.String `tfsdk:"system_router_id"`
	TimeCreated    types.String `tfsdk:"time_created"`
	TimeModified   types.String `tfsdk:"time_modified"`
}

func (d *DataSource) Metadata(
	_ context.Context,
	_ datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = "oxide_vpcs"
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
		MarkdownDescription: `
Retrieve a list of VPCs in a project.
`,
		Attributes: map[string]schema.Attribute{
			"project": schema.StringAttribute{
				Required:    true,
				Description: "Name or ID of the project that contains the VPCs.",
			},
			"id": schema.StringAttribute{
				Computed: true,
			},
			"timeouts": timeouts.Attributes(ctx),
			"vpcs": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"description": schema.StringAttribute{
							Computed:    true,
							Description: "Description for the VPC.",
						},
						"dns_name": schema.StringAttribute{
							Computed:    true,
							Description: "DNS name of the VPC.",
						},
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "Unique, immutable, system-controlled identifier of the VPC.",
						},
						"ipv6_prefix": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "All IPv6 subnets created from this VPC must be taken from this range, which should be a unique local address in the range `fd00::/48`.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Name of the VPC.",
						},
						"project_id": schema.StringAttribute{
							Computed:    true,
							Description: "ID of the project that contains the VPC.",
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
				},
			},
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

	vpcs, err := d.client.VpcListAllPages(ctx, oxide.VpcListParams{
		Project: oxide.NameOrId(state.Project.ValueString()),
		SortBy:  oxide.NameOrIdSortModeIdAscending,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read VPCs:",
			"API error: "+err.Error(),
		)
		return
	}

	tflog.Trace(
		ctx,
		"read all VPCs from project",
		map[string]any{"project": state.Project.ValueString()},
	)

	state.ID = types.StringValue(uuid.New().String())
	state.VPCs = make([]VPCDataSourceModel, len(vpcs))
	for i, vpc := range vpcs {
		state.VPCs[i] = VPCDataSourceModel{
			Description:    types.StringValue(vpc.Description),
			DNSName:        types.StringValue(string(vpc.DnsName)),
			ID:             types.StringValue(vpc.Id),
			IPV6Prefix:     types.StringValue(string(vpc.Ipv6Prefix)),
			Name:           types.StringValue(string(vpc.Name)),
			ProjectID:      types.StringValue(vpc.ProjectId),
			SystemRouterID: types.StringValue(vpc.SystemRouterId),
			TimeCreated:    types.StringValue(vpc.TimeCreated.String()),
			TimeModified:   types.StringValue(vpc.TimeModified.String()),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

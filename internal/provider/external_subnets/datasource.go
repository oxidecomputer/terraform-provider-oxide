// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package externalsubnets

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
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

type DataSource struct {
	client *oxide.Client
}

type DataSourceModel struct {
	ID              types.String                    `tfsdk:"id"`
	Project         types.String                    `tfsdk:"project"`
	ExternalSubnets []ExternalSubnetDataSourceModel `tfsdk:"external_subnets"`
	Timeouts        timeouts.Value                  `tfsdk:"timeouts"`
}

type ExternalSubnetDataSourceModel struct {
	Description        types.String       `tfsdk:"description"`
	ID                 types.String       `tfsdk:"id"`
	InstanceID         types.String       `tfsdk:"instance_id"`
	Name               types.String       `tfsdk:"name"`
	ProjectID          types.String       `tfsdk:"project_id"`
	Subnet             cidrtypes.IPPrefix `tfsdk:"subnet"`
	SubnetPoolID       types.String       `tfsdk:"subnet_pool_id"`
	SubnetPoolMemberID types.String       `tfsdk:"subnet_pool_member_id"`
	TimeCreated        types.String       `tfsdk:"time_created"`
	TimeModified       types.String       `tfsdk:"time_modified"`
}

// NewDataSource initialises an external_subnets data source.
func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

// Metadata returns the data source type name.
func (d *DataSource) Metadata(
	_ context.Context,
	_ datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = "oxide_external_subnets"
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
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieve all external subnets in a project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"project": schema.StringAttribute{
				Required:    true,
				Description: "Name or ID of the project that contains the external subnets.",
			},
			"external_subnets": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"description": schema.StringAttribute{
							Computed:    true,
							Description: "Human-readable free-form text about the external subnet.",
						},
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "Unique, immutable, system-controlled identifier of the external subnet.",
						},
						"instance_id": schema.StringAttribute{
							Computed:    true,
							Description: "Instance ID this external subnet is attached to, if any.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Unique, mutable, user-controlled identifier for the external subnet.",
						},
						"project_id": schema.StringAttribute{
							Computed:    true,
							Description: "Project ID where this external subnet is located.",
						},
						"subnet": schema.StringAttribute{
							Computed:    true,
							CustomType:  cidrtypes.IPPrefixType{},
							Description: "The allocated subnet CIDR.",
						},
						"subnet_pool_id": schema.StringAttribute{
							Computed:    true,
							Description: "The subnet pool this external subnet was allocated from.",
						},
						"subnet_pool_member_id": schema.StringAttribute{
							Computed:    true,
							Description: "The subnet pool member this external subnet was allocated from.",
						},
						"time_created": schema.StringAttribute{
							Computed:    true,
							Description: "Timestamp when this external subnet was created.",
						},
						"time_modified": schema.StringAttribute{
							Computed:    true,
							Description: "Timestamp when this external subnet was last modified.",
						},
					},
				},
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

	params := oxide.ExternalSubnetListParams{
		Project: oxide.NameOrId(state.Project.ValueString()),
		SortBy:  oxide.NameOrIdSortModeIdAscending,
	}
	externalSubnets, err := d.client.ExternalSubnetListAllPages(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read external subnets list:",
			err.Error(),
		)
		return
	}

	tflog.Trace(
		ctx,
		fmt.Sprintf("read all external subnets from project: %v", state.Project.ValueString()),
	)

	state.ID = types.StringValue(uuid.New().String())
	state.ExternalSubnets = make([]ExternalSubnetDataSourceModel, len(externalSubnets))
	for i, externalSubnet := range externalSubnets {
		state.ExternalSubnets[i] = ExternalSubnetDataSourceModel{
			Description:        types.StringValue(externalSubnet.Description),
			ID:                 types.StringValue(externalSubnet.Id),
			InstanceID:         types.StringValue(externalSubnet.InstanceId),
			Name:               types.StringValue(string(externalSubnet.Name)),
			ProjectID:          types.StringValue(externalSubnet.ProjectId),
			Subnet:             cidrtypes.NewIPPrefixValue(externalSubnet.Subnet.String()),
			SubnetPoolID:       types.StringValue(externalSubnet.SubnetPoolId),
			SubnetPoolMemberID: types.StringValue(externalSubnet.SubnetPoolMemberId),
			TimeCreated:        types.StringValue(externalSubnet.TimeCreated.String()),
			TimeModified:       types.StringValue(externalSubnet.TimeModified.String()),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

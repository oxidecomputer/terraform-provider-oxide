// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package floatingips

import (
	"context"
	"fmt"

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

type DataSource struct {
	client *oxide.Client
}

type DataSourceModel struct {
	ID          types.String                `tfsdk:"id"`
	Project     types.String                `tfsdk:"project"`
	Timeouts    timeouts.Value              `tfsdk:"timeouts"`
	FloatingIPs []FloatingIPDataSourceModel `tfsdk:"floating_ips"`
}

type FloatingIPDataSourceModel struct {
	Description  types.String `tfsdk:"description"`
	ID           types.String `tfsdk:"id"`
	InstanceID   types.String `tfsdk:"instance_id"`
	IP           types.String `tfsdk:"ip"`
	IPPoolID     types.String `tfsdk:"ip_pool_id"`
	Name         types.String `tfsdk:"name"`
	ProjectID    types.String `tfsdk:"project_id"`
	TimeCreated  types.String `tfsdk:"time_created"`
	TimeModified types.String `tfsdk:"time_modified"`
}

// NewDataSource initialises a floating_ips data source.
func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

func (d *DataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = "oxide_floating_ips"
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
	resp.Schema = schema.Schema{
		MarkdownDescription: `
Retrieve a list of floating IPs in a project.
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"project": schema.StringAttribute{
				Required:    true,
				Description: "Name or ID of the project containing the floating IPs.",
			},
			"timeouts": timeouts.Attributes(ctx),
			"floating_ips": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"description": schema.StringAttribute{
							Computed:    true,
							Description: "Human-readable free-form text about the floating IP.",
						},
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "Unique, immutable, system-controlled identifier for the floating IP.",
						},
						"instance_id": schema.StringAttribute{
							Computed:    true,
							Description: "Instance ID that this floating IP is attached to, if presently attached.",
						},
						"ip": schema.StringAttribute{
							Computed:    true,
							Description: "IP address for this floating IP.",
						},
						"ip_pool_id": schema.StringAttribute{
							Computed:    true,
							Description: "ID of the IP pool containing this floating IP.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Unique, mutable, user-controlled identifier for the floating IP.",
						},
						"project_id": schema.StringAttribute{
							Computed:    true,
							Description: "Project ID where this floating IP is located.",
						},
						"time_created": schema.StringAttribute{
							Computed:    true,
							Description: "Timestamp when this floating IP was created.",
						},
						"time_modified": schema.StringAttribute{
							Computed:    true,
							Description: "Timestamp when this floating IP was last modified.",
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

	params := oxide.FloatingIpListParams{
		Project: oxide.NameOrId(state.Project.ValueString()),
		SortBy:  oxide.NameOrIdSortModeIdAscending,
	}
	floatingIPs, err := d.client.FloatingIpListAllPages(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read floating IPs:",
			"API error: "+err.Error(),
		)
		return
	}

	tflog.Trace(
		ctx,
		fmt.Sprintf("read all floating IPs from project: %v", state.Project.ValueString()),
		map[string]any{"success": true},
	)

	state.ID = types.StringValue(uuid.New().String())

	state.FloatingIPs = make([]FloatingIPDataSourceModel, len(floatingIPs))
	for i, floatingIP := range floatingIPs {
		state.FloatingIPs[i] = FloatingIPDataSourceModel{
			Description:  types.StringValue(floatingIP.Description),
			ID:           types.StringValue(floatingIP.Id),
			InstanceID:   types.StringValue(floatingIP.InstanceId),
			IP:           types.StringValue(floatingIP.Ip),
			IPPoolID:     types.StringValue(floatingIP.IpPoolId),
			Name:         types.StringValue(string(floatingIP.Name)),
			ProjectID:    types.StringValue(floatingIP.ProjectId),
			TimeCreated:  types.StringValue(floatingIP.TimeCreated.String()),
			TimeModified: types.StringValue(floatingIP.TimeModified.String()),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

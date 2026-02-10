// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/oxidecomputer/oxide.go/oxide"
)

// floatingIPDataSourceModel represents the Terraform configuration and state
// for the Oxide floating IP resource.
type floatingIPDataSourceModel struct {
	ID           types.String      `tfsdk:"id"`
	Name         types.String      `tfsdk:"name"`
	Description  types.String      `tfsdk:"description"`
	IP           types.String      `tfsdk:"ip"`
	InstanceID   types.String      `tfsdk:"instance_id"`
	IPPoolID     types.String      `tfsdk:"ip_pool_id"`
	ProjectID    types.String      `tfsdk:"project_id"`
	ProjectName  types.String      `tfsdk:"project_name"`
	TimeCreated  timetypes.RFC3339 `tfsdk:"time_created"`
	TimeModified timetypes.RFC3339 `tfsdk:"time_modified"`
	Timeouts     timeouts.Value    `tfsdk:"timeouts"`
}

// Compile-time assertions to check that the floatingIPDataSource implements the
// necessary Terraform data source interfaces.
var (
	_ datasource.DataSource              = (*floatingIPDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*floatingIPDataSource)(nil)
)

// floatingIPResource is the concrete type that implements the necessary
// Terraform resource interfaces. It holds state to interact with the Oxide API.
type floatingIPDataSource struct {
	client *oxide.Client
}

// NewFloatingIPDataSource is a helper to easily construct a
// floatingIPDataSource as a type that implements the Terraform data source
// interface.
func NewFloatingIPDataSource() datasource.DataSource {
	return &floatingIPDataSource{}
}

// Metadata configures the Terraform data source name for the Oxide floating IP
// data source.
func (f *floatingIPDataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = "oxide_floating_ip"
}

// Configure sets up necessary data or clients needed by the floatingIPDataSource.
func (f *floatingIPDataSource) Configure(
	ctx context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	f.client = req.ProviderData.(*oxide.Client)
}

// Schema defines the attributes for this Oxide floating IP data source.
func (f *floatingIPDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
Retrieve information about a specified floating IP.
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier for the floating IP.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Unique, mutable, user-controlled identifier for the floating IP.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Human-readable free-form text about the floating IP.",
			},
			"instance_id": schema.StringAttribute{
				Computed:    true,
				Description: "Instance ID that this floating IP is attached to, if presently attached.",
			},
			"ip": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "IP address for this floating IP. If unset an IP address will be chosen from the given `ip_pool_id`.",
			},
			"ip_pool_id": schema.StringAttribute{
				Computed:    true,
				Description: "IP pool ID to allocate this floating IP from. If unset the silo's default IP pool is used.",
			},
			"project_id": schema.StringAttribute{
				Computed:    true,
				Description: "Project ID where this floating IP is located.",
			},
			"project_name": schema.StringAttribute{
				Required:    true,
				Description: "Project name where this floating IP is located.",
			},
			"time_created": schema.StringAttribute{
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
				Description: "Timestamp when this floating IP was created.",
			},
			"time_modified": schema.StringAttribute{
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
				Description: "Timestamp when this floating IP was last modified.",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Read: true,
			}),
		},
	}
}

// Read fetches an Oxide floating IP from the Oxide API.
func (f *floatingIPDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state floatingIPDataSourceModel

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

	params := oxide.FloatingIpViewParams{
		FloatingIp: oxide.NameOrId(state.Name.ValueString()),
		Project:    oxide.NameOrId(state.ProjectName.ValueString()),
	}

	floatingIP, err := f.client.FloatingIpView(ctx, params)
	if err != nil {
		if is404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to read floating IP:",
			"API error: "+err.Error(),
		)
		return
	}

	tflog.Trace(
		ctx,
		fmt.Sprintf("read floating IP with ID: %v", floatingIP.Id),
		map[string]any{"success": true},
	)

	state.ID = types.StringValue(floatingIP.Id)
	state.Name = types.StringValue(string(floatingIP.Name))
	state.Description = types.StringValue(floatingIP.Description)
	state.IP = types.StringValue(floatingIP.Ip)
	state.InstanceID = types.StringValue(floatingIP.InstanceId)
	state.IPPoolID = types.StringValue(floatingIP.IpPoolId)
	state.ProjectID = types.StringValue(floatingIP.ProjectId)
	state.TimeCreated = timetypes.NewRFC3339TimeValue(floatingIP.TimeCreated.UTC())
	state.TimeModified = timetypes.NewRFC3339TimeValue(floatingIP.TimeModified.UTC())

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

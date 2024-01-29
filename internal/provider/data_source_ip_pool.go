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

var _ datasource.DataSource = (*ipPoolDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*ipPoolDataSource)(nil)

type ipPoolDataSource struct {
	client *oxide.Client
}

type ipPoolDataSourceModel struct {
	Description  types.String   `tfsdk:"description"`
	ID           types.String   `tfsdk:"id"`
	IsDefault    types.Bool     `tfsdk:"is_default"`
	Name         types.String   `tfsdk:"name"`
	Timeouts     timeouts.Value `tfsdk:"timeouts"`
	TimeCreated  types.String   `tfsdk:"time_created"`
	TimeModified types.String   `tfsdk:"time_modified"`
}

// NewIpPoolDataSource initialises an ip_pool datasource
func NewIpPoolDataSource() datasource.DataSource {
	return &ipPoolDataSource{}
}

func (d *ipPoolDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "oxide_ip_pool"
}

// Configure adds the provider configured client to the data source.
func (d *ipPoolDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*oxide.Client)
}

func (d *ipPoolDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the IP pool.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description for the IP pool.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier of the IP pool.",
			},
			"is_default": schema.BoolAttribute{
				Computed:    true,
				Description: "If a pool is the default for a silo, floating IPs and instance ephemeral IPs will come from that pool when no other pool is specified. There can be at most one default for a given silo.",
			},
			"time_created": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this IP pool was created.",
			},
			"time_modified": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this IP pool was last modified.",
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

func (d *ipPoolDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ipPoolDataSourceModel

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

	params := oxide.ProjectIpPoolViewParams{
		Pool: oxide.NameOrId(state.Name.ValueString()),
	}
	ipPool, err := d.client.ProjectIpPoolView(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read IP pool:",
			err.Error(),
		)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("read IP pool with ID: %v", ipPool.Id), map[string]any{"success": true})

	// Map response body to model
	state.Description = types.StringValue(ipPool.Description)
	state.ID = types.StringValue(ipPool.Id)
	state.IsDefault = types.BoolPointerValue(ipPool.IsDefault)
	state.Name = types.StringValue(string(ipPool.Name))
	state.TimeCreated = types.StringValue(ipPool.TimeCreated.String())
	state.TimeModified = types.StringValue(ipPool.TimeCreated.String())

	// Save state into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

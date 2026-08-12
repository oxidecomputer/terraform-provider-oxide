// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package systemippool

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/oxidecomputer/oxide.go/oxide"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/shared"
)

var _ datasource.DataSource = (*DataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*DataSource)(nil)

type DataSource struct {
	client *oxide.Client
}

type DataSourceModel struct {
	Assignment   types.String   `tfsdk:"assignment"`
	Description  types.String   `tfsdk:"description"`
	ID           types.String   `tfsdk:"id"`
	IPVersion    types.String   `tfsdk:"ip_version"`
	Name         types.String   `tfsdk:"name"`
	Pool         types.String   `tfsdk:"pool"`
	PoolType     types.String   `tfsdk:"pool_type"`
	Timeouts     timeouts.Value `tfsdk:"timeouts"`
	TimeCreated  types.String   `tfsdk:"time_created"`
	TimeModified types.String   `tfsdk:"time_modified"`
}

// NewDataSource constructs the data source.
func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

// Metadata returns the data source name.
func (d *DataSource) Metadata(
	_ context.Context,
	_ datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = "oxide_system_ip_pool"
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
		MarkdownDescription: "Retrieve information about a specified system IP pool.",
		Attributes: map[string]schema.Attribute{
			"pool": schema.StringAttribute{
				Required:    true,
				Description: "Name or ID of the IP pool.",
			},
			"assignment": schema.StringAttribute{
				Computed:    true,
				Description: "What this pool is currently assigned to.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Human-readable free-form text about the IP pool.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier for the IP pool.",
			},
			"ip_version": schema.StringAttribute{
				Computed:    true,
				Description: "The IP version for the pool.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, mutable, user-controlled identifier for the IP pool.",
			},
			"pool_type": schema.StringAttribute{
				Computed:    true,
				Description: "Type of IP pool (unicast or multicast).",
			},
			"time_created": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when this IP pool was created.",
			},
			"time_modified": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when this IP pool was last modified.",
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

	ipPool, err := d.client.SystemIpPoolView(ctx, oxide.SystemIpPoolViewParams{
		Pool: oxide.NameOrId(state.Pool.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read system IP pool:",
			err.Error(),
		)
		return
	}

	tflog.Trace(
		ctx,
		fmt.Sprintf("read system IP pool with ID: %v", ipPool.Id),
		map[string]any{"success": true},
	)

	state.Assignment = types.StringValue(string(ipPool.Assignment))
	state.Description = types.StringValue(ipPool.Description)
	state.ID = types.StringValue(ipPool.Id)
	state.IPVersion = types.StringValue(string(ipPool.IpVersion))
	state.Name = types.StringValue(string(ipPool.Name))
	state.PoolType = types.StringValue(string(ipPool.PoolType))
	state.TimeCreated = types.StringValue(ipPool.TimeCreated.String())
	state.TimeModified = types.StringValue(ipPool.TimeModified.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

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

var _ datasource.DataSource = (*siloDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*siloDataSource)(nil)

type siloDataSource struct {
	client *oxide.Client
}

type siloDataSourceModel struct {
	Description  types.String   `tfsdk:"description"`
	Discoverable types.Bool     `tfsdk:"discoverable"`
	ID           types.String   `tfsdk:"id"`
	IdentityMode types.String   `tfsdk:"identity_mode"`
	Name         types.String   `tfsdk:"name"`
	Timeouts     timeouts.Value `tfsdk:"timeouts"`
	TimeCreated  types.String   `tfsdk:"time_created"`
	TimeModified types.String   `tfsdk:"time_modified"`
}

// NewSiloDataSource initialises a silo datasource
func NewSiloDataSource() datasource.DataSource {
	return &siloDataSource{}
}

func (d *siloDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "oxide_silo"
}

// Configure adds the provider configured client to the data source.
func (d *siloDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*oxide.Client)
}

func (d *siloDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
Retrieve information about a specified silo.
`,
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the silo.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description for the silo.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier of the silo.",
			},
			"identity_mode": schema.StringAttribute{
				Computed:    true,
				Description: "How users and groups are managed in this silo.",
			},
			"discoverable": schema.BoolAttribute{
				Computed:    true,
				Description: "A silo where discoverable is false can be retrieved only by its ID - it will not be part of the 'list all silos' output.",
			},
			"time_created": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this silo was created.",
			},
			"time_modified": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this silo was last modified.",
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

func (d *siloDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state siloDataSourceModel

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

	params := oxide.SiloViewParams{
		Silo: oxide.NameOrId(state.Name.ValueString()),
	}
	silo, err := d.client.SiloView(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read Silo:",
			err.Error(),
		)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("read Silo with ID: %v", silo.Id), map[string]any{"success": true})

	// Map response body to model
	state.Description = types.StringValue(silo.Description)
	state.ID = types.StringValue(silo.Id)
	state.Discoverable = types.BoolPointerValue(silo.Discoverable)
	state.IdentityMode = types.StringValue(string(silo.IdentityMode))
	state.Name = types.StringValue(string(silo.Name))
	state.TimeCreated = types.StringValue(silo.TimeCreated.String())
	state.TimeModified = types.StringValue(silo.TimeCreated.String())

	// Save state into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

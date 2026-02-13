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

var _ datasource.DataSource = (*addressLotDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*addressLotDataSource)(nil)

type addressLotDataSource struct {
	client *oxide.Client
}

type addressLotDataSourceModel struct {
	Blocks       []addressLotDataSourceBlockModel `tfsdk:"blocks"`
	Description  types.String                     `tfsdk:"description"`
	Kind         types.String                     `tfsdk:"kind"`
	Name         types.String                     `tfsdk:"name"`
	ID           types.String                     `tfsdk:"id"`
	TimeCreated  types.String                     `tfsdk:"time_created"`
	TimeModified types.String                     `tfsdk:"time_modified"`
	Timeouts     timeouts.Value                   `tfsdk:"timeouts"`
}

type addressLotDataSourceBlockModel struct {
	ID           types.String `tfsdk:"id"`
	FirstAddress types.String `tfsdk:"first_address"`
	LastAddress  types.String `tfsdk:"last_address"`
}

// NewAddressLotDataSource initialises an address_lot datasource.
func NewAddressLotDataSource() datasource.DataSource {
	return &addressLotDataSource{}
}

func (d *addressLotDataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = "oxide_address_lot"
}

// Configure adds the provider configured client to the data source.
func (d *addressLotDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	_ *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*oxide.Client)
}

func (d *addressLotDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the address lot.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description for the address lot.",
			},
			"kind": schema.StringAttribute{
				Computed:    true,
				Description: "Kind for the address lot.",
			},
			"blocks": schema.SetNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "ID of the address lot block.",
							Computed:    true,
						},
						"first_address": schema.StringAttribute{
							Description: "First address in the lot.",
							Computed:    true,
						},
						"last_address": schema.StringAttribute{
							Description: "Last address in the lot.",
							Computed:    true,
						},
					},
				},
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier of the address lot.",
			},
			"time_created": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this address lot was created.",
			},
			"time_modified": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this address lot was last modified.",
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

func (d *addressLotDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state addressLotDataSourceModel

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

	addressLot, err := d.client.NetworkingAddressLotView(ctx, oxide.NetworkingAddressLotViewParams{
		AddressLot: oxide.NameOrId(state.Name.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read address lot:",
			"API error: "+err.Error(),
		)
		return
	}
	lot := addressLot.Lot
	tflog.Trace(
		ctx,
		fmt.Sprintf("read address lot with ID: %v", lot.Id),
		map[string]any{"success": true},
	)

	state.ID = types.StringValue(lot.Id)
	state.Name = types.StringValue(string(lot.Name))
	state.Kind = types.StringValue(string(lot.Kind))
	state.Description = types.StringValue(lot.Description)
	state.TimeCreated = types.StringValue(lot.TimeCreated.String())
	state.TimeModified = types.StringValue(lot.TimeModified.String())

	blockModels := make([]addressLotDataSourceBlockModel, len(addressLot.Blocks))
	for index, item := range addressLot.Blocks {
		blockModels[index] = addressLotDataSourceBlockModel{
			ID:           types.StringValue(item.Id),
			FirstAddress: types.StringValue(item.FirstAddress),
			LastAddress:  types.StringValue(item.LastAddress),
		}
	}
	state.Blocks = blockModels

	// Save state into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

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
)

var (
	_ datasource.DataSource              = (*networkingAddressLotsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*networkingAddressLotsDataSource)(nil)
)

// NewNetworkingAddressLotsDataSource initialises an images datasource
func NewNetworkingAddressLotsDataSource() datasource.DataSource {
	return &networkingAddressLotsDataSource{}
}

type networkingAddressLotsDataSource struct {
	client *oxide.Client
}

type networkingAddressLotsDatasourceModel struct {
	ID          types.String                `tfsdk:"id"`
	Timeouts    timeouts.Value              `tfsdk:"timeouts"`
	AddressLots []addressLotDatasourceModel `tfsdk:"address_lots"`
}

type addressLotDatasourceModel struct {
	ID          types.String `tfsdk:"id"`
	Description types.String `tfsdk:"description"`
	Kind        types.String `tfsdk:"kind"`
	Name        types.String `tfsdk:"name"`
}

func (d *networkingAddressLotsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "oxide_networking_address_lots"
}

// Configure adds the provider configured client to the data source.
func (d *networkingAddressLotsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*oxide.Client)
}

func (d *networkingAddressLotsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"timeouts": timeouts.Attributes(ctx),
			"address_lots": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "Address Lot ID",
						},
						"description": schema.StringAttribute{
							Computed:    true,
							Description: "human-readable free-form text about an Address Lot",
						},
						"kind": schema.StringAttribute{
							Computed:    true,
							Description: "Desired use of Address Lot",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "unique, mutable, user-controlled identifier for each resource",
						},
					},
				},
			},
		},
	}
}

func (d *networkingAddressLotsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state networkingAddressLotsDatasourceModel

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

	params := oxide.NetworkingAddressLotListParams{}

	addressLots, err := d.client.NetworkingAddressLotListAllPages(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read address lots:",
			"API error: "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("read all address lots"), map[string]any{"success": true})

	// Set a unique ID for the datasource payload
	state.ID = types.StringValue(uuid.New().String())

	// Map response body to model
	for _, lot := range addressLots {
		addressLotState := addressLotDatasourceModel{
			ID:          types.StringValue(lot.Id),
			Description: types.StringValue(lot.Description),
			Kind:        types.StringValue(string(lot.Kind)),
			Name:        types.StringValue(string(lot.Name)),
		}
		state.AddressLots = append(state.AddressLots, addressLotState)
	}

	// Save state into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

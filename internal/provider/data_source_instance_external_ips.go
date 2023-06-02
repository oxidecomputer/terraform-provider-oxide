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
	_ datasource.DataSource              = (*instanceExternalIPsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*instanceExternalIPsDataSource)(nil)
)

// NewInstanceExternalIPsDataSource initialises an images datasource
func NewInstanceExternalIPsDataSource() datasource.DataSource {
	return &instanceExternalIPsDataSource{}
}

type instanceExternalIPsDataSource struct {
	client *oxide.Client
}

type instanceExternalIPsDatasourceModel struct {
	ID          types.String                `tfsdk:"id"`
	InstanceID  types.String                `tfsdk:"instance_id"`
	Timeouts    timeouts.Value              `tfsdk:"timeouts"`
	ExternalIPs []externalIPDatasourceModel `tfsdk:"external_ips"`
}

type externalIPDatasourceModel struct {
	IP   types.String `tfsdk:"ip"`
	Kind types.String `tfsdk:"kind"`
}

func (d *instanceExternalIPsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "oxide_instance_external_ips"
}

// Configure adds the provider configured client to the data source.
func (d *instanceExternalIPsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*oxide.Client)
}

func (d *instanceExternalIPsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"instance_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the instance to which the external IPs belong to.",
			},
			"id": schema.StringAttribute{
				Computed: true,
			},
			"timeouts": timeouts.Attributes(ctx),
			"external_ips": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"ip": schema.StringAttribute{
							Description: "External IP address.",
							Computed:    true,
						},
						"kind": schema.StringAttribute{
							Computed:    true,
							Description: "Kind of external IP address.",
						},
					},
				},
			},
		},
	}
}

func (d *instanceExternalIPsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state instanceExternalIPsDatasourceModel

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

	ips, err := d.client.InstanceExternalIpList(
		oxide.InstanceExternalIpListParams{
			Instance: oxide.NameOrId(state.InstanceID.ValueString()),
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read external ips:",
			"API error: "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("read all external IPs from instance: %v", state.InstanceID.ValueString()), map[string]any{"success": true})

	// Set a unique ID for the datasource payload
	state.ID = types.StringValue(uuid.New().String())

	// Map response body to model
	for _, ip := range ips.Items {
		externalIPState := externalIPDatasourceModel{
			IP:   types.StringValue(ip.Ip),
			Kind: types.StringValue(string(ip.Kind)),
		}
		state.ExternalIPs = append(state.ExternalIPs, externalIPState)
	}

	// Save state into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

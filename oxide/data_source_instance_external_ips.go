// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	oxideSDK "github.com/oxidecomputer/oxide.go/oxide"
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
	client *oxideSDK.Client
}

type instanceExternalIPsModel struct {
	ID          types.String      `tfsdk:"id"`
	InstanceID  types.String      `tfsdk:"instance_id"`
	Timeouts    timeouts.Value    `tfsdk:"timeouts"`
	ExternalIPs []externalIPModel `tfsdk:"external_ips"`
}

type externalIPModel struct {
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

	d.client = req.ProviderData.(*oxideSDK.Client)
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
	var state instanceExternalIPsModel

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
		oxideSDK.InstanceExternalIpListParams{
			Instance: oxideSDK.NameOrId(state.InstanceID.ValueString()),
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
		externalIPState := externalIPModel{
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

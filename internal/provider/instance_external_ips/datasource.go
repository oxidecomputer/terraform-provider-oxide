// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package instanceexternalips

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

// NewDataSource initialises an images datasource
func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

type DataSource struct {
	client *oxide.Client
}

type DataSourceModel struct {
	ID          types.String      `tfsdk:"id"`
	InstanceID  types.String      `tfsdk:"instance_id"`
	Timeouts    timeouts.Value    `tfsdk:"timeouts"`
	ExternalIPs []ExternalIPModel `tfsdk:"external_ips"`
}

type ExternalIPModel struct {
	IP   types.String `tfsdk:"ip"`
	Kind types.String `tfsdk:"kind"`
}

func (d *DataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = "oxide_instance_external_ips"
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
Retrieve information of all external IPs associated to an instance.
`,
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

func (d *DataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state DataSourceModel

	// Read Terraform configuration data into the model
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

	params := oxide.InstanceExternalIpListParams{
		Instance: oxide.NameOrId(state.InstanceID.ValueString()),
	}
	ips, err := d.client.InstanceExternalIpList(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read external ips:",
			"API error: "+err.Error(),
		)
		return
	}

	tflog.Trace(
		ctx,
		fmt.Sprintf("read all external IPs from instance: %v", state.InstanceID.ValueString()),
		map[string]any{"success": true},
	)

	// Set a unique ID for the datasource payload
	state.ID = types.StringValue(uuid.New().String())

	// Map response body to model
	for _, ip := range ips.Items {
		var ipAddr string
		switch v := ip.Value.(type) {
		case *oxide.ExternalIpEphemeral:
			ipAddr = v.Ip
		case *oxide.ExternalIpFloating:
			ipAddr = v.Ip
		case *oxide.ExternalIpSnat:
			ipAddr = v.Ip
		default:
			resp.Diagnostics.AddError(
				"Unexpected external IP type",
				fmt.Sprintf("Encountered unexpected external IP type: %T", ip.Value),
			)
			return
		}
		externalIPState := ExternalIPModel{
			IP:   types.StringValue(ipAddr),
			Kind: types.StringValue(string(ip.Kind())),
		}
		state.ExternalIPs = append(state.ExternalIPs, externalIPState)
	}

	// Save state into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

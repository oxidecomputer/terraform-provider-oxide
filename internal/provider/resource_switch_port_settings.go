// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/oxidecomputer/oxide.go/oxide"
)

var (
	_ resource.Resource              = (*switchPortSettingsResource)(nil)
	_ resource.ResourceWithConfigure = (*switchPortSettingsResource)(nil)
)

type switchPortSettingsResource struct {
	client *oxide.Client
}

type switchPortSettingsModel struct {
	ID           types.String                       `tfsdk:"id"`
	Name         types.String                       `tfsdk:"name"`
	Description  types.String                       `tfsdk:"description"`
	Addresses    []switchPortSettingsAddressModel   `tfsdk:"addresses"`
	BGPPeers     []switchPortSettingsBGPPeerModel   `tfsdk:"bgp_peers"`
	Groups       []types.String                     `tfsdk:"groups"`
	Interfaces   []switchPortSettingsInterfaceModel `tfsdk:"interfaces"`
	Links        []switchPortSettingsLinkModel      `tfsdk:"links"`
	PortConfig   *switchPortSettingsPortConfigModel `tfsdk:"port_config"`
	Routes       []switchPortSettingsRouteModel     `tfsdk:"routes"`
	TimeCreated  types.String                       `tfsdk:"time_created"`
	TimeModified types.String                       `tfsdk:"time_modified"`
	Timeouts     timeouts.Value                     `tfsdk:"timeouts"`
}

type switchPortSettingsAddressModel struct {
	Addresses []switchPortSettingsAddressAddressModel `tfsdk:"addresses"`
	LinkName  types.String                            `tfsdk:"link_name"`
}

type switchPortSettingsAddressAddressModel struct {
	Address    types.String `tfsdk:"address"`
	AddressLot types.String `tfsdk:"address_lot"`
	VlanID     types.Int32  `tfsdk:"vlan_id"`
}

type switchPortSettingsBGPPeerModel struct {
	LinkName types.String                         `tfsdk:"link_name"`
	Peers    []switchPortSettingsBGPPeerPeerModel `tfsdk:"peers"`
}

type switchPortSettingsBGPPeerPeerModel struct {
	Addr                   types.String                                     `tfsdk:"addr"`
	AllowedExport          *switchPortSettingsBGPPeerPeerAllowedExportModel `tfsdk:"allow_export"`
	AllowedImport          *switchPortSettingsBGPPeerPeerAllowedImportModel `tfsdk:"allow_import"`
	BGPConfig              types.String                                     `tfsdk:"bgp_config"`
	Communities            []types.Int64                                    `tfsdk:"communities"`
	ConnectRetry           types.Int64                                      `tfsdk:"connect_retry"`
	DelayOpen              types.Int64                                      `tfsdk:"delay_open"`
	EnforceFirstAs         types.Bool                                       `tfsdk:"enforce_first_as"`
	HoldTime               types.Int64                                      `tfsdk:"hold_time"`
	IdleHoldTime           types.Int64                                      `tfsdk:"idle_hold_time"`
	InterfaceName          types.String                                     `tfsdk:"interface_name"`
	Keepalive              types.Int64                                      `tfsdk:"keepalive"`
	LocalPref              types.Int64                                      `tfsdk:"local_pref"`
	MD5AuthKey             types.String                                     `tfsdk:"md5_auth_key"`
	MinTTL                 types.Int32                                      `tfsdk:"min_ttl"`
	MultiExitDiscriminator types.Int64                                      `tfsdk:"multi_exit_discriminator"`
	RemoteASN              types.Int64                                      `tfsdk:"remote_asn"`
	VlanID                 types.Int32                                      `tfsdk:"vlan_id"`
}

type switchPortSettingsBGPPeerPeerAllowedExportModel struct {
	Type  types.String   `tfsdk:"type"`
	Value []types.String `tfsdk:"value"`
}

type switchPortSettingsBGPPeerPeerAllowedImportModel struct {
	Type  types.String   `tfsdk:"type"`
	Value []types.String `tfsdk:"value"`
}

type switchPortSettingsInterfaceModel struct {
	Kind      *switchPortSettingsInterfaceKindModel `tfsdk:"kind"`
	LinkName  types.String                          `tfsdk:"link_name"`
	V6Enabled types.Bool                            `tfsdk:"v6_enabled"`
}

type switchPortSettingsInterfaceKindModel struct {
	Type types.String `tfsdk:"type"`
	VID  types.Int32  `tfsdk:"vid"`
}

type switchPortSettingsLinkModel struct {
	Autoneg  types.Bool                       `tfsdk:"autoneg"`
	FEC      types.String                     `tfsdk:"fec"`
	LinkName types.String                     `tfsdk:"link_name"`
	LLDP     *switchPortSettingsLinkLLDPModel `tfsdk:"lldp"`
	MTU      types.Int32                      `tfsdk:"mtu"`
	Speed    types.String                     `tfsdk:"speed"`
	TxEq     *switchPortSettingsLinkTxEqModel `tfsdk:"tx_eq"`
}

type switchPortSettingsLinkLLDPModel struct {
	ChassisID         types.String `tfsdk:"chassis_id"`
	Enabled           types.Bool   `tfsdk:"enabled"`
	LinkDescription   types.String `tfsdk:"link_description"`
	LinkName          types.String `tfsdk:"link_name"`
	ManagementIP      types.String `tfsdk:"management_ip"`
	SystemDescription types.String `tfsdk:"system_description"`
	SystemName        types.String `tfsdk:"system_name"`
}

type switchPortSettingsLinkTxEqModel struct {
	Main  types.Int32 `tfsdk:"main"`
	Post1 types.Int32 `tfsdk:"post1"`
	Post2 types.Int32 `tfsdk:"post2"`
	Pre1  types.Int32 `tfsdk:"pre1"`
	Pre2  types.Int32 `tfsdk:"pre2"`
}

type switchPortSettingsPortConfigModel struct {
	Geometry types.String `tfsdk:"geometry"`
}

type switchPortSettingsRouteModel struct {
	LinkName types.String                        `tfsdk:"link_name"`
	Routes   []switchPortSettingsRouteRouteModel `tfsdk:"routes"`
}

type switchPortSettingsRouteRouteModel struct {
	Dst         types.String `tfsdk:"dst"`
	GW          types.String `tfsdk:"gw"`
	RIBPriority types.Int32  `tfsdk:"rib_priority"`
	VID         types.Int32  `tfsdk:"vid"`
}

func NewSwitchPortSettingsResource() resource.Resource {
	return &switchPortSettingsResource{}
}

func (r *switchPortSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "oxide_switch_port_settings"
}

func (r *switchPortSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*oxide.Client)
}

func (r *switchPortSettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *switchPortSettingsResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier of the switch port settings.",
			},
			"addresses": schema.SetNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"link_name": schema.StringAttribute{
							Required: true,
						},
						"addresses": schema.SetNestedAttribute{
							Required: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"address": schema.StringAttribute{
										Required: true,
									},
									"address_lot": schema.StringAttribute{
										Required: true,
									},
									"vlan_id": schema.Int32Attribute{
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"bgp_peers": schema.SetNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"link_name": schema.StringAttribute{
							Required: true,
						},
						"peers": schema.SetNestedAttribute{
							Required: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"addr": schema.StringAttribute{
										Required: true,
									},
									"allowed_export": schema.SingleNestedAttribute{
										Required: true,
										Attributes: map[string]schema.Attribute{
											"type": schema.StringAttribute{
												Required: true,
											},
											"value": schema.SetAttribute{
												Optional:    true,
												ElementType: types.StringType,
											},
										},
									},
									"allowed_import": schema.SingleNestedAttribute{
										Required: true,
										Attributes: map[string]schema.Attribute{
											"type": schema.StringAttribute{
												Required: true,
											},
											"value": schema.SetAttribute{
												Optional:    true,
												ElementType: types.StringType,
											},
										},
									},
									"bgp_config": schema.StringAttribute{
										Required: true,
									},
									"communities": schema.SetAttribute{
										Required:    true,
										ElementType: types.Int64Type,
									},
									"connect_retry": schema.Int64Attribute{
										Required: true,
									},
									"delay_open": schema.Int64Attribute{
										Required: true,
									},
									"enforce_first_as": schema.BoolAttribute{
										Required: true,
									},
									"hold_time": schema.Int64Attribute{
										Required: true,
									},
									"idle_hold_time": schema.Int64Attribute{
										Required: true,
									},
									"interface_name": schema.StringAttribute{
										Required: true,
									},
									"keepalive": schema.Int64Attribute{
										Required: true,
									},
									"local_pref": schema.Int64Attribute{
										Optional: true,
									},
									"md5_auth_key": schema.StringAttribute{
										Optional: true,
									},
									"min_ttl": schema.Int32Attribute{
										Optional: true,
									},
									"multi_exit_discriminator": schema.Int64Attribute{
										Optional: true,
									},
									"remote_asn": schema.Int64Attribute{
										Optional: true,
									},
									"vlan_id": schema.Int32Attribute{
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"description": schema.StringAttribute{
				Required: true,
			},
			"groups": schema.SetAttribute{
				Optional:    true,
				ElementType: types.StringType,
			},
			"interfaces": schema.SetNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"kind": schema.SingleNestedAttribute{
							Required: true,
							Attributes: map[string]schema.Attribute{
								"type": schema.StringAttribute{
									Required: true,
								},
								"vid": schema.Int32Attribute{
									Optional: true,
								},
							},
						},
						"link_name": schema.StringAttribute{
							Required: true,
						},
						"v6_enabled": schema.BoolAttribute{
							Optional: true,
						},
					},
				},
			},
			"links": schema.SetNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"autoneg": schema.BoolAttribute{
							Required: true,
						},
						"fec": schema.StringAttribute{
							Optional: true,
						},
						"link_name": schema.StringAttribute{
							Required: true,
						},
						"lldp": schema.SingleNestedAttribute{
							Required: true,
							Attributes: map[string]schema.Attribute{
								"chassis_id": schema.StringAttribute{
									Optional: true,
								},
								"enabled": schema.BoolAttribute{
									Required: true,
								},
								"link_description": schema.StringAttribute{
									Optional: true,
								},
								"link_name": schema.StringAttribute{
									Optional: true,
								},
								"management_ip": schema.StringAttribute{
									Optional: true,
								},
								"system_description": schema.StringAttribute{
									Optional: true,
								},
								"system_name": schema.StringAttribute{
									Optional: true,
								},
							},
						},
						"mtu": schema.Int32Attribute{
							Required: true,
						},
						"speed": schema.StringAttribute{
							Required: true,
						},
						"tx_eq": schema.SingleNestedAttribute{
							Optional: true,
							Attributes: map[string]schema.Attribute{
								"main": schema.Int32Attribute{
									Optional: true,
								},
								"post1": schema.Int32Attribute{
									Optional: true,
								},
								"post2": schema.Int32Attribute{
									Optional: true,
								},
								"pre1": schema.Int32Attribute{
									Optional: true,
								},
								"pre2": schema.Int32Attribute{
									Optional: true,
								},
							},
						},
					},
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"port_config": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"geometry": schema.StringAttribute{
						Required: true,
					},
				},
			},
			"routes": schema.SetNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"link_name": schema.StringAttribute{
							Required: true,
						},
						"routes": schema.SetNestedAttribute{
							Required: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"dst": schema.StringAttribute{
										Required: true,
									},
									"gw": schema.StringAttribute{
										Required: true,
									},
									"rib_priority": schema.Int32Attribute{
										Optional: true,
									},
									"vid": schema.Int32Attribute{
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Update: true,
				Delete: true,
			}),
			"time_created": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this Switch Port Configuration was created.",
			},
			"time_modified": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this Switch Port Configuration was last modified.",
			},
		},
	}
}

func (r *switchPortSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan switchPortSettingsModel

	// Read Terraform plan data into the model.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := plan.Timeouts.Create(ctx, defaultTimeout())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	params, _ := toOxideParams(plan)

	settings, err := r.client.NetworkingSwitchPortSettingsCreate(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating switch port settings",
			"API error: "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("created switch port settings with ID: %v", settings.Id), map[string]any{"success": true})

	// Map response body to schema and populate computed attribute values.
	plan.ID = types.StringValue(settings.Id)
	plan.TimeCreated = types.StringValue(settings.TimeCreated.String())
	plan.TimeModified = types.StringValue(settings.TimeModified.String())

	// Save plan into Terraform state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *switchPortSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state switchPortSettingsModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
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

	settings, err := r.client.NetworkingSwitchPortSettingsView(ctx, oxide.NetworkingSwitchPortSettingsViewParams{
		Port: oxide.NameOrId(state.ID.ValueString()),
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read Switch Port Settings:",
			"API error: "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("read Switch Port Settings with ID: %v", settings.Id), map[string]any{"success": true})

	// Map response body to schema
	state.ID = types.StringValue(settings.Id)
	state.Name = types.StringValue(string(settings.Name))
	state.Description = types.StringValue(settings.Description)
	state.TimeCreated = types.StringValue(settings.TimeCreated.String())
	state.TimeModified = types.StringValue(settings.TimeModified.String())

	model, diags := toTerraformModel(settings)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Addresses = model.Addresses
	state.BGPPeers = model.BGPPeers
	state.Groups = model.Groups
	state.Interfaces = model.Interfaces
	state.Links = model.Links
	state.PortConfig = model.PortConfig
	state.Routes = model.Routes

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *switchPortSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) { // plan is the resource data model for the update request.
	var plan switchPortSettingsModel
	// state is the resource data model for the current state.
	var state switchPortSettingsModel

	// Read Terraform plan data into the plan model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read Terraform prior state data into the state model to retrieve ID
	// which is a computed attribute, so it won't show up in the plan.
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout, diags := plan.Timeouts.Update(ctx, defaultTimeout())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	params, _ := toOxideParams(plan)

	settings, err := r.client.NetworkingSwitchPortSettingsCreate(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating switch port settings",
			"API error: "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("updated switch port settings with ID: %v", settings.Id), map[string]any{"success": true})

	// Map response body to schema and populate computed attribute values.
	plan.ID = types.StringValue(settings.Id)
	plan.TimeCreated = types.StringValue(settings.TimeCreated.String())
	plan.TimeModified = types.StringValue(settings.TimeModified.String())

	// Save plan into Terraform state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *switchPortSettingsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// state is the resource data model for the current state.
	var state switchPortSettingsModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := state.Timeouts.Delete(ctx, defaultTimeout())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	if err := r.client.NetworkingSwitchPortSettingsDelete(
		ctx,
		oxide.NetworkingSwitchPortSettingsDeleteParams{
			PortSettings: oxide.NameOrId(state.ID.ValueString()),
		}); err != nil {
		if !is404(err) {
			resp.Diagnostics.AddError(
				"Error deleting Switch Port Settings:",
				"API error: "+err.Error(),
			)
			return
		}
	}
	tflog.Trace(ctx, fmt.Sprintf("deleted Switch Port Settings with ID: %v", state.ID.ValueString()), map[string]any{"success": true})

}

func toTerraformModel(settings *oxide.SwitchPortSettings) (switchPortSettingsModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := switchPortSettingsModel{
		ID:          types.StringValue(settings.Id),
		Name:        types.StringValue(string(settings.Name)),
		Description: types.StringValue(settings.Description),
		PortConfig: &switchPortSettingsPortConfigModel{
			Geometry: types.StringValue(string(settings.Port.Geometry)),
		},
		TimeCreated:  types.StringValue(settings.TimeCreated.String()),
		TimeModified: types.StringValue(settings.TimeModified.String()),
	}

	addressesMap := make(map[string][]switchPortSettingsAddressAddressModel)
	for _, address := range settings.Addresses {
		if _, ok := addressesMap[string(address.InterfaceName)]; !ok {
			addressesMap[string(address.InterfaceName)] = make([]switchPortSettingsAddressAddressModel, 0)
		}

		addressesMap[string(address.InterfaceName)] = append(
			addressesMap[string(address.InterfaceName)],
			switchPortSettingsAddressAddressModel{
				Address:    types.StringValue(address.Address.(string)),
				AddressLot: types.StringValue(string(address.AddressLotId)),
				VlanID: func() types.Int32 {
					if address.VlanId != nil {
						return types.Int32Value(int32(*address.VlanId))
					}
					return types.Int32Null()
				}(),
			},
		)
	}
	addresses := make([]switchPortSettingsAddressModel, 0)
	for linkName, addrs := range addressesMap {
		addresses = append(addresses, switchPortSettingsAddressModel{
			Addresses: addrs,
			LinkName:  types.StringValue(linkName),
		})
	}
	model.Addresses = addresses

	bgpPeersMap := make(map[string][]switchPortSettingsBGPPeerPeerModel)
	for _, bgpPeer := range settings.BgpPeers {
		if _, ok := bgpPeersMap[string(bgpPeer.InterfaceName)]; !ok {
			bgpPeersMap[string(bgpPeer.InterfaceName)] = make([]switchPortSettingsBGPPeerPeerModel, 0)
		}

		allowedExportValue := make([]types.String, 0)
		for _, elem := range bgpPeer.AllowedExport.Value {
			allowedExportValue = append(allowedExportValue, types.StringValue(elem.(string)))
		}

		allowedImportValue := make([]types.String, 0)
		for _, elem := range bgpPeer.AllowedImport.Value {
			allowedImportValue = append(allowedImportValue, types.StringValue(elem.(string)))
		}

		communities := make([]types.Int64, 0)
		for _, communityStr := range bgpPeer.Communities {
			community, err := strconv.ParseInt(communityStr, 10, 64)
			if err != nil {
				diags.AddError(
					"Error parsing community element",
					fmt.Sprintf("Could not parse %s as int64: %v", communityStr, err),
				)
			}
			communities = append(communities, types.Int64Value(community))
		}

		bgpPeersMap[string(bgpPeer.InterfaceName)] = append(
			bgpPeersMap[string(bgpPeer.InterfaceName)],
			switchPortSettingsBGPPeerPeerModel{
				Addr: types.StringValue(bgpPeer.Addr),
				AllowedExport: &switchPortSettingsBGPPeerPeerAllowedExportModel{
					Type:  types.StringValue(string(bgpPeer.AllowedExport.Type)),
					Value: allowedExportValue,
				},
				AllowedImport: &switchPortSettingsBGPPeerPeerAllowedImportModel{
					Type:  types.StringValue(string(bgpPeer.AllowedImport.Type)),
					Value: allowedImportValue,
				},
				BGPConfig:              types.StringValue(string(bgpPeer.BgpConfig)),
				Communities:            communities,
				ConnectRetry:           types.Int64Value(int64(*bgpPeer.ConnectRetry)),
				DelayOpen:              types.Int64Value(int64(*bgpPeer.DelayOpen)),
				EnforceFirstAs:         types.BoolPointerValue(bgpPeer.EnforceFirstAs),
				HoldTime:               types.Int64Value(int64(*bgpPeer.HoldTime)),
				IdleHoldTime:           types.Int64Value(int64(*bgpPeer.IdleHoldTime)),
				InterfaceName:          types.StringValue(string(bgpPeer.InterfaceName)),
				Keepalive:              types.Int64Value(int64(*bgpPeer.Keepalive)),
				LocalPref:              types.Int64Value(int64(*bgpPeer.LocalPref)),
				MD5AuthKey:             types.StringValue(bgpPeer.Md5AuthKey),
				MinTTL:                 types.Int32Value(int32(*bgpPeer.MinTtl)),
				MultiExitDiscriminator: types.Int64Value(int64(*bgpPeer.MultiExitDiscriminator)),
				RemoteASN:              types.Int64Value(int64(*bgpPeer.RemoteAsn)),
				VlanID:                 types.Int32Value(int32(*bgpPeer.VlanId)),
			},
		)
	}
	bgpPeers := make([]switchPortSettingsBGPPeerModel, 0)
	for linkName, peers := range bgpPeersMap {
		bgpPeers = append(bgpPeers, switchPortSettingsBGPPeerModel{
			Peers:    peers,
			LinkName: types.StringValue(linkName),
		})
	}
	model.BGPPeers = bgpPeers

	groups := make([]types.String, 0)
	for _, group := range settings.Groups {
		groups = append(groups, types.StringValue(group.PortSettingsGroupId))
	}
	model.Groups = groups

	interfaces := make([]switchPortSettingsInterfaceModel, 0)
	for _, iface := range settings.Interfaces {
		interfaces = append(interfaces, switchPortSettingsInterfaceModel{
			Kind: &switchPortSettingsInterfaceKindModel{
				Type: types.StringValue(string(iface.Kind)),
			},
			LinkName:  types.StringValue(string(iface.InterfaceName)),
			V6Enabled: types.BoolPointerValue(iface.V6Enabled),
		})
	}
	model.Interfaces = interfaces

	links := make([]switchPortSettingsLinkModel, 0)
	for _, link := range settings.Links {
		lldp := &switchPortSettingsLinkLLDPModel{}
		if link.LldpLinkConfig != nil {
			lldp.ChassisID = types.StringValue(link.LldpLinkConfig.ChassisId)
			lldp.Enabled = types.BoolPointerValue(link.LldpLinkConfig.Enabled)
			lldp.LinkDescription = types.StringValue(link.LldpLinkConfig.LinkDescription)
			lldp.LinkName = types.StringValue(link.LldpLinkConfig.LinkName)
			lldp.ManagementIP = types.StringValue(link.LldpLinkConfig.ManagementIp)
			lldp.SystemDescription = types.StringValue(link.LldpLinkConfig.SystemDescription)
			lldp.SystemName = types.StringValue(link.LldpLinkConfig.SystemName)
		}

		txEq := &switchPortSettingsLinkTxEqModel{}
		if link.TxEqConfig != nil {
			txEq.Main = func() types.Int32 {
				if link.TxEqConfig.Main != nil {
					return types.Int32Value(int32(*link.TxEqConfig.Main))
				}
				return types.Int32Null()
			}()
			txEq.Post1 = func() types.Int32 {
				if link.TxEqConfig.Post1 != nil {
					return types.Int32Value(int32(*link.TxEqConfig.Post1))
				}
				return types.Int32Null()
			}()
			txEq.Post2 = func() types.Int32 {
				if link.TxEqConfig.Post2 != nil {
					return types.Int32Value(int32(*link.TxEqConfig.Post2))
				}
				return types.Int32Null()
			}()
			txEq.Pre1 = func() types.Int32 {
				if link.TxEqConfig.Pre1 != nil {
					return types.Int32Value(int32(*link.TxEqConfig.Pre1))
				}
				return types.Int32Null()
			}()
			txEq.Pre2 = func() types.Int32 {
				if link.TxEqConfig.Pre2 != nil {
					return types.Int32Value(int32(*link.TxEqConfig.Pre2))
				}
				return types.Int32Null()
			}()
		}

		links = append(links, switchPortSettingsLinkModel{
			Autoneg:  types.BoolPointerValue(link.Autoneg),
			FEC:      types.StringValue(string(link.Fec)),
			LinkName: types.StringValue(string(link.LinkName)),
			LLDP:     lldp,
			MTU:      types.Int32Value(int32(*link.Mtu)),
			Speed:    types.StringValue(string(link.Speed)),
			TxEq:     txEq,
		})
	}
	model.Links = links

	routesMap := make(map[string][]switchPortSettingsRouteRouteModel)
	for _, route := range settings.Routes {
		if _, ok := routesMap[string(route.InterfaceName)]; !ok {
			routesMap[string(route.InterfaceName)] = make([]switchPortSettingsRouteRouteModel, 0)
		}

		routesMap[string(route.InterfaceName)] = append(
			routesMap[string(route.InterfaceName)],
			switchPortSettingsRouteRouteModel{
				Dst: types.StringValue(route.Dst.(string)),
				GW:  types.StringValue(route.Gw),
				RIBPriority: func() types.Int32 {
					if route.RibPriority != nil {
						return types.Int32Value(int32(*route.RibPriority))
					}
					return types.Int32Null()
				}(),
				VID: func() types.Int32 {
					if route.VlanId != nil {
						return types.Int32Value(int32(*route.VlanId))
					}
					return types.Int32Null()
				}(),
			},
		)
	}
	routes := make([]switchPortSettingsRouteModel, 0)
	for linkName, rts := range routesMap {
		routes = append(routes, switchPortSettingsRouteModel{
			Routes:   rts,
			LinkName: types.StringValue(linkName),
		})
	}
	model.Routes = routes

	return model, diags
}

func toOxideParams(model switchPortSettingsModel) (oxide.NetworkingSwitchPortSettingsCreateParams, diag.Diagnostics) {

	params := oxide.NetworkingSwitchPortSettingsCreateParams{
		Body: &oxide.SwitchPortSettingsCreate{
			Name:        oxide.Name(model.Name.ValueString()),
			Description: model.Description.ValueString(),
			PortConfig: oxide.SwitchPortConfigCreate{
				Geometry: oxide.SwitchPortGeometry(model.PortConfig.Geometry.ValueString()),
			},
		},
	}

	addresses := make([]oxide.AddressConfig, 0)
	for _, address := range model.Addresses {
		addrs := make([]oxide.Address, 0)
		for _, addr := range address.Addresses {
			addrs = append(addrs, oxide.Address{
				Address:    oxide.IpNet(addr.Address.ValueString()),
				AddressLot: oxide.NameOrId(addr.AddressLot.ValueString()),
				VlanId: func() *int {
					if !addr.VlanID.IsNull() {
						return oxide.NewPointer(int(addr.VlanID.ValueInt32()))
					}
					return nil
				}(),
			})
		}

		addresses = append(addresses, oxide.AddressConfig{
			LinkName:  oxide.Name(address.LinkName.ValueString()),
			Addresses: addrs,
		})
	}
	params.Body.Addresses = addresses

	bgpPeers := make([]oxide.BgpPeerConfig, 0)
	for _, bgpPeer := range model.BGPPeers {
		peers := make([]oxide.BgpPeer, 0)
		for _, peer := range bgpPeer.Peers {
			allowedExportValue := make([]oxide.IpNet, 0)
			for _, value := range peer.AllowedExport.Value {
				allowedExportValue = append(allowedExportValue, oxide.IpNet(value.ValueString()))
			}

			allowedImportValue := make([]oxide.IpNet, 0)
			for _, value := range peer.AllowedImport.Value {
				allowedImportValue = append(allowedImportValue, oxide.IpNet(value.ValueString()))
			}

			communities := make([]string, 0)
			for _, community := range peer.Communities {
				communities = append(communities, fmt.Sprintf("%d", community.ValueInt64()))
			}

			peers = append(peers, oxide.BgpPeer{
				Addr: peer.Addr.ValueString(),
				AllowedExport: oxide.ImportExportPolicy{
					Type:  oxide.ImportExportPolicyType(peer.AllowedExport.Type.ValueString()),
					Value: allowedExportValue,
				},
				AllowedImport: oxide.ImportExportPolicy{
					Type:  oxide.ImportExportPolicyType(peer.AllowedImport.Type.ValueString()),
					Value: allowedImportValue,
				},
				BgpConfig:              oxide.NameOrId(peer.BGPConfig.ValueString()),
				Communities:            communities,
				ConnectRetry:           oxide.NewPointer(int(peer.ConnectRetry.ValueInt64())),
				DelayOpen:              oxide.NewPointer(int(peer.DelayOpen.ValueInt64())),
				EnforceFirstAs:         oxide.NewPointer(peer.EnforceFirstAs.ValueBool()),
				HoldTime:               oxide.NewPointer(int(peer.HoldTime.ValueInt64())),
				IdleHoldTime:           oxide.NewPointer(int(peer.IdleHoldTime.ValueInt64())),
				InterfaceName:          oxide.Name(peer.InterfaceName.ValueString()),
				Keepalive:              oxide.NewPointer(int(peer.Keepalive.ValueInt64())),
				LocalPref:              oxide.NewPointer(int(peer.LocalPref.ValueInt64())),
				Md5AuthKey:             peer.MD5AuthKey.ValueString(),
				MinTtl:                 oxide.NewPointer(int(peer.MinTTL.ValueInt32())),
				MultiExitDiscriminator: oxide.NewPointer(int(peer.MultiExitDiscriminator.ValueInt64())),
				RemoteAsn:              oxide.NewPointer(int(peer.RemoteASN.ValueInt64())),
				VlanId:                 oxide.NewPointer(int(peer.VlanID.ValueInt32())),
			})
		}

		bgpPeers = append(bgpPeers, oxide.BgpPeerConfig{
			LinkName: oxide.Name(bgpPeer.LinkName.ValueString()),
			Peers:    peers,
		})
	}
	params.Body.BgpPeers = bgpPeers

	groups := make([]oxide.NameOrId, 0)
	for _, group := range model.Groups {
		groups = append(groups, oxide.NameOrId(group.ValueString()))
	}
	params.Body.Groups = groups

	interfaces := make([]oxide.SwitchInterfaceConfigCreate, 0)
	for _, iface := range model.Interfaces {
		interfaces = append(interfaces, oxide.SwitchInterfaceConfigCreate{
			Kind: oxide.SwitchInterfaceKind{
				Type: oxide.SwitchInterfaceKindType(iface.Kind.Type.ValueString()),
				Vid: func() *int {
					if !iface.Kind.VID.IsNull() {
						return oxide.NewPointer(int(iface.Kind.VID.ValueInt32()))
					}
					return nil
				}(),
			},
			LinkName:  oxide.Name(iface.LinkName.ValueString()),
			V6Enabled: oxide.NewPointer(iface.V6Enabled.ValueBool()),
		})
	}
	params.Body.Interfaces = interfaces

	links := make([]oxide.LinkConfigCreate, 0)
	for _, link := range model.Links {
		var txeq *oxide.TxEqConfig
		if link.TxEq != nil {
			txeq = &oxide.TxEqConfig{
				Main:  oxide.NewPointer(int(link.TxEq.Main.ValueInt32())),
				Post1: oxide.NewPointer(int(link.TxEq.Post1.ValueInt32())),
				Post2: oxide.NewPointer(int(link.TxEq.Post2.ValueInt32())),
				Pre1:  oxide.NewPointer(int(link.TxEq.Pre1.ValueInt32())),
				Pre2:  oxide.NewPointer(int(link.TxEq.Pre2.ValueInt32())),
			}
		}

		links = append(links, oxide.LinkConfigCreate{
			Autoneg:  link.Autoneg.ValueBoolPointer(),
			Fec:      oxide.LinkFec(link.FEC.ValueString()),
			LinkName: oxide.Name(link.LinkName.ValueString()),
			Lldp: oxide.LldpLinkConfigCreate{
				ChassisId:         link.LLDP.ChassisID.ValueString(),
				Enabled:           link.LLDP.Enabled.ValueBoolPointer(),
				LinkDescription:   link.LLDP.LinkDescription.ValueString(),
				LinkName:          link.LLDP.LinkName.ValueString(),
				ManagementIp:      link.LLDP.ManagementIP.ValueString(),
				SystemDescription: link.LLDP.SystemDescription.ValueString(),
				SystemName:        link.LLDP.SystemName.ValueString(),
			},
			Mtu:   oxide.NewPointer(int(link.MTU.ValueInt32())),
			Speed: oxide.LinkSpeed(link.Speed.ValueString()),
			TxEq:  txeq,
		})
	}
	params.Body.Links = links

	routes := make([]oxide.RouteConfig, 0)
	for _, route := range model.Routes {
		rts := make([]oxide.Route, 0)
		for _, rt := range route.Routes {
			var ribPriority *int
			if rt.RIBPriority.ValueInt32Pointer() != nil {
				ribPriority = oxide.NewPointer(int(rt.RIBPriority.ValueInt32()))
			}

			var vid *int
			if rt.VID.ValueInt32Pointer() != nil {
				vid = oxide.NewPointer(int(rt.VID.ValueInt32()))
			}

			rts = append(rts, oxide.Route{
				Dst:         oxide.IpNet(rt.Dst.ValueString()),
				Gw:          rt.GW.ValueString(),
				RibPriority: ribPriority,
				Vid:         vid,
			})
		}

		routes = append(routes, oxide.RouteConfig{
			LinkName: oxide.Name(route.LinkName.ValueString()),
			Routes:   rts,
		})
	}
	params.Body.Routes = routes

	return params, nil
}

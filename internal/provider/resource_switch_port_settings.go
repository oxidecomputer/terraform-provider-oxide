// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
	Address      types.String `tfsdk:"address"`
	AddressLotID types.String `tfsdk:"address_lot_id"`
	VlanID       types.Int32  `tfsdk:"vlan_id"`
}

type switchPortSettingsBGPPeerModel struct {
	LinkName types.String                         `tfsdk:"link_name"`
	Peers    []switchPortSettingsBGPPeerPeerModel `tfsdk:"peers"`
}

type switchPortSettingsBGPPeerPeerModel struct {
	Address                types.String                                     `tfsdk:"address"`
	AllowedExport          *switchPortSettingsBGPPeerPeerAllowedExportModel `tfsdk:"allowed_export"`
	AllowedImport          *switchPortSettingsBGPPeerPeerAllowedImportModel `tfsdk:"allowed_import"`
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

// NewSwitchPortSettingsResource contructs a Terraform resource.
func NewSwitchPortSettingsResource() resource.Resource {
	return &switchPortSettingsResource{}
}

// Metadata sets the metadata for the resource.
func (r *switchPortSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "oxide_switch_port_settings"
}

// Configure sets data needed by other methods for this resources.
func (r *switchPortSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*oxide.Client)
}

// ImportState contains logic on how to import the resource.
func (r *switchPortSettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Schema defines the Terraform configuration for this resource.
func (r *switchPortSettingsResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier of the switch port settings.",
			},
			"addresses": schema.SetNestedAttribute{
				Required:    true,
				Description: "Address configuration for the switch port.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"link_name": schema.StringAttribute{
							Required:    true,
							Description: "Name of the link for the address configuration.",
						},
						"addresses": schema.SetNestedAttribute{
							Required:    true,
							Description: "Set of addresses to assign to the link.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"address": schema.StringAttribute{
										Required:    true,
										Description: "IPv4 or IPv6 address, including the subnet mask.",
									},
									"address_lot_id": schema.StringAttribute{
										Required:    true,
										Description: "Address lot the address is allocated from.",
									},
									"vlan_id": schema.Int32Attribute{
										Optional:    true,
										Description: "VLAN ID for the address.",
									},
								},
							},
						},
					},
				},
			},
			"bgp_peers": schema.SetNestedAttribute{
				Optional:    true,
				Description: "BGP peer configuration for the switch port.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"link_name": schema.StringAttribute{
							Required:    true,
							Description: "Name of the link for the BGP peers configuration.",
						},
						"peers": schema.SetNestedAttribute{
							Required:    true,
							Description: "Set of BGP peers configuration to assign to the link.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"address": schema.StringAttribute{
										Required:    true,
										Description: "Address of the host to peer with.",
									},
									"allowed_export": schema.SingleNestedAttribute{
										Required:    true,
										Description: "Export policy for the peer.",
										Attributes: map[string]schema.Attribute{
											"type": schema.StringAttribute{
												Required:    true,
												Description: "Type of filter to apply.",
												Validators: []validator.String{
													stringvalidator.OneOf(
														string(oxide.ImportExportPolicyTypeNoFiltering),
														string(oxide.ImportExportPolicyTypeAllow),
													),
												},
											},
											"value": schema.SetAttribute{
												Optional:    true,
												ElementType: types.StringType,
												Description: "IPv4 or IPv6 address to apply the filter to, including the subnet mask.",
											},
										},
									},
									"allowed_import": schema.SingleNestedAttribute{
										Required:    true,
										Description: "Import policy for the peer.",
										Attributes: map[string]schema.Attribute{
											"type": schema.StringAttribute{
												Required:    true,
												Description: "Type of filter to apply.",
												Validators: []validator.String{
													stringvalidator.OneOf(
														string(oxide.ImportExportPolicyTypeNoFiltering),
														string(oxide.ImportExportPolicyTypeAllow),
													),
												},
											},
											"value": schema.SetAttribute{
												Optional:    true,
												ElementType: types.StringType,
												Description: "IPv4 or IPv6 address to apply the filter to, including the subnet mask.",
											},
										},
									},
									"bgp_config": schema.StringAttribute{
										Required:    true,
										Description: "Name or ID of the global BGP configuration used for establishing a session with this peer.",
									},
									"communities": schema.SetAttribute{
										Required:    true,
										ElementType: types.Int64Type,
										Description: "BGP communities to apply to this peer's routes.",
									},
									"connect_retry": schema.Int64Attribute{
										Required:    true,
										Description: "Number of seconds to wait before retrying a TCP connection.",
									},
									"delay_open": schema.Int64Attribute{
										Required:    true,
										Description: "Number of seconds to delay sending an open request after establishing a TCP session.",
									},
									"enforce_first_as": schema.BoolAttribute{
										Required:    true,
										Description: "Whether to enforce that the first autonomous system in paths received from this peer is the peer's autonomous system.",
									},
									"hold_time": schema.Int64Attribute{
										Required:    true,
										Description: "Number of seconds to hold peer connections between keepalives.",
									},
									"idle_hold_time": schema.Int64Attribute{
										Required:    true,
										Description: "Number of seconds to hold a peer in idle before attempting a new session.",
									},
									"interface_name": schema.StringAttribute{
										Required:    true,
										Description: "Name of the interface to use for this BGP peer session.",
									},
									"keepalive": schema.Int64Attribute{
										Required:    true,
										Description: "Number of seconds between sending BGP keepalive requests.",
									},
									"local_pref": schema.Int64Attribute{
										Optional:    true,
										Description: "BGP local preference value for routes received from this peer.",
									},
									"md5_auth_key": schema.StringAttribute{
										Optional:    true,
										Description: "MD5 authentication key for this BGP session.",
									},
									"min_ttl": schema.Int32Attribute{
										Optional:    true,
										Description: "Minimum acceptable TTL for BGP packets from this peer.",
									},
									"multi_exit_discriminator": schema.Int64Attribute{
										Optional:    true,
										Description: "Multi-exit discriminator (MED) to advertise to this peer.",
									},
									"remote_asn": schema.Int64Attribute{
										Optional:    true,
										Description: "Remote autonomous system number for this BGP peer.",
									},
									"vlan_id": schema.Int32Attribute{
										Optional:    true,
										Description: "VLAN ID for this BGP peer session.",
									},
								},
							},
						},
					},
				},
			},
			"description": schema.StringAttribute{
				Required:    true,
				Description: "Human-readable description of the switch port settings.",
			},
			"groups": schema.SetAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Set of port settings group IDs to include in these settings.",
			},
			"interfaces": schema.SetNestedAttribute{
				Optional:    true,
				Description: "Interface configuration for the switch port.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"kind": schema.SingleNestedAttribute{
							Required:    true,
							Description: "The kind of interface this configuration represents.",
							Attributes: map[string]schema.Attribute{
								"type": schema.StringAttribute{
									Required:    true,
									Description: "Type of the interface.",
									Validators: []validator.String{
										stringvalidator.OneOf(
											string(oxide.SwitchInterfaceKindTypePrimary),
											string(oxide.SwitchInterfaceKindTypeVlan),
											string(oxide.SwitchInterfaceKindTypeLoopback),
										),
									},
								},
								"vid": schema.Int32Attribute{
									Optional:    true,
									Description: "VLAN ID for the interfaces.",
								},
							},
						},
						"link_name": schema.StringAttribute{
							Required:    true,
							Description: "Name of the link this interface is associated with.",
						},
						"v6_enabled": schema.BoolAttribute{
							Optional:    true,
							Description: "Enable IPv6 on this interface.",
						},
					},
				},
			},
			"links": schema.SetNestedAttribute{
				Required:    true,
				Description: "Link configuration for the switch port.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"autoneg": schema.BoolAttribute{
							Required:    true,
							Description: "Whether to enable auto-negotiation for this link.",
						},
						"fec": schema.StringAttribute{
							Optional:    true,
							Description: "Forward error correction (FEC) type.",
							Validators: []validator.String{
								stringvalidator.OneOf(
									string(oxide.LinkFecFirecode),
									string(oxide.LinkFecNone),
									string(oxide.LinkFecRs),
								),
							},
						},
						"link_name": schema.StringAttribute{
							Required:    true,
							Description: "Name of the link.",
						},
						"lldp": schema.SingleNestedAttribute{
							Required:    true,
							Description: "Link Layer Discovery Protocol (LLDP) configuration.",
							Attributes: map[string]schema.Attribute{
								"chassis_id": schema.StringAttribute{
									Optional:    true,
									Description: "LLDP chassis ID.",
								},
								"enabled": schema.BoolAttribute{
									Required:    true,
									Description: "Whether to enable LLDP on this link.",
								},
								"link_description": schema.StringAttribute{
									Optional:    true,
									Description: "LLDP link description.",
								},
								"link_name": schema.StringAttribute{
									Optional:    true,
									Description: "LLDP link name.",
								},
								"management_ip": schema.StringAttribute{
									Optional:    true,
									Description: "LLDP management IP address.",
								},
								"system_description": schema.StringAttribute{
									Optional:    true,
									Description: "LLDP system description.",
								},
								"system_name": schema.StringAttribute{
									Optional:    true,
									Description: "LLDP system name.",
								},
							},
						},
						"mtu": schema.Int32Attribute{
							Required:    true,
							Description: "Maximum Transmission Unit (MTU) for this link.",
						},
						"speed": schema.StringAttribute{
							Required:    true,
							Description: "Link speed.",
							Validators: []validator.String{
								stringvalidator.OneOf(
									string(oxide.LinkSpeedSpeed0G),
									string(oxide.LinkSpeedSpeed1G),
									string(oxide.LinkSpeedSpeed10G),
									string(oxide.LinkSpeedSpeed25G),
									string(oxide.LinkSpeedSpeed40G),
									string(oxide.LinkSpeedSpeed50G),
									string(oxide.LinkSpeedSpeed100G),
									string(oxide.LinkSpeedSpeed200G),
									string(oxide.LinkSpeedSpeed400G),
								),
							},
						},
						"tx_eq": schema.SingleNestedAttribute{
							Optional:    true,
							Description: "Transceiver equalization settings.",
							Attributes: map[string]schema.Attribute{
								"main": schema.Int32Attribute{
									Optional:    true,
									Description: "Main tap equalization value.",
								},
								"post1": schema.Int32Attribute{
									Optional:    true,
									Description: "Post-cursor tap1 equalization value.",
								},
								"post2": schema.Int32Attribute{
									Optional:    true,
									Description: "Post-cursor tap2 equalization value.",
								},
								"pre1": schema.Int32Attribute{
									Optional:    true,
									Description: "Pre-cursor tap1 equalization value.",
								},
								"pre2": schema.Int32Attribute{
									Optional:    true,
									Description: "Pre-cursor tap2 equalization value.",
								},
							},
						},
					},
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the switch port settings.",
			},
			"port_config": schema.SingleNestedAttribute{
				Required:    true,
				Description: "Physical port configuration.",
				Attributes: map[string]schema.Attribute{
					"geometry": schema.StringAttribute{
						Required:    true,
						Description: "Port geometry.",
						Validators: []validator.String{
							stringvalidator.OneOf(
								string(oxide.SwitchPortGeometryQsfp28X1),
								string(oxide.SwitchPortGeometryQsfp28X2),
								string(oxide.SwitchPortGeometrySfp28X4),
							),
						},
					},
				},
			},
			"routes": schema.SetNestedAttribute{
				Optional:    true,
				Description: "Static route configuration.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"link_name": schema.StringAttribute{
							Required:    true,
							Description: "Name of the link for these routes.",
						},
						"routes": schema.SetNestedAttribute{
							Required:    true,
							Description: "Set of static routes for this link.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"dst": schema.StringAttribute{
										Required:    true,
										Description: "Destination network in CIDR notation.",
									},
									"gw": schema.StringAttribute{
										Required:    true,
										Description: "Gateway IP address for this route.",
									},
									"rib_priority": schema.Int32Attribute{
										Optional:    true,
										Description: "Routing Information Base (RIB) priority for this route.",
									},
									"vid": schema.Int32Attribute{
										Optional:    true,
										Description: "VLAN ID for this route.",
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
				Description: "Timestamp of when the switch port settings were created.",
			},
			"time_modified": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when the switch port settings were last modified.",
			},
		},
	}
}

// Create sets the switch port settings.
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

	params, diags := toNetworkingSwitchPortSettingsCreateParams(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

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

	// Read Terraform prior state data into the model.
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

	// Map response body to schema.
	state.ID = types.StringValue(settings.Id)
	state.Name = types.StringValue(string(settings.Name))
	state.Description = types.StringValue(settings.Description)
	state.TimeCreated = types.StringValue(settings.TimeCreated.String())
	state.TimeModified = types.StringValue(settings.TimeModified.String())

	model, diags := toSwitchPortSettingsModel(settings)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// These fields are populated from the built Terraform model to handle the
	// asymmetry of the API and edge cases with mapping Oxide API types to Terraform
	// types.
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
// There is no Oxide API to update switch port settings so all the switch port
// settings are overwritten using the `switch_port_settings_create` API.
func (r *switchPortSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) { // plan is the resource data model for the update request.
	// Read the Terraform plan data.
	var plan switchPortSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the Terraform state data to retreive compute attributes.
	var state switchPortSettingsModel
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

	params, diags := toNetworkingSwitchPortSettingsCreateParams(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

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
	// Read the Terraform state.
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

	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
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

// toSwitchPortSettingsModel converts [oxide.SwitchPortSettings]
// to [switchPortSettingsModel]. It's used deserialize the Oxide
// `switch_port_settings_create` API response into the Terraform data model used
// by this resource.
//
// This function is quite long so let's break down the core of what it's doing.
//
// 1. This function handles the asymmetrical nature of the
// `switch_port_settings_create` API. For example, the request body
// accepts attributes such as `addresses[].addresses[].address` but returns
// `addresses[].address`. This function handles that conversion by creating a
// map of all the respective configurations for a given link name and using it
// to populate the nested model.
//
// 2. This function assumes null values by default to prevent Terraform from
// having a non-empty refresh plan right after a successful apply. For example,
// assume the Terraform configuration omits the `bgp_peers` attribute. After an
// apply, the Oxide `switch_port_settings_create` API returns `"bgp_peers": []`
// and Terraform plans must see this as a null value to match the configuration.
// Otherwise, the Terraform refresh plan will read the value `[]` and attempt to
// set it back to null since the configuration had `bgp_peers` omitted. This is
// why you'll see `if len(settings.Addresses) > 0` conditions.
//
// 3. The above point about assuming null values applies to other nested
// attributes as well. That's why you'll see a bunch of anonymous functions to
// set values to null if they are either null or their zero value as retrieved
// from Oxide.
func toSwitchPortSettingsModel(settings *oxide.SwitchPortSettings) (switchPortSettingsModel, diag.Diagnostics) {
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

	//
	// Addresses
	//
	if len(settings.Addresses) > 0 {
		linkToAddrs := make(map[string][]switchPortSettingsAddressAddressModel)
		for _, address := range settings.Addresses {
			link := string(address.InterfaceName)

			if _, ok := linkToAddrs[link]; !ok {
				linkToAddrs[link] = make([]switchPortSettingsAddressAddressModel, 0)
			}

			addressModel := switchPortSettingsAddressAddressModel{
				Address:      types.StringValue(address.Address.(string)),
				AddressLotID: types.StringValue(string(address.AddressLotId)),
				VlanID: func() types.Int32 {
					if address.VlanId == nil {
						return types.Int32Null()
					}
					return types.Int32Value(int32(*address.VlanId))
				}(),
			}

			linkToAddrs[link] = append(linkToAddrs[link], addressModel)
		}

		addressModels := make([]switchPortSettingsAddressModel, 0)
		for linkName, addrModels := range linkToAddrs {
			addressModel := switchPortSettingsAddressModel{
				Addresses: addrModels,
				LinkName:  types.StringValue(linkName),
			}
			addressModels = append(addressModels, addressModel)
		}

		model.Addresses = addressModels
	}

	//
	// BGP Peers
	//
	if len(settings.BgpPeers) > 0 {
		linkToBGPPeer := make(map[string][]switchPortSettingsBGPPeerPeerModel)
		for _, bgpPeer := range settings.BgpPeers {
			link := string(bgpPeer.InterfaceName)

			if _, ok := linkToBGPPeer[link]; !ok {
				linkToBGPPeer[link] = make([]switchPortSettingsBGPPeerPeerModel, 0)
			}

			bgpPeerModel := switchPortSettingsBGPPeerPeerModel{
				Address:        types.StringValue(bgpPeer.Addr),
				BGPConfig:      types.StringValue(string(bgpPeer.BgpConfig)),
				ConnectRetry:   types.Int64Value(int64(*bgpPeer.ConnectRetry)),
				DelayOpen:      types.Int64Value(int64(*bgpPeer.DelayOpen)),
				EnforceFirstAs: types.BoolPointerValue(bgpPeer.EnforceFirstAs),
				HoldTime:       types.Int64Value(int64(*bgpPeer.HoldTime)),
				IdleHoldTime:   types.Int64Value(int64(*bgpPeer.IdleHoldTime)),
				InterfaceName:  types.StringValue(string(bgpPeer.InterfaceName)),
				Keepalive:      types.Int64Value(int64(*bgpPeer.Keepalive)),

				// The fields below are nullable so we handle them specially.
				LocalPref: func() types.Int64 {
					if bgpPeer.LocalPref == nil {
						return types.Int64Null()
					}
					return types.Int64Value(int64(*bgpPeer.LocalPref))
				}(),
				MD5AuthKey: func() types.String {
					if bgpPeer.Md5AuthKey == "" {
						return types.StringNull()
					}
					return types.StringValue(bgpPeer.Md5AuthKey)
				}(),
				MinTTL: func() types.Int32 {
					if bgpPeer.MinTtl == nil {
						return types.Int32Null()
					}
					return types.Int32Value(int32(*bgpPeer.MinTtl))
				}(),
				MultiExitDiscriminator: func() types.Int64 {
					if bgpPeer.MultiExitDiscriminator == nil {
						return types.Int64Null()
					}
					return types.Int64Value(int64(*bgpPeer.MultiExitDiscriminator))
				}(),
				RemoteASN: func() types.Int64 {
					if bgpPeer.RemoteAsn == nil {
						return types.Int64Null()
					}
					return types.Int64Value(int64(*bgpPeer.RemoteAsn))
				}(),
				VlanID: func() types.Int32 {
					if bgpPeer.VlanId == nil {
						return types.Int32Null()
					}
					return types.Int32Value(int32(*bgpPeer.VlanId))
				}(),
			}

			bgpPeerModel.AllowedExport = &switchPortSettingsBGPPeerPeerAllowedExportModel{
				Type: types.StringValue(string(bgpPeer.AllowedExport.Type)),
				Value: func() []types.String {
					res := make([]types.String, 0)
					for _, elem := range bgpPeer.AllowedExport.Value {
						res = append(res, types.StringValue(elem.(string)))
					}
					return res
				}(),
			}

			bgpPeerModel.AllowedImport = &switchPortSettingsBGPPeerPeerAllowedImportModel{
				Type: types.StringValue(string(bgpPeer.AllowedImport.Type)),
				Value: func() []types.String {
					res := make([]types.String, 0)
					for _, elem := range bgpPeer.AllowedImport.Value {
						res = append(res, types.StringValue(elem.(string)))
					}
					return res
				}(),
			}

			bgpPeerModel.Communities = func() []types.Int64 {
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
				return communities
			}()

			linkToBGPPeer[link] = append(linkToBGPPeer[link], bgpPeerModel)
		}

		bgpPeersModels := make([]switchPortSettingsBGPPeerModel, 0)
		for linkName, bgpPeers := range linkToBGPPeer {
			bgpPeerModel := switchPortSettingsBGPPeerModel{
				Peers:    bgpPeers,
				LinkName: types.StringValue(linkName),
			}
			bgpPeersModels = append(bgpPeersModels, bgpPeerModel)
		}

		model.BGPPeers = bgpPeersModels
	}

	//
	// Groups
	//
	if len(settings.Groups) > 0 {
		groupModels := make([]types.String, 0)
		for _, group := range settings.Groups {
			groupModel := types.StringValue(group.PortSettingsGroupId)
			groupModels = append(groupModels, groupModel)
		}
		model.Groups = groupModels
	}

	//
	// Interfaces
	//
	if len(settings.Interfaces) > 0 {
		interfaceModels := make([]switchPortSettingsInterfaceModel, 0)
		for _, iface := range settings.Interfaces {
			interfaceModel := switchPortSettingsInterfaceModel{
				Kind: &switchPortSettingsInterfaceKindModel{
					Type: types.StringValue(string(iface.Kind)),
				},
				LinkName:  types.StringValue(string(iface.InterfaceName)),
				V6Enabled: types.BoolPointerValue(iface.V6Enabled),
			}
			interfaceModels = append(interfaceModels, interfaceModel)
		}
		model.Interfaces = interfaceModels
	}

	//
	// Links
	//
	if len(settings.Links) > 0 {
		linkModels := make([]switchPortSettingsLinkModel, 0)
		for _, link := range settings.Links {
			linkModel := switchPortSettingsLinkModel{
				Autoneg: func() types.Bool {
					if link.Autoneg == nil {
						return types.BoolNull()
					}
					return types.BoolPointerValue(link.Autoneg)
				}(),
				FEC: func() types.String {
					if link.Fec == "" {
						return types.StringNull()
					}
					return types.StringValue(string(link.Fec))
				}(),
				LinkName: types.StringValue(string(link.LinkName)),
				MTU: func() types.Int32 {
					if link.Mtu == nil {
						return types.Int32Null()
					}
					return types.Int32Value(int32(*link.Mtu))
				}(),
				Speed: types.StringValue(string(link.Speed)),
			}

			if link.LldpLinkConfig != nil {
				linkModel.LLDP = &switchPortSettingsLinkLLDPModel{
					Enabled: types.BoolPointerValue(link.LldpLinkConfig.Enabled),
				}

				if *link.LldpLinkConfig.Enabled {
					linkModel.LLDP.ChassisID = func() types.String {
						if link.LldpLinkConfig.ChassisId == "" {
							return types.StringNull()
						}
						return types.StringValue(link.LldpLinkConfig.ChassisId)
					}()
					linkModel.LLDP.LinkDescription = func() types.String {
						if link.LldpLinkConfig.LinkDescription == "" {
							return types.StringNull()
						}
						return types.StringValue(link.LldpLinkConfig.LinkDescription)
					}()
					linkModel.LLDP.LinkName = func() types.String {
						if link.LldpLinkConfig.LinkName == "" {
							return types.StringNull()
						}
						return types.StringValue(link.LldpLinkConfig.LinkName)
					}()
					linkModel.LLDP.ManagementIP = func() types.String {
						if link.LldpLinkConfig.ManagementIp == "" {
							return types.StringNull()
						}
						return types.StringValue(link.LldpLinkConfig.ManagementIp)
					}()
					linkModel.LLDP.SystemDescription = func() types.String {
						if link.LldpLinkConfig.SystemDescription == "" {
							return types.StringNull()
						}
						return types.StringValue(link.LldpLinkConfig.SystemDescription)
					}()
					linkModel.LLDP.SystemName = func() types.String {
						if link.LldpLinkConfig.SystemName == "" {
							return types.StringNull()
						}
						return types.StringValue(link.LldpLinkConfig.SystemName)
					}()
				}
			}

			if link.TxEqConfig != nil {
				linkModel.TxEq = &switchPortSettingsLinkTxEqModel{
					Main: func() types.Int32 {
						if link.TxEqConfig.Main == nil {
							return types.Int32Null()
						}
						return types.Int32Value(int32(*link.TxEqConfig.Main))
					}(),
					Post1: func() types.Int32 {
						if link.TxEqConfig.Post1 == nil {
							return types.Int32Null()
						}
						return types.Int32Value(int32(*link.TxEqConfig.Post1))
					}(),
					Post2: func() types.Int32 {
						if link.TxEqConfig.Post2 == nil {
							return types.Int32Null()
						}
						return types.Int32Value(int32(*link.TxEqConfig.Post2))
					}(),
					Pre1: func() types.Int32 {
						if link.TxEqConfig.Pre1 == nil {
							return types.Int32Null()
						}
						return types.Int32Value(int32(*link.TxEqConfig.Pre1))
					}(),
					Pre2: func() types.Int32 {
						if link.TxEqConfig.Pre2 == nil {
							return types.Int32Null()
						}
						return types.Int32Value(int32(*link.TxEqConfig.Pre2))
					}(),
				}
			}

			linkModels = append(linkModels, linkModel)
		}

		model.Links = linkModels
	}

	//
	// Routes
	//
	if len(settings.Routes) > 0 {
		linkToRoutes := make(map[string][]switchPortSettingsRouteRouteModel)
		for _, route := range settings.Routes {
			link := string(route.InterfaceName)

			if _, ok := linkToRoutes[link]; !ok {
				linkToRoutes[link] = make([]switchPortSettingsRouteRouteModel, 0)
			}

			routeModel := switchPortSettingsRouteRouteModel{
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
			}

			linkToRoutes[link] = append(linkToRoutes[link], routeModel)
		}

		routeModels := make([]switchPortSettingsRouteModel, 0)
		for linkName, rts := range linkToRoutes {
			routeModel := switchPortSettingsRouteModel{
				Routes:   rts,
				LinkName: types.StringValue(linkName),
			}
			routeModels = append(routeModels, routeModel)
		}

		model.Routes = routeModels
	}

	return model, diags
}

// toNetworkingSwitchPortSettingsCreateParams converts [switchPortSettingsModel`
// to [oxide.NetworkingSwitchPortSettingsCreateParams]. This is far simpler than
// [toSwitchPortSettingsModel] since the Oxide `switch_port_settings_create` API
// request body matches the Terraform schema.
func toNetworkingSwitchPortSettingsCreateParams(model switchPortSettingsModel) (oxide.NetworkingSwitchPortSettingsCreateParams, diag.Diagnostics) {
	params := oxide.NetworkingSwitchPortSettingsCreateParams{
		Body: &oxide.SwitchPortSettingsCreate{
			Name:        oxide.Name(model.Name.ValueString()),
			Description: model.Description.ValueString(),
			PortConfig: oxide.SwitchPortConfigCreate{
				Geometry: oxide.SwitchPortGeometry(model.PortConfig.Geometry.ValueString()),
			},
		},
	}

	//
	// Addresses
	//
	addressConfigs := make([]oxide.AddressConfig, 0)
	for _, addressModel := range model.Addresses {
		addresses := make([]oxide.Address, 0)
		for _, addressModelNested := range addressModel.Addresses {
			address := oxide.Address{
				Address:    oxide.IpNet(addressModelNested.Address.ValueString()),
				AddressLot: oxide.NameOrId(addressModelNested.AddressLotID.ValueString()),
				VlanId: func() *int {
					if addressModelNested.VlanID.IsNull() {
						return nil
					}
					return oxide.NewPointer(int(addressModelNested.VlanID.ValueInt32()))
				}(),
			}

			addresses = append(addresses, address)
		}

		addressConfig := oxide.AddressConfig{
			LinkName:  oxide.Name(addressModel.LinkName.ValueString()),
			Addresses: addresses,
		}

		addressConfigs = append(addressConfigs, addressConfig)
	}
	params.Body.Addresses = addressConfigs

	//
	// BGPPeers
	//
	bgpPeerConfigs := make([]oxide.BgpPeerConfig, 0)
	for _, bgpPeerModel := range model.BGPPeers {
		bgpPeers := make([]oxide.BgpPeer, 0)
		for _, bgpModelNested := range bgpPeerModel.Peers {
			bgpPeer := oxide.BgpPeer{
				Addr:           bgpModelNested.Address.ValueString(),
				BgpConfig:      oxide.NameOrId(bgpModelNested.BGPConfig.ValueString()),
				ConnectRetry:   oxide.NewPointer(int(bgpModelNested.ConnectRetry.ValueInt64())),
				DelayOpen:      oxide.NewPointer(int(bgpModelNested.DelayOpen.ValueInt64())),
				EnforceFirstAs: oxide.NewPointer(bgpModelNested.EnforceFirstAs.ValueBool()),
				HoldTime:       oxide.NewPointer(int(bgpModelNested.HoldTime.ValueInt64())),
				IdleHoldTime:   oxide.NewPointer(int(bgpModelNested.IdleHoldTime.ValueInt64())),
				InterfaceName:  oxide.Name(bgpModelNested.InterfaceName.ValueString()),
				Keepalive:      oxide.NewPointer(int(bgpModelNested.Keepalive.ValueInt64())),
				LocalPref: func() *int {
					if bgpModelNested.LocalPref.IsNull() {
						return nil
					}
					return oxide.NewPointer(int(bgpModelNested.LocalPref.ValueInt64()))
				}(),
				Md5AuthKey: bgpModelNested.MD5AuthKey.ValueString(),
				MinTtl: func() *int {
					if bgpModelNested.MinTTL.IsNull() {
						return nil
					}
					return oxide.NewPointer(int(bgpModelNested.MinTTL.ValueInt32()))
				}(),
				MultiExitDiscriminator: func() *int {
					if bgpModelNested.MultiExitDiscriminator.IsNull() {
						return nil
					}
					return oxide.NewPointer(int(bgpModelNested.MultiExitDiscriminator.ValueInt64()))
				}(),
				RemoteAsn: func() *int {
					if bgpModelNested.RemoteASN.IsNull() {
						return nil
					}
					return oxide.NewPointer(int(bgpModelNested.RemoteASN.ValueInt64()))
				}(),
				VlanId: func() *int {
					if bgpModelNested.VlanID.IsNull() {
						return nil
					}
					return oxide.NewPointer(int(bgpModelNested.VlanID.ValueInt32()))
				}(),
			}

			bgpPeer.AllowedExport = oxide.ImportExportPolicy{
				Type: oxide.ImportExportPolicyType(bgpModelNested.AllowedExport.Type.ValueString()),
				Value: func() []oxide.IpNet {
					if len(bgpModelNested.AllowedExport.Value) == 0 {
						return nil
					}

					values := make([]oxide.IpNet, 0)
					for _, value := range bgpModelNested.AllowedExport.Value {
						values = append(values, oxide.IpNet(value.ValueString()))
					}

					return values
				}(),
			}

			bgpPeer.AllowedImport = oxide.ImportExportPolicy{
				Type: oxide.ImportExportPolicyType(bgpModelNested.AllowedImport.Type.ValueString()),
				Value: func() []oxide.IpNet {
					if len(bgpModelNested.AllowedImport.Value) == 0 {
						return nil
					}

					values := make([]oxide.IpNet, 0)
					for _, value := range bgpModelNested.AllowedImport.Value {
						values = append(values, oxide.IpNet(value.ValueString()))
					}

					return values
				}(),
			}

			bgpPeer.Communities = func() []string {
				communities := make([]string, 0)
				for _, community := range bgpModelNested.Communities {
					communities = append(communities, fmt.Sprintf("%d", community.ValueInt64()))
				}
				return communities
			}()

			bgpPeers = append(bgpPeers, bgpPeer)
		}

		bgpPeerConfig := oxide.BgpPeerConfig{
			LinkName: oxide.Name(bgpPeerModel.LinkName.ValueString()),
			Peers:    bgpPeers,
		}

		bgpPeerConfigs = append(bgpPeerConfigs, bgpPeerConfig)
	}
	params.Body.BgpPeers = bgpPeerConfigs

	//
	// Groups
	//
	groups := make([]oxide.NameOrId, 0)
	for _, group := range model.Groups {
		groups = append(groups, oxide.NameOrId(group.ValueString()))
	}
	params.Body.Groups = groups

	//
	// Interfaces
	//
	interfaceConfigs := make([]oxide.SwitchInterfaceConfigCreate, 0)
	for _, interfaceModel := range model.Interfaces {
		interfaceConfig := oxide.SwitchInterfaceConfigCreate{
			Kind: oxide.SwitchInterfaceKind{
				Type: oxide.SwitchInterfaceKindType(interfaceModel.Kind.Type.ValueString()),
				Vid: func() *int {
					if interfaceModel.Kind.VID.IsNull() {
						return nil
					}
					return oxide.NewPointer(int(interfaceModel.Kind.VID.ValueInt32()))
				}(),
			},
			LinkName:  oxide.Name(interfaceModel.LinkName.ValueString()),
			V6Enabled: oxide.NewPointer(interfaceModel.V6Enabled.ValueBool()),
		}

		interfaceConfigs = append(interfaceConfigs, interfaceConfig)
	}
	params.Body.Interfaces = interfaceConfigs

	//
	// Links
	//
	linkConfigs := make([]oxide.LinkConfigCreate, 0)
	for _, link := range model.Links {
		linkConfig := oxide.LinkConfigCreate{
			Autoneg:  link.Autoneg.ValueBoolPointer(),
			Fec:      oxide.LinkFec(link.FEC.ValueString()),
			LinkName: oxide.Name(link.LinkName.ValueString()),
			Mtu:      oxide.NewPointer(int(link.MTU.ValueInt32())),
			Speed:    oxide.LinkSpeed(link.Speed.ValueString()),
			Lldp: oxide.LldpLinkConfigCreate{
				ChassisId:         link.LLDP.ChassisID.ValueString(),
				Enabled:           link.LLDP.Enabled.ValueBoolPointer(),
				LinkDescription:   link.LLDP.LinkDescription.ValueString(),
				LinkName:          link.LLDP.LinkName.ValueString(),
				ManagementIp:      link.LLDP.ManagementIP.ValueString(),
				SystemDescription: link.LLDP.SystemDescription.ValueString(),
				SystemName:        link.LLDP.SystemName.ValueString(),
			},
		}

		if link.TxEq != nil {
			linkConfig.TxEq = &oxide.TxEqConfig{
				Main: func() *int {
					if link.TxEq.Main.IsNull() {
						return nil
					}
					return oxide.NewPointer(int(link.TxEq.Main.ValueInt32()))
				}(),
				Post1: func() *int {
					if link.TxEq.Post1.IsNull() {
						return nil
					}
					return oxide.NewPointer(int(link.TxEq.Post1.ValueInt32()))
				}(),
				Post2: func() *int {
					if link.TxEq.Post2.IsNull() {
						return nil
					}
					return oxide.NewPointer(int(link.TxEq.Post2.ValueInt32()))
				}(),
				Pre1: func() *int {
					if link.TxEq.Pre1.IsNull() {
						return nil
					}
					return oxide.NewPointer(int(link.TxEq.Pre1.ValueInt32()))
				}(),
				Pre2: func() *int {
					if link.TxEq.Pre2.IsNull() {
						return nil
					}
					return oxide.NewPointer(int(link.TxEq.Pre2.ValueInt32()))
				}(),
			}
		}

		linkConfigs = append(linkConfigs, linkConfig)
	}
	params.Body.Links = linkConfigs

	//
	// Routes
	//
	routeConfigs := make([]oxide.RouteConfig, 0)
	for _, routeModel := range model.Routes {
		routes := make([]oxide.Route, 0)
		for _, routeModel := range routeModel.Routes {
			route := oxide.Route{
				Dst: oxide.IpNet(routeModel.Dst.ValueString()),
				Gw:  routeModel.GW.ValueString(),
				RibPriority: func() *int {
					if routeModel.RIBPriority.IsNull() {
						return nil
					}
					return oxide.NewPointer(int(routeModel.RIBPriority.ValueInt32()))
				}(),
				Vid: func() *int {
					if routeModel.VID.IsNull() {
						return nil
					}
					return oxide.NewPointer(int(routeModel.VID.ValueInt32()))
				}(),
			}

			routes = append(routes, route)
		}

		routeConfig := oxide.RouteConfig{
			LinkName: oxide.Name(routeModel.LinkName.ValueString()),
			Routes:   routes,
		}

		routeConfigs = append(routeConfigs, routeConfig)
	}
	params.Body.Routes = routeConfigs

	return params, nil
}

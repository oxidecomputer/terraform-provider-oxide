// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package switchportsettings

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	"github.com/hashicorp/terraform-plugin-framework-nettypes/iptypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/oxidecomputer/oxide.go/oxide"
)

// Schemas and model definitions for upgrading the oxide_switch_port_settings
// resource from version 0.
//
// Version 1 of the schema replaced the BGP peer `address` (string) and
// `interface_name` (string) attributes with a single `addr` object that models
// the [oxide.RouterPeerType] tagged union. The schema and structs below are
// copied from their version 0 sources so prior state can be decoded and
// migrated to the current schema.

// stateUpgraderV1 is a StateUpgrader function that upgrades an
// oxide_switch_port_settings resource from schema version 0 to the latest
// schema version.
func (r *Resource) stateUpgraderV1(
	ctx context.Context,
	req resource.UpgradeStateRequest,
	resp *resource.UpgradeStateResponse,
) {
	schemaV0 := r.schemaV0(ctx)

	rawStateValue, err := req.RawState.Unmarshal(
		schemaV0.Type().TerraformType(ctx),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Upgrade State From Version 0",
			fmt.Sprintf("failed to convert state to schema version 0: %v", err),
		)
		return
	}

	stateV0 := &tfsdk.State{
		Raw:    rawStateValue,
		Schema: schemaV0,
	}

	var modelV0 modelV0
	resp.Diagnostics.Append(stateV0.Get(ctx, &modelV0)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, modelV0.upgrade())...)
}

func (r *Resource) schemaV0(ctx context.Context) schema.Schema {
	return schema.Schema{
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
										CustomType:  cidrtypes.IPPrefixType{},
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
										Optional:    true,
										CustomType:  iptypes.IPAddressType{},
										Description: "Address of the host to peer with. If not provided, this is an unnumbered BGP session that will be established over the interface specified by `interface_name`.",
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
														string(
															oxide.ImportExportPolicyTypeNoFiltering,
														),
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
														string(
															oxide.ImportExportPolicyTypeNoFiltering,
														),
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
									CustomType:  iptypes.IPAddressType{},
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
										CustomType:  iptypes.IPAddressType{},
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

type modelV0 struct {
	ID           types.String             `tfsdk:"id"`
	Name         types.String             `tfsdk:"name"`
	Description  types.String             `tfsdk:"description"`
	Addresses    []AddressResourceModel   `tfsdk:"addresses"`
	BGPPeers     []bgpPeerModelV0         `tfsdk:"bgp_peers"`
	Links        []LinkResourceModel      `tfsdk:"links"`
	PortConfig   *PortConfigResourceModel `tfsdk:"port_config"`
	Routes       []RouteResourceModel     `tfsdk:"routes"`
	TimeCreated  types.String             `tfsdk:"time_created"`
	TimeModified types.String             `tfsdk:"time_modified"`
	Timeouts     timeouts.Value           `tfsdk:"timeouts"`
}

func (m modelV0) upgrade() ResourceModel {
	upgraded := ResourceModel{
		ID:           m.ID,
		Name:         m.Name,
		Description:  m.Description,
		Addresses:    m.Addresses,
		Links:        m.Links,
		PortConfig:   m.PortConfig,
		Routes:       m.Routes,
		TimeCreated:  m.TimeCreated,
		TimeModified: m.TimeModified,
		Timeouts:     m.Timeouts,
	}

	if m.BGPPeers != nil {
		upgraded.BGPPeers = make([]BGPPeerResourceModel, len(m.BGPPeers))
		for i, group := range m.BGPPeers {
			upgraded.BGPPeers[i] = group.upgrade()
		}
	}

	return upgraded
}

type bgpPeerModelV0 struct {
	LinkName types.String         `tfsdk:"link_name"`
	Peers    []bgpPeerPeerModelV0 `tfsdk:"peers"`
}

func (m bgpPeerModelV0) upgrade() BGPPeerResourceModel {
	peers := make([]BGPPeerPeerResourceModel, len(m.Peers))
	for i, p := range m.Peers {
		peers[i] = p.upgrade()
	}

	return BGPPeerResourceModel{
		LinkName: m.LinkName,
		Peers:    peers,
	}
}

type bgpPeerPeerModelV0 struct {
	Address                iptypes.IPAddress                      `tfsdk:"address"`
	AllowedExport          *BGPPeerPeerAllowedExportResourceModel `tfsdk:"allowed_export"`
	AllowedImport          *BGPPeerPeerAllowedImportResourceModel `tfsdk:"allowed_import"`
	BGPConfig              types.String                           `tfsdk:"bgp_config"`
	Communities            []types.Int64                          `tfsdk:"communities"`
	ConnectRetry           types.Int64                            `tfsdk:"connect_retry"`
	DelayOpen              types.Int64                            `tfsdk:"delay_open"`
	EnforceFirstAs         types.Bool                             `tfsdk:"enforce_first_as"`
	HoldTime               types.Int64                            `tfsdk:"hold_time"`
	IdleHoldTime           types.Int64                            `tfsdk:"idle_hold_time"`
	InterfaceName          types.String                           `tfsdk:"interface_name"`
	Keepalive              types.Int64                            `tfsdk:"keepalive"`
	LocalPref              types.Int64                            `tfsdk:"local_pref"`
	MD5AuthKey             types.String                           `tfsdk:"md5_auth_key"`
	MinTTL                 types.Int32                            `tfsdk:"min_ttl"`
	MultiExitDiscriminator types.Int64                            `tfsdk:"multi_exit_discriminator"`
	RemoteASN              types.Int64                            `tfsdk:"remote_asn"`
	VlanID                 types.Int32                            `tfsdk:"vlan_id"`
}

func (m bgpPeerPeerModelV0) upgrade() BGPPeerPeerResourceModel {
	return BGPPeerPeerResourceModel{
		Addr:                   m.addr(),
		AllowedExport:          m.AllowedExport,
		AllowedImport:          m.AllowedImport,
		BGPConfig:              m.BGPConfig,
		Communities:            m.Communities,
		ConnectRetry:           m.ConnectRetry,
		DelayOpen:              m.DelayOpen,
		EnforceFirstAs:         m.EnforceFirstAs,
		HoldTime:               m.HoldTime,
		IdleHoldTime:           m.IdleHoldTime,
		Keepalive:              m.Keepalive,
		LocalPref:              m.LocalPref,
		MD5AuthKey:             m.MD5AuthKey,
		MinTTL:                 m.MinTTL,
		MultiExitDiscriminator: m.MultiExitDiscriminator,
		RemoteASN:              m.RemoteASN,
		VlanID:                 m.VlanID,
	}
}

// addr maps a schema version 0 BGP peer `address` into the version 1 `addr`
// object. A non-empty address becomes a `numbered` peer; an unset address
// becomes an `unnumbered` peer. The previous schema had no equivalent of
// `router_lifetime`, so it is left null for unnumbered peers and may produce a
// diff against configuration on the next plan.
func (m bgpPeerPeerModelV0) addr() *BGPPeerAddrResourceModel {
	if !m.Address.IsNull() && !m.Address.IsUnknown() &&
		m.Address.ValueString() != "" {
		return &BGPPeerAddrResourceModel{
			Type:           types.StringValue(string(oxide.RouterPeerTypeTypeNumbered)),
			IP:             m.Address,
			RouterLifetime: types.Int64Null(),
		}
	}

	return &BGPPeerAddrResourceModel{
		Type:           types.StringValue(string(oxide.RouterPeerTypeTypeUnnumbered)),
		IP:             iptypes.NewIPAddressNull(),
		RouterLifetime: types.Int64Null(),
	}
}

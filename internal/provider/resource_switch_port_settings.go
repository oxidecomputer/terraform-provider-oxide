package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	_ "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	_ "github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	_ "github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/oxidecomputer/oxide.go/oxide"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = (*switchPortConfigurationResource)(nil)
	_ resource.ResourceWithConfigure = (*switchPortConfigurationResource)(nil)
)

type switchPortConfigurationResource struct {
	client *oxide.Client
}

type addressLotModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type switchPortAddressModel struct {
	Address    types.String    `tfsdk:"address"`
	AddressLot addressLotModel `tfsdk:"address_lot"`
	Vlan       types.Int32     `tfsdk:"vlan"`
}

type switchPortAddressConfigModel struct {
	LinkName  types.String             `tfsdk:"link_name"`
	Addresses []switchPortAddressModel `tfsdk:"addresses"`
}

type routeModel struct {
	Destination types.String `tfsdk:"destination"`
	Gateway     types.String `tfsdk:"gateway"`
	RibPriority types.Int32  `tfsdk:"rib_priority"`
	VlanID      types.Int32  `tfsdk:"vlan_id"`
}

type routeConfigModel struct {
	LinkName types.String `tfsdk:"link_name"`
	Routes   []routeModel `tfsdk:"routes"`
}

// type lldpLinkConfigCreateModel struct {
// 	ChassisId         types.String `tfsdk:"chassis_id"`
// 	Enabled           types.Bool   `tfsdk:"enabled"`
// 	LinkDescription   types.String `tfsdk:"link_description"`
// 	LinkName          types.String `tfsdk:"link_name"`
// 	ManagementIp      types.String `tfsdk:"management_ip"`
// 	SystemDescription types.String `tfsdk:"system_description"`
// 	SystemName        types.String `tfsdk:"system_name"`
// }

// type txEqConfigModel struct {
// 	Main  types.Int32 `tfsdk:"main"`
// 	Post1 types.Int32 `tfsdk:"post1"`
// 	Post2 types.Int32 `tfsdk:"post2"`
// 	Pre1  types.Int32 `tfsdk:"pre1"`
// 	Pre2  types.Int32 `tfsdk:"pre2"`
// }

type linkConfigModel struct {
	Name    types.String `tfsdk:"name"`
	Autoneg bool         `tfsdk:"autoneg"`
	Fec     types.String `tfsdk:"fec"`
	// TODO: Resolve lldp information
	// Lldp    lldpLinkConfigCreateModel `tfsdk:"lldp"`
	Mtu   types.Int32  `tfsdk:"mtu"`
	Speed types.String `tfsdk:"speed"`
	// TODO: Resolve tx_eq information
	// TxEq  txEqConfigModel `tfsdk:"tx_eq"`
}

type importExportPolicyModel struct {
	Type  types.String   `tfsdk:"policy_type"`
	Value []types.String `tfsdk:"value"`
}

type bgpPeerConfigModel struct {
	LinkName types.String   `tfsdk:"link_name"`
	Peers    []bgpPeerModel `tfsdk:"peers"`
}

type bgpPeerModel struct {
	Addr                   types.String            `tfsdk:"addr"`
	AllowedExport          importExportPolicyModel `tfsdk:"allowed_export"`
	AllowedImport          importExportPolicyModel `tfsdk:"allowed_import"`
	BgpConfig              types.String            `tfsdk:"bgp_config"`
	Communities            []types.String          `tfsdk:"communities"`
	ConnectRetry           types.Int32             `tfsdk:"connect_retry"`
	DelayOpen              types.Int32             `tfsdk:"delay_open"`
	EnforceFirstAs         types.Bool              `tfsdk:"enforce_first_as"`
	HoldTime               types.Int32             `tfsdk:"hold_time"`
	IdleHoldTime           types.Int32             `tfsdk:"idle_hold_time"`
	InterfaceName          types.String            `tfsdk:"interface_name"`
	Keepalive              types.Int32             `tfsdk:"keepalive"`
	LocalPref              types.Int32             `tfsdk:"local_pref"`
	Md5AuthKey             types.String            `tfsdk:"md5_auth_key"`
	MinTtl                 types.Int32             `tfsdk:"min_ttl"`
	MultiExitDiscriminator types.Int32             `tfsdk:"multi_exit_discriminator"`
	RemoteAsn              types.Int32             `tfsdk:"remote_asn"`
	VlanId                 types.Int32             `tfsdk:"vlan_id"`
}

type switchPortSettingsModel struct {
	ID           types.String                   `tfsdk:"id"`
	Name         types.String                   `tfsdk:"name"`
	Description  types.String                   `tfsdk:"description"`
	Addresses    []switchPortAddressConfigModel `tfsdk:"addresses"`
	BgpPeers     []bgpPeerConfigModel           `tfsdk:"bgp_peers"`
	Links        []linkConfigModel              `tfsdk:"links"`
	PortConfig   types.String                   `tfsdk:"port_config"`
	Routes       []routeConfigModel             `tfsdk:"routes"`
	TimeCreated  types.String                   `tfsdk:"time_created"`
	TimeModified types.String                   `tfsdk:"time_modified"`
	Timeouts     timeouts.Value                 `tfsdk:"timeouts"`
}

// NewSwitchPortConfigurationResource is a helper function to simplify the provider implementation.
func NewSwitchPortConfigurationResource() resource.Resource {
	return &switchPortConfigurationResource{}
}

// Metadata returns the resource type name.
func (r *switchPortConfigurationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "oxide_switch_port_configuration"
}

// Configure adds the provider configured client to the data source.
func (r *switchPortConfigurationResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*oxide.Client)
}

func (r *switchPortConfigurationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Schema defines the schema for the resource.
func (r *switchPortConfigurationResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier of the Switch Port Configuration.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the Switch Port Configuration.",
			},
			"description": schema.StringAttribute{
				Required:    true,
				Description: "Description for the Switch Port Configuration.",
			},

			"addresses": schema.SetNestedAttribute{
				Optional:    true,
				Description: "List of addresses for the Switch Port Configuration.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"link_name": schema.StringAttribute{
							Required:    true,
							Description: "Name of the link for the Switch Port Configuration.",
						},
						"addresses": schema.SetNestedAttribute{
							Required:    true,
							Description: "List of addresses for the Switch Port Configuration.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"address": schema.StringAttribute{
										Required:    true,
										Description: "Address for the Switch Port Configuration.",
									},
									"address_lot": schema.SingleNestedAttribute{
										Required: true,
										Attributes: map[string]schema.Attribute{
											"id": schema.StringAttribute{
												Computed:    true,
												Description: "ID of the address lot.",
											},
											"name": schema.StringAttribute{
												Required:    true,
												Description: "Name of the address lot.",
											},
										},
									},
									"vlan": schema.Int32Attribute{
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
				Description: "List of BGP peers for the Switch Port Configuration.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"link_name": schema.StringAttribute{
							Required:    true,
							Description: "Name of the link for the BGP peer configuration.",
						},
						"peers": schema.SetNestedAttribute{
							Required:    true,
							Description: "List of BGP peers for the link.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"addr": schema.StringAttribute{
										Required:    true,
										Description: "Address of the BGP peer.",
									},
									"allowed_export": schema.SetNestedAttribute{
										Required:    true,
										Description: "Export policy for the BGP peer.",
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"policy_type": schema.StringAttribute{
													Required:    true,
													Description: "Type of the export policy.",
												},
												"value": schema.ListAttribute{
													Required:    true,
													Description: "Values for the export policy.",
													ElementType: types.StringType,
												},
											},
										},
									},
									"allowed_import": schema.SetNestedAttribute{
										Required:    true,
										Description: "Import policy for the BGP peer.",
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"policy_type": schema.StringAttribute{
													Required:    true,
													Description: "Type of the import policy.",
												},
												"value": schema.ListAttribute{
													Required:    true,
													Description: "Values for the import policy.",
													ElementType: types.StringType,
												},
											},
										},
									},
									"bgp_config": schema.StringAttribute{
										Optional:    true,
										Description: "BGP configuration for the peer.",
									},
									"communities": schema.ListAttribute{
										Optional:    true,
										Description: "List of communities for the BGP peer.",
										ElementType: types.StringType,
									},
									"connect_retry": schema.Int32Attribute{
										Optional:    true,
										Description: "Connect retry interval for the BGP peer.",
									},
									"delay_open": schema.Int32Attribute{
										Optional:    true,
										Description: "Delay open interval for the BGP peer.",
									},
									"enforce_first_as": schema.BoolAttribute{
										Optional:    true,
										Description: "Whether to enforce the first AS for the BGP peer.",
									},
									"hold_time": schema.Int32Attribute{
										Optional:    true,
										Description: "Hold time for the BGP peer.",
									},
									"idle_hold_time": schema.Int32Attribute{
										Optional:    true,
										Description: "Idle hold time for the BGP peer.",
									},
									"interface_name": schema.StringAttribute{
										Required:    true,
										Description: "Interface name for the BGP peer.",
									},
									"keepalive": schema.Int32Attribute{
										Optional:    true,
										Description: "Keepalive interval for the BGP peer.",
									},
									"local_pref": schema.Int32Attribute{
										Optional:    true,
										Description: "Local preference for the BGP peer.",
									},
									"md5_auth_key": schema.StringAttribute{
										Optional:    true,
										Description: "MD5 authentication key for the BGP peer.",
									},
									"min_ttl": schema.Int32Attribute{
										Optional:    true,
										Description: "Minimum TTL for the BGP peer.",
									},
									"multi_exit_discriminator": schema.Int32Attribute{
										Optional:    true,
										Description: "Multi-exit discriminator for the BGP peer.",
									},
									"remote_asn": schema.Int32Attribute{
										Required:    true,
										Description: "Remote ASN for the BGP peer.",
									},
									"vlan_id": schema.Int32Attribute{
										Optional:    true,
										Description: "VLAN ID for the BGP peer.",
									},
								},
							},
						},
					},
				},
			},

			"links": schema.SetNestedAttribute{
				Optional:    true,
				Description: "List of links for the Switch Port Configuration.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "Name of the link.",
						},
						"autoneg": schema.BoolAttribute{
							Required:    true,
							Description: "Whether autonegotiation is enabled for the link.",
						},
						"fec": schema.StringAttribute{
							Required:    true,
							Description: "Forward error correction (FEC) mode for the link.",
						},
						"mtu": schema.Int32Attribute{
							Required:    true,
							Description: "Maximum transmission unit (MTU) for the link.",
						},
						"speed": schema.StringAttribute{
							Required:    true,
							Description: "Speed of the link.",
						},
					},
				},
			},

			"routes": schema.SetNestedAttribute{
				Optional:    true,
				Description: "List of routes for the Switch Port Configuration.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"link_name": schema.StringAttribute{
							Required:    true,
							Description: "Name of the link for the route configuration.",
						},
						"routes": schema.SetNestedAttribute{
							Required:    true,
							Description: "List of routes for the link.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"destination": schema.StringAttribute{
										Required:    true,
										Description: "Destination address for the route.",
									},
									"gateway": schema.StringAttribute{
										Required:    true,
										Description: "Gateway address for the route.",
									},
									"rib_priority": schema.Int32Attribute{
										Optional:    true,
										Description: "RIB priority for the route.",
									},
									"vlan_id": schema.Int32Attribute{
										Optional:    true,
										Description: "VLAN ID for the route.",
									},
								},
							},
						},
					},
				},
			},

			"port_config": schema.StringAttribute{
				Required:    true,
				Description: "Port configuration for the Switch Port Configuration.",
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

// Create creates the resource and sets the initial Terraform state.
func (r *switchPortConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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

	params := buildParams(&plan)

	response, err := r.client.NetworkingSwitchPortSettingsCreate(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating switch port settings",
			"API error: "+err.Error(),
		)
		return
	}

	settings := response.Settings

	tflog.Trace(ctx, fmt.Sprintf("created switch port settings with ID: %v", settings.Id), map[string]any{"success": true})

	// Map response body to schema and populate computed attribute values.
	plan.ID = types.StringValue(settings.Id)
	plan.Description = types.StringValue(settings.Description)
	plan.TimeCreated = types.StringValue(settings.TimeCreated.String())
	plan.TimeModified = types.StringValue(settings.TimeModified.String())

	// Save plan into Terraform state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *switchPortConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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

	response, err := r.client.NetworkingSwitchPortSettingsView(ctx, oxide.NetworkingSwitchPortSettingsViewParams{
		Port: oxide.NameOrId(state.ID.ValueString()),
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read Switch Port Settings:",
			"API error: "+err.Error(),
		)
		return
	}

	switchPortSettings := response.Settings

	tflog.Trace(ctx, fmt.Sprintf("read Switch Port Settings with ID: %v", switchPortSettings.Id), map[string]any{"success": true})

	// Map response body to schema
	state.Name = types.StringValue(string(switchPortSettings.Name))
	state.ID = types.StringValue(switchPortSettings.Id)
	state.Description = types.StringValue(switchPortSettings.Description)

	state.Addresses = []switchPortAddressConfigModel{}
	state.BgpPeers = []bgpPeerConfigModel{}
	state.Links = []linkConfigModel{}
	state.PortConfig = types.StringValue(string(response.Port.Geometry))
	state.Routes = []routeConfigModel{}

	state.TimeCreated = types.StringValue(switchPortSettings.TimeCreated.String())
	state.TimeModified = types.StringValue(switchPortSettings.TimeModified.String())

	addressMappings := make(map[string]switchPortAddressConfigModel)

	for _, item := range response.Addresses {
		// fetch the address config from the map
		val, ok := addressMappings[item.InterfaceName]

		// If the address is not already in the map, create a new entry
		if !ok {
			val = switchPortAddressConfigModel{}
			val.LinkName = types.StringValue(item.InterfaceName)
		}

		// Add the address to the existing entry
		newAddress := switchPortAddressModel{
			Address: types.StringValue(fmt.Sprintf("%v", item.Address)),
		}

		if item.VlanId != nil {
			newAddress.Vlan = types.Int32Value(int32(*item.VlanId))
		} else {
			newAddress.Vlan = types.Int32Null()
		}

		newAddress.AddressLot.ID = types.StringValue(item.AddressLotId)
		newAddress.AddressLot.Name = types.StringValue(string(item.AddressLotName))

		val.Addresses = append(addressMappings[item.InterfaceName].Addresses, newAddress)

		// update the value stored in the map
		addressMappings[item.InterfaceName] = val
	}

	for _, value := range addressMappings {
		state.Addresses = append(state.Addresses, value)
	}

	bgpPeerMappings := make(map[string]bgpPeerConfigModel)

	for _, item := range response.BgpPeers {
		val, ok := bgpPeerMappings[item.InterfaceName]

		if !ok {
			val = bgpPeerConfigModel{}
			val.LinkName = types.StringValue(item.InterfaceName)
		}

		newPeer := bgpPeerModel{
			Addr: types.StringValue(item.Addr),
			AllowedExport: importExportPolicyModel{
				Type:  types.StringValue(string(item.AllowedExport.Type)),
				Value: make([]types.String, len(item.AllowedExport.Value)),
			},
			AllowedImport: importExportPolicyModel{
				Type:  types.StringValue(string(item.AllowedImport.Type)),
				Value: make([]types.String, len(item.AllowedImport.Value)),
			},
			BgpConfig:              types.StringValue(string(item.BgpConfig)),
			Communities:            make([]types.String, len(item.Communities)),
			ConnectRetry:           types.Int32Null(),
			DelayOpen:              types.Int32Null(),
			EnforceFirstAs:         types.BoolNull(),
			HoldTime:               types.Int32Null(),
			IdleHoldTime:           types.Int32Null(),
			InterfaceName:          types.StringValue(item.InterfaceName),
			Keepalive:              types.Int32Null(),
			LocalPref:              types.Int32Null(),
			Md5AuthKey:             types.StringValue(item.Md5AuthKey),
			MinTtl:                 types.Int32Null(),
			MultiExitDiscriminator: types.Int32Null(),
			RemoteAsn:              types.Int32Null(),
			VlanId:                 types.Int32Null(),
		}

		if item.ConnectRetry != nil {
			newPeer.ConnectRetry = types.Int32Value(int32(*item.ConnectRetry))
		}
		if item.DelayOpen != nil {
			newPeer.DelayOpen = types.Int32Value(int32(*item.DelayOpen))
		}
		if item.EnforceFirstAs != nil {
			newPeer.EnforceFirstAs = types.BoolValue(*item.EnforceFirstAs)
		}
		if item.HoldTime != nil {
			newPeer.HoldTime = types.Int32Value(int32(*item.HoldTime))
		}
		if item.IdleHoldTime != nil {
			newPeer.IdleHoldTime = types.Int32Value(int32(*item.IdleHoldTime))
		}
		if item.Keepalive != nil {
			newPeer.Keepalive = types.Int32Value(int32(*item.Keepalive))
		}
		if item.LocalPref != nil {
			newPeer.LocalPref = types.Int32Value(int32(*item.LocalPref))
		}
		if item.MinTtl != nil {
			newPeer.MinTtl = types.Int32Value(int32(*item.MinTtl))
		}
		if item.MultiExitDiscriminator != nil {
			newPeer.MultiExitDiscriminator = types.Int32Value(int32(*item.MultiExitDiscriminator))
		}
		if item.RemoteAsn != nil {
			newPeer.RemoteAsn = types.Int32Value(int32(*item.RemoteAsn))
		}
		if item.VlanId != nil {
			newPeer.VlanId = types.Int32Value(int32(*item.VlanId))
		}

		for i, value := range item.AllowedExport.Value {
			newPeer.AllowedExport.Value[i] = types.StringValue(fmt.Sprintf("%v", value))
		}

		for i, value := range item.AllowedImport.Value {
			newPeer.AllowedImport.Value[i] = types.StringValue(fmt.Sprintf("%v", value))
		}

		for i, community := range item.Communities {
			newPeer.Communities[i] = types.StringValue(community)
		}

		val.Peers = append(val.Peers, newPeer)
		bgpPeerMappings[item.InterfaceName] = val
	}

	for _, value := range bgpPeerMappings {
		state.BgpPeers = append(state.BgpPeers, value)
	}

	linkMappings := make(map[string]linkConfigModel)

	for _, item := range response.Links {
		val, ok := linkMappings[item.LinkName]

		if !ok {
			val = linkConfigModel{}
			val.Name = types.StringValue(item.LinkName)
			if item.Autoneg != nil {
				val.Autoneg = *item.Autoneg
			} else {
				val.Autoneg = true // or a default value if applicable
			}
			val.Fec = types.StringValue(string(item.Fec))
			if item.Mtu != nil {
				val.Mtu = types.Int32Value(int32(*item.Mtu))
			} else {
				val.Mtu = types.Int32Null()
			}
			val.Speed = types.StringValue(string(item.Speed))

			// TODO: Resolve lldp information
			// Currently the API is disjointed. The SwitchPortSettings api returns an LLDP config id,
			// but the LLDP apis do not allow you to look them up by id.

			// TODO: Resolve tx_eq information
			// The go client smushes the tx_eq information into a slice of strings, but the API returns
			// a collection of structs. This will need to be resturctured to be more amicable to various
			// clients
		}

		linkMappings[item.LinkName] = val
	}

	for _, value := range linkMappings {
		state.Links = append(state.Links, value)
	}

	routeMappings := make(map[string]routeConfigModel)

	for _, item := range response.Routes {
		val, ok := routeMappings[item.InterfaceName]

		if !ok {
			val = routeConfigModel{}
			val.LinkName = types.StringValue(item.InterfaceName)
		}

		newRoute := routeModel{
			Destination: types.StringValue(fmt.Sprintf("%v", item.Dst)),
			Gateway:     types.StringValue(fmt.Sprintf("%v", item.Gw)),
		}

		if item.RibPriority != nil {
			newRoute.RibPriority = types.Int32Value(int32(*item.RibPriority))
		} else {
			newRoute.RibPriority = types.Int32Null()
		}

		if item.VlanId != nil {
			newRoute.VlanID = types.Int32Value(int32(*item.VlanId))
		} else {
			newRoute.VlanID = types.Int32Null()
		}

		val.Routes = append(val.Routes, newRoute)
		routeMappings[item.InterfaceName] = val
	}

	for _, value := range routeMappings {
		state.Routes = append(state.Routes, value)
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *switchPortConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// plan is the resource data model for the update request.
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

	params := buildParams(&plan)

	// TODO: currently the switch port settings API performs update using the same endpoint
	// as create.

	response, err := r.client.NetworkingSwitchPortSettingsCreate(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating switch port settings",
			"API error: "+err.Error(),
		)
		return
	}

	settings := response.Settings

	tflog.Trace(ctx, fmt.Sprintf("updated switch port settings with ID: %v", settings.Id), map[string]any{"success": true})

	// Map response body to schema and populate computed attribute values.
	plan.ID = types.StringValue(settings.Id)
	plan.Description = types.StringValue(settings.Description)
	plan.TimeCreated = types.StringValue(settings.TimeCreated.String())
	plan.TimeModified = types.StringValue(settings.TimeModified.String())

	// Save plan into Terraform state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *switchPortConfigurationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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

func buildParams(plan *switchPortSettingsModel) oxide.NetworkingSwitchPortSettingsCreateParams {
	addressConfigs := []oxide.AddressConfig{}

	for _, config := range plan.Addresses {
		newConfig := oxide.AddressConfig{
			LinkName:  oxide.Name(config.LinkName.ValueString()),
			Addresses: []oxide.Address{},
		}

		for _, addr := range config.Addresses {
			address := oxide.Address{
				Address:    addr.Address.ValueString(),
				AddressLot: oxide.NameOrId(addr.AddressLot.Name.ValueString()),
				VlanId:     new(int),
			}
			if !addr.Vlan.IsNull() {
				*address.VlanId = int(addr.Vlan.ValueInt32())
			}
			newConfig.Addresses = append(newConfig.Addresses, address)
		}

		addressConfigs = append(addressConfigs, newConfig)
	}

	routes := []oxide.RouteConfig{}

	for _, config := range plan.Routes {
		newConfig := oxide.RouteConfig{
			LinkName: oxide.Name(config.LinkName.ValueString()),
			Routes:   []oxide.Route{},
		}

		for _, route := range config.Routes {
			newRoute := oxide.Route{
				Dst:         route.Destination.ValueString(),
				Gw:          route.Gateway.ValueString(),
				RibPriority: new(int),
				Vid:         new(int),
			}
			if !route.RibPriority.IsNull() {
				*newRoute.RibPriority = int(route.RibPriority.ValueInt32())
			}
			if !route.VlanID.IsNull() {
				*newRoute.Vid = int(route.VlanID.ValueInt32())
			}

			newConfig.Routes = append(newConfig.Routes, newRoute)
		}

		routes = append(routes, newConfig)
	}

	bgpPeerConfigs := []oxide.BgpPeerConfig{}

	for _, config := range plan.BgpPeers {
		newConfig := oxide.BgpPeerConfig{
			LinkName: oxide.Name(config.LinkName.ValueString()),
			Peers:    []oxide.BgpPeer{},
		}

		for _, peer := range config.Peers {
			newPeer := oxide.BgpPeer{
				Addr: peer.Addr.ValueString(),
				AllowedExport: oxide.ImportExportPolicy{
					Type:  oxide.ImportExportPolicyType(peer.AllowedExport.Type.ValueString()),
					Value: make([]oxide.IpNet, len(peer.AllowedExport.Value)),
				},
				AllowedImport: oxide.ImportExportPolicy{
					Type:  oxide.ImportExportPolicyType(peer.AllowedImport.Type.ValueString()),
					Value: make([]oxide.IpNet, len(peer.AllowedImport.Value)),
				},
				BgpConfig:              oxide.NameOrId(peer.BgpConfig.ValueString()),
				Communities:            make([]string, len(peer.Communities)),
				ConnectRetry:           new(int),
				DelayOpen:              new(int),
				EnforceFirstAs:         new(bool),
				HoldTime:               new(int),
				IdleHoldTime:           new(int),
				InterfaceName:          peer.InterfaceName.ValueString(),
				Keepalive:              new(int),
				LocalPref:              new(int),
				Md5AuthKey:             peer.Md5AuthKey.ValueString(),
				MinTtl:                 new(int),
				MultiExitDiscriminator: new(int),
				RemoteAsn:              new(int),
				VlanId:                 new(int),
			}

			for i, value := range peer.AllowedExport.Value {
				newPeer.AllowedExport.Value[i] = value.ValueString()
			}

			for i, value := range peer.AllowedImport.Value {
				newPeer.AllowedImport.Value[i] = value.ValueString()
			}

			for i, community := range peer.Communities {
				newPeer.Communities[i] = community.ValueString()
			}

			if !peer.ConnectRetry.IsNull() {
				*newPeer.ConnectRetry = int(peer.ConnectRetry.ValueInt32())
			}
			if !peer.DelayOpen.IsNull() {
				*newPeer.DelayOpen = int(peer.DelayOpen.ValueInt32())
			}
			if !peer.EnforceFirstAs.IsNull() {
				*newPeer.EnforceFirstAs = peer.EnforceFirstAs.ValueBool()
			}
			if !peer.HoldTime.IsNull() {
				*newPeer.HoldTime = int(peer.HoldTime.ValueInt32())
			}
			if !peer.IdleHoldTime.IsNull() {
				*newPeer.IdleHoldTime = int(peer.IdleHoldTime.ValueInt32())
			}
			if !peer.Keepalive.IsNull() {
				*newPeer.Keepalive = int(peer.Keepalive.ValueInt32())
			}
			if !peer.LocalPref.IsNull() {
				*newPeer.LocalPref = int(peer.LocalPref.ValueInt32())
			}
			if !peer.MinTtl.IsNull() {
				*newPeer.MinTtl = int(peer.MinTtl.ValueInt32())
			}
			if !peer.MultiExitDiscriminator.IsNull() {
				*newPeer.MultiExitDiscriminator = int(peer.MultiExitDiscriminator.ValueInt32())
			}
			if !peer.RemoteAsn.IsNull() {
				*newPeer.RemoteAsn = int(peer.RemoteAsn.ValueInt32())
			}
			if !peer.VlanId.IsNull() {
				*newPeer.VlanId = int(peer.VlanId.ValueInt32())
			}

			newConfig.Peers = append(newConfig.Peers, newPeer)
		}

		bgpPeerConfigs = append(bgpPeerConfigs, newConfig)
	}

	linkConfigs := []oxide.LinkConfigCreate{}

	for _, link := range plan.Links {
		newLink := oxide.LinkConfigCreate{
			LinkName: oxide.Name(link.Name.ValueString()),
			Autoneg:  &link.Autoneg,
			Fec:      oxide.LinkFec(link.Fec.ValueString()),
			Mtu:      new(int),
			Speed:    oxide.LinkSpeed(link.Speed.ValueString()),
		}

		if !link.Mtu.IsNull() {
			*newLink.Mtu = int(link.Mtu.ValueInt32())
		}

		linkConfigs = append(linkConfigs, newLink)
	}

	return oxide.NetworkingSwitchPortSettingsCreateParams{
		Body: &oxide.SwitchPortSettingsCreate{
			Addresses:   addressConfigs,
			BgpPeers:    bgpPeerConfigs,
			Description: plan.Description.String(),
			Links:       linkConfigs,
			Name:        oxide.Name(plan.Name.String()),
			PortConfig: oxide.SwitchPortConfigCreate{
				Geometry: oxide.SwitchPortGeometry(plan.PortConfig.String()),
			},
			Routes: routes,
		},
	}
}

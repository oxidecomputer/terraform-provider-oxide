// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"slices"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/oxidecomputer/oxide.go/oxide"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                 = (*instanceResource)(nil)
	_ resource.ResourceWithConfigure    = (*instanceResource)(nil)
	_ resource.ResourceWithUpgradeState = (*instanceResource)(nil)
)

// NewInstanceResource is a helper function to simplify the provider implementation.
func NewInstanceResource() resource.Resource {
	return &instanceResource{}
}

// instanceResource is the resource implementation.
type instanceResource struct {
	client *oxide.Client
}

type instanceResourceModel struct {
	AntiAffinityGroups        types.Set                        `tfsdk:"anti_affinity_groups"`
	AutoRestartPolicy         types.String                     `tfsdk:"auto_restart_policy"`
	BootDiskID                types.String                     `tfsdk:"boot_disk_id"`
	Description               types.String                     `tfsdk:"description"`
	DiskAttachments           types.Set                        `tfsdk:"disk_attachments"`
	ExternalIPs               *instanceResourceExternalIPModel `tfsdk:"external_ips"`
	HostnameDeprecated        types.String                     `tfsdk:"host_name"`
	Hostname                  types.String                     `tfsdk:"hostname"`
	ID                        types.String                     `tfsdk:"id"`
	Memory                    types.Int64                      `tfsdk:"memory"`
	Name                      types.String                     `tfsdk:"name"`
	NetworkInterfaces         []instanceResourceNICModel       `tfsdk:"network_interfaces"`
	AttachedNetworkInterfaces types.Map                        `tfsdk:"attached_network_interfaces"`
	NCPUs                     types.Int64                      `tfsdk:"ncpus"`
	ProjectID                 types.String                     `tfsdk:"project_id"`
	SSHPublicKeys             types.Set                        `tfsdk:"ssh_public_keys"`
	StartOnCreate             types.Bool                       `tfsdk:"start_on_create"`
	TimeCreated               types.String                     `tfsdk:"time_created"`
	TimeModified              types.String                     `tfsdk:"time_modified"`
	Timeouts                  timeouts.Value                   `tfsdk:"timeouts"`
	UserData                  types.String                     `tfsdk:"user_data"`
}

type instanceResourceNICModel struct {
	Description  types.String                   `tfsdk:"description"`
	ID           types.String                   `tfsdk:"id"`
	IPAddr       types.String                   `tfsdk:"ip_address"`
	IPConfig     *instanceResourceIPConfigModel `tfsdk:"ip_config"`
	MAC          types.String                   `tfsdk:"mac_address"`
	Name         types.String                   `tfsdk:"name"`
	Primary      types.Bool                     `tfsdk:"primary"`
	SubnetID     types.String                   `tfsdk:"subnet_id"`
	TimeCreated  types.String                   `tfsdk:"time_created"`
	TimeModified types.String                   `tfsdk:"time_modified"`
	VPCID        types.String                   `tfsdk:"vpc_id"`
}

func (nic instanceResourceNICModel) Hash() string {
	h := md5.New()

	// Hash user-provided values to detect changes to the configuration.
	io.WriteString(h, nic.Name.ValueString())
	io.WriteString(h, nic.Description.ValueString())
	io.WriteString(h, nic.SubnetID.ValueString())
	io.WriteString(h, nic.VPCID.ValueString())
	if nic.IPConfig != nil {
		if nic.IPConfig.V4 != nil {
			io.WriteString(h, nic.IPConfig.V4.IP.ValueString())
		}
		if nic.IPConfig.V6 != nil {
			io.WriteString(h, nic.IPConfig.V6.IP.ValueString())
		}
	}
	io.WriteString(h, nic.IPAddr.ValueString())

	return string(h.Sum(nil))
}

type instanceResourceIPConfigModel struct {
	V4 *instanceResourceIPConfigV4Model `tfsdk:"v4"`
	V6 *instanceResourceIPConfigV6Model `tfsdk:"v6"`
}

func (ip *instanceResourceIPConfigModel) Equal(other *instanceResourceIPConfigModel) bool {
	if ip == nil || other == nil {
		return ip == other
	}

	return ip.V4.Equal(other.V4) && ip.V6.Equal(other.V6)
}

type instanceResourceIPConfigV4Model struct {
	IP types.String `tfsdk:"ip"`
}

func (ip *instanceResourceIPConfigV4Model) Equal(other *instanceResourceIPConfigV4Model) bool {
	if ip == nil || other == nil {
		return ip == other
	}

	return ip.IP.Equal(other.IP)
}

type instanceResourceIPConfigV6Model struct {
	IP types.String `tfsdk:"ip"`
}

func (ip *instanceResourceIPConfigV6Model) Equal(other *instanceResourceIPConfigV6Model) bool {
	if ip == nil || other == nil {
		return ip == other
	}

	return ip.IP.Equal(other.IP)
}

type instanceResourceAttachedNICModel struct {
	ID           types.String                 `tfsdk:"id"`
	Name         types.String                 `tfsdk:"name"`
	Description  types.String                 `tfsdk:"description"`
	SubnetID     types.String                 `tfsdk:"subnet_id"`
	VPCID        types.String                 `tfsdk:"vpc_id"`
	InstanceID   types.String                 `tfsdk:"instance_id"`
	Primary      types.Bool                   `tfsdk:"primary"`
	MAC          types.String                 `tfsdk:"mac_address"`
	IPStack      instanceResourceIPStackModel `tfsdk:"ip_stack"`
	TimeCreated  types.String                 `tfsdk:"time_created"`
	TimeModified types.String                 `tfsdk:"time_modified"`
}

type instanceResourceIPStackModel struct {
	V4 *instanceResourceIPStackV4Model `tfsdk:"v4"`
	V6 *instanceResourceIPStackV6Model `tfsdk:"v6"`
}

type instanceResourceIPStackV4Model struct {
	IP types.String `tfsdk:"ip"`
}

type instanceResourceIPStackV6Model struct {
	IP types.String `tfsdk:"ip"`
}

type instanceResourceExternalIPModel struct {
	Ephemeral []instanceResourceEphemeralIPModel `tfsdk:"ephemeral"`
	Floating  []instanceResourceFloatingIPModel  `tfsdk:"floating"`
}

func (ip *instanceResourceExternalIPModel) Empty() bool {
	return ip == nil || len(ip.Ephemeral) == 0 && len(ip.Floating) == 0
}

type instanceResourceEphemeralIPModel struct {
	PoolID    types.String `tfsdk:"pool_id"`
	IPVersion types.String `tfsdk:"ip_version"`
}

type instanceResourceFloatingIPModel struct {
	ID types.String `tfsdk:"id"`
}

var instanceResourceNICType = types.ObjectType{}.WithAttributeTypes(map[string]attr.Type{
	"name":        types.StringType,
	"description": types.StringType,
	"subnet_id":   types.StringType,
	"vpc_id":      types.StringType,
	"ip_config": types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"v4": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"ip": types.StringType,
				},
			},
			"v6": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"ip": types.StringType,
				},
			},
		},
	},
	"ip_address":    types.StringType,
	"mac_address":   types.StringType,
	"id":            types.StringType,
	"primary":       types.BoolType,
	"time_created":  types.StringType,
	"time_modified": types.StringType,
},
)

var instanceResourceAttachedNICType = types.ObjectType{}.WithAttributeTypes(
	map[string]attr.Type{
		"id":          types.StringType,
		"name":        types.StringType,
		"description": types.StringType,
		"subnet_id":   types.StringType,
		"vpc_id":      types.StringType,
		"instance_id": types.StringType,
		"primary":     types.BoolType,
		"mac_address": types.StringType,
		"ip_stack": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"v4": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"ip": types.StringType,
					},
				},
				"v6": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"ip": types.StringType,
					},
				},
			},
		},
		"time_created":  types.StringType,
		"time_modified": types.StringType,
	},
)

// Metadata returns the resource type name.
func (r *instanceResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = "oxide_instance"
}

// Configure adds the provider configured client to the data source.
func (r *instanceResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*oxide.Client)
}

// ImportState imports an existing instance resource into Terraform state.
func (r *instanceResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Schema defines the schema for the resource.
func (r *instanceResource) Schema(
	ctx context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Version: 1,
		MarkdownDescription: replaceBackticks(`
This resource manages instances.

!> Updates will stop and start the instance.

-> When setting a boot disk using ''boot_disk_id'', the boot disk ID must also be present in ''disk_attachments''.
`),
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the project that will contain the instance.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the instance.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Required:    true,
				Description: "Description for the instance.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"host_name": schema.StringAttribute{
				Optional:           true,
				Computed:           true,
				DeprecationMessage: "Use hostname instead. This attribute will be removed in the next minor version of the provider.",
				Description:        "Hostname of the instance.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIf(
						ModifyPlanForHostnameDeprecation, "", "",
					),
				},
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("hostname"),
					),
				},
			},
			"hostname": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Hostname of the instance.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIf(
						ModifyPlanForHostnameDeprecation, "", "",
					),
				},
			},
			"memory": schema.Int64Attribute{
				Required:    true,
				Description: "Instance memory in bytes.",
			},
			"ncpus": schema.Int64Attribute{
				Required:    true,
				Description: "Number of CPUs allocated for this instance.",
			},
			"auto_restart_policy": schema.StringAttribute{
				Optional:    true,
				Description: "The auto-restart policy for this instance.",
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(oxide.InstanceAutoRestartPolicyBestEffort),
						string(oxide.InstanceAutoRestartPolicyNever),
					),
				},
			},
			"anti_affinity_groups": schema.SetAttribute{
				Optional:    true,
				Description: "IDs of the anti-affinity groups this instance should belong to.",
				ElementType: types.StringType,
			},
			"boot_disk_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "ID of the disk the instance should be booted from. When provided, this ID must also be present in `disk_attachments`.",
				Validators: []validator.String{
					stringvalidator.AlsoRequires(
						path.MatchRoot("disk_attachments"),
					),
				},
			},
			"start_on_create": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether to start the instance on creation.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"disk_attachments": schema.SetAttribute{
				Optional:            true,
				MarkdownDescription: "IDs of the disks to be attached to the instance. When multiple disk IDs are provided, set `boot_disk_id` to specify the boot disk for the instance. Otherwise, a boot disk will be chosen randomly.",
				ElementType:         types.StringType,
			},
			"ssh_public_keys": schema.SetAttribute{
				Optional:    true,
				Description: "An allowlist of IDs of the SSH public keys to be transferred to the instance via cloud-init during instance creation.",
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
				Validators: []validator.Set{
					setvalidator.NoNullValues(),
					setvalidator.ValueStringsAre(
						stringvalidator.NoneOf(""),
					),
				},
			},
			// Changes to network_interfaces need to be tracked manually.
			// When adding a new attribute, check if they also need to be
			// added to instanceResourceNICModel.Hash() and
			// instanceNetworkInterfacesPlanModifier.PlanModifySet().
			"network_interfaces": schema.SetNestedAttribute{
				Optional:    true,
				Description: "Network interface devices attached to the instance.",
				PlanModifiers: []planmodifier.Set{
					instanceNetworkInterfacesPlanModifier{},
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "Name of the instance network interface.",
							// TODO: Remove once update is implemented
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplaceIf(
									RequiresReplaceUnlessEmptyStringOrNull(), "", "",
								),
							},
						},
						"description": schema.StringAttribute{
							Required:    true,
							Description: "Description for the instance network interface.",
							// TODO: Remove once update is implemented
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplaceIf(
									RequiresReplaceUnlessEmptyStringOrNull(), "", "",
								),
							},
						},
						"subnet_id": schema.StringAttribute{
							Required:    true,
							Description: "ID of the VPC subnet in which to create the instance network interface.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplaceIf(
									RequiresReplaceUnlessEmptyStringOrNull(), "", "",
								),
							},
						},
						"vpc_id": schema.StringAttribute{
							Required:    true,
							Description: "ID of the VPC in which to create the instance network interface.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplaceIf(
									RequiresReplaceUnlessEmptyStringOrNull(), "", "",
								),
							},
						},
						"ip_config": schema.SingleNestedAttribute{
							// Make this attribute optional to support zero-change provider updates
							// and instance imports. It should be marked as required once the
							// deprecated attributes are removed.
							Optional:    true,
							Description: "IP stack to create for the instance network interface.",
							Validators: []validator.Object{
								instanceIPConfigValidator{},
							},
							Attributes: map[string]schema.Attribute{
								"v4": schema.SingleNestedAttribute{
									Optional:    true,
									Description: "Creates an IPv4 stack for the instance network interface.",
									Attributes: map[string]schema.Attribute{
										"ip": schema.StringAttribute{
											Required: true,
											Validators: []validator.String{
												ipConfigValidator{oxide.IpVersionV4},
											},
											Description: `The IPv4 address for the instance network interface or "auto" to auto-assign one.`,
										},
									},
								},
								"v6": schema.SingleNestedAttribute{
									Optional: true,
									Attributes: map[string]schema.Attribute{
										"ip": schema.StringAttribute{
											Required: true,
											Validators: []validator.String{
												ipConfigValidator{oxide.IpVersionV6},
											},
											Description: `The IPv6 address for the instance network interface or "auto" to auto-assign one.`,
										},
									},
								},
							},
						},
						"ip_address": schema.StringAttribute{
							DeprecationMessage: "Use ip_config to set the instance network interface IP address and attached_network_interfaces[<name>].ip_stack to retrieve its value. This attribute will be removed in the next minor version of the provider.",
							Optional:           true,
							Computed:           true,
							Description: "IP address for the instance network interface. " +
								"One will be auto-assigned if not provided.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplaceIfConfigured(),
							},
						},
						"mac_address": schema.StringAttribute{
							DeprecationMessage: "Use attached_network_interfaces[<name>].mac_address instead.",
							Computed:           true,
							Description:        "MAC address assigned to the instance network interface.",
						},
						"id": schema.StringAttribute{
							DeprecationMessage: "Use attached_network_interfaces[<name>].id instead.",
							Computed:           true,
							Description:        "Unique, immutable, system-controlled identifier of the instance network interface.",
						},
						"primary": schema.BoolAttribute{
							DeprecationMessage: "Use attached_network_interfaces[<name>].primary instead.",
							Computed:           true,
							Description:        "True if this is the primary network interface for the instance to which it's attached to.",
						},
						"time_created": schema.StringAttribute{
							DeprecationMessage: "Use attached_network_interfaces[<name>].time_created instead.",
							Computed:           true,
							Description:        "Timestamp of when this instance network interface was created.",
						},
						"time_modified": schema.StringAttribute{
							DeprecationMessage: "Use attached_network_interfaces[<name>].time_modified instead.",
							Computed:           true,
							Description:        "Timestamp of when this instance network interface was last modified.",
						},
					},
				},
			},
			"attached_network_interfaces": schema.MapNestedAttribute{
				Computed:    true,
				Description: "Network interfaces attached to the instance.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "Unique, immutable, system-controlled identifier of the instance network interface.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Name of the instance network interface.",
						},
						"description": schema.StringAttribute{
							Computed:    true,
							Description: "Description of the instance network interface.",
						},
						"subnet_id": schema.StringAttribute{
							Computed:    true,
							Description: "VPC subnet ID of the instance network interface.",
						},
						"vpc_id": schema.StringAttribute{
							Computed:    true,
							Description: "VPC ID of the instance network interface.",
						},
						"instance_id": schema.StringAttribute{
							Computed:    true,
							Description: "Instance ID of the network interface.",
						},
						"primary": schema.BoolAttribute{
							Computed:    true,
							Description: "True if this is the primary network interface for the instance to which it's attached to.",
						},
						"mac_address": schema.StringAttribute{
							Computed:    true,
							Description: "MAC address of the instance network interface.",
						},
						"ip_stack": schema.SingleNestedAttribute{
							Computed:    true,
							Description: "IP stack of the instance network interface.",
							Attributes: map[string]schema.Attribute{
								"v4": schema.SingleNestedAttribute{
									Computed:    true,
									Description: "IPv4 stack of the instance network interface.",
									Attributes: map[string]schema.Attribute{
										"ip": schema.StringAttribute{
											Computed:    true,
											Description: "IPv4 address of the instance network interface.",
										},
									},
								},
								"v6": schema.SingleNestedAttribute{
									Computed:    true,
									Description: "IPv6 stack of the instance network interface.",
									Attributes: map[string]schema.Attribute{
										"ip": schema.StringAttribute{
											Computed:    true,
											Description: "IPv6 address of the instance network interface.",
										},
									},
								},
							},
						},
						"time_created": schema.StringAttribute{
							Computed:    true,
							Description: "Timestamp of when this instance network interface was created.",
						},
						"time_modified": schema.StringAttribute{
							Computed:    true,
							Description: "Timestamp of when this instance network interface was last modified.",
						},
					},
				},
			},
			"external_ips": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "External IP addresses provided to this instance.",
				Validators: []validator.Object{
					instanceExternalIPValidator{},
					objectvalidator.AlsoRequires(path.MatchRoot("network_interfaces")),
				},
				Attributes: map[string]schema.Attribute{
					"ephemeral": schema.SetNestedAttribute{
						Optional:    true,
						Description: "External ephemeral IPs to attach to the instance. Each instance can have at most one IPv4 and one IPv6 ephemeral IP.",
						Validators: []validator.Set{
							setvalidator.SizeBetween(1, 2),
						},
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"pool_id": schema.StringAttribute{
									Optional:            true,
									Computed:            true,
									MarkdownDescription: "ID of the IP pool to allocate from. Conflicts with `ip_version`.",
									Validators: []validator.String{
										stringvalidator.ConflictsWith(
											path.MatchRelative().AtParent().AtName("ip_version"),
										),
									},
								},
								"ip_version": schema.StringAttribute{
									Optional:            true,
									Computed:            true,
									MarkdownDescription: "IP version to use when multiple default pools exist. Conflicts with `pool_id`.",
									Validators: []validator.String{
										stringvalidator.ConflictsWith(
											path.MatchRelative().AtParent().AtName("pool_id"),
										),
										stringvalidator.OneOf(
											string(oxide.IpVersionV4),
											string(oxide.IpVersionV6),
										),
									},
								},
							},
						},
					},
					"floating": schema.SetNestedAttribute{
						Optional:    true,
						Description: "External floating IPs to attach to the instance.",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"id": schema.StringAttribute{
									Required:    true,
									Description: "The external floating IP ID.",
								},
							},
						},
						Validators: []validator.Set{
							setvalidator.SizeAtLeast(1),
						},
					},
				},
			},
			"user_data": schema.StringAttribute{
				Optional: true,
				MarkdownDescription: `
User data for instance initialization systems (such as cloud-init).
Must be a Base64-encoded string, as specified in [RFC 4648 § 4](https://datatracker.ietf.org/doc/html/rfc4648#section-4).
Maximum 32 KiB unencoded data.`,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Update: true,
				Delete: true,
			}),
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier of the instance.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"time_created": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this instance was created.",
			},
			"time_modified": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this instance was last modified.",
			},
		},
	}
}

func (r *instanceResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: r.schemaV0(ctx),
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var oldState instanceResourceModelV0

				resp.Diagnostics.Append(req.State.Get(ctx, &oldState)...)
				if resp.Diagnostics.HasError() {
					return
				}

				// Migrate network interfaces.
				var newNICs []instanceResourceNICModel
				for _, oldNIC := range oldState.NetworkInterfaces {
					newNIC := instanceResourceNICModel{
						Description:  oldNIC.Description,
						ID:           oldNIC.ID,
						IPAddr:       oldNIC.IPAddr,
						MAC:          oldNIC.MAC,
						Name:         oldNIC.Name,
						Primary:      oldNIC.Primary,
						SubnetID:     oldNIC.SubnetID,
						TimeCreated:  oldNIC.TimeCreated,
						TimeModified: oldNIC.TimeModified,
						VPCID:        oldNIC.VPCID,
					}

					if oldNIC.IPConfig != nil {
						newNIC.IPConfig = &instanceResourceIPConfigModel{}

						if oldNIC.IPConfig.V4 != nil {
							newNIC.IPConfig.V4 = &instanceResourceIPConfigV4Model{
								IP: oldNIC.IPConfig.V4.IP,
							}
						}

						if oldNIC.IPConfig.V6 != nil {
							newNIC.IPConfig.V6 = &instanceResourceIPConfigV6Model{
								IP: oldNIC.IPConfig.V6.IP,
							}
						}
					}

					newNICs = append(newNICs, newNIC)
				}

				// Migrate external IPs.
				var newExtIPs *instanceResourceExternalIPModel
				if len(oldState.ExternalIPs) > 0 {
					newExtIPs = &instanceResourceExternalIPModel{}
					for _, oldExtIP := range oldState.ExternalIPs {
						switch oxide.ExternalIpKind(oldExtIP.Type.ValueString()) {
						case oxide.ExternalIpKindEphemeral:
							newExtIPs.Ephemeral = append(
								newExtIPs.Ephemeral,
								instanceResourceEphemeralIPModel{
									PoolID: oldExtIP.ID,
								},
							)

						case oxide.ExternalIpKindFloating:
							newExtIPs.Floating = append(
								newExtIPs.Floating,
								instanceResourceFloatingIPModel{
									ID: oldExtIP.ID,
								},
							)
						}
					}
				}

				newState := instanceResourceModel{
					AntiAffinityGroups:        oldState.AntiAffinityGroups,
					AutoRestartPolicy:         oldState.AutoRestartPolicy,
					BootDiskID:                oldState.BootDiskID,
					Description:               oldState.Description,
					DiskAttachments:           oldState.DiskAttachments,
					ExternalIPs:               newExtIPs,
					HostnameDeprecated:        oldState.HostnameDeprecated,
					Hostname:                  oldState.Hostname,
					ID:                        oldState.ID,
					Memory:                    oldState.Memory,
					Name:                      oldState.Name,
					NetworkInterfaces:         newNICs,
					AttachedNetworkInterfaces: oldState.AttachedNetworkInterfaces,
					NCPUs:                     oldState.NCPUs,
					ProjectID:                 oldState.ProjectID,
					SSHPublicKeys:             oldState.SSHPublicKeys,
					StartOnCreate:             oldState.StartOnCreate,
					TimeCreated:               oldState.TimeCreated,
					TimeModified:              oldState.TimeModified,
					Timeouts:                  oldState.Timeouts,
					UserData:                  oldState.UserData,
				}

				resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *instanceResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan instanceResourceModel

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

	// Determine the hostname value from the new hostname attribute or the
	// deprecated host_name attribute.
	var hostnameValue string
	if !plan.Hostname.IsNull() && !plan.Hostname.IsUnknown() {
		hostnameValue = plan.Hostname.ValueString()
	} else {
		hostnameValue = plan.HostnameDeprecated.ValueString()
	}

	params := oxide.InstanceCreateParams{
		Project: oxide.NameOrId(plan.ProjectID.ValueString()),
		Body: &oxide.InstanceCreate{
			Description: plan.Description.ValueString(),
			Name:        oxide.Name(plan.Name.ValueString()),
			Hostname:    oxide.Hostname(hostnameValue),
			Memory:      oxide.ByteCount(plan.Memory.ValueInt64()),
			Ncpus:       oxide.InstanceCpuCount(plan.NCPUs.ValueInt64()),
			Start:       plan.StartOnCreate.ValueBoolPointer(),
			UserData:    plan.UserData.ValueString(),
		},
	}

	// Add auto-restart policy if any.
	if !plan.AutoRestartPolicy.IsNull() {
		params.Body.AutoRestartPolicy = oxide.InstanceAutoRestartPolicy(
			plan.AutoRestartPolicy.ValueString(),
		)
	}

	// Add boot disk if any.
	if !plan.BootDiskID.IsNull() {
		// Validate whether the boot disk ID is included in `attachments`
		// This is necessary as the response from InstanceDiskList includes
		// the boot disk and would result in an inconsistent state in terraform
		isBootIDPresent, err := attrValueSliceContains(
			plan.DiskAttachments.Elements(),
			plan.BootDiskID.ValueString(),
		)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error unquoting disk attachments",
				"Validation error: "+err.Error(),
			)
			return
		}

		if !isBootIDPresent {
			resp.Diagnostics.AddError(
				"Validation error",
				"Boot disk ID should be part of `disk_attachments`",
			)
			return
		}

		diskParams := oxide.DiskViewParams{
			Disk: oxide.NameOrId(plan.BootDiskID.ValueString()),
		}
		diskView, err := r.client.DiskView(ctx, diskParams)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error retrieving boot disk information",
				"API error: "+err.Error(),
			)
			return
		}
		params.Body.BootDisk = &oxide.InstanceDiskAttachment{
			Name: diskView.Name,
			Type: oxide.InstanceDiskAttachmentTypeAttach,
		}
	}

	sshKeys, diags := newNameOrIdList(plan.SSHPublicKeys)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	params.Body.SshPublicKeys = sshKeys

	antiAffinityGroupIDs, diags := newNameOrIdList(plan.AntiAffinityGroups)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	params.Body.AntiAffinityGroups = antiAffinityGroupIDs

	disks, diags := newDiskAttachmentsOnCreate(ctx, r.client, plan.DiskAttachments)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The control plane API counts the BootDisk and the Disk attachments when it calculates the
	// limit on disk attachments. If bootdisk is set explicitly, we don't want it to be in the API
	// call, but we need it in the state entry.
	params.Body.Disks = filterBootDiskFromDisks(disks, params.Body.BootDisk)

	externalIPs := newExternalIPsOnCreate(plan.ExternalIPs)
	params.Body.ExternalIps = externalIPs

	nics, diags := newNetworkInterfaceAttachment(ctx, r.client, plan.NetworkInterfaces)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	params.Body.NetworkInterfaces = nics

	instance, err := r.client.InstanceCreate(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating instance",
			"API error: "+err.Error(),
		)
		return
	}

	tflog.Trace(
		ctx,
		fmt.Sprintf("created instance with ID: %v", instance.Id),
		map[string]any{"success": true},
	)

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(instance.Id)
	plan.Hostname = types.StringValue(instance.Hostname)
	plan.HostnameDeprecated = types.StringValue(instance.Hostname)
	plan.TimeCreated = types.StringValue(instance.TimeCreated.String())
	plan.TimeModified = types.StringValue(instance.TimeModified.String())

	// Populate Computed attribute values about external IPs.
	instExternalIPs, diags := newAttachedExternalIPModel(ctx, r.client, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	for i, ip := range instExternalIPs.Ephemeral {
		plan.ExternalIPs.Ephemeral[i].PoolID = ip.PoolID
		plan.ExternalIPs.Ephemeral[i].IPVersion = ip.IPVersion
	}

	// Populate Computed attribute values about network interfaces.
	_, attachedNICs, diags := newAttachedNetworkInterfacesModel(ctx, r.client, plan)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	plan.AttachedNetworkInterfaces, diags = types.MapValueFrom(
		ctx, instanceResourceAttachedNICType, attachedNICs,
	)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	for i, nic := range plan.NetworkInterfaces {
		attachedNIC, ok := attachedNICs[nic.Name.ValueString()]
		if !ok {
			diags.AddWarning(
				"Missing network interface",
				fmt.Sprintf(
					"Network interface %s is not attached to instance.",
					nic.Name.ValueString(),
				),
			)
			continue
		}

		// Populated deprecated Computed attributes.
		plan.NetworkInterfaces[i].ID = attachedNIC.ID
		plan.NetworkInterfaces[i].TimeCreated = attachedNIC.TimeCreated
		plan.NetworkInterfaces[i].TimeModified = attachedNIC.TimeModified
		plan.NetworkInterfaces[i].MAC = attachedNIC.MAC
		plan.NetworkInterfaces[i].Primary = attachedNIC.Primary

		var ipAddr string
		if attachedNIC.IPStack.V4 != nil {
			ipAddr = attachedNIC.IPStack.V4.IP.ValueString()
		}
		plan.NetworkInterfaces[i].IPAddr = types.StringValue(ipAddr)
	}

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *instanceResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state instanceResourceModel

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

	params := oxide.InstanceViewParams{
		Instance: oxide.NameOrId(state.ID.ValueString()),
	}
	instance, err := r.client.InstanceView(ctx, params)
	if err != nil {
		if is404(err) {
			// Remove resource from state during a refresh
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to read instance:",
			"API error: "+err.Error(),
		)
		return
	}

	tflog.Trace(
		ctx,
		fmt.Sprintf("read instance with ID: %v", instance.Id),
		map[string]any{"success": true},
	)

	if instance.BootDiskId != "" {
		state.BootDiskID = types.StringValue(instance.BootDiskId)
	}
	if instance.AutoRestartPolicy != "" {
		state.AutoRestartPolicy = types.StringValue(string(instance.AutoRestartPolicy))
	}
	state.Description = types.StringValue(instance.Description)

	// Set both attributes to the same value to facilitate migration across
	// attributes.
	state.Hostname = types.StringValue(string(instance.Hostname))
	state.HostnameDeprecated = types.StringValue(string(instance.Hostname))

	state.ID = types.StringValue(instance.Id)
	state.Memory = types.Int64Value(int64(instance.Memory))
	state.Name = types.StringValue(string(instance.Name))
	state.NCPUs = types.Int64Value(int64(instance.Ncpus))
	state.ProjectID = types.StringValue(instance.ProjectId)
	state.TimeCreated = types.StringValue(instance.TimeCreated.String())
	state.TimeModified = types.StringValue(instance.TimeModified.String())

	externalIPs, diags := newAttachedExternalIPModel(ctx, r.client, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Only set the external IPs if there are any to avoid drift.
	if !externalIPs.Empty() {
		state.ExternalIPs = externalIPs
	}

	keySet, diags := newAssociatedSSHKeysOnCreateSet(ctx, r.client, state.ID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Only set the SSH key list if there are any associated keys
	if len(keySet.Elements()) > 0 {
		state.SSHPublicKeys = keySet
	}

	antiAffinityGroupSet, diags := newAssociatedAntiAffinityGroupsOnCreateSet(
		ctx,
		r.client,
		state.ID.ValueString(),
	)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Only set the anti-affinity group list if there are any associated groups
	if len(antiAffinityGroupSet.Elements()) > 0 {
		state.AntiAffinityGroups = antiAffinityGroupSet
	}

	diskSet, diags := newAttachedDisksSet(ctx, r.client, state.ID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Only set the disk list if there are disk attachments
	if len(diskSet.Elements()) > 0 {
		state.DiskAttachments = diskSet
	}

	nicSet, attachedNICs, diags := newAttachedNetworkInterfacesModel(ctx, r.client, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.AttachedNetworkInterfaces, diags = types.MapValueFrom(
		ctx, instanceResourceAttachedNICType, attachedNICs,
	)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Only populate NICs if there are associated NICs to avoid drift
	if len(nicSet) > 0 {
		state.NetworkInterfaces = nicSet
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *instanceResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan instanceResourceModel
	var state instanceResourceModel

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

	updateTimeout, diags := state.Timeouts.Update(ctx, defaultTimeout())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	// The instance must be stopped for all updates
	stopParams := oxide.InstanceStopParams{
		Instance: oxide.NameOrId(state.ID.ValueString()),
	}
	_, err := r.client.InstanceStop(ctx, stopParams)
	if err != nil {
		if !is404(err) {
			resp.Diagnostics.AddError(
				"Unable to stop instance:",
				"API error: "+err.Error(),
			)
			return
		}
	}

	diags = waitForInstanceStop(ctx, r.client, updateTimeout, state.ID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Trace(
		ctx,
		fmt.Sprintf("stopped instance with ID: %v", state.ID.ValueString()),
		map[string]any{"success": true},
	)

	// Update disk attachments
	//
	// We attach new disks first in case the new boot disk is one of the newly added
	// disks
	planDisks := plan.DiskAttachments.Elements()
	stateDisks := state.DiskAttachments.Elements()

	// Check plan and if it has an ID that the state doesn't then attach it
	disksToAttach := sliceDiff(planDisks, stateDisks)
	resp.Diagnostics.Append(attachDisks(ctx, r.client, disksToAttach, state.ID.ValueString())...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update instance only if configurable instance params change.
	// Due to the design of the API, when updating an instance all
	// parameters must be set. This means we set all of the parameters
	// even if only a single one changed.
	if state.AutoRestartPolicy != plan.AutoRestartPolicy ||
		state.BootDiskID != plan.BootDiskID ||
		state.Memory != plan.Memory ||
		state.NCPUs != plan.NCPUs {

		params := oxide.InstanceUpdateParams{
			Instance: oxide.NameOrId(state.ID.ValueString()),
			Body: &oxide.InstanceUpdate{
				Memory: oxide.ByteCount(plan.Memory.ValueInt64()),
				Ncpus:  oxide.InstanceCpuCount(plan.NCPUs.ValueInt64()),
			},
		}
		if !plan.AutoRestartPolicy.IsNull() {
			params.Body.AutoRestartPolicy = (*oxide.InstanceAutoRestartPolicy)(
				plan.AutoRestartPolicy.ValueStringPointer(),
			)
		}
		if !plan.BootDiskID.IsNull() {
			params.Body.BootDisk = (*oxide.NameOrId)(plan.BootDiskID.ValueStringPointer())
		}
		instance, err := r.client.InstanceUpdate(ctx, params)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to read instance:",
				"API error: "+err.Error(),
			)
			return
		}

		tflog.Trace(ctx, fmt.Sprintf(
			"updated boot disk forinstance with ID: %v",
			instance.Id,
		), map[string]any{"success": true},
		)
	}

	// Check state and if it has an ID that the plan doesn't then detach it
	//
	// We only detach disks once we have made changes to the boot disk (if any)
	// in case we need to remove the previous boot disk
	disksToDetach := sliceDiff(stateDisks, planDisks)
	resp.Diagnostics.Append(detachDisks(ctx, r.client, disksToDetach, state.ID.ValueString())...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update NICs and external IPs.
	//
	// These attributes are related, and must be updated in a particular order,
	// because the external IP versions that can be attached to an instance
	// depends on the network interface IP stack version.

	// Detach external IPs to free the network interface.
	//
	// For example, if there are IPv6 external IPs left, we can't remove IPv6
	// network interfaces.
	resp.Diagnostics.Append(
		detachExternalIPs(
			ctx, r.client, state.ID.ValueString(),
			state.ExternalIPs, plan.ExternalIPs,
		)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Remove network interfaces to allow potentially conflicting adds.
	//
	// For example, each network interface needs to be in a separate subnet, so
	// we must first make room for new network interface if it is replacing one
	// in the same subnet.
	nicsToDelete := sliceDiffByID(
		state.NetworkInterfaces,
		plan.NetworkInterfaces,
		func(e instanceResourceNICModel) any {
			return e.Hash()
		},
	)
	resp.Diagnostics.Append(deleteNICs(ctx, r.client, nicsToDelete)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Add new network interfaces.
	nicsToCreate := sliceDiffByID(
		plan.NetworkInterfaces,
		state.NetworkInterfaces,
		func(e instanceResourceNICModel) any {
			return e.Hash()
		},
	)
	resp.Diagnostics.Append(createNICs(ctx, r.client, nicsToCreate, state.ID.ValueString())...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Add new external IPs.
	resp.Diagnostics.Append(
		attachExternalIPs(
			ctx,
			r.client,
			state.ID.ValueString(),
			state.ExternalIPs,
			plan.ExternalIPs,
		)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update anti-affinity groups
	planAntiAffinityGroups := plan.AntiAffinityGroups.Elements()
	stateAntiAffinityGroups := state.AntiAffinityGroups.Elements()

	// Check plan and if it has an ID that the state doesn't then add it
	antiAffinityGroupsToAdd := sliceDiff(planAntiAffinityGroups, stateAntiAffinityGroups)
	resp.Diagnostics.Append(
		addAntiAffinityGroups(ctx, r.client, antiAffinityGroupsToAdd, state.ID.ValueString())...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check state and if it has an ID that the plan doesn't then remove it
	antiAffinityGroupsToRemove := sliceDiff(stateAntiAffinityGroups, planAntiAffinityGroups)
	resp.Diagnostics.Append(
		removeAntiAffinityGroups(
			ctx,
			r.client,
			antiAffinityGroupsToRemove,
			state.ID.ValueString(),
		)...)
	if resp.Diagnostics.HasError() {
		return
	}

	startParams := oxide.InstanceStartParams{Instance: oxide.NameOrId(state.ID.ValueString())}
	_, err = r.client.InstanceStart(ctx, startParams)
	if err != nil {
		if !is404(err) {
			resp.Diagnostics.AddError(
				"Unable to start instance:",
				"API error: "+err.Error(),
			)
			return
		}
	}

	// Read instance to retrieve modified time value if this is the only update we are doing
	params := oxide.InstanceViewParams{
		Instance: oxide.NameOrId(state.ID.ValueString()),
	}
	instance, err := r.client.InstanceView(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read instance:",
			"API error: "+err.Error(),
		)
		return
	}

	tflog.Trace(
		ctx,
		fmt.Sprintf("read instance with ID: %v", instance.Id),
		map[string]any{"success": true},
	)

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(instance.Id)
	plan.Hostname = types.StringValue(instance.Hostname)
	plan.HostnameDeprecated = types.StringValue(instance.Hostname)
	plan.ProjectID = types.StringValue(instance.ProjectId)
	plan.TimeCreated = types.StringValue(instance.TimeCreated.String())
	plan.TimeModified = types.StringValue(instance.TimeModified.String())

	// We use the plan here instead of the state to capture the desired IP pool ID
	// value for the ephemeral external IP rather than the previous value.
	externalIPs, diags := newAttachedExternalIPModel(ctx, r.client, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !externalIPs.Empty() {
		plan.ExternalIPs = externalIPs
	}

	// TODO: should I do this or read from the newly created ones?
	diskSet, diags := newAttachedDisksSet(ctx, r.client, state.ID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Only set the disk list if there are disk attachments
	if len(diskSet.Elements()) > 0 {
		plan.DiskAttachments = diskSet
	}

	// TODO: should I do this or read from the newly created ones?
	nicModel, attachedNICs, diags := newAttachedNetworkInterfacesModel(ctx, r.client, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.AttachedNetworkInterfaces, diags = types.MapValueFrom(
		ctx, instanceResourceAttachedNICType, attachedNICs,
	)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Only populate NICs if there are associated NICs to avoid drift
	if len(nicModel) > 0 {
		plan.NetworkInterfaces = nicModel
	}

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *instanceResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state instanceResourceModel

	// Read Terraform prior state data into the model
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

	params := oxide.InstanceStopParams{
		Instance: oxide.NameOrId(state.ID.ValueString()),
	}
	_, err := r.client.InstanceStop(ctx, params)
	if err != nil {
		if !is404(err) {
			resp.Diagnostics.AddError(
				"Unable to stop instance:",
				"API error: "+err.Error(),
			)
			return
		}
	}

	diags = waitForInstanceStop(ctx, r.client, deleteTimeout, state.ID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Trace(
		ctx,
		fmt.Sprintf("stopped instance with ID: %v", state.ID.ValueString()),
		map[string]any{"success": true},
	)

	params2 := oxide.InstanceDeleteParams{
		Instance: oxide.NameOrId(state.ID.ValueString()),
	}
	if err := r.client.InstanceDelete(ctx, params2); err != nil {
		if !is404(err) {
			resp.Diagnostics.AddError(
				"Unable to delete instance:",
				"API error: "+err.Error(),
			)
			return
		}
	}
	tflog.Trace(
		ctx,
		fmt.Sprintf("deleted instance with ID: %v", state.ID.ValueString()),
		map[string]any{"success": true},
	)
}

func waitForInstanceStop(
	ctx context.Context,
	client *oxide.Client,
	timeout time.Duration,
	instanceID string,
) diag.Diagnostics {
	var diags diag.Diagnostics

	stateConfig := retry.StateChangeConf{
		PollInterval: time.Second,
		Delay:        time.Second,
		Pending: []string{
			string(oxide.InstanceStateCreating),
			string(oxide.InstanceStateStarting),
			string(oxide.InstanceStateRunning),
			string(oxide.InstanceStateStopping),
			string(oxide.InstanceStateRebooting),
			string(oxide.InstanceStateMigrating),
			string(oxide.InstanceStateRepairing),
		},
		Target:  []string{string(oxide.InstanceStateStopped)},
		Timeout: timeout,
		Refresh: func() (any, string, error) {
			tflog.Info(ctx, fmt.Sprintf("checking on state of instance: %v", instanceID))
			params := oxide.InstanceViewParams{
				Instance: oxide.NameOrId(instanceID),
			}
			instance, err := client.InstanceView(ctx, params)
			if err != nil {
				if !is404(err) {
					return nil, "nil", fmt.Errorf(
						"while polling for the status of instance %v: %v",
						instanceID,
						err,
					)
				}
				return instance, "", nil
			}
			tflog.Trace(
				ctx,
				fmt.Sprintf("read instance with ID: %v", instanceID),
				map[string]any{"success": true},
			)
			return instance, string(instance.RunState), nil
		},
	}
	if _, err := stateConfig.WaitForStateContext(ctx); err != nil {
		if !is404(err) {
			diags.AddError(
				"Error stopping instance",
				"API error: "+err.Error(),
			)
		}
		return diags
	}

	return nil
}

func newAttachedDisksSet(
	ctx context.Context,
	client *oxide.Client,
	instanceID string,
) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics

	params := oxide.InstanceDiskListParams{
		Limit:    oxide.NewPointer(1000000000),
		Instance: oxide.NameOrId(instanceID),
	}
	disks, err := client.InstanceDiskList(ctx, params)
	if err != nil {
		diags.AddError(
			"Unable to list attached disks:",
			"API error: "+err.Error(),
		)
		return types.SetNull(types.StringType), diags
	}

	d := []attr.Value{}
	for _, disk := range disks.Items {
		id := types.StringValue(disk.Id)
		d = append(d, id)
	}
	diskSet, diags := types.SetValue(types.StringType, d)
	diags.Append(diags...)
	if diags.HasError() {
		return types.SetNull(types.StringType), diags
	}

	return diskSet, nil
}

func newAssociatedSSHKeysOnCreateSet(
	ctx context.Context,
	client *oxide.Client,
	instanceID string,
) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics

	params := oxide.InstanceSshPublicKeyListParams{
		Limit:    oxide.NewPointer(1000000000),
		Instance: oxide.NameOrId(instanceID),
	}
	keys, err := client.InstanceSshPublicKeyList(ctx, params)
	if err != nil {
		diags.AddError(
			"Unable to list associated SSH keys:",
			"API error: "+err.Error(),
		)
		return types.SetNull(types.StringType), diags
	}

	d := []attr.Value{}
	for _, key := range keys.Items {
		id := types.StringValue(key.Id)
		d = append(d, id)
	}
	keySet, diags := types.SetValue(types.StringType, d)
	diags.Append(diags...)
	if diags.HasError() {
		return types.SetNull(types.StringType), diags
	}

	return keySet, nil
}

func newAssociatedAntiAffinityGroupsOnCreateSet(
	ctx context.Context,
	client *oxide.Client,
	instanceID string,
) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics

	params := oxide.InstanceAntiAffinityGroupListParams{
		Limit:    oxide.NewPointer(1000000000),
		Instance: oxide.NameOrId(instanceID),
	}
	groups, err := client.InstanceAntiAffinityGroupList(ctx, params)
	if err != nil {
		diags.AddError(
			"Unable to list associated anti-affinity groups:",
			"API error: "+err.Error(),
		)
		return types.SetNull(types.StringType), diags
	}

	d := []attr.Value{}
	for _, group := range groups.Items {
		id := types.StringValue(group.Id)
		d = append(d, id)
	}
	groupSet, diags := types.SetValue(types.StringType, d)
	diags.Append(diags...)
	if diags.HasError() {
		return types.SetNull(types.StringType), diags
	}

	return groupSet, nil
}

func newNetworkInterfaceAttachment(
	ctx context.Context,
	client *oxide.Client,
	model []instanceResourceNICModel,
) (
	oxide.InstanceNetworkInterfaceAttachment, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(model) == 0 {
		return oxide.InstanceNetworkInterfaceAttachment{
			Type: oxide.InstanceNetworkInterfaceAttachmentTypeNone,
		}, diags
	}

	var nicParams []oxide.InstanceNetworkInterfaceCreate
	for _, planNIC := range model {
		names, diags := retrieveVPCandSubnetNames(ctx, client, planNIC.VPCID.ValueString(),
			planNIC.SubnetID.ValueString())
		diags.Append(diags...)
		if diags.HasError() {
			return oxide.InstanceNetworkInterfaceAttachment{}, diags
		}

		nic := oxide.InstanceNetworkInterfaceCreate{
			Description: planNIC.Description.ValueString(),
			Name:        oxide.Name(planNIC.Name.ValueString()),
			SubnetName:  oxide.Name(names.subnet),
			VpcName:     oxide.Name(names.vpc),
			IpConfig:    newIPStackCreate(planNIC),
		}
		nicParams = append(nicParams, nic)
	}

	nicAttachment := oxide.InstanceNetworkInterfaceAttachment{
		Type:   oxide.InstanceNetworkInterfaceAttachmentTypeCreate,
		Params: nicParams,
	}
	return nicAttachment, nil
}

func newAttachedNetworkInterfacesModel(
	ctx context.Context,
	client *oxide.Client,
	state instanceResourceModel,
) (
	[]instanceResourceNICModel, map[string]instanceResourceAttachedNICModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Store network interfaces from state or plan in a map for quick retrieval
	// of write-only attribute values that are preserved from the state or plan
	// instead of read from the API.
	stateIPConfigs := make(map[string]*instanceResourceIPConfigModel)
	for _, nic := range state.NetworkInterfaces {
		if nic.IPConfig != nil {
			stateIPConfigs[nic.Name.ValueString()] = nic.IPConfig
		}
	}

	params := oxide.InstanceNetworkInterfaceListParams{
		Instance: oxide.NameOrId(state.ID.ValueString()),
		Limit:    oxide.NewPointer(1000000000),
	}
	nics, err := client.InstanceNetworkInterfaceList(ctx, params)
	if err != nil {
		diags.AddError(
			"Unable to read instance network interfaces:",
			"API error: "+err.Error(),
		)
		return []instanceResourceNICModel{}, nil, diags
	}

	nicSet := []instanceResourceNICModel{}
	attachedNICs := make(map[string]instanceResourceAttachedNICModel)
	for _, nic := range nics.Items {
		ipStack, err := newAttachedNetworkInterfacesIPStackModel(nic.IpStack)
		if err != nil {
			diags.AddError(
				"Unable to read instance network interfaces:",
				"API error: "+err.Error(),
			)
			return []instanceResourceNICModel{}, nil, diags
		}

		var ipAddr string
		if ipStack.V4 != nil {
			ipAddr = ipStack.V4.IP.ValueString()
		}

		nicSet = append(nicSet, instanceResourceNICModel{
			Description:  types.StringValue(nic.Description),
			ID:           types.StringValue(nic.Id),
			IPConfig:     stateIPConfigs[string(nic.Name)],
			IPAddr:       types.StringValue(ipAddr),
			MAC:          types.StringValue(string(nic.Mac)),
			Name:         types.StringValue(string(nic.Name)),
			Primary:      types.BoolPointerValue(nic.Primary),
			SubnetID:     types.StringValue(nic.SubnetId),
			TimeCreated:  types.StringValue(nic.TimeCreated.String()),
			TimeModified: types.StringValue(nic.TimeModified.String()),
			VPCID:        types.StringValue(nic.VpcId),
		})

		attachedNICs[string(nic.Name)] = instanceResourceAttachedNICModel{
			ID:           types.StringValue(nic.Id),
			Name:         types.StringValue(string(nic.Name)),
			Description:  types.StringValue(nic.Description),
			SubnetID:     types.StringValue(nic.SubnetId),
			VPCID:        types.StringValue(nic.VpcId),
			InstanceID:   types.StringValue(nic.InstanceId),
			Primary:      types.BoolPointerValue(nic.Primary),
			MAC:          types.StringValue(string(nic.Mac)),
			IPStack:      ipStack,
			TimeCreated:  types.StringValue(nic.TimeCreated.String()),
			TimeModified: types.StringValue(nic.TimeModified.String()),
		}
	}
	if diags.HasError() {
		return []instanceResourceNICModel{}, nil, diags
	}

	return nicSet, attachedNICs, nil
}

// newAttachedNetworkInterfacesIPStackModel parses a network interface IP stack
// from the API to a resource model.
func newAttachedNetworkInterfacesIPStackModel(
	stack oxide.PrivateIpStack,
) (instanceResourceIPStackModel, error) {
	switch s := stack.Value.(type) {
	case *oxide.PrivateIpStackV4:
		return instanceResourceIPStackModel{
			V4: &instanceResourceIPStackV4Model{
				IP: types.StringValue(s.Value.Ip),
			},
		}, nil

	case *oxide.PrivateIpStackV6:
		return instanceResourceIPStackModel{
			V6: &instanceResourceIPStackV6Model{
				IP: types.StringValue(s.Value.Ip),
			},
		}, nil

	case *oxide.PrivateIpStackDualStack:
		return instanceResourceIPStackModel{
			V4: &instanceResourceIPStackV4Model{
				IP: types.StringValue(s.Value.V4.Ip),
			},
			V6: &instanceResourceIPStackV6Model{
				IP: types.StringValue(s.Value.V6.Ip),
			},
		}, nil

	default:
		return instanceResourceIPStackModel{}, fmt.Errorf(
			"unexpected IP stack type %T",
			stack.Value,
		)
	}
}

// newAttachedExternalIPModel fetches the external IP addresses for the instance
// specified by model, keeping the IP pool ID from the ephemeral external IP
// from the model if one is present.
func newAttachedExternalIPModel(
	ctx context.Context,
	client *oxide.Client,
	model instanceResourceModel,
) (
	*instanceResourceExternalIPModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	externalIPResponse, err := client.InstanceExternalIpList(
		ctx,
		oxide.InstanceExternalIpListParams{
			Instance: oxide.NameOrId(model.ID.ValueString()),
		},
	)
	if err != nil {
		diags.AddError(
			"Unable to list instance external ips:",
			"API error: "+err.Error(),
		)
		return nil, diags
	}

	externalIPs := &instanceResourceExternalIPModel{}
	for _, ip := range externalIPResponse.Items {
		switch ip.Kind {
		case oxide.ExternalIpKindEphemeral:
			// The API requires the IP version to delete ephemeral IPs, but it
			// doesn't return the original value, so infer it from the shape of
			// the IP address so it is always present in state, even if the
			// user does not provide it.
			var ipVersion string
			if isIPv4(ip.Ip) {
				ipVersion = string(oxide.IpVersionV4)
			} else if isIPv6(ip.Ip) {
				ipVersion = string(oxide.IpVersionV6)
			}

			externalIPs.Ephemeral = append(externalIPs.Ephemeral, instanceResourceEphemeralIPModel{
				PoolID:    types.StringValue(ip.IpPoolId),
				IPVersion: types.StringValue(ipVersion),
			})

		case oxide.ExternalIpKindFloating:
			externalIPs.Floating = append(externalIPs.Floating, instanceResourceFloatingIPModel{
				ID: types.StringValue(ip.Id),
			})
		// Skipped until the schema is updated to support SNAT external IPs.
		case oxide.ExternalIpKindSnat:
			continue
		default:
			diags.AddError(
				"Invalid external IP kind:",
				fmt.Sprintf("Encountered unexpected external IP kind: %s", ip.Kind),
			)
		}
	}

	return externalIPs, nil
}

type vpcAndSubnetNames struct {
	vpc    string
	subnet string
}

func retrieveVPCandSubnetNames(ctx context.Context, client *oxide.Client, vpcID, subnetID string) (
	vpcAndSubnetNames, diag.Diagnostics) {
	var diags diag.Diagnostics
	// This is an unfortunate result of having the create body use names as identifiers
	// but the body return IDs. making two API calls to retrieve VPC and subnet names.
	// Using IDs only for the provider schema as names are mutable.
	params := oxide.VpcViewParams{
		Vpc: oxide.NameOrId(vpcID),
	}
	vpc, err := client.VpcView(ctx, params)
	if err != nil {
		diags.AddError(
			"Unable to read information about corresponding VPC:",
			"API error: "+err.Error(),
		)
		return vpcAndSubnetNames{}, diags
	}
	tflog.Trace(ctx, fmt.Sprintf("read VPC with ID: %v", vpcID), map[string]any{"success": true})

	params2 := oxide.VpcSubnetViewParams{
		Subnet: oxide.NameOrId(subnetID),
	}
	subnet, err := client.VpcSubnetView(ctx, params2)
	if err != nil {
		diags.AddError(
			"Unable to read information about corresponding subnet:",
			"API error: "+err.Error(),
		)
		return vpcAndSubnetNames{}, diags
	}
	tflog.Trace(ctx, fmt.Sprintf("read subnet with ID: %v", subnetID),
		map[string]any{"success": true})

	return vpcAndSubnetNames{
		vpc:    string(vpc.Name),
		subnet: string(subnet.Name),
	}, nil
}

func newDiskAttachmentsOnCreate(
	ctx context.Context,
	client *oxide.Client,
	diskIDs types.Set,
) ([]oxide.InstanceDiskAttachment, diag.Diagnostics) {
	var diags diag.Diagnostics
	var disks = []oxide.InstanceDiskAttachment{}
	for _, diskAttch := range diskIDs.Elements() {
		diskID, err := strconv.Unquote(diskAttch.String())
		if err != nil {
			diags.AddError(
				"Error retrieving disk information",
				"Disk ID parse error: "+err.Error(),
			)
			return []oxide.InstanceDiskAttachment{}, diags
		}

		params := oxide.DiskViewParams{
			Disk: oxide.NameOrId(diskID),
		}
		disk, err := client.DiskView(ctx, params)
		if err != nil {
			diags.AddError(
				"Error retrieving disk information",
				"API error: "+err.Error(),
			)
			return []oxide.InstanceDiskAttachment{}, diags
		}

		da := oxide.InstanceDiskAttachment{
			Name: disk.Name,
			// Only allow attach (no disk create on instance create)
			Type: oxide.InstanceDiskAttachmentTypeAttach,
		}
		disks = append(disks, da)
	}

	return disks, diags
}

func filterBootDiskFromDisks(
	disks []oxide.InstanceDiskAttachment,
	boot_disk *oxide.InstanceDiskAttachment,
) []oxide.InstanceDiskAttachment {
	if boot_disk == nil {
		return disks
	}

	var filtered_disks = []oxide.InstanceDiskAttachment{}
	for _, disk := range disks {
		if disk == *boot_disk {
			continue
		}
		filtered_disks = append(filtered_disks, disk)
	}
	return filtered_disks
}

func newExternalIPsOnCreate(externalIPs *instanceResourceExternalIPModel) []oxide.ExternalIpCreate {
	if externalIPs == nil {
		return nil
	}

	var ips []oxide.ExternalIpCreate

	for _, ip := range externalIPs.Ephemeral {
		if pool := ip.PoolID.ValueString(); pool != "" {
			ips = append(ips, oxide.ExternalIpCreate{
				Type: oxide.ExternalIpCreateTypeEphemeral,
				PoolSelector: oxide.PoolSelector{
					Type: oxide.PoolSelectorTypeExplicit,
					Pool: oxide.NameOrId(pool),
				},
			})
		} else {
			ips = append(ips, oxide.ExternalIpCreate{
				Type: oxide.ExternalIpCreateTypeEphemeral,
				PoolSelector: oxide.PoolSelector{
					Type:      oxide.PoolSelectorTypeAuto,
					IpVersion: oxide.IpVersion(ip.IPVersion.ValueString()),
				},
			})
		}
	}

	for _, ip := range externalIPs.Floating {
		ips = append(ips, oxide.ExternalIpCreate{
			Type:       oxide.ExternalIpCreateTypeFloating,
			FloatingIp: oxide.NameOrId(ip.ID.ValueString()),
		})
	}

	return ips
}

func newIPStackCreate(model instanceResourceNICModel) oxide.PrivateIpStackCreate {
	// Fallback to the original behaviour if ip_config is not set.
	if model.IPConfig == nil {
		if ip := model.IPAddr.ValueString(); ip != "" {
			return oxide.PrivateIpStackCreate{
				Value: &oxide.PrivateIpStackCreateV4{
					Value: oxide.PrivateIpv4StackCreate{
						Ip: oxide.Ipv4Assignment{
							Type:  oxide.Ipv4AssignmentTypeExplicit,
							Value: ip,
						},
					},
				},
			}
		} else {
			return oxide.PrivateIpStackCreate{
				Value: &oxide.PrivateIpStackCreateV4{
					Value: oxide.PrivateIpv4StackCreate{
						Ip: oxide.Ipv4Assignment{
							Type: oxide.Ipv4AssignmentTypeAuto,
						},
					},
				},
			}
		}
	}

	if model.IPConfig.V4 != nil && model.IPConfig.V6 != nil {
		return oxide.PrivateIpStackCreate{
			Value: &oxide.PrivateIpStackCreateDualStack{
				Value: newIPStackCreateDualStack(model.IPConfig.V4, model.IPConfig.V6),
			},
		}
	}

	if model.IPConfig.V6 != nil {
		return oxide.PrivateIpStackCreate{
			Value: &oxide.PrivateIpStackCreateV6{
				Value: newIPStackCreateV6(model.IPConfig.V6),
			},
		}
	}

	return oxide.PrivateIpStackCreate{
		Value: &oxide.PrivateIpStackCreateV4{
			Value: newIPStackCreateV4(model.IPConfig.V4),
		},
	}
}

func newIPStackCreateV4(stack *instanceResourceIPConfigV4Model) oxide.PrivateIpv4StackCreate {
	ip := stack.IP.ValueString()

	if ip == string(oxide.Ipv4AssignmentTypeAuto) {
		return oxide.PrivateIpv4StackCreate{
			Ip: oxide.Ipv4Assignment{
				Type: oxide.Ipv4AssignmentTypeAuto,
			},
		}
	}

	return oxide.PrivateIpv4StackCreate{
		Ip: oxide.Ipv4Assignment{
			Type:  oxide.Ipv4AssignmentTypeExplicit,
			Value: ip,
		},
	}
}

func newIPStackCreateV6(stack *instanceResourceIPConfigV6Model) oxide.PrivateIpv6StackCreate {
	ip := stack.IP.ValueString()

	if ip == string(oxide.Ipv6AssignmentTypeAuto) {
		return oxide.PrivateIpv6StackCreate{
			Ip: oxide.Ipv6Assignment{
				Type: oxide.Ipv6AssignmentTypeAuto,
			},
		}
	}

	return oxide.PrivateIpv6StackCreate{
		Ip: oxide.Ipv6Assignment{
			Type:  oxide.Ipv6AssignmentTypeExplicit,
			Value: ip,
		},
	}
}

func newIPStackCreateDualStack(
	stackV4 *instanceResourceIPConfigV4Model,
	stackV6 *instanceResourceIPConfigV6Model,
) oxide.PrivateIpStackCreateValue {
	return oxide.PrivateIpStackCreateValue{
		V4: newIPStackCreateV4(stackV4),
		V6: newIPStackCreateV6(stackV6),
	}
}

func createNICs(
	ctx context.Context,
	client *oxide.Client,
	models []instanceResourceNICModel,
	instanceID string,
) diag.Diagnostics {
	for _, model := range models {
		names, diags := retrieveVPCandSubnetNames(ctx, client, model.VPCID.ValueString(),
			model.SubnetID.ValueString())
		diags.Append(diags...)
		if diags.HasError() {
			return diags
		}

		params := oxide.InstanceNetworkInterfaceCreateParams{
			Instance: oxide.NameOrId(instanceID),
			Body: &oxide.InstanceNetworkInterfaceCreate{
				Description: model.Description.ValueString(),
				Name:        oxide.Name(model.Name.ValueString()),
				SubnetName:  oxide.Name(names.subnet),
				VpcName:     oxide.Name(names.vpc),
				IpConfig:    newIPStackCreate(model),
			},
		}

		nic, err := client.InstanceNetworkInterfaceCreate(ctx, params)
		if err != nil {
			diags.AddError(
				"Error creating instance network interface",
				"API error: "+err.Error(),
			)
			return diags
		}
		tflog.Trace(ctx, fmt.Sprintf("created instance network interface with ID: %v", nic.Id),
			map[string]any{"success": true})
	}

	return nil
}

func deleteNICs(
	ctx context.Context,
	client *oxide.Client,
	models []instanceResourceNICModel,
) diag.Diagnostics {
	var diags diag.Diagnostics

	// The API doesn't allow deleting the primary network interface
	// while other interfaces are still attached, so leave it for last.
	slices.SortStableFunc(models, func(a, b instanceResourceNICModel) int {
		if a.Primary.ValueBool() && !b.Primary.ValueBool() {
			return 1
		} else if !a.Primary.ValueBool() && b.Primary.ValueBool() {
			return -1
		}
		return 0
	})

	for _, model := range models {
		params := oxide.InstanceNetworkInterfaceDeleteParams{
			Interface: oxide.NameOrId(model.ID.ValueString()),
		}
		if err := client.InstanceNetworkInterfaceDelete(ctx, params); err != nil {
			if !is404(err) {
				diags.AddError(
					"Error deleting instance network interface:",
					"API error: "+err.Error(),
				)
				// TODO: Should this be a return or a continue?
				return diags
			}
		}
		tflog.Trace(
			ctx,
			fmt.Sprintf("deleted instance network interface with ID: %v", model.ID.ValueString()),
			map[string]any{"success": true},
		)
	}

	return nil
}

// attachExternalIPs attaches the external IPs that are present in the plan but
// are not in the state.
func attachExternalIPs(
	ctx context.Context,
	client *oxide.Client,
	instanceID string,
	state *instanceResourceExternalIPModel,
	plan *instanceResourceExternalIPModel,
) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(attachEphemeralIPs(ctx, client, instanceID, state, plan)...)
	diags.Append(attachFloatingIPs(ctx, client, instanceID, state, plan)...)

	return diags
}

// attachEphemeralIPs attaches the external ephemeral IPs that are present in
// the plan but are not in the state.
func attachEphemeralIPs(
	ctx context.Context,
	client *oxide.Client,
	instanceID string,
	state, plan *instanceResourceExternalIPModel,
) diag.Diagnostics {
	// No need to attach ephemeral IPs if the plan doesn't have any external IP.
	if plan == nil {
		return nil
	}

	var ipsToAttach []instanceResourceEphemeralIPModel
	if state == nil {
		ipsToAttach = plan.Ephemeral
	} else {
		ipsToAttach = sliceDiff(plan.Ephemeral, state.Ephemeral)
	}

	var diags diag.Diagnostics

	for _, ip := range ipsToAttach {
		var params oxide.InstanceEphemeralIpAttachParams

		if ipPool := ip.PoolID.ValueString(); ipPool != "" {
			params = oxide.InstanceEphemeralIpAttachParams{
				Instance: oxide.NameOrId(instanceID),
				Body: &oxide.EphemeralIpCreate{
					PoolSelector: oxide.PoolSelector{
						Type: oxide.PoolSelectorTypeExplicit,
						Pool: oxide.NameOrId(ipPool),
					},
				},
			}
		} else {
			params = oxide.InstanceEphemeralIpAttachParams{
				Instance: oxide.NameOrId(instanceID),
				Body: &oxide.EphemeralIpCreate{
					PoolSelector: oxide.PoolSelector{
						Type:      oxide.PoolSelectorTypeAuto,
						IpVersion: oxide.IpVersion(ip.IPVersion.ValueString()),
					},
				},
			}
		}

		if _, err := client.InstanceEphemeralIpAttach(ctx, params); err != nil {
			diags.AddError(
				fmt.Sprintf("Error attaching ephemeral external IP to instance %s", instanceID),
				"API error: "+err.Error(),
			)
			continue
		}

		tflog.Trace(
			ctx,
			fmt.Sprintf("successfully attached ephemeral external IP to instance %s", instanceID),
			map[string]any{"success": true},
		)
	}

	return diags
}

// attachFloatingIPs attaches the external floating IPs that are present in the
// plan but are not in the state.
func attachFloatingIPs(
	ctx context.Context,
	client *oxide.Client,
	instanceID string,
	state, plan *instanceResourceExternalIPModel,
) diag.Diagnostics {
	// No need to attach floating IPs if the plan doesn't have any external IP.
	if plan == nil {
		return nil
	}

	var ipsToAttach []instanceResourceFloatingIPModel
	if state == nil {
		ipsToAttach = plan.Floating
	} else {
		ipsToAttach = sliceDiff(plan.Floating, state.Floating)
	}

	var diags diag.Diagnostics

	for _, ip := range ipsToAttach {
		params := oxide.FloatingIpAttachParams{
			FloatingIp: oxide.NameOrId(ip.ID.ValueString()),
			Body: &oxide.FloatingIpAttach{
				Kind:   oxide.FloatingIpParentKindInstance,
				Parent: oxide.NameOrId(instanceID),
			},
		}

		if _, err := client.FloatingIpAttach(ctx, params); err != nil {
			diags.AddError(
				fmt.Sprintf("Error attaching floating external IP with ID %s", ip.ID.ValueString()),
				"API error: "+err.Error(),
			)

			return diags
		}

		tflog.Trace(
			ctx,
			fmt.Sprintf(
				"successfully attached floating external IP with ID %s",
				ip.ID.ValueString(),
			),
			map[string]any{"success": true},
		)
	}

	return nil
}

// detachExternalIPs detaches the external IPs that are present in state but
// not in the plan.
func detachExternalIPs(
	ctx context.Context,
	client *oxide.Client,
	instanceID string,
	state *instanceResourceExternalIPModel,
	plan *instanceResourceExternalIPModel,
) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(detachEphemeralIPs(ctx, client, instanceID, state, plan)...)
	diags.Append(detachFloatingIPs(ctx, client, state, plan)...)

	return diags
}

// detachEphemeralIPs detaches the external ephemeral IPs that are present in
// state but not in the plan.
func detachEphemeralIPs(
	ctx context.Context,
	client *oxide.Client,
	instanceID string,
	state, plan *instanceResourceExternalIPModel,
) diag.Diagnostics {
	// No need to detach ephemeral IPs if the state doesn't have any external IP.
	if state == nil {
		return nil
	}

	var ipsToDetach []instanceResourceEphemeralIPModel
	if plan == nil {
		ipsToDetach = state.Ephemeral
	} else {
		ipsToDetach = sliceDiff(state.Ephemeral, plan.Ephemeral)
	}

	var diags diag.Diagnostics

	for _, ip := range ipsToDetach {
		params := oxide.InstanceEphemeralIpDetachParams{
			Instance:  oxide.NameOrId(instanceID),
			IpVersion: oxide.IpVersion(ip.IPVersion.ValueString()),
		}

		if err := client.InstanceEphemeralIpDetach(ctx, params); err != nil {
			diags.AddError(
				fmt.Sprintf(
					"Error detaching ephemeral external IP%s from instance %s",
					ip.IPVersion.ValueString(),
					instanceID,
				),
				"API error: "+err.Error(),
			)
			continue
		}

		tflog.Trace(
			ctx,
			fmt.Sprintf(
				"successfully detached ephemeral external IP%s from instance %s",
				ip.IPVersion.ValueString(),
				instanceID,
			),
			map[string]any{"success": true},
		)
	}

	return diags
}

// detachFloatingIPs detaches the external floating IPs that are present in
// state but not in the plan.
func detachFloatingIPs(
	ctx context.Context,
	client *oxide.Client,
	state, plan *instanceResourceExternalIPModel,
) diag.Diagnostics {
	// No need to detach floating IPs if the state doesn't have any external IP.
	if state == nil {
		return nil
	}

	var ipsToDetach []instanceResourceFloatingIPModel
	if plan == nil {
		ipsToDetach = state.Floating
	} else {
		ipsToDetach = sliceDiff(state.Floating, plan.Floating)
	}

	var diags diag.Diagnostics

	for _, ip := range ipsToDetach {
		params := oxide.FloatingIpDetachParams{
			FloatingIp: oxide.NameOrId(ip.ID.ValueString()),
		}

		if _, err := client.FloatingIpDetach(ctx, params); err != nil {
			diags.AddError(
				fmt.Sprintf("Error detaching floating external IP with ID %s", ip.ID.ValueString()),
				"API error: "+err.Error(),
			)
			continue
		}

		tflog.Trace(
			ctx,
			fmt.Sprintf(
				"successfully detached floating external IP with ID %s",
				ip.ID.ValueString(),
			),
			map[string]any{"success": true},
		)
	}

	return diags
}

func attachDisks(
	ctx context.Context,
	client *oxide.Client,
	disks []attr.Value,
	instanceID string,
) diag.Diagnostics {
	var diags diag.Diagnostics

	for _, v := range disks {
		diskID, err := strconv.Unquote(v.String())
		if err != nil {
			diags.AddError(
				"Error attaching disk",
				"Disk ID parse error: "+err.Error(),
			)
			return diags
		}

		params := oxide.InstanceDiskAttachParams{
			Instance: oxide.NameOrId(instanceID),
			Body: &oxide.DiskPath{
				Disk: oxide.NameOrId(diskID),
			},
		}
		_, err = client.InstanceDiskAttach(ctx, params)
		if err != nil {
			diags.AddError(
				"Error attaching disk",
				"API error: "+err.Error(),
			)
			// TODO: Should this return here or should I continue trying to attach the other disks?
			return diags
		}
		tflog.Trace(
			ctx,
			fmt.Sprintf("attached disk with ID: %v", v),
			map[string]any{"success": true},
		)
	}

	return nil
}

func detachDisks(
	ctx context.Context,
	client *oxide.Client,
	disks []attr.Value,
	instanceID string,
) diag.Diagnostics {
	var diags diag.Diagnostics

	for _, v := range disks {
		diskID, err := strconv.Unquote(v.String())
		if err != nil {
			diags.AddError(
				"Error detaching disk",
				"Disk ID parse error: "+err.Error(),
			)
			return diags
		}

		params := oxide.InstanceDiskDetachParams{
			Instance: oxide.NameOrId(instanceID),
			Body: &oxide.DiskPath{
				Disk: oxide.NameOrId(diskID),
			},
		}
		_, err = client.InstanceDiskDetach(ctx, params)
		if err != nil {
			diags.AddError(
				"Error detaching disk",
				"API error: "+err.Error(),
			)
			// TODO: Should this return here or should I continue trying to detach the other disks?
			return diags
		}
		tflog.Trace(
			ctx,
			fmt.Sprintf("detached disk with ID: %v", v),
			map[string]any{"success": true},
		)
	}

	return nil
}

func addAntiAffinityGroups(
	ctx context.Context,
	client *oxide.Client,
	groups []attr.Value,
	instanceID string,
) diag.Diagnostics {
	var diags diag.Diagnostics

	for _, v := range groups {
		id, err := strconv.Unquote(v.String())
		if err != nil {
			diags.AddError(
				"Error adding anti-affinity group to instance",
				"anti-affinity group ID parse error: "+err.Error(),
			)
			return diags
		}

		params := oxide.AntiAffinityGroupMemberInstanceAddParams{
			Instance:          oxide.NameOrId(instanceID),
			AntiAffinityGroup: oxide.NameOrId(id),
		}
		_, err = client.AntiAffinityGroupMemberInstanceAdd(ctx, params)
		if err != nil {
			diags.AddError(
				"Error adding anti-affinity group to instance",
				"API error: "+err.Error(),
			)
			return diags
		}
		tflog.Trace(
			ctx,
			fmt.Sprintf(
				"added anti-affinity group with ID: %v to instance with ID: %v",
				id,
				instanceID,
			),
			map[string]any{"success": true},
		)
	}

	return nil
}

func removeAntiAffinityGroups(
	ctx context.Context,
	client *oxide.Client,
	groups []attr.Value,
	instanceID string,
) diag.Diagnostics {
	var diags diag.Diagnostics

	for _, v := range groups {
		id, err := strconv.Unquote(v.String())
		if err != nil {
			diags.AddError(
				"Error removing anti-affinity group from instance",
				"anti-affinity group ID parse error: "+err.Error(),
			)
			return diags
		}

		params := oxide.AntiAffinityGroupMemberInstanceDeleteParams{
			Instance:          oxide.NameOrId(instanceID),
			AntiAffinityGroup: oxide.NameOrId(id),
		}
		err = client.AntiAffinityGroupMemberInstanceDelete(ctx, params)
		if err != nil {
			// If the anti-affinity group doesn't exist anymore, it means
			// the instance isn't part of it. We can just return.
			if is404(err) {
				return nil
			}
			diags.AddError(
				"Error removing anti-affinity group from instance",
				"API error: "+err.Error(),
			)
			return diags
		}
		tflog.Trace(
			ctx,
			fmt.Sprintf(
				"removed anti-affinity group with ID %v to instance with ID %v",
				id,
				instanceID,
			),
			map[string]any{"success": true},
		)
	}

	return nil
}

func attrValueSliceContains(s []attr.Value, str string) (bool, error) {
	for _, a := range s {
		v, err := strconv.Unquote(a.String())
		if err != nil {
			return false, err
		}
		if v == str {
			return true, nil
		}
	}
	return false, nil
}

var _ validator.String = ipConfigValidator{}

// ipConfigValidator validates that a string is a valid IP configuration.
type ipConfigValidator struct {
	ipVersion oxide.IpVersion
}

func (v ipConfigValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v ipConfigValidator) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("value must be a valid IP%s address or `auto`.", v.ipVersion)
}

func (v ipConfigValidator) ValidateString(
	_ context.Context,
	req validator.StringRequest,
	res *validator.StringResponse,
) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var valid bool
	value := req.ConfigValue.ValueString()

	switch v.ipVersion {
	case oxide.IpVersionV4:
		valid = (value == string(oxide.Ipv4AssignmentTypeAuto) ||
			isIPv4(req.ConfigValue.ValueString()))

	case oxide.IpVersionV6:
		valid = (value == string(oxide.Ipv6AssignmentTypeAuto) ||
			isIPv6(req.ConfigValue.ValueString()))
	}

	if !valid {
		res.Diagnostics.AddAttributeError(
			req.Path,
			fmt.Sprintf("Invalid IP%s configuration", v.ipVersion),
			fmt.Sprintf(`Attribute %s must be an IP%s address or "auto".`, req.Path, v.ipVersion),
		)
	}
}

// instanceIPConfigValidator is a custom validator that validates the ip_config
// attribute of network_interfaces.
type instanceIPConfigValidator struct{}

var _ validator.Object = instanceIPConfigValidator{}

func (v instanceIPConfigValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v instanceIPConfigValidator) MarkdownDescription(_ context.Context) string {
	return ""
}

func (f instanceIPConfigValidator) ValidateObject(
	ctx context.Context,
	req validator.ObjectRequest,
	resp *validator.ObjectResponse,
) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var ipConfig *instanceResourceIPConfigModel
	diags := req.ConfigValue.As(ctx, &ipConfig, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if ipConfig.V4 == nil && ipConfig.V6 == nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid network interface IP configuration",
			"At least one of v4 or v6 must be defined.",
		)
		return
	}
}

// Ensure the concrete validator satisfies the [validator.Set] interface.
var _ validator.Object = instanceExternalIPValidator{}

// instanceExternalIPValidator is a custom validator that validates the
// external_ips attribute on an oxide_instance resource.
type instanceExternalIPValidator struct{}

// Description returns the validation description in plain text formatting.
func (f instanceExternalIPValidator) Description(ctx context.Context) string {
	return f.MarkdownDescription(ctx)
}

// MarkdownDescription returns the validation description in Markdown formatting.
func (f instanceExternalIPValidator) MarkdownDescription(context.Context) string {
	return "cannot have more than one ephemeral external ip"
}

func (f instanceExternalIPValidator) ValidateObject(
	ctx context.Context,
	req validator.ObjectRequest,
	resp *validator.ObjectResponse,
) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var externalIPs *instanceResourceExternalIPModel
	diags := req.ConfigValue.As(ctx, &externalIPs, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if externalIPs.Empty() {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid external IP configuration",
			"At least one ephemeral or floating IP must be defined.",
		)
		return
	}
}

// ModifyPlanForHostnameDeprecation modifies the plan to support the
// deprecation of `host_name` in favor of `hostname`. This must be
// added to both the deprecated `host_name` attribute and the new
// `hostname` attribute. This function is responsible for setting
// [stringplanmodifier.RequiresReplaceIfFuncResponse.RequiresReplace] to `true`
// when it's deemed the user is actually changing the hostname value in the
// configuration or `false` when the user is just updating their configuration
// to comply with the deprecation.
func ModifyPlanForHostnameDeprecation(
	ctx context.Context,
	req planmodifier.StringRequest,
	resp *stringplanmodifier.RequiresReplaceIfFuncResponse,
) {
	// Check which attribute this modifier function is being called on as the logic
	// is vice versa for `host_name` and `hostname`.
	switch attribute := req.Path.String(); attribute {
	case "hostname":
		var hostnameDeprecated types.String
		diags := req.Config.GetAttribute(ctx, path.Root("host_name"), &hostnameDeprecated)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// The deprecated `host_name` attribute has a value. We do not need to replace
		// the resource just because the new `hostname` attribute doesn't have a value.
		if !hostnameDeprecated.IsNull() && !hostnameDeprecated.IsUnknown() {
			return
		}
	case "host_name":
		var hostname types.String
		diags := req.Plan.GetAttribute(ctx, path.Root("hostname"), &hostname)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// The new `hostname` attribute has a value. We do not need to replace the
		// resource just because the deprecated `host_name` attribute doesn't have
		// a value.
		if !hostname.IsUnknown() && !hostname.IsNull() {
			return
		}
	default:
		resp.Diagnostics.AddAttributeError(
			req.Path,
			fmt.Sprintf("Invalid plan modifier for attribute %s", attribute),
			"ModifyPlanForHostnameDeprecation can only be used for instance hostname attributes.",
		)
	}

	// If we've reached this point, it's because the actual value of either
	// `host_name` or `hostname` was modified, which must result in the resource
	// being replaced.
	resp.RequiresReplace = true
}

// instanceNetworkInterfacesPlanModifier is a plan modifier that detects
// changes to the network interfaces that Terraform can't know how to handle
// and modifies the plan to take them into  account.
type instanceNetworkInterfacesPlanModifier struct{}

var _ planmodifier.Set = instanceNetworkInterfacesPlanModifier{}

func (v instanceNetworkInterfacesPlanModifier) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v instanceNetworkInterfacesPlanModifier) MarkdownDescription(_ context.Context) string {
	return "instance network interface modified"
}

func (v instanceNetworkInterfacesPlanModifier) PlanModifySet(
	ctx context.Context,
	req planmodifier.SetRequest,
	resp *planmodifier.SetResponse,
) {
	var diags diag.Diagnostics

	var state []instanceResourceNICModel
	var plan []instanceResourceNICModel

	resp.Diagnostics.Append(req.StateValue.ElementsAs(ctx, &state, true)...)
	resp.Diagnostics.Append(req.PlanValue.ElementsAs(ctx, &plan, true)...)
	if resp.Diagnostics.HasError() {
		return
	}

	stateMap := make(map[string]instanceResourceNICModel)
	for _, nic := range state {
		stateMap[nic.ID.ValueString()] = nic
	}

	// Invalidate Computed attributes if the network interface hash changes
	// because it will be recreated on Update().
	for i, nic := range plan {
		stateNIC, ok := stateMap[nic.ID.ValueString()]
		if !ok {
			// Ignore network interface that are not in state since they are
			// new ones.
			continue
		}

		if nic.Hash() != stateNIC.Hash() {
			plan[i].ID = types.StringUnknown()
			plan[i].IPAddr = types.StringUnknown()
			plan[i].Primary = types.BoolUnknown()
			plan[i].TimeCreated = types.StringUnknown()
			plan[i].TimeModified = types.StringUnknown()
		}
	}

	resp.PlanValue, diags = types.SetValueFrom(ctx, instanceResourceNICType, plan)
	resp.Diagnostics.Append(diags...)
}

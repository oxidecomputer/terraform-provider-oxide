// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	AntiAffinityGroups types.Set                         `tfsdk:"anti_affinity_groups"`
	AutoRestartPolicy  types.String                      `tfsdk:"auto_restart_policy"`
	BootDiskID         types.String                      `tfsdk:"boot_disk_id"`
	Description        types.String                      `tfsdk:"description"`
	DiskAttachments    types.Set                         `tfsdk:"disk_attachments"`
	ExternalIPs        []instanceResourceExternalIPModel `tfsdk:"external_ips"`
	HostnameDeprecated types.String                      `tfsdk:"host_name"`
	Hostname           types.String                      `tfsdk:"hostname"`
	ID                 types.String                      `tfsdk:"id"`
	Memory             types.Int64                       `tfsdk:"memory"`
	Name               types.String                      `tfsdk:"name"`
	NetworkInterfaces  []instanceResourceNICModel        `tfsdk:"network_interfaces"`
	NCPUs              types.Int64                       `tfsdk:"ncpus"`
	ProjectID          types.String                      `tfsdk:"project_id"`
	SSHPublicKeys      types.Set                         `tfsdk:"ssh_public_keys"`
	StartOnCreate      types.Bool                        `tfsdk:"start_on_create"`
	TimeCreated        types.String                      `tfsdk:"time_created"`
	TimeModified       types.String                      `tfsdk:"time_modified"`
	Timeouts           timeouts.Value                    `tfsdk:"timeouts"`
	UserData           types.String                      `tfsdk:"user_data"`
}

type instanceResourceNICModel struct {
	Description  types.String                  `tfsdk:"description"`
	ID           types.String                  `tfsdk:"id"`
	IPAddr       types.String                  `tfsdk:"ip_address"`
	IPStack      *instanceResourceIPStackModel `tfsdk:"ip_stack"`
	MAC          types.String                  `tfsdk:"mac_address"`
	Name         types.String                  `tfsdk:"name"`
	Primary      types.Bool                    `tfsdk:"primary"`
	SubnetID     types.String                  `tfsdk:"subnet_id"`
	TimeCreated  types.String                  `tfsdk:"time_created"`
	TimeModified types.String                  `tfsdk:"time_modified"`
	VPCID        types.String                  `tfsdk:"vpc_id"`
}

type instanceResourceIPStackModel struct {
	Type types.String                    `tfsdk:"type"`
	V4   *instanceResourceIPStackV4Model `tfsdk:"v4"`
	V6   *instanceResourceIPStackV6Model `tfsdk:"v6"`
}

type instanceResourceIPStackV4Model struct {
	IP           types.String   `tfsdk:"ip"`
	IPAssignment types.String   `tfsdk:"ip_assignment"`
	TransitIPs   []types.String `tfsdk:"transit_ips"`
}

type instanceResourceIPStackV6Model struct {
	IP           types.String   `tfsdk:"ip"`
	IPAssignment types.String   `tfsdk:"ip_assignment"`
	TransitIPs   []types.String `tfsdk:"transit_ips"`
}

type instanceResourceExternalIPModel struct {
	ID   types.String `tfsdk:"id"`
	Type types.String `tfsdk:"type"`
}

// Metadata returns the resource type name.
func (r *instanceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "oxide_instance"
}

// Configure adds the provider configured client to the data source.
func (r *instanceResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*oxide.Client)
}

// ImportState imports an existing instance resource into Terraform state.
func (r *instanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Schema defines the schema for the resource.
func (r *instanceResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"network_interfaces": schema.SetNestedAttribute{
				Optional:    true,
				Description: "Network interface devices attached to the instance.",
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
						"ip_stack": schema.SingleNestedAttribute{
							Optional:    true,
							Description: "IP stack for the instance network interface. Defaults to dual stack with auto-assigned addresses if not provided.",
							Attributes: map[string]schema.Attribute{
								"type": schema.StringAttribute{
									Computed:    true,
									Description: "The IP stack type.",
								},
								"v4": schema.SingleNestedAttribute{
									Optional:    true,
									Description: "Creates an IPv4 stack for the instance network interface.",
									Attributes: map[string]schema.Attribute{
										"ip_assignment": schema.StringAttribute{
											Required:    true,
											Description: `IPv4 address for the instance network interface or "auto" to auto-assigned one.`,
											Validators: []validator.String{
												ipAssignmentValidator{oxide.IpVersionV4},
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.RequiresReplaceIfConfigured(),
											},
										},
										"transit_ips": schema.ListAttribute{
											Optional:    true,
											Description: "Additional networks on which the interface can send / receive traffic.",
											ElementType: types.StringType,
										},
										"ip": schema.StringAttribute{
											Computed:    true,
											Description: "IPv4 address of the instance.",
										},
									},
								},
								"v6": schema.SingleNestedAttribute{
									Optional: true,
									Attributes: map[string]schema.Attribute{
										"ip_assignment": schema.StringAttribute{
											Required:    true,
											Description: `IPv6 address for the instance network interface or "auto" to auto-assigned one.`,
											Validators: []validator.String{
												ipAssignmentValidator{oxide.IpVersionV6},
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.RequiresReplaceIfConfigured(),
											},
										},
										"transit_ips": schema.ListAttribute{
											Optional:    true,
											Description: "Additional networks on which the interface can send / receive traffic.",
											ElementType: types.StringType,
										},
										"ip": schema.StringAttribute{
											Computed:    true,
											Description: "IPv6 address of the instance.",
										},
									},
								},
							},
						},
						"ip_address": schema.StringAttribute{
							Optional:           true,
							Computed:           true,
							DeprecationMessage: "Use ip_stack instead. This attribute will be removed in the next minor version of the provider.",
							Description: "IP address for the instance network interface. " +
								"One will be auto-assigned if not provided.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplaceIfConfigured(),
							},
						},
						"mac_address": schema.StringAttribute{
							Computed:    true,
							Description: "MAC address assigned to the instance network interface.",
						},
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "Unique, immutable, system-controlled identifier of the instance network interface.",
						},
						"primary": schema.BoolAttribute{
							Computed:    true,
							Description: "True if this is the primary network interface for the instance to which it's attached to.",
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
			"external_ips": schema.SetNestedAttribute{
				Optional:    true,
				Description: "External IP addresses provided to this instance.",
				Validators: []validator.Set{
					instanceExternalIPValidator{},
					setvalidator.AlsoRequires(path.MatchRoot("network_interfaces")),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						// The id attribute is optional, computed, and has a default to account for the
						// case where an instance created with an external IP using the default IP pool
						// (i.e., id = null) would drift when read (e.g., id = "") and require updating
						// in place.
						"id": schema.StringAttribute{
							MarkdownDescription: "If `type` is `ephemeral`, ID of the IP pool to retrieve addresses from, or all available pools if not specified. If `type` is `floating`, ID of the floating IP.",
							Optional:            true,
							Computed:            true,
							Default:             stringdefault.StaticString(""),
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "Type of external IP. Must be one of `ephemeral` or `floating`.",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.OneOf(
									string(oxide.ExternalIpCreateTypeEphemeral),
									string(oxide.ExternalIpCreateTypeFloating),
								),
							},
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

// UpgradeState upgrades the Terraform state for the oxide_instance
// resource from a previous schema version to the current version.
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
						MAC:          oldNIC.MAC,
						Name:         oldNIC.Name,
						Primary:      oldNIC.Primary,
						SubnetID:     oldNIC.SubnetID,
						TimeCreated:  oldNIC.TimeCreated,
						TimeModified: oldNIC.TimeModified,
						VPCID:        oldNIC.VPCID,

						// TODO: remove once ip_address is fully deprecated.
						IPAddr: oldNIC.IPAddr,
					}

					// Convert ip_address to ip_stack structure if there's an IP address
					if ipAddr := oldNIC.IPAddr.ValueString(); ipAddr != "" {
						newNIC.IPStack = &instanceResourceIPStackModel{
							Type: types.StringValue("v4"),
							V4: &instanceResourceIPStackV4Model{
								IP:           oldNIC.IPAddr,
								IPAssignment: types.StringValue(ipAddr),
								TransitIPs:   nil,
							},
						}
					}

					newNICs = append(newNICs, newNIC)
				}

				// Migrate external IPs.
				var newExtIPs []instanceResourceExternalIPModel
				for _, oldExtIP := range oldState.ExternalIPs {
					newExtIPs = append(newExtIPs, instanceResourceExternalIPModel(oldExtIP))
				}

				newState := instanceResourceModel{
					AntiAffinityGroups: oldState.AntiAffinityGroups,
					AutoRestartPolicy:  oldState.AutoRestartPolicy,
					BootDiskID:         oldState.BootDiskID,
					Description:        oldState.Description,
					DiskAttachments:    oldState.DiskAttachments,
					ExternalIPs:        newExtIPs,
					HostnameDeprecated: oldState.HostnameDeprecated,
					Hostname:           oldState.Hostname,
					ID:                 oldState.ID,
					Memory:             oldState.Memory,
					Name:               oldState.Name,
					NetworkInterfaces:  newNICs,
					NCPUs:              oldState.NCPUs,
					ProjectID:          oldState.ProjectID,
					SSHPublicKeys:      oldState.SSHPublicKeys,
					StartOnCreate:      oldState.StartOnCreate,
					TimeCreated:        oldState.TimeCreated,
					TimeModified:       oldState.TimeModified,
					Timeouts:           oldState.Timeouts,
					UserData:           oldState.UserData,
				}

				resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *instanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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
		params.Body.AutoRestartPolicy = oxide.InstanceAutoRestartPolicy(plan.AutoRestartPolicy.ValueString())
	}

	// Add boot disk if any.
	if !plan.BootDiskID.IsNull() {
		// Validate whether the boot disk ID is included in `attachments`
		// This is necessary as the response from InstanceDiskList includes
		// the boot disk and would result in an inconsistent state in terraform
		isBootIDPresent, err := attrValueSliceContains(plan.DiskAttachments.Elements(), plan.BootDiskID.ValueString())
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

	// The control plane API counts the BootDisk and the Disk attachments when it calculates the limit on disk attachments.
	// If bootdisk is set explicitly, we don't want it to be in the API call, but we need it in the state entry.
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

	tflog.Trace(ctx, fmt.Sprintf("created instance with ID: %v", instance.Id), map[string]any{"success": true})

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(instance.Id)
	plan.Hostname = types.StringValue(instance.Hostname)
	plan.HostnameDeprecated = types.StringValue(instance.Hostname)
	plan.TimeCreated = types.StringValue(instance.TimeCreated.String())
	plan.TimeModified = types.StringValue(instance.TimeModified.String())

	// Populate NIC information
	for i := range plan.NetworkInterfaces {
		params := oxide.InstanceNetworkInterfaceViewParams{
			Interface: oxide.NameOrId(plan.NetworkInterfaces[i].Name.ValueString()),
			Instance:  oxide.NameOrId(instance.Id),
		}
		nic, err := r.client.InstanceNetworkInterfaceView(ctx, params)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to read instance network interface:",
				"API error: "+err.Error(),
			)
			// Don't return here as the instance has already been created.
			// Otherwise the state won't be saved.
			continue
		}
		tflog.Trace(ctx, fmt.Sprintf("read instance network interface with ID: %v", nic.Id), map[string]any{"success": true})

		// Map response body to schema and populate Computed attribute values
		plan.NetworkInterfaces[i].ID = types.StringValue(nic.Id)
		plan.NetworkInterfaces[i].TimeCreated = types.StringValue(nic.TimeCreated.String())
		plan.NetworkInterfaces[i].TimeModified = types.StringValue(nic.TimeModified.String())
		plan.NetworkInterfaces[i].MAC = types.StringValue(string(nic.Mac))
		plan.NetworkInterfaces[i].Primary = types.BoolPointerValue(nic.Primary)

		// Handle Computed attributes for the IP stack.
		planIPStack := plan.NetworkInterfaces[i].IPStack
		remoteIPStack, err := instanceIPStackToModel(nic.IpStack)
		if err != nil {
			diags.AddError(
				"Unable to parse network interface IP stack:",
				fmt.Sprintf("Invalid IP stack for network interface %d: %v", i, err.Error()),
			)

			// Don't return here as the instance has already been created.
			// Otherwise the state won't be saved.
			continue
		}

		if planIPStack != nil && remoteIPStack != nil {
			planIPStack.Type = remoteIPStack.Type

			if planIPStack.V4 != nil && remoteIPStack.V4 != nil {
				planIPStack.V4.IP = remoteIPStack.V4.IP

				// TODO: remove once ip_address is fully deprecated.
				plan.NetworkInterfaces[i].IPAddr = remoteIPStack.V4.IP
			}

			if planIPStack.V6 != nil && remoteIPStack.V6 != nil {
				planIPStack.V6.IP = remoteIPStack.V6.IP
			}
		}
	}

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *instanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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

	tflog.Trace(ctx, fmt.Sprintf("read instance with ID: %v", instance.Id), map[string]any{"success": true})

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
	if len(externalIPs) > 0 {
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

	antiAffinityGroupSet, diags := newAssociatedAntiAffinityGroupsOnCreateSet(ctx, r.client, state.ID.ValueString())
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

	nicSet, diags := newAttachedNetworkInterfacesModel(ctx, r.client, state)
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
func (r *instanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
	tflog.Trace(ctx, fmt.Sprintf("stopped instance with ID: %v", state.ID.ValueString()), map[string]any{"success": true})

	// Update external IPs.
	// We detach external IPs first to account for the case where an ephemeral
	// external IP's IP Pool is modified. This is because there can only be one
	// ephemeral external IP attached to an instance at a given time and the
	// last detachment/attachment wins.
	{
		externalIPsToDetach := sliceDiff(state.ExternalIPs, plan.ExternalIPs)
		resp.Diagnostics.Append(detachExternalIPs(ctx, r.client, externalIPsToDetach, state.ID.ValueString())...)
		if resp.Diagnostics.HasError() {
			return
		}

		externalIPsToAttach := sliceDiff(plan.ExternalIPs, state.ExternalIPs)
		resp.Diagnostics.Append(attachExternalIPs(ctx, r.client, externalIPsToAttach, state.ID.ValueString())...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

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
			params.Body.AutoRestartPolicy = (*oxide.InstanceAutoRestartPolicy)(plan.AutoRestartPolicy.ValueStringPointer())
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
			"updated boot disk forinstance with ID: %v", instance.Id), map[string]any{"success": true},
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

	// Update NICs
	planNICs := plan.NetworkInterfaces
	stateNICs := state.NetworkInterfaces

	// Check plan and if it has an ID that the state doesn't then attach it
	nicsToCreate := sliceDiffByID(
		planNICs, stateNICs,
		func(e instanceResourceNICModel) any {
			return e.ID.ValueString()
		},
	)
	resp.Diagnostics.Append(createNICs(ctx, r.client, nicsToCreate, state.ID.ValueString())...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check state and if it has an ID that the plan doesn't then delete it
	nicsToDelete := sliceDiffByID(
		stateNICs, planNICs,
		func(e instanceResourceNICModel) any {
			return e.ID.ValueString()
		},
	)
	resp.Diagnostics.Append(deleteNICs(ctx, r.client, nicsToDelete)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update anti-affinity groups
	planAntiAffinityGroups := plan.AntiAffinityGroups.Elements()
	stateAntiAffinityGroups := state.AntiAffinityGroups.Elements()

	// Check plan and if it has an ID that the state doesn't then add it
	antiAffinityGroupsToAdd := sliceDiff(planAntiAffinityGroups, stateAntiAffinityGroups)
	resp.Diagnostics.Append(addAntiAffinityGroups(ctx, r.client, antiAffinityGroupsToAdd, state.ID.ValueString())...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check state and if it has an ID that the plan doesn't then remove it
	antiAffinityGroupsToRemove := sliceDiff(stateAntiAffinityGroups, planAntiAffinityGroups)
	resp.Diagnostics.Append(removeAntiAffinityGroups(ctx, r.client, antiAffinityGroupsToRemove, state.ID.ValueString())...)
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

	tflog.Trace(ctx, fmt.Sprintf("read instance with ID: %v", instance.Id), map[string]any{"success": true})

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
	if len(externalIPs) > 0 {
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
	nicModel, diags := newAttachedNetworkInterfacesModel(ctx, r.client, state)
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
func (r *instanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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
	tflog.Trace(ctx, fmt.Sprintf("stopped instance with ID: %v", state.ID.ValueString()), map[string]any{"success": true})

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
	tflog.Trace(ctx, fmt.Sprintf("deleted instance with ID: %v", state.ID.ValueString()), map[string]any{"success": true})
}

func waitForInstanceStop(ctx context.Context, client *oxide.Client, timeout time.Duration, instanceID string) diag.Diagnostics {
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
					return nil, "nil", fmt.Errorf("while polling for the status of instance %v: %v", instanceID, err)
				}
				return instance, "", nil
			}
			tflog.Trace(ctx, fmt.Sprintf("read instance with ID: %v", instanceID), map[string]any{"success": true})
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

func newAttachedDisksSet(ctx context.Context, client *oxide.Client, instanceID string) (types.Set, diag.Diagnostics) {
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

func newAssociatedSSHKeysOnCreateSet(ctx context.Context, client *oxide.Client, instanceID string) (types.Set, diag.Diagnostics) {
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

func newAssociatedAntiAffinityGroupsOnCreateSet(ctx context.Context, client *oxide.Client, instanceID string) (types.Set, diag.Diagnostics) {
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

func newNetworkInterfaceAttachment(ctx context.Context, client *oxide.Client, model []instanceResourceNICModel) (
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
			IpConfig:    newPrivateIpStackCreate(planNIC),
			SubnetName:  oxide.Name(names.subnet),
			VpcName:     oxide.Name(names.vpc),
		}
		nicParams = append(nicParams, nic)
	}

	nicAttachment := oxide.InstanceNetworkInterfaceAttachment{
		Type:   oxide.InstanceNetworkInterfaceAttachmentTypeCreate,
		Params: nicParams,
	}
	return nicAttachment, nil
}

func newPrivateIpStackCreate(nic instanceResourceNICModel) oxide.PrivateIpStackCreate {
	// Falback to previous behaviour of creating an IPv4 stack using
	// ip_address if ip_stack is not defined.
	// TODO: remove once ip_address is fully deprecated.
	if nic.IPStack == nil {
		return oxide.PrivateIpStackCreate{
			Type: oxide.PrivateIpStackCreateTypeV4,
			Value: oxide.PrivateIpv4StackCreate{
				Ip: oxide.Ipv4Assignment{
					Type:  oxide.Ipv4AssignmentTypeExplicit,
					Value: nic.IPAddr.ValueString(),
				},
			},
		}
	}

	if nic.IPStack.V4 != nil && nic.IPStack.V6 != nil {
		return oxide.PrivateIpStackCreate{
			Type:  oxide.PrivateIpStackCreateTypeDualStack,
			Value: newPrivateIpStackCreateDualStack(nic.IPStack.V4, nic.IPStack.V6),
		}

	}

	if nic.IPStack.V4 != nil {
		return oxide.PrivateIpStackCreate{
			Type:  oxide.PrivateIpStackCreateTypeV4,
			Value: newPrivateIpStackCreateV4(nic.IPStack.V4),
		}

	}

	return oxide.PrivateIpStackCreate{
		Type:  oxide.PrivateIpStackCreateTypeV6,
		Value: newPrivateIpStackCreateV6(nic.IPStack.V6),
	}
}

func newPrivateIpStackCreateV4(stack *instanceResourceIPStackV4Model) oxide.PrivateIpv4StackCreate {
	var stackCreate oxide.PrivateIpv4StackCreate

	ip := stack.IPAssignment.ValueString()
	if ip == string(oxide.Ipv4AssignmentTypeAuto) {
		stackCreate.Ip = oxide.Ipv4Assignment{
			Type: oxide.Ipv4AssignmentTypeAuto,
		}
	} else {
		stackCreate.Ip = oxide.Ipv4Assignment{
			Type:  oxide.Ipv4AssignmentTypeExplicit,
			Value: ip,
		}
	}

	for _, ip := range stack.TransitIPs {
		stackCreate.TransitIps = append(stackCreate.TransitIps, oxide.Ipv4Net(ip.ValueString()))
	}

	return stackCreate
}

func newPrivateIpStackCreateV6(stack *instanceResourceIPStackV6Model) oxide.PrivateIpv6StackCreate {
	var stackCreate oxide.PrivateIpv6StackCreate

	ip := stack.IPAssignment.ValueString()
	if ip == string(oxide.Ipv6AssignmentTypeAuto) {
		stackCreate.Ip = oxide.Ipv6Assignment{
			Type: oxide.Ipv6AssignmentTypeAuto,
		}
	} else {
		stackCreate.Ip = oxide.Ipv6Assignment{
			Type:  oxide.Ipv6AssignmentTypeExplicit,
			Value: ip,
		}
	}

	for _, ip := range stack.TransitIPs {
		stackCreate.TransitIps = append(stackCreate.TransitIps, oxide.Ipv6Net(ip.ValueString()))
	}

	return stackCreate
}

func newPrivateIpStackCreateDualStack(
	stackV4 *instanceResourceIPStackV4Model,
	stackV6 *instanceResourceIPStackV6Model,
) oxide.PrivateIpStackCreateValue {
	return oxide.PrivateIpStackCreateValue{
		V4: newPrivateIpStackCreateV4(stackV4),
		V6: newPrivateIpStackCreateV6(stackV6),
	}
}

func newAttachedNetworkInterfacesModel(ctx context.Context, client *oxide.Client, state instanceResourceModel) (
	[]instanceResourceNICModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Store existing network interfaces in a map for quick retrieval by ID.
	stateIPStacks := make(map[string]*instanceResourceIPStackModel)
	for _, nic := range state.NetworkInterfaces {
		if nic.IPStack != nil {
			stateIPStacks[nic.ID.ValueString()] = nic.IPStack
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
		return []instanceResourceNICModel{}, diags
	}

	nicSet := []instanceResourceNICModel{}
	for _, nic := range nics.Items {
		n := instanceResourceNICModel{
			Description:  types.StringValue(nic.Description),
			ID:           types.StringValue(nic.Id),
			MAC:          types.StringValue(string(nic.Mac)),
			Name:         types.StringValue(string(nic.Name)),
			Primary:      types.BoolPointerValue(nic.Primary),
			SubnetID:     types.StringValue(nic.SubnetId),
			TimeCreated:  types.StringValue(nic.TimeCreated.String()),
			TimeModified: types.StringValue(nic.TimeModified.String()),
			VPCID:        types.StringValue(nic.VpcId),
		}

		// Build network interface IP stack.
		ipStack, err := instanceIPStackToModel(nic.IpStack)
		if err != nil {
			diags.AddError(
				"Unable to parse network interface IP stack:",
				fmt.Sprintf("Invalid IP stack for network interface %s: %v", n.ID, err.Error()),
			)
			return nil, diags
		}

		if ipStack != nil {
			// Read ip_assignment from state if present to avoid drifts. This
			// value is only used when creating the instance so it is not
			// returned by the API.
			if stateIPStack, ok := stateIPStacks[nic.Id]; ok {
				if ipStack.V4 != nil && stateIPStack.V4 != nil {
					ipStack.V4.IPAssignment = stateIPStack.V4.IPAssignment
				}
				if ipStack.V6 != nil && stateIPStack.V6 != nil {
					ipStack.V6.IPAssignment = stateIPStack.V6.IPAssignment
				}
			}

			// TODO: remove once ip_address is fully deprecated.
			if ipStack.V4 != nil {
				n.IPAddr = ipStack.V4.IP
			}

			n.IPStack = ipStack
		}

		nicSet = append(nicSet, n)
	}

	return nicSet, nil
}

// instanceIPStackToModel converts an instance network interface IP stack as
// returned by the Oxide API to an internal provider model representation.
//
// The returned model will only be populated with values that exist in the
// remote API. Values that only exist in the resource schema may need to be
// filled from existing state.
func instanceIPStackToModel(stack oxide.PrivateIpStack) (*instanceResourceIPStackModel, error) {
	switch stackType := stack.Type; stackType {
	case oxide.PrivateIpStackTypeV4:
		var parsedStack oxide.PrivateIpv4Stack
		if err := parseInstanceIPStack(stack, &parsedStack); err != nil {
			return nil, err
		}

		var transitIPs []types.String
		for _, ip := range parsedStack.TransitIps {
			transitIPs = append(transitIPs, types.StringValue(string(ip)))
		}

		return &instanceResourceIPStackModel{
			Type: types.StringValue(string(stackType)),
			V4: &instanceResourceIPStackV4Model{
				IP:         types.StringValue(parsedStack.Ip),
				TransitIPs: transitIPs,
			},
		}, nil

	case oxide.PrivateIpStackTypeV6:
		var parsedStack oxide.PrivateIpv6Stack
		if err := parseInstanceIPStack(stack, &parsedStack); err != nil {
			return nil, err
		}

		var transitIPs []types.String
		for _, ip := range parsedStack.TransitIps {
			transitIPs = append(transitIPs, types.StringValue(string(ip)))
		}

		return &instanceResourceIPStackModel{
			Type: types.StringValue(string(stackType)),
			V6: &instanceResourceIPStackV6Model{
				IP:         types.StringValue(parsedStack.Ip),
				TransitIPs: transitIPs,
			},
		}, nil

	case oxide.PrivateIpStackTypeDualStack:
		var parsedStack oxide.PrivateIpStackValue
		if err := parseInstanceIPStack(stack, &parsedStack); err != nil {
			return nil, err
		}

		v4 := &instanceResourceIPStackV4Model{
			IP: types.StringValue(parsedStack.V4.Ip),
		}
		for _, ip := range parsedStack.V4.TransitIps {
			v4.TransitIPs = append(v4.TransitIPs, types.StringValue(string(ip)))
		}

		v6 := &instanceResourceIPStackV6Model{
			IP: types.StringValue(parsedStack.V6.Ip),
		}
		for _, ip := range parsedStack.V6.TransitIps {
			v6.TransitIPs = append(v6.TransitIPs, types.StringValue(string(ip)))
		}

		return &instanceResourceIPStackModel{
			Type: types.StringValue(string(stackType)),
			V4:   v4,
			V6:   v6,
		}, nil

	default:
		return nil, fmt.Errorf("unexpected IP stack type %s", stackType)
	}
}

func parseInstanceIPStack(stack oxide.PrivateIpStack, val any) error {
	// stack.Value is an any, so marshal it into bytes so we can unmarshal into
	// the right struct depending on its type.
	marshalled, err := json.Marshal(stack.Value)
	if err != nil {
		return fmt.Errorf("failed to marshal IP stack value: %w", err)
	}

	if err := json.Unmarshal(marshalled, val); err != nil {
		return fmt.Errorf("failed to unmarshal IP stack value: %w", err)
	}

	return nil
}

// newAttachedExternalIPModel fetches the external IP addresses for the instance
// specified by model, keeping the IP pool ID from the ephemeral external IP
// from the model if one is present.
func newAttachedExternalIPModel(ctx context.Context, client *oxide.Client, model instanceResourceModel) (
	[]instanceResourceExternalIPModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	externalIPs := make([]instanceResourceExternalIPModel, 0)

	// The [oxide.Client.InstanceExternalIpList] method does not return the IP pool
	// ID for ephemeral external IPs. See https://github.com/oxidecomputer/omicron/issues/6825.
	//
	// Pull the IP pool ID out of the model to populate the external IP model that
	// we're building to prevent erroneous Terraform diffs (e.g., state contains
	// IP pool ID and refresh doesn't). It's safe to stop at the first ephemeral
	// external IP encountered since the configuration enforces at most one
	// ephemeral external IP.
	ephemeralIPPoolID := ""
	for _, externalIP := range model.ExternalIPs {
		if externalIP.Type.ValueString() == string(oxide.ExternalIpCreateTypeEphemeral) {
			ephemeralIPPoolID = externalIP.ID.ValueString()
			break
		}
	}

	externalIPResponse, err := client.InstanceExternalIpList(ctx, oxide.InstanceExternalIpListParams{
		Instance: oxide.NameOrId(model.ID.ValueString()),
	})
	if err != nil {
		diags.AddError(
			"Unable to list instance external ips:",
			"API error: "+err.Error(),
		)
		return nil, diags
	}

	for _, externalIP := range externalIPResponse.Items {
		switch externalIP.Kind {
		case oxide.ExternalIpKindEphemeral:
			externalIPs = append(externalIPs, instanceResourceExternalIPModel{
				ID:   types.StringValue(ephemeralIPPoolID),
				Type: types.StringValue(string(externalIP.Kind)),
			})
		case oxide.ExternalIpKindFloating:
			externalIPs = append(externalIPs, instanceResourceExternalIPModel{
				ID:   types.StringValue(externalIP.Id),
				Type: types.StringValue(string(externalIP.Kind)),
			})
		// Skipped until the schema is updated to support SNAT external IPs.
		case oxide.ExternalIpKindSnat:
			continue
		default:
			diags.AddError(
				"Invalid external IP kind:",
				fmt.Sprintf("Encountered unexpected external IP kind: %s", externalIP.Kind),
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

func newDiskAttachmentsOnCreate(ctx context.Context, client *oxide.Client, diskIDs types.Set) ([]oxide.InstanceDiskAttachment, diag.Diagnostics) {
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

func filterBootDiskFromDisks(disks []oxide.InstanceDiskAttachment, boot_disk *oxide.InstanceDiskAttachment) []oxide.InstanceDiskAttachment {
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

func newExternalIPsOnCreate(externalIPs []instanceResourceExternalIPModel) []oxide.ExternalIpCreate {
	var ips []oxide.ExternalIpCreate

	for _, ip := range externalIPs {
		var eIP oxide.ExternalIpCreate

		if ip.Type.ValueString() == string(oxide.ExternalIpCreateTypeEphemeral) {
			if ip.ID.ValueString() != "" {
				eIP.Pool = oxide.NameOrId(ip.ID.ValueString())
			}
			eIP.Type = oxide.ExternalIpCreateType(ip.Type.ValueString())
		}

		if ip.Type.ValueString() == string(oxide.ExternalIpCreateTypeFloating) {
			eIP.FloatingIp = oxide.NameOrId(ip.ID.ValueString())
			eIP.Type = oxide.ExternalIpCreateType(ip.Type.ValueString())
		}

		ips = append(ips, eIP)
	}

	return ips
}

func createNICs(ctx context.Context, client *oxide.Client, models []instanceResourceNICModel, instanceID string) diag.Diagnostics {
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
				IpConfig:    newPrivateIpStackCreate(model),
				SubnetName:  oxide.Name(names.subnet),
				VpcName:     oxide.Name(names.vpc),
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

func deleteNICs(ctx context.Context, client *oxide.Client, models []instanceResourceNICModel) diag.Diagnostics {
	var diags diag.Diagnostics

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
		tflog.Trace(ctx, fmt.Sprintf("deleted instance network interface with ID: %v", model.ID.ValueString()),
			map[string]any{"success": true})
	}

	return nil
}

// attachExternalIPs attaches the external IPs specified by externalIPs to the
// instance specified by instanceID.
func attachExternalIPs(ctx context.Context, client *oxide.Client, externalIPs []instanceResourceExternalIPModel, instanceID string) diag.Diagnostics {
	var diags diag.Diagnostics

	for _, v := range externalIPs {
		externalIPID := v.ID
		externalIPType := v.Type

		switch oxide.ExternalIpKind(externalIPType.ValueString()) {
		case oxide.ExternalIpKindEphemeral:
			params := oxide.InstanceEphemeralIpAttachParams{
				Instance: oxide.NameOrId(instanceID),
				Body: &oxide.EphemeralIpCreate{
					Pool: oxide.NameOrId(externalIPID.ValueString()),
				},
			}

			if _, err := client.InstanceEphemeralIpAttach(ctx, params); err != nil {
				diags.AddError(
					fmt.Sprintf("Error attaching ephemeral external IP with ID %s", externalIPID.ValueString()),
					"API error: "+err.Error(),
				)

				return diags
			}

		case oxide.ExternalIpKindFloating:
			params := oxide.FloatingIpAttachParams{
				FloatingIp: oxide.NameOrId(externalIPID.ValueString()),
				Body: &oxide.FloatingIpAttach{
					Kind:   oxide.FloatingIpParentKindInstance,
					Parent: oxide.NameOrId(instanceID),
				},
			}

			if _, err := client.FloatingIpAttach(ctx, params); err != nil {
				diags.AddError(
					fmt.Sprintf("Error attaching floating external IP with ID %s", externalIPID.ValueString()),
					"API error: "+err.Error(),
				)

				return diags
			}
		default:
			diags.AddError(
				fmt.Sprintf("Cannot attach invalid external IP type %q", externalIPType.ValueString()),
				fmt.Sprintf("The external IP type must be one of: %q, %q", oxide.ExternalIpCreateTypeEphemeral, oxide.ExternalIpCreateTypeFloating),
			)
			return diags
		}

		tflog.Trace(ctx, fmt.Sprintf("successfully attached %s external IP with ID %s", externalIPType.ValueString(), externalIPID.ValueString()), map[string]any{"success": true})
	}

	return nil
}

// detachExternalIPs detaches the external IPs specified by externalIPs from the
// instance specified by instanceID.
func detachExternalIPs(ctx context.Context, client *oxide.Client, externalIPs []instanceResourceExternalIPModel, instanceID string) diag.Diagnostics {
	var diags diag.Diagnostics

	for _, v := range externalIPs {
		externalIPID := v.ID
		externalIPType := v.Type

		switch oxide.ExternalIpKind(externalIPType.ValueString()) {
		case oxide.ExternalIpKindEphemeral:
			params := oxide.InstanceEphemeralIpDetachParams{
				Instance: oxide.NameOrId(instanceID),
			}

			if err := client.InstanceEphemeralIpDetach(ctx, params); err != nil {
				diags.AddError(
					fmt.Sprintf("Error detaching ephemeral external IP with ID %s", externalIPID.ValueString()),
					"API error: "+err.Error(),
				)

				return diags
			}

		case oxide.ExternalIpKindFloating:
			params := oxide.FloatingIpDetachParams{
				FloatingIp: oxide.NameOrId(externalIPID.ValueString()),
			}

			if _, err := client.FloatingIpDetach(ctx, params); err != nil {
				diags.AddError(
					fmt.Sprintf("Error detaching floating external IP with ID %s", externalIPID.ValueString()),
					"API error: "+err.Error(),
				)

				return diags
			}
		// It's not possible to detach an SNAT external IP. Skip it.
		case oxide.ExternalIpKindSnat:
			continue
		default:
			diags.AddError(
				fmt.Sprintf("Cannot detach invalid external IP type %q", externalIPType.ValueString()),
				fmt.Sprintf("The external IP type must be one of: %q, %q", oxide.ExternalIpCreateTypeEphemeral, oxide.ExternalIpCreateTypeFloating),
			)
			return diags
		}

		tflog.Trace(ctx, fmt.Sprintf("successfully detached %s external IP with ID %s", externalIPType.ValueString(), externalIPID.ValueString()), map[string]any{"success": true})
	}

	return nil
}

func attachDisks(ctx context.Context, client *oxide.Client, disks []attr.Value, instanceID string) diag.Diagnostics {
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
		tflog.Trace(ctx, fmt.Sprintf("attached disk with ID: %v", v), map[string]any{"success": true})
	}

	return nil
}

func detachDisks(ctx context.Context, client *oxide.Client, disks []attr.Value, instanceID string) diag.Diagnostics {
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
		tflog.Trace(ctx, fmt.Sprintf("detached disk with ID: %v", v), map[string]any{"success": true})
	}

	return nil
}

func addAntiAffinityGroups(
	ctx context.Context, client *oxide.Client, groups []attr.Value, instanceID string) diag.Diagnostics {
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
		tflog.Trace(ctx, fmt.Sprintf("added anti-affinity group with ID: %v to instance with ID: %v", id, instanceID), map[string]any{"success": true})
	}

	return nil
}

func removeAntiAffinityGroups(
	ctx context.Context, client *oxide.Client, groups []attr.Value, instanceID string) diag.Diagnostics {
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
		tflog.Trace(ctx, fmt.Sprintf("removed anti-affinity group with ID %v to instance with ID %v", id, instanceID), map[string]any{"success": true})
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

// Ensure the concrete validator satisfies the [validator.Set] interface.
var _ validator.Set = instanceExternalIPValidator{}

// instanceExternalIPValidator is a custom validator that validates the
// external_ips attribute on an oxide_instance resource.
type instanceExternalIPValidator struct{}

// Description returns the validation description in plain text formatting.
func (f instanceExternalIPValidator) Description(context.Context) string {
	return "cannot have more than one ephemeral external ip"
}

// MarkdownDescription returns the validation description in Markdown formatting.
func (f instanceExternalIPValidator) MarkdownDescription(context.Context) string {
	return "cannot have more than one ephemeral external ip"
}

// ValidateSet validates whether a set of [instanceResourceExternalIPModel]
// objects has at most one object with type = "ephemeral". The Terraform SDK
// already deduplicates sets within configuration. For example, the following
// configuration in Terraform results in a single ephemeral external IP.
//
//	resource "oxide_instance" "example" {
//	  external_ips = [
//	    { type = "ephemeral"},
//	    { type = "ephemeral"},
//	  ]
//	}
//
// However, that deduplication does not extend to sets that contain different
// attributes, like so.
//
//	resource "oxide_instance" "example" {
//	  external_ips = [
//	    { type = "ephemeral", id = "a58dc21d-896d-4e5a-bb77-b0922a04e553"},
//	    { type = "ephemeral"},
//	  ]
//	}
//
// That's where this validator comes in. This validator errors with the above
// configuration, preventing a user from using multiple ephemeral external IPs.
func (f instanceExternalIPValidator) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var externalIPs []instanceResourceExternalIPModel
	diags := req.ConfigValue.ElementsAs(ctx, &externalIPs, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var ephemeralExternalIPs int
	for _, externalIP := range externalIPs {
		if externalIP.Type.ValueString() == string(oxide.ExternalIpCreateTypeEphemeral) {
			ephemeralExternalIPs++
		}
	}
	if ephemeralExternalIPs > 1 {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Too many external IPs with type = %q", oxide.ExternalIpCreateTypeEphemeral),
			fmt.Sprintf("Only 1 external IP with type = %q is allowed, but found %d.", oxide.ExternalIpCreateTypeEphemeral, ephemeralExternalIPs),
		)
		return
	}
}

var _ validator.String = ipAssignmentValidator{}

// ipAssignmentValidator validates that a string is a valid ip_assignment.
type ipAssignmentValidator struct {
	ipVersion oxide.IpVersion
}

func (v ipAssignmentValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v ipAssignmentValidator) MarkdownDescription(_ context.Context) string {
	return "value must be an IPv4 address or auto."
}

func (v ipAssignmentValidator) ValidateString(_ context.Context, req validator.StringRequest, res *validator.StringResponse) {
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
			fmt.Sprintf("Invalid IP%s assignment", v.ipVersion),
			fmt.Sprintf("Attribute %s must be an IP%s address or auto.", req.Path, v.ipVersion),
		)
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
func ModifyPlanForHostnameDeprecation(ctx context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
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

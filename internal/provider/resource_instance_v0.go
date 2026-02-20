// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/oxidecomputer/oxide.go/oxide"
)

// Schema and structs source:
// https://github.com/oxidecomputer/terraform-provider-oxide/blob/v0.17.0/internal/provider/resource_instance.go

type instanceResourceModelV0 struct {
	AntiAffinityGroups types.Set                           `tfsdk:"anti_affinity_groups"`
	AutoRestartPolicy  types.String                        `tfsdk:"auto_restart_policy"`
	BootDiskID         types.String                        `tfsdk:"boot_disk_id"`
	Description        types.String                        `tfsdk:"description"`
	DiskAttachments    types.Set                           `tfsdk:"disk_attachments"`
	ExternalIPs        []instanceResourceExternalIPModelV0 `tfsdk:"external_ips"`
	HostName           types.String                        `tfsdk:"host_name"`
	ID                 types.String                        `tfsdk:"id"`
	Memory             types.Int64                         `tfsdk:"memory"`
	Name               types.String                        `tfsdk:"name"`
	NetworkInterfaces  []instanceResourceNICModelV0        `tfsdk:"network_interfaces"`
	NCPUs              types.Int64                         `tfsdk:"ncpus"`
	ProjectID          types.String                        `tfsdk:"project_id"`
	SSHPublicKeys      types.Set                           `tfsdk:"ssh_public_keys"`
	StartOnCreate      types.Bool                          `tfsdk:"start_on_create"`
	TimeCreated        types.String                        `tfsdk:"time_created"`
	TimeModified       types.String                        `tfsdk:"time_modified"`
	Timeouts           timeouts.Value                      `tfsdk:"timeouts"`
	UserData           types.String                        `tfsdk:"user_data"`
}

type instanceResourceNICModelV0 struct {
	Description  types.String `tfsdk:"description"`
	ID           types.String `tfsdk:"id"`
	IPAddr       types.String `tfsdk:"ip_address"`
	MAC          types.String `tfsdk:"mac_address"`
	Name         types.String `tfsdk:"name"`
	Primary      types.Bool   `tfsdk:"primary"`
	SubnetID     types.String `tfsdk:"subnet_id"`
	TimeCreated  types.String `tfsdk:"time_created"`
	TimeModified types.String `tfsdk:"time_modified"`
	VPCID        types.String `tfsdk:"vpc_id"`
}

type instanceResourceExternalIPModelV0 struct {
	ID   types.String `tfsdk:"id"`
	Type types.String `tfsdk:"type"`
}

func (r *instanceResource) schemaV0(ctx context.Context) *schema.Schema {
	return &schema.Schema{
		Version: 0,
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
				Required:    true,
				Description: "Host name of the instance.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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
						"ip_address": schema.StringAttribute{
							Optional: true,
							Computed: true,
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
				Validators:  []validator.Set{},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						// The id attribute is optional, computed, and has a default to account for
						// the case where an instance created with an external IP using the default
						// IP pool (i.e., id = null) would drift when read (e.g., id = "") and
						// require updating
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

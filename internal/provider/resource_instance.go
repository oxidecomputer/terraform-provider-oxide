// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"
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
	_ resource.Resource              = (*instanceResource)(nil)
	_ resource.ResourceWithConfigure = (*instanceResource)(nil)
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
	BootDiskID         types.String                      `tfsdk:"boot_disk_id"`
	Description        types.String                      `tfsdk:"description"`
	DiskAttachments    types.Set                         `tfsdk:"disk_attachments"`
	ExternalIPs        []instanceResourceExternalIPModel `tfsdk:"external_ips"`
	HostName           types.String                      `tfsdk:"host_name"`
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

func (r *instanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Schema defines the schema for the resource.
func (r *instanceResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
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
				Description: "Host name of the instance",
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
			"anti_affinity_groups": schema.SetAttribute{
				Optional:    true,
				Description: "IDs of the anti-affinity groups this instance should belong to.",
				ElementType: types.StringType,
			},
			"boot_disk_id": schema.StringAttribute{
				Optional:    true,
				Description: "ID of the disk the instance should be booted from.",
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
				Description: "Starts the instance on creation",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"disk_attachments": schema.SetAttribute{
				Optional:    true,
				Description: "IDs of the disks to be attached to the instance.",
				ElementType: types.StringType,
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
				Description: "Associated Network Interfaces.",
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
							Description: "ID of the VPC in which to create the instance network interface",
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
				Validators: []validator.Set{
					instanceExternalIPValidator{},
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						// The id attribute is optional, computed, and has a default to account for the
						// case where an instance created with an external IP using the default IP pool
						// (i.e., id = null) would drift when read (e.g., id = "") and require updating
						// in place.
						"id": schema.StringAttribute{
							Description: "If type is ephemeral, ID of the IP pool to retrieve addresses from, or all available pools if not specified. If type is floating, ID of the floating IP",
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString(""),
						},
						"type": schema.StringAttribute{
							Description: "Type of external IP.",
							Required:    true,
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
				Description: "User data for instance initialization systems (such as cloud-init). " +
					"Must be a Base64-encoded string, as specified in RFC 4648 § 4 (+ and / characters with padding). " +
					"Maximum 32 KiB unencoded data.",
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

	params := oxide.InstanceCreateParams{
		Project: oxide.NameOrId(plan.ProjectID.ValueString()),
		Body: &oxide.InstanceCreate{
			Description: plan.Description.ValueString(),
			Name:        oxide.Name(plan.Name.ValueString()),
			Hostname:    oxide.Hostname(plan.HostName.ValueString()),
			Memory:      oxide.ByteCount(plan.Memory.ValueInt64()),
			Ncpus:       oxide.InstanceCpuCount(plan.NCPUs.ValueInt64()),
			Start:       plan.StartOnCreate.ValueBoolPointer(),
			UserData:    plan.UserData.ValueString(),
		},
	}

	// Add boot disk if any
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
		// Setting IPAddress as it is both computed and optional
		plan.NetworkInterfaces[i].IPAddr = types.StringValue(nic.Ip)
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
	state.Description = types.StringValue(instance.Description)
	state.HostName = types.StringValue(string(instance.Hostname))
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

	nicSet, diags := newAttachedNetworkInterfacesModel(ctx, r.client, state.ID.ValueString())
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
	if state.BootDiskID != plan.BootDiskID ||
		state.Memory != plan.Memory ||
		state.NCPUs != plan.NCPUs {

		params := oxide.InstanceUpdateParams{
			Instance: oxide.NameOrId(state.ID.ValueString()),
			Body: &oxide.InstanceUpdate{
				BootDisk: oxide.NameOrId(plan.BootDiskID.ValueString()),
				Memory:   oxide.ByteCount(plan.Memory.ValueInt64()),
				Ncpus:    oxide.InstanceCpuCount(plan.NCPUs.ValueInt64()),
			},
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
	nicsToCreate := sliceDiff(planNICs, stateNICs)
	resp.Diagnostics.Append(createNICs(ctx, r.client, nicsToCreate, state.ID.ValueString())...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check state and if it has an ID that the plan doesn't then delete it
	nicsToDelete := sliceDiff(stateNICs, planNICs)
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
	nicModel, diags := newAttachedNetworkInterfacesModel(ctx, r.client, state.ID.ValueString())
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
		Refresh: func() (interface{}, string, error) {
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
			Ip:          planNIC.IPAddr.ValueString(),
			Name:        oxide.Name(planNIC.Name.ValueString()),
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

func newAttachedNetworkInterfacesModel(ctx context.Context, client *oxide.Client, instanceID string) (
	[]instanceResourceNICModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	params := oxide.InstanceNetworkInterfaceListParams{
		Instance: oxide.NameOrId(instanceID),
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
			IPAddr:       types.StringValue(nic.Ip),
			MAC:          types.StringValue(string(nic.Mac)),
			Name:         types.StringValue(string(nic.Name)),
			Primary:      types.BoolPointerValue(nic.Primary),
			SubnetID:     types.StringValue(nic.SubnetId),
			TimeCreated:  types.StringValue(nic.TimeCreated.String()),
			TimeModified: types.StringValue(nic.TimeModified.String()),
			VPCID:        types.StringValue(nic.VpcId),
		}
		nicSet = append(nicSet, n)
	}

	return nicSet, nil
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
				Ip:          model.IPAddr.ValueString(),
				Name:        oxide.Name(model.Name.ValueString()),
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

// Description describes the validation in plain text formatting.
func (f instanceExternalIPValidator) Description(context.Context) string {
	return "cannot have more than one ephemeral external ip"
}

// MarkdownDescription describes the validation in Markdown formatting.
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

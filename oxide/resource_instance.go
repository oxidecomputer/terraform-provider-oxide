// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	oxideSDK "github.com/oxidecomputer/oxide.go/oxide"
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
	client *oxideSDK.Client
}

type instanceResourceModel struct {
	Description       types.String               `tfsdk:"description"`
	DiskAttachments   types.Set                  `tfsdk:"disk_attachments"`
	ExternalIPs       types.List                 `tfsdk:"external_ips"`
	HostName          types.String               `tfsdk:"host_name"`
	ID                types.String               `tfsdk:"id"`
	Memory            types.Int64                `tfsdk:"memory"`
	Name              types.String               `tfsdk:"name"`
	NetworkInterfaces []instanceResourceNICModel `tfsdk:"network_interfaces"`
	NCPUs             types.Int64                `tfsdk:"ncpus"`
	ProjectID         types.String               `tfsdk:"project_id"`
	StartOnCreate     types.Bool                 `tfsdk:"start_on_create"`
	TimeCreated       types.String               `tfsdk:"time_created"`
	TimeModified      types.String               `tfsdk:"time_modified"`
	Timeouts          timeouts.Value             `tfsdk:"timeouts"`
	UserData          types.String               `tfsdk:"user_data"`
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

// Metadata returns the resource type name.
func (r *instanceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "oxide_instance"
}

// Configure adds the provider configured client to the data source.
func (r *instanceResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*oxideSDK.Client)
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
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"ncpus": schema.Int64Attribute{
				Required:    true,
				Description: "Number of CPUs allocated for this instance.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
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
			"external_ips": schema.ListAttribute{
				Optional:    true,
				Description: "External IP addresses provided to this instance. List of IP pools from which to draw addresses.",
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
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

	params := oxideSDK.InstanceCreateParams{
		Project: oxideSDK.NameOrId(plan.ProjectID.ValueString()),
		Body: &oxideSDK.InstanceCreate{
			Description: plan.Description.ValueString(),
			Name:        oxideSDK.Name(plan.Name.ValueString()),
			Hostname:    plan.HostName.ValueString(),
			Memory:      oxideSDK.ByteCount(plan.Memory.ValueInt64()),
			Ncpus:       oxideSDK.InstanceCpuCount(plan.NCPUs.ValueInt64()),
			Start:       plan.StartOnCreate.ValueBoolPointer(),
			UserData:    plan.UserData.ValueString(),
		},
	}

	// Retrieve names of all disks based on their provided IDs
	var disks = []oxideSDK.InstanceDiskAttachment{}
	for _, diskAttch := range plan.DiskAttachments.Elements() {
		diskID, err := strconv.Unquote(diskAttch.String())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error retrieving disk information",
				"Disk ID parse error: "+err.Error(),
			)
			return
		}

		disk, err := r.client.DiskView(oxideSDK.DiskViewParams{
			Disk: oxideSDK.NameOrId(diskID),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error retrieving disk information",
				"API error: "+err.Error(),
			)
			return
		}

		da := oxideSDK.InstanceDiskAttachment{
			Name: disk.Name,
			// Only allow attach (no disk create on instance create)
			Type: oxideSDK.InstanceDiskAttachmentTypeAttach,
		}
		disks = append(disks, da)
	}
	params.Body.Disks = disks

	var externalIPs = []oxideSDK.ExternalIpCreate{}
	for _, ip := range plan.ExternalIPs.Elements() {
		poolName, err := strconv.Unquote(ip.String())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating external IP addresses",
				"IP pool name parse error: "+err.Error(),
			)
			return
		}
		eIP := oxideSDK.ExternalIpCreate{
			PoolName: oxideSDK.Name(poolName),
			// TODO: Implement other types when these are supported.
			Type: oxideSDK.ExternalIpCreateTypeEphemeral,
		}

		externalIPs = append(externalIPs, eIP)
	}
	params.Body.ExternalIps = externalIPs

	nics, diags := newNetworkInterfaceAttachment(ctx, r.client, plan.NetworkInterfaces)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	params.Body.NetworkInterfaces = nics

	instance, err := r.client.InstanceCreate(params)
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
		nic, err := r.client.InstanceNetworkInterfaceView(oxideSDK.InstanceNetworkInterfaceViewParams{
			Interface: oxideSDK.NameOrId(plan.NetworkInterfaces[i].Name.ValueString()),
			Instance:  oxideSDK.NameOrId(instance.Id),
		})
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

	instance, err := r.client.InstanceView(oxideSDK.InstanceViewParams{
		Instance: oxideSDK.NameOrId(state.ID.ValueString()),
	})
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

	state.Description = types.StringValue(instance.Description)
	state.HostName = types.StringValue(string(instance.Hostname))
	state.ID = types.StringValue(instance.Id)
	state.Memory = types.Int64Value(int64(instance.Memory))
	state.Name = types.StringValue(string(instance.Name))
	state.NCPUs = types.Int64Value(int64(instance.Ncpus))
	state.ProjectID = types.StringValue(instance.ProjectId)
	state.TimeCreated = types.StringValue(instance.TimeCreated.String())
	state.TimeModified = types.StringValue(instance.TimeModified.String())

	diskSet, diags := newAttachedDisksSet(r.client, state.ID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Only set the disk list if there are disk attachments
	if len(diskSet.Elements()) > 0 {
		state.DiskAttachments = diskSet
	}

	nicSet, diags := newAttachedNetworkInterfacesSet(r.client, state.ID.ValueString())
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

	planDisks := plan.DiskAttachments.Elements()
	stateDisks := state.DiskAttachments.Elements()

	// Check plan and if it has an ID that the state doesn't then attach it
	disksToAttach := sliceDiff(planDisks, stateDisks)
	for _, v := range disksToAttach {
		diskID, err := strconv.Unquote(v.String())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error attaching disk",
				"Disk ID parse error: "+err.Error(),
			)
			return
		}
		_, err = r.client.InstanceDiskAttach(oxideSDK.InstanceDiskAttachParams{
			Instance: oxideSDK.NameOrId(state.ID.ValueString()),
			Body: &oxideSDK.DiskPath{
				Disk: oxideSDK.NameOrId(diskID),
			},
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error attaching disk",
				"API error: "+err.Error(),
			)
			return
		}
		tflog.Trace(ctx, fmt.Sprintf("attached disk with ID: %v", v), map[string]any{"success": true})
	}

	// Check state and if it has an ID that the plan doesn't then detach it
	disksToDetach := sliceDiff(stateDisks, planDisks)
	for _, v := range disksToDetach {
		diskID, err := strconv.Unquote(v.String())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error detaching disk",
				"Disk ID parse error: "+err.Error(),
			)
			return
		}
		_, err = r.client.InstanceDiskDetach(oxideSDK.InstanceDiskDetachParams{
			Instance: oxideSDK.NameOrId(state.ID.ValueString()),
			Body: &oxideSDK.DiskPath{
				Disk: oxideSDK.NameOrId(diskID),
			},
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error detaching disk",
				"API error: "+err.Error(),
			)
			return
		}
		tflog.Trace(ctx, fmt.Sprintf("detached disk with ID: %v", v), map[string]any{"success": true})
	}

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

	// Read instance to retrieve modified time value
	instance, err := r.client.InstanceView(oxideSDK.InstanceViewParams{
		Instance: oxideSDK.NameOrId(state.ID.ValueString()),
	})
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

	// Retrieve attached disks
	diskSet, diags := newAttachedDisksSet(r.client, state.ID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Only set the disk list if there are disk attachments
	if len(diskSet.Elements()) > 0 {
		plan.DiskAttachments = diskSet
	}

	nicSet, diags := newAttachedNetworkInterfacesSet(r.client, state.ID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Only populate NICs if there are associated NICs to avoid drift
	if len(nicSet) > 0 {
		plan.NetworkInterfaces = nicSet
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
	_, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	// Detach all disks
	for _, diskAttch := range state.DiskAttachments.Elements() {
		diskID, err := strconv.Unquote(diskAttch.String())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error detaching disk",
				"Disk ID parse error: "+err.Error(),
			)
			return
		}

		_, err = r.client.InstanceDiskDetach(oxideSDK.InstanceDiskDetachParams{
			Instance: oxideSDK.NameOrId(state.ID.ValueString()),
			Body: &oxideSDK.DiskPath{
				Disk: oxideSDK.NameOrId(diskID),
			},
		})
		if err != nil {
			if !is404(err) {
				resp.Diagnostics.AddError(
					"Error detaching disk",
					"API error: "+err.Error(),
				)
				return
			}
			continue
		}
		tflog.Trace(ctx, fmt.Sprintf("detached disk with ID: %v", diskID), map[string]any{"success": true})
	}

	_, err := r.client.InstanceStop(oxideSDK.InstanceStopParams{
		Instance: oxideSDK.NameOrId(state.ID.ValueString()),
	})
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

	if err := r.client.InstanceDelete(oxideSDK.InstanceDeleteParams{
		Instance: oxideSDK.NameOrId(state.ID.ValueString()),
	}); err != nil {
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

func waitForInstanceStop(ctx context.Context, client *oxideSDK.Client, timeout time.Duration, instanceID string) diag.Diagnostics {
	var diags diag.Diagnostics

	stateConfig := retry.StateChangeConf{
		PollInterval: time.Second,
		Delay:        time.Second,
		Pending: []string{
			string(oxideSDK.InstanceStateCreating),
			string(oxideSDK.InstanceStateStarting),
			string(oxideSDK.InstanceStateStopping),
			string(oxideSDK.InstanceStateRebooting),
		},
		Target:  []string{string(oxideSDK.InstanceStateStopped)},
		Timeout: timeout,
		Refresh: func() (interface{}, string, error) {
			tflog.Info(ctx, fmt.Sprintf("checking on state of instance: %v", instanceID))
			instance, err := client.InstanceView(oxideSDK.InstanceViewParams{
				Instance: oxideSDK.NameOrId(instanceID),
			})
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

func newAttachedDisksSet(client *oxideSDK.Client, instanceID string) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics

	disks, err := client.InstanceDiskList(oxideSDK.InstanceDiskListParams{
		Limit:    1000000000,
		Instance: oxideSDK.NameOrId(instanceID),
	})
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

func newNetworkInterfaceAttachment(ctx context.Context, client *oxideSDK.Client, model []instanceResourceNICModel) (
	oxideSDK.InstanceNetworkInterfaceAttachment, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(model) == 0 {
		return oxideSDK.InstanceNetworkInterfaceAttachment{
			Type: oxideSDK.InstanceNetworkInterfaceAttachmentTypeNone,
		}, diags
	}

	var nicParams []oxideSDK.InstanceNetworkInterfaceCreate
	for _, planNIC := range model {
		names, diags := retrieveVPCandSubnetNames(ctx, client, planNIC.VPCID.ValueString(),
			planNIC.SubnetID.ValueString())
		diags.Append(diags...)
		if diags.HasError() {
			return oxideSDK.InstanceNetworkInterfaceAttachment{}, diags
		}

		nic := oxideSDK.InstanceNetworkInterfaceCreate{
			Description: planNIC.Description.ValueString(),
			Ip:          planNIC.IPAddr.ValueString(),
			Name:        oxideSDK.Name(planNIC.Name.ValueString()),
			SubnetName:  oxideSDK.Name(names.subnet),
			VpcName:     oxideSDK.Name(names.vpc),
		}
		nicParams = append(nicParams, nic)
	}

	nicAttachment := oxideSDK.InstanceNetworkInterfaceAttachment{
		Type:   oxideSDK.InstanceNetworkInterfaceAttachmentTypeCreate,
		Params: nicParams,
	}
	return nicAttachment, nil
}

func newAttachedNetworkInterfacesSet(client *oxideSDK.Client, instanceID string) (
	[]instanceResourceNICModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	nics, err := client.InstanceNetworkInterfaceList(oxideSDK.InstanceNetworkInterfaceListParams{
		Instance: oxideSDK.NameOrId(instanceID),
		Limit:    1000000000,
	})
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
			Description: types.StringValue(nic.Description),
			ID:          types.StringValue(nic.Id),
			IPAddr:      types.StringValue(nic.Ip),
			MAC:         types.StringValue(string(nic.Mac)),
			Name:        types.StringValue(string(nic.Name)),
			// TODO: Should I check for nil before assigning this one?
			Primary:      types.BoolValue(*nic.Primary),
			SubnetID:     types.StringValue(nic.SubnetId),
			TimeCreated:  types.StringValue(nic.TimeCreated.String()),
			TimeModified: types.StringValue(nic.TimeModified.String()),
			VPCID:        types.StringValue(nic.VpcId),
		}
		nicSet = append(nicSet, n)
	}

	return nicSet, nil
}

type vpcAndSubnetNames struct {
	vpc    string
	subnet string
}

func retrieveVPCandSubnetNames(ctx context.Context, client *oxideSDK.Client, vpcID, subnetID string) (
	vpcAndSubnetNames, diag.Diagnostics) {
	var diags diag.Diagnostics
	// This is an unfortunate result of having the create body use names as identifiers
	// but the body return IDs. making two API calls to retrieve VPC and subnet names.
	// Using IDs only for the provider schema as names are mutable.
	vpc, err := client.VpcView(oxideSDK.VpcViewParams{
		Vpc: oxideSDK.NameOrId(vpcID),
	})
	if err != nil {
		diags.AddError(
			"Unable to read information about corresponding VPC:",
			"API error: "+err.Error(),
		)
		return vpcAndSubnetNames{}, diags
	}
	tflog.Trace(ctx, fmt.Sprintf("read VPC with ID: %v", vpcID), map[string]any{"success": true})

	subnet, err := client.VpcSubnetView(oxideSDK.VpcSubnetViewParams{
		Subnet: oxideSDK.NameOrId(subnetID),
	})
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

func createNICs(ctx context.Context, client *oxideSDK.Client, models []instanceResourceNICModel, instanceID string) diag.Diagnostics {
	for _, model := range models {
		names, diags := retrieveVPCandSubnetNames(ctx, client, model.VPCID.ValueString(),
			model.SubnetID.ValueString())
		diags.Append(diags...)
		if diags.HasError() {
			return diags
		}

		params := oxideSDK.InstanceNetworkInterfaceCreateParams{
			Instance: oxideSDK.NameOrId(instanceID),
			Body: &oxideSDK.InstanceNetworkInterfaceCreate{
				Description: model.Description.ValueString(),
				Ip:          model.IPAddr.ValueString(),
				Name:        oxideSDK.Name(model.Name.ValueString()),
				SubnetName:  oxideSDK.Name(names.subnet),
				VpcName:     oxideSDK.Name(names.vpc),
			},
		}

		nic, err := client.InstanceNetworkInterfaceCreate(params)
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

func deleteNICs(ctx context.Context, client *oxideSDK.Client, models []instanceResourceNICModel) diag.Diagnostics {
	var diags diag.Diagnostics

	for _, model := range models {
		if err := client.InstanceNetworkInterfaceDelete(oxideSDK.InstanceNetworkInterfaceDeleteParams{
			Interface: oxideSDK.NameOrId(model.ID.ValueString()),
		}); err != nil {
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

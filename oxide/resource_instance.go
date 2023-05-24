// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"context"
	"fmt"
	"strconv"

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
	Description types.String `tfsdk:"description"`
	ID          types.String `tfsdk:"id"`
	IPAddr      types.String `tfsdk:"ip_address"`
	// InstanceID   types.String   `tfsdk:"instance_id"`
	MAC          types.String `tfsdk:"mac_address"`
	Name         types.String `tfsdk:"name"`
	Primary      types.Bool   `tfsdk:"primary"`
	SubnetID     types.String `tfsdk:"subnet_id"`
	TimeCreated  types.String `tfsdk:"time_created"`
	TimeModified types.String `tfsdk:"time_modified"`
	//Timeouts     timeouts.Value `tfsdk:"timeouts"`
	VPCID types.String `tfsdk:"vpc_id"`
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
						},
						"description": schema.StringAttribute{
							Required:    true,
							Description: "Description for the instance network interface.",
						},
						"subnet_id": schema.StringAttribute{
							Required:    true,
							Description: "ID of the VPC subnet in which to create the instance network interface.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"vpc_id": schema.StringAttribute{
							Required:    true,
							Description: "ID of the VPC in which to create the instance network interface",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
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
				// TODO: Restore once updates are enabled
				// Update: true,
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

	//NICs
	nics := oxideSDK.InstanceNetworkInterfaceAttachment{}
	if len(plan.NetworkInterfaces) > 0 {
		nics.Type = oxideSDK.InstanceNetworkInterfaceAttachmentTypeCreate
	} else {
		nics.Type = oxideSDK.InstanceNetworkInterfaceAttachmentTypeNone
	}

	nicParams := []oxideSDK.InstanceNetworkInterfaceCreate{}
	for _, planNIC := range plan.NetworkInterfaces {
		// This is an unfortunate result of having the create body use names as identifiers
		// but the body return IDs. making two API calls to retrieve VPC and subnet names
		// Using IDs only for the provider schema as names are mutable.
		vpc, err := r.client.VpcView(oxideSDK.VpcViewParams{
			Vpc: oxideSDK.NameOrId(planNIC.VPCID.ValueString()),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to read information about corresponding VPC:",
				"API error: "+err.Error(),
			)
			return
		}
		tflog.Trace(ctx, fmt.Sprintf("read VPC with ID: %v", planNIC.VPCID.ValueString()), map[string]any{"success": true})

		subnet, err := r.client.VpcSubnetView(oxideSDK.VpcSubnetViewParams{
			Subnet: oxideSDK.NameOrId(planNIC.SubnetID.ValueString()),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to read information about corresponding subnet:",
				"API error: "+err.Error(),
			)
			return
		}
		tflog.Trace(ctx, fmt.Sprintf("read subnet with ID: %v", planNIC.SubnetID.ValueString()), map[string]any{"success": true})

		nic := oxideSDK.InstanceNetworkInterfaceCreate{
			Description: planNIC.Description.ValueString(),
			Ip:          planNIC.IPAddr.ValueString(),
			Name:        oxideSDK.Name(planNIC.Name.ValueString()),
			SubnetName:  subnet.Name,
			VpcName:     vpc.Name,
		}
		nicParams = append(nicParams, nic)
	}
	nics.Params = nicParams

	params.Body.NetworkInterfaces = nics
	//

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
	if len(plan.NetworkInterfaces) > 0 {
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

	//	nics, err := r.client.InstanceNetworkInterfaceList(oxideSDK.InstanceNetworkInterfaceListParams{
	//		Instance: oxideSDK.NameOrId(instance.Id),
	//		Limit:    1000000000,
	//	})
	//	if err != nil {
	//		resp.Diagnostics.AddError(
	//			"Unable to read instance network interfaces:",
	//			"API error: "+err.Error(),
	//		)
	//		return
	//	}
	//	n := []attr.Value{}
	//	for _, nic := range nics.Items {
	//		id := types.StringValue(nic.Id)
	//		n = append(n, id)
	//	}

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
			resp.Diagnostics.AddError(
				"Error detaching disk",
				"API error: "+err.Error(),
			)
			return
		}
		tflog.Trace(ctx, fmt.Sprintf("detached disk with ID: %v", diskID), map[string]any{"success": true})
	}

	// TODO: Double check if this is necessary, could be an optional feature?
	//_, err = r.client.InstanceStop(oxideSDK.InstanceStopParams{
	//	Instance: oxideSDK.NameOrId(state.ID.ValueString()),
	//})
	//if err != nil {
	//	if !is404(err) {
	//		resp.Diagnostics.AddError(
	//			"Unable to stop instance:",
	//			"API error: "+err.Error(),
	//		)
	//		return
	//	}
	//}
	//
	//	ch := make(chan error)
	//	go waitForStoppedInstance(r.client, oxideSDK.NameOrId(state.ID.ValueString()), ch)
	//	e := <-ch
	//	if !is404(e) {
	//		resp.Diagnostics.AddError(
	//			"Unable to stop instance:",
	//			"API error: "+e.Error(),
	//		)
	//		return
	//	}
	// tflog.Trace(ctx, fmt.Sprintf("stopped instance with ID: %v", state.ID.ValueString()), map[string]any{"success": true})

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

// func waitForStoppedInstance(client *oxideSDK.Client, instanceID oxideSDK.NameOrId, ch chan error) {
// 	for {
// 		params := oxideSDK.InstanceViewParams{Instance: instanceID}
// 		resp, err := client.InstanceView(params)
// 		if err != nil {
// 			ch <- err
// 		}
// 		if resp.RunState == oxideSDK.InstanceStateStopped {
// 			break
// 		}
// 		// Suggested alternatives suggested by linter are not fit for purpose
// 		//lintignore:R018
// 		time.Sleep(time.Second)
// 	}
// 	ch <- nil
// }

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

	return diskSet, diags
}

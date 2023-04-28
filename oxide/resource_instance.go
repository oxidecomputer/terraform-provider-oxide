// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

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
	// TODO: Evaluate if we wish to create disks as well, and the implications of this
	AttachToDisks types.List   `tfsdk:"attach_to_disks"`
	Description   types.String `tfsdk:"description"`
	ExternalIPs   types.List   `tfsdk:"external_ips"`
	HostName      types.String `tfsdk:"host_name"`
	ID            types.String `tfsdk:"id"`
	Memory        types.Int64  `tfsdk:"memory"`
	Name          types.String `tfsdk:"name"`
	NCPUs         types.Int64  `tfsdk:"ncpus"`
	// TODO: This should be plural
	NetworkInterface    []instanceResourceNetworkInterfaceModel `tfsdk:"network_interface"`
	ProjectID           types.String                            `tfsdk:"project_id"`
	RunState            types.String                            `tfsdk:"run_state"`
	TimeCreated         types.String                            `tfsdk:"time_created"`
	TimeModified        types.String                            `tfsdk:"time_modified"`
	TimeRunStateUpdated types.String                            `tfsdk:"time_run_state_updated"`
	Timeouts            timeouts.Value                          `tfsdk:"timeouts"`
}

type instanceResourceNetworkInterfaceModel struct {
	Description types.String `tfsdk:"description"`
	// TODO: Return the ID of the nic
	IP       types.String `tfsdk:"ip"`
	Name     types.String `tfsdk:"name"`
	SubnetID types.String `tfsdk:"subnet_id"`
	VPCID    types.String `tfsdk:"vpc_id"`
	// TODO: We shouldn't need to use these, only IDs.
	// Fix the API
	SubnetName types.String `tfsdk:"subnet_name"`
	VPCName    types.String `tfsdk:"vpc_name"`
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

// Schema defines the schema for the resource.
func (r *instanceResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the project that will contain the instance.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the instance.",
			},
			"description": schema.StringAttribute{
				Required:    true,
				Description: "Description for the instance.",
			},
			"host_name": schema.StringAttribute{
				Required:    true,
				Description: "Host name of the instance",
			},
			"memory": schema.Int64Attribute{
				Required:    true,
				Description: "Instance memory in bytes.",
			},
			"ncpus": schema.Int64Attribute{
				Required:    true,
				Description: "Number of CPUs allocated for this instance.",
			},
			"attach_to_disks": schema.ListAttribute{
				Optional:    true,
				Description: "Disks to be attached to this instance.",
				ElementType: types.StringType,
			},
			"external_ips": schema.ListAttribute{
				Optional:    true,
				Description: "External IP addresses provided to this instance. List of IP pools from which to draw addresses.",
				ElementType: types.StringType,
			},
			"network_interface": schema.ListNestedAttribute{
				Optional:    true,
				Description: "Attaches network interfaces to an instance at the time the instance is created.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "Name of the network interface.",
						},
						"description": schema.StringAttribute{
							Required:    true,
							Description: "Description for the network interface.",
						},
						"subnet_name": schema.StringAttribute{
							Required:    true,
							Description: "Name of the VPC Subnet in which to create the network interface.",
						},
						"vpc_name": schema.StringAttribute{
							Required:    true,
							Description: "Name of the VPC in which to create the network interface.",
						},
						"ip": schema.StringAttribute{
							// TODO: For the purposes of this demo we will stick to
							// auto-assigned IP addresses. In the future we will want
							// this value to be computed/optional or required.
							Computed:    true,
							Description: "IP address for the network interface.",
						},
						"subnet_id": schema.StringAttribute{
							Computed:    true,
							Description: "ID of the VPC Subnet to which the interface belongs.",
						},
						"vpc_id": schema.StringAttribute{
							Computed:    true,
							Description: "ID of the VPC to which the interface belongs.",
						},
					},
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
			"run_state": schema.StringAttribute{
				Computed: true,
				Description: "Running state of an Instance (primarily: booted or stopped)." +
					" This typically reflects whether it's starting, running, stopping, or stopped," +
					" but also includes states related to the instance's lifecycle.",
			},
			"time_created": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this instance was created.",
			},
			"time_modified": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this instance was last modified.",
			},
			"time_run_state_updated": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when the run state of this instance was last modified.",
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
		},
	}

	var diskAttachements = []oxideSDK.InstanceDiskAttachment{}
	for _, disk := range plan.AttachToDisks.Elements() {
		diskName, err := strconv.Unquote(disk.String())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error attaching instance to disk",
				"IP pool name parse error: "+err.Error(),
			)
			return
		}
		ds := oxideSDK.InstanceDiskAttachment{
			Name: oxideSDK.Name(diskName),
			// TODO: For now we are only attaching. Verify if it makes sense to create
			// as well. Probably not, there would be no way to delete that disk via
			// TF
			Type: oxideSDK.InstanceDiskAttachmentTypeAttach,
		}

		diskAttachements = append(diskAttachements, ds)
	}
	params.Body.Disks = diskAttachements

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

	var nicAttachment = oxideSDK.InstanceNetworkInterfaceAttachment{}
	if len(plan.NetworkInterface) < 1 {
		// When not attaching any nics the API requires that you explicitly declare it
		nicAttachment.Type = oxideSDK.InstanceNetworkInterfaceAttachmentTypeNone
	} else {
		nics := []oxideSDK.InstanceNetworkInterfaceCreate{}
		for _, nic := range plan.NetworkInterface {
			nicCreate := oxideSDK.InstanceNetworkInterfaceCreate{
				Description: nic.Description.ValueString(),
				Name:        oxideSDK.Name(nic.Name.ValueString()),
				// TODO: Ideally from the API we should be able to create with IDs, not names
				SubnetName: oxideSDK.Name(nic.SubnetName.ValueString()),
				VpcName:    oxideSDK.Name(nic.VPCName.ValueString()),
			}

			nics = append(nics, nicCreate)
		}
		nicAttachment.Params = nics
		nicAttachment.Type = oxideSDK.InstanceNetworkInterfaceAttachmentTypeCreate
	}
	params.Body.NetworkInterfaces = nicAttachment

	instance, err := r.client.InstanceCreate(params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating instance",
			"API error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(instance.Id)
	plan.RunState = types.StringValue(string(instance.RunState))
	plan.TimeCreated = types.StringValue(instance.TimeCreated.String())
	plan.TimeModified = types.StringValue(instance.TimeCreated.String())
	// TODO: Would it be problematic to have this reported? Likely changes randomly,
	// could make for a flaky resource
	plan.TimeRunStateUpdated = types.StringValue(instance.TimeRunStateUpdated.String())

	// The instance response does not include associated NICs, so we need additional
	// API calls to retrieve them
	if len(plan.NetworkInterface) > 1 {
		for index, nic := range plan.NetworkInterface {
			subnet, err := r.client.VpcSubnetView(oxideSDK.VpcSubnetViewParams{
				Subnet:  oxideSDK.NameOrId(nic.SubnetName.ValueString()),
				Vpc:     oxideSDK.NameOrId(nic.VPCName.ValueString()),
				Project: oxideSDK.NameOrId(plan.ProjectID.ValueString()),
			})
			if err != nil {
				resp.Diagnostics.AddError(
					"Error reading associated subnet information during create",
					"API error: "+err.Error(),
				)
				return
			}

			plan.NetworkInterface[index].SubnetID = types.StringValue(subnet.Id)
			plan.NetworkInterface[index].VPCID = types.StringValue(subnet.VpcId)
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

	state.Description = types.StringValue(instance.Description)
	state.HostName = types.StringValue(string(instance.Hostname))
	state.ID = types.StringValue(instance.Id)
	state.Memory = types.Int64Value(int64(instance.Memory))
	state.Name = types.StringValue(string(instance.Name))
	state.NCPUs = types.Int64Value(int64(instance.Ncpus))
	state.ProjectID = types.StringValue(instance.ProjectId)
	// Should this be something we have as part of the schema? likely to make for a buggy provider
	state.RunState = types.StringValue(string(instance.RunState))
	state.TimeCreated = types.StringValue(instance.TimeCreated.String())
	state.TimeModified = types.StringValue(instance.TimeCreated.String())
	state.TimeRunStateUpdated = types.StringValue(instance.TimeRunStateUpdated.String())

	//state.AttachToDisks = TODO
	//state.ExternalIPs = TODO

	nics, err := r.client.InstanceNetworkInterfaceList(
		oxideSDK.InstanceNetworkInterfaceListParams{
			Instance: oxideSDK.NameOrId(instance.Id),
			Limit:    1000000000,
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read network interface attachments:",
			"API error: "+err.Error(),
		)
		return
	}
	for index, item := range nics.Items {
		nic := instanceResourceNetworkInterfaceModel{
			Description: types.StringValue(item.Description),
			IP:          types.StringValue(item.Ip),
			Name:        types.StringValue(string(item.Name)),
			SubnetID:    types.StringValue(item.SubnetId),
			VPCID:       types.StringValue(item.VpcId),
		}

		// Ideally the NetworkInterface struct would contain the names of the VPC and subnet.
		// For now they only give the ID so we'll retrieve the names separately.
		vpc, err := r.client.VpcView(oxideSDK.VpcViewParams{
			Vpc: oxideSDK.NameOrId(item.VpcId),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to read information about corresponding VPC:",
				"API error: "+err.Error(),
			)
			return
		}

		subnet, err := r.client.VpcSubnetView(oxideSDK.VpcSubnetViewParams{
			Subnet: oxideSDK.NameOrId(item.SubnetId),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to read information about corresponding subnet:",
				"API error: "+err.Error(),
			)
			return
		}

		nic.SubnetName = types.StringValue(string(subnet.Name))
		nic.VPCName = types.StringValue(string(vpc.Name))

		state.NetworkInterface[index] = nic
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *instanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Error updating instance",
		"the oxide API currently does not support updating instances")
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

	// TODO: Check if I actually need the "wait for stopped instance function"
	// Wait for instance to be stopped before attempting to destroy
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
}

// func waitForStoppedInstance(client *oxideSDK.Client, instanceId oxideSDK.NameOrId, ch chan error) {
// 	for {
// 		params := oxideSDK.InstanceViewParams{Instance: instanceId}
// 		resp, err := client.InstanceView(params)
// 		if err != nil {
// 			ch <- err
// 		}
// 		if resp.RunState == oxideSDK.InstanceStateStopped {
// 			break
// 		}
// 		time.Sleep(time.Second)
// 	}
// 	ch <- nil
// }

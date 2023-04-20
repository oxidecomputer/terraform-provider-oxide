// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"context"
	"strconv"

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
func (r *instanceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
		ds := oxideSDK.InstanceDiskAttachment{
			Name: oxideSDK.Name(disk.String()),
			// TODO: For now we are only attaching. Verify if it makes sense to create
			// as well. Probably not, there would be no way to delete that disk via
			// TF
			Type: oxideSDK.InstanceDiskAttachmentTypeAttach,
		}

		diskAttachements = append(diskAttachements, ds)
	}
	params.Body.Disks = diskAttachements

	var externalIPs = []oxideSDK.ExternalIpCreate{}
	for _, ip := range plan.AttachToDisks.Elements() {
		eIP := oxideSDK.ExternalIpCreate{
			PoolName: oxideSDK.Name(strconv.Unquote(ip.String())),
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
			"API error: "+err.Error()+" "+string(params.Body.Disks[0].Name),
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

// func oldinstanceResource() *schema.Resource {
// 	return &schema.Resource{
// 		Description:   "",
// 		Schema:        newInstanceSchema(),
// 		CreateContext: createInstance,
// 		ReadContext:   readInstance,
// 		UpdateContext: updateInstance,
// 		DeleteContext: deleteInstance,
// 		Importer: &schema.ResourceImporter{
// 			StateContext: schema.ImportStatePassthroughContext,
// 		},
// 		Timeouts: &schema.ResourceTimeout{
// 			Default: schema.DefaultTimeout(5 * time.Minute),
// 		},
// 	}
// }
//
// func newInstanceSchema() map[string]*schema.Schema {
// 	return map[string]*schema.Schema{
// 		"project_id": {
// 			Type:        schema.TypeString,
// 			Description: "ID of the project that will contain the instance.",
// 			Required:    true,
// 		},
// 		"name": {
// 			Type:        schema.TypeString,
// 			Description: "Name of the instance.",
// 			Required:    true,
// 		},
// 		"description": {
// 			Type:        schema.TypeString,
// 			Description: "Description for the instance.",
// 			Required:    true,
// 		},
// 		"host_name": {
// 			Type:        schema.TypeString,
// 			Description: "Host name of the instance.",
// 			Required:    true,
// 		},
// 		"memory": {
// 			Type:        schema.TypeInt,
// 			Description: "Instance memory in bytes.",
// 			Required:    true,
// 		},
// 		"ncpus": {
// 			Type:        schema.TypeInt,
// 			Description: "Number of CPUs allocated for this instance.",
// 			Required:    true,
// 		},
// 		"attach_to_disks": {
// 			Type:        schema.TypeList,
// 			Description: "Disks to be attached to this instance.",
// 			Optional:    true,
// 			Elem: &schema.Schema{
// 				Type: schema.TypeString,
// 			},
// 		},
// 		"external_ips": {
// 			Type:        schema.TypeList,
// 			Description: "External IP addresses provided to this instance. List of IP pools from which to draw addresses.",
// 			Optional:    true,
// 			Elem: &schema.Schema{
// 				Type: schema.TypeString,
// 			},
// 		},
// 		"network_interface": {
// 			Type:        schema.TypeList,
// 			Description: "Attaches network interfaces to an instance at the time the instance is created.",
// 			Optional:    true,
// 			Elem:        newNetworkInterfaceResource(),
// 		},
// 		"id": {
// 			Type:        schema.TypeString,
// 			Description: "Unique, immutable, system-controlled identifier.",
// 			Computed:    true,
// 		},
// 		"run_state": {
// 			Type:        schema.TypeString,
// 			Description: "Running state of an Instance (primarily: booted or stopped). This typically reflects whether it's starting, running, stopping, or stopped, but also includes states related to the instance's lifecycle.",
// 			Computed:    true,
// 		},
// 		"time_created": {
// 			Type:        schema.TypeString,
// 			Description: "Timestamp of when this instance was created.",
// 			Computed:    true,
// 		},
// 		"time_modified": {
// 			Type:        schema.TypeString,
// 			Description: "Timestamp of when this instance was last modified.",
// 			Computed:    true,
// 		},
// 		"time_run_state_updated": {
// 			Type:        schema.TypeString,
// 			Description: "Timestamp of when the run state of this instance was last modified.",
// 			Computed:    true,
// 		},
// 	}
// }
//
// func newNetworkInterfaceResource() *schema.Resource {
// 	return &schema.Resource{
// 		Schema: map[string]*schema.Schema{
// 			"description": {
// 				Type:        schema.TypeString,
// 				Description: "Description for the network interface.",
// 				Required:    true,
// 			},
// 			"name": {
// 				Type:        schema.TypeString,
// 				Description: "Name for the network interface.",
// 				Required:    true,
// 			},
// 			"subnet_name": {
// 				Type:        schema.TypeString,
// 				Description: "Name of the VPC Subnet in which to create the network interface.",
// 				Required:    true,
// 			},
// 			"vpc_name": {
// 				Type:        schema.TypeString,
// 				Description: "Name of the VPC in which to create the network interface.",
// 				Required:    true,
// 			},
// 			"ip": {
// 				Type:        schema.TypeString,
// 				Description: "IP address for the network interface.",
// 				// TODO: For the purposes of this demo we will stick to
// 				// auto-assigned IP addresses. In the future we will want
// 				// this value to be computed/optional or required.
// 				Computed: true,
// 			},
// 			"subnet_id": {
// 				Type:        schema.TypeString,
// 				Description: "ID of the VPC Subnet to which the interface belongs.",
// 				Computed:    true,
// 			},
// 			"vpc_id": {
// 				Type:        schema.TypeString,
// 				Description: "ID of the VPC in which to which the interface belongs.",
// 				Computed:    true,
// 			},
// 		},
// 	}
// }
//
// func createInstance(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
// 	client := meta.(*oxideSDK.Client)
//
// 	projectId := d.Get("project_id").(string)
// 	description := d.Get("description").(string)
// 	name := d.Get("name").(string)
// 	hostName := d.Get("host_name").(string)
// 	memory := d.Get("memory").(int)
// 	ncpus := d.Get("ncpus").(int)
//
// 	params := oxideSDK.InstanceCreateParams{
// 		Project: oxideSDK.NameOrId(projectId),
// 		Body: &oxideSDK.InstanceCreate{
// 			Description:       description,
// 			Name:              oxideSDK.Name(name),
// 			Hostname:          hostName,
// 			Memory:            oxideSDK.ByteCount(memory),
// 			Ncpus:             oxideSDK.InstanceCpuCount(ncpus),
// 			Disks:             newInstanceDiskAttach(d),
// 			ExternalIps:       newInstanceExternalIps(d),
// 			NetworkInterfaces: newNetworkInterface(d),
// 		},
// 	}
//
// 	resp, err := client.InstanceCreate(params)
// 	if err != nil {
// 		return diag.FromErr(err)
// 	}
//
// 	d.SetId(resp.Id)
//
// 	return readInstance(ctx, d, meta)
// }
//
// func readInstance(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
// 	client := meta.(*oxideSDK.Client)
// 	instanceId := d.Get("id").(string)
//
// 	params := oxideSDK.InstanceViewParams{Instance: oxideSDK.NameOrId(instanceId)}
// 	resp, err := client.InstanceView(params)
// 	if err != nil {
// 		return diag.FromErr(err)
// 	}
//
// 	if err := instanceToState(d, resp); err != nil {
// 		return diag.FromErr(err)
// 	}
//
// 	nis := d.Get("network_interface").([]interface{})
// 	if len(nis) > 0 {
// 		nicParams := oxideSDK.InstanceNetworkInterfaceListParams{
// 			Instance: oxideSDK.NameOrId(instanceId),
// 			Limit:    1000000000,
// 			SortBy:   oxideSDK.NameOrIdSortModeNameAscending,
// 		}
// 		resp2, err := client.InstanceNetworkInterfaceList(nicParams)
// 		if err != nil {
// 			return diag.FromErr(err)
// 		}
//
// 		networkInterfaces, err := networkInterfaceToState(client, *resp2)
// 		if err != nil {
// 			return diag.FromErr(err)
// 		}
//
// 		if err := d.Set("network_interface", networkInterfaces); err != nil {
// 			return diag.FromErr(err)
// 		}
// 	}
//
// 	return nil
// }
//
// func updateInstance(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
// 	// TODO: Currently there is no endpoint to update an instance. Update this function when such endpoint exists
// 	return diag.FromErr(errors.New("the oxide_instance resource currently does not support updates"))
// }
//
// func deleteInstance(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
// 	client := meta.(*oxideSDK.Client)
// 	instanceId := d.Get("id").(string)
//
// 	params := oxideSDK.InstanceStopParams{Instance: oxideSDK.NameOrId(instanceId)}
// 	_, err := client.InstanceStop(params)
// 	if err != nil {
// 		if is404(err) {
// 			d.SetId("")
// 			return nil
// 		}
// 		return diag.FromErr(err)
// 	}
//
// 	// Wait for instance to be stopped before attempting to destroy
// 	ch := make(chan error)
// 	go waitForStoppedInstance(client, oxideSDK.NameOrId(instanceId), ch)
// 	e := <-ch
// 	if e != nil {
// 		return diag.FromErr(e)
// 	}
//
// 	delParams := oxideSDK.InstanceDeleteParams{Instance: oxideSDK.NameOrId(instanceId)}
// 	if err := client.InstanceDelete(delParams); err != nil {
// 		if is404(err) {
// 			d.SetId("")
// 			return nil
// 		}
// 		return diag.FromErr(err)
// 	}
//
// 	d.SetId("")
// 	return nil
// }
//
// func instanceToState(d *schema.ResourceData, instance *oxideSDK.Instance) error {
// 	if err := d.Set("name", instance.Name); err != nil {
// 		return err
// 	}
// 	if err := d.Set("description", instance.Description); err != nil {
// 		return err
// 	}
// 	if err := d.Set("host_name", instance.Hostname); err != nil {
// 		return err
// 	}
// 	if err := d.Set("memory", instance.Memory); err != nil {
// 		return err
// 	}
// 	if err := d.Set("ncpus", instance.Ncpus); err != nil {
// 		return err
// 	}
// 	if err := d.Set("id", instance.Id); err != nil {
// 		return err
// 	}
// 	if err := d.Set("project_id", instance.ProjectId); err != nil {
// 		return err
// 	}
// 	if err := d.Set("run_state", instance.RunState); err != nil {
// 		return err
// 	}
// 	if err := d.Set("time_created", instance.TimeCreated.String()); err != nil {
// 		return err
// 	}
// 	if err := d.Set("time_modified", instance.TimeModified.String()); err != nil {
// 		return err
// 	}
// 	if err := d.Set("time_run_state_updated", instance.TimeRunStateUpdated.String()); err != nil {
// 		return err
// 	}
//
// 	return nil
// }
//
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
// 		time.Sleep(100 * time.Millisecond)
// 	}
// 	ch <- nil
// }
//
// func newInstanceExternalIps(d *schema.ResourceData) []oxideSDK.ExternalIpCreate {
// 	var externalIps = []oxideSDK.ExternalIpCreate{}
// 	ips := d.Get("external_ips").([]interface{})
//
// 	if len(ips) < 1 {
// 		return externalIps
// 	}
// 	for _, ip := range ips {
// 		ds := oxideSDK.ExternalIpCreate{
// 			PoolName: oxideSDK.Name(ip.(string)),
// 			// TODO: Implement other types when these are supported.
// 			Type: oxideSDK.ExternalIpCreateTypeEphemeral,
// 		}
//
// 		externalIps = append(externalIps, ds)
// 	}
//
// 	return externalIps
// }
//
// func newInstanceDiskAttach(d *schema.ResourceData) []oxideSDK.InstanceDiskAttachment {
// 	var diskAttachement = []oxideSDK.InstanceDiskAttachment{}
// 	disks := d.Get("attach_to_disks").([]interface{})
//
// 	if len(disks) < 1 {
// 		return diskAttachement
// 	}
// 	for _, disk := range disks {
// 		ds := oxideSDK.InstanceDiskAttachment{
// 			Name: oxideSDK.Name(disk.(string)),
// 			Type: "attach",
// 		}
//
// 		diskAttachement = append(diskAttachement, ds)
// 	}
//
// 	return diskAttachement
// }
//
// func newNetworkInterface(d *schema.ResourceData) oxideSDK.InstanceNetworkInterfaceAttachment {
// 	nis := d.Get("network_interface").([]interface{})
//
// 	if len(nis) < 1 {
// 		return oxideSDK.InstanceNetworkInterfaceAttachment{
// 			Type: "none",
// 		}
// 	}
//
// 	var interfaceCreate = []oxideSDK.InstanceNetworkInterfaceCreate{}
// 	for _, ni := range nis {
// 		nwInterface := ni.(map[string]interface{})
//
// 		nwInterfaceCreate := oxideSDK.InstanceNetworkInterfaceCreate{
// 			Description: nwInterface["description"].(string),
// 			Name:        oxideSDK.Name(nwInterface["name"].(string)),
// 			// TODO: Ideally from the API we should be able to create with IDs, not names
// 			SubnetName: oxideSDK.Name(nwInterface["subnet_name"].(string)),
// 			VpcName:    oxideSDK.Name(nwInterface["vpc_name"].(string)),
// 		}
//
// 		interfaceCreate = append(interfaceCreate, nwInterfaceCreate)
// 	}
//
// 	return oxideSDK.InstanceNetworkInterfaceAttachment{
// 		Params: interfaceCreate,
// 		Type:   "create",
// 	}
// }
//
// func networkInterfaceToState(client *oxideSDK.Client, nwInterface oxideSDK.InstanceNetworkInterfaceResultsPage) ([]interface{}, error) {
// 	items := nwInterface.Items
// 	var result = make([]interface{}, 0, len(items))
// 	for _, item := range items {
// 		var m = make(map[string]interface{})
//
// 		m["description"] = item.Description
// 		m["ip"] = item.Ip
// 		m["name"] = item.Name
// 		m["subnet_id"] = item.SubnetId
// 		m["vpc_id"] = item.VpcId
//
// 		// Ideally the NetworkInterface struct would contain the names of the VPC and subnet.
// 		// For now they only give the ID so we'll retrieve the names separately.
// 		params := oxideSDK.VpcViewParams{Vpc: oxideSDK.NameOrId(item.VpcId)}
// 		vpcResp, err := client.VpcView(params)
// 		if err != nil {
// 			return nil, err
// 		}
//
// 		subnetParams := oxideSDK.VpcSubnetViewParams{Subnet: oxideSDK.NameOrId(item.SubnetId)}
// 		subnetResp, err := client.VpcSubnetView(subnetParams)
// 		if err != nil {
// 			return nil, err
// 		}
//
// 		m["subnet_name"] = subnetResp.Name
// 		m["vpc_name"] = vpcResp.Name
//
// 		result = append(result, m)
// 	}
//
// 	return result, nil
// }
//

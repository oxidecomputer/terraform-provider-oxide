// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	oxideSDK "github.com/oxidecomputer/oxide.go/oxide"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = (*instanceNICResource)(nil)
	_ resource.ResourceWithConfigure = (*instanceNICResource)(nil)
)

// NewInstanceNetworkInterfaceResource is a helper function to simplify the provider implementation.
func NewInstanceNetworkInterfaceResource() resource.Resource {
	return &instanceNICResource{}
}

// instanceNICResource is the resource implementation.
type instanceNICResource struct {
	client *oxideSDK.Client
}

type instanceNICResourceModel struct {
	Description  types.String   `tfsdk:"description"`
	ID           types.String   `tfsdk:"id"`
	IPAddr       types.String   `tfsdk:"ip_address"`
	InstanceID   types.String   `tfsdk:"instance_id"`
	MAC          types.String   `tfsdk:"mac_address"`
	Name         types.String   `tfsdk:"name"`
	Primary      types.Bool     `tfsdk:"primary"`
	SubnetID     types.String   `tfsdk:"subnet_id"`
	TimeCreated  types.String   `tfsdk:"time_created"`
	TimeModified types.String   `tfsdk:"time_modified"`
	Timeouts     timeouts.Value `tfsdk:"timeouts"`
	VPCID        types.String   `tfsdk:"vpc_id"`
}

// Metadata returns the resource type name.
func (r *instanceNICResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "oxide_instance_network_interface"
}

// Configure adds the provider configured client to the data source.
func (r *instanceNICResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*oxideSDK.Client)
}

func (r *instanceNICResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Schema defines the schema for the resource.
func (r *instanceNICResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"instance_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the instance to which the network interface will belong to.",
			},
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
			},
			"vpc_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the VPC in which to create the instance network interface",
			},
			"ip_address": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Description: "IP address for the instance network interface. " +
					"One will be auto-assigned if not provided.",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Update: true,
				Delete: true,
			}),
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
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *instanceNICResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan instanceNICResourceModel

	// Read Terraform plan data into the model
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

	// This is an unfortunate result of having the create body use names as identifiers
	// but the body return IDs. making two API calls to retrieve VPC and subnet names
	// Using IDs only for the provider schema as names are mutable.
	vpc, err := r.client.VpcView(oxideSDK.VpcViewParams{
		Vpc: oxideSDK.NameOrId(plan.VPCID.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read information about corresponding VPC:",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("read VPC with ID: %v", plan.VPCID.ValueString()), map[string]any{"success": true})

	subnet, err := r.client.VpcSubnetView(oxideSDK.VpcSubnetViewParams{
		Subnet: oxideSDK.NameOrId(plan.SubnetID.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read information about corresponding subnet:",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("read subnet with ID: %v", plan.SubnetID.ValueString()), map[string]any{"success": true})

	// Stop instance so we can attach network interface
	_, err = r.client.InstanceStop(oxideSDK.InstanceStopParams{
		Instance: oxideSDK.NameOrId(plan.InstanceID.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to stop associated instance:",
			"API error: "+err.Error(),
		)
		return
	}

	// Wait for instance to be stopped before attempting to create NIC
	ch := make(chan error)
	go waitForStoppedInstance(r.client, oxideSDK.NameOrId(plan.InstanceID.ValueString()), ch)
	e := <-ch
	if !is404(e) {
		resp.Diagnostics.AddError(
			"Unable to stop instance:",
			"API error: "+e.Error(),
		)
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("stopped instance with ID: %v", plan.InstanceID.ValueString()), map[string]any{"success": true})

	params := oxideSDK.InstanceNetworkInterfaceCreateParams{
		Instance: oxideSDK.NameOrId(plan.InstanceID.ValueString()),
		Body: &oxideSDK.InstanceNetworkInterfaceCreate{
			Description: plan.Description.ValueString(),
			Ip:          plan.IPAddr.ValueString(),
			Name:        oxideSDK.Name(plan.Name.ValueString()),
			SubnetName:  subnet.Name,
			VpcName:     vpc.Name,
		},
	}
	nic, err := r.client.InstanceNetworkInterfaceCreate(params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating instance network interface",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("created instance network interface with ID: %v", nic.Id), map[string]any{"success": true})

	// Start instance again after attaching NIC
	_, err = r.client.InstanceStart(oxideSDK.InstanceStartParams{
		Instance: oxideSDK.NameOrId(plan.InstanceID.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to start associated instance:",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("started instance with ID: %v", plan.InstanceID.ValueString()), map[string]any{"success": true})

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(nic.Id)
	plan.TimeCreated = types.StringValue(nic.TimeCreated.String())
	plan.TimeModified = types.StringValue(nic.TimeCreated.String())
	plan.MAC = types.StringValue(string(nic.Mac))
	plan.Primary = types.BoolValue(nic.Primary)
	// Setting IPAddress as it is both computed and optional
	plan.IPAddr = types.StringValue(nic.Ip)

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *instanceNICResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state instanceNICResourceModel

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

	nic, err := r.client.InstanceNetworkInterfaceView(oxideSDK.InstanceNetworkInterfaceViewParams{
		Interface: oxideSDK.NameOrId(state.ID.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read instance network interface:",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("read instance network interface with ID: %v", nic.Id), map[string]any{"success": true})

	state.Description = types.StringValue(nic.Description)
	state.ID = types.StringValue(nic.Id)
	state.IPAddr = types.StringValue(nic.Ip)
	state.InstanceID = types.StringValue(nic.InstanceId)
	state.MAC = types.StringValue(nic.Id)
	state.Name = types.StringValue(string(nic.Name))
	state.Primary = types.BoolValue(nic.Primary)
	state.SubnetID = types.StringValue(nic.SubnetId)
	state.TimeCreated = types.StringValue(nic.TimeCreated.String())
	state.TimeModified = types.StringValue(nic.TimeCreated.String())
	state.VPCID = types.StringValue(nic.VpcId)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *instanceNICResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan instanceNICResourceModel
	var state instanceNICResourceModel

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

	// TODO: Look into plan modifiers to see if they are fit for purpose
	// Check if plan has changed and return error for fields that cannot
	// be changed
	if !plan.VPCID.Equal(state.VPCID) {
		resp.Diagnostics.AddError(
			"Error updating instance network interface:",
			"vpc_id cannot be modified",
		)
		return
	}
	if !plan.SubnetID.Equal(state.SubnetID) {
		resp.Diagnostics.AddError(
			"Error updating instance network interface:",
			"subnet_id cannot be modified",
		)
		return
	}
	// TODO: This doesn't work the same as the old diff validators.
	// It registers a change if the user chose not to specify an IPv6
	// prefix at all. Investigate how to validate this change

	//if !plan.IPAddr.Equal(state.IPAddr) {
	//	resp.Diagnostics.AddError(
	//		"Error updating instance network interface::",
	//		"ip_address cannot be modified",
	//	)
	//	return
	// }

	params := oxideSDK.InstanceNetworkInterfaceUpdateParams{
		Interface: oxideSDK.NameOrId(state.ID.ValueString()),
		Body: &oxideSDK.InstanceNetworkInterfaceUpdate{
			Description: plan.Description.ValueString(),
			Name:        oxideSDK.Name(plan.Name.ValueString()),
		},
	}
	nic, err := r.client.InstanceNetworkInterfaceUpdate(params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating instance network interface",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("updated instance network interface with ID: %v", nic.Id), map[string]any{"success": true})

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(nic.Id)
	plan.TimeCreated = types.StringValue(nic.TimeCreated.String())
	plan.TimeModified = types.StringValue(nic.TimeCreated.String())
	plan.MAC = types.StringValue(string(nic.Mac))
	plan.Primary = types.BoolValue(nic.Primary)
	// Setting IPAddress as it is both computed and optional
	plan.IPAddr = types.StringValue(nic.Ip)

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *instanceNICResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state instanceNICResourceModel

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

	if err := r.client.InstanceNetworkInterfaceDelete(oxideSDK.InstanceNetworkInterfaceDeleteParams{
		Interface: oxideSDK.NameOrId(state.ID.ValueString()),
	}); err != nil {
		if !is404(err) {
			resp.Diagnostics.AddError(
				"Error deleting instance network interface:",
				"API error: "+err.Error(),
			)
			return
		}
	}
	tflog.Trace(ctx, fmt.Sprintf("deleted instance network interface with ID: %v", state.ID.ValueString()), map[string]any{"success": true})
}

func waitForStoppedInstance(client *oxideSDK.Client, instanceID oxideSDK.NameOrId, ch chan error) {
	for {
		params := oxideSDK.InstanceViewParams{Instance: instanceID}
		resp, err := client.InstanceView(params)
		if err != nil {
			ch <- err
		}
		if resp.RunState == oxideSDK.InstanceStateStopped {
			break
		}
		// Suggested alternatives suggested by linter are not fit for purpose
		//lintignore:R018
		time.Sleep(time.Second)
	}
	ch <- nil
}

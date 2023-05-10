// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
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
	AttachToDisks types.List     `tfsdk:"attach_to_disks"`
	Description   types.String   `tfsdk:"description"`
	ExternalIPs   types.List     `tfsdk:"external_ips"`
	HostName      types.String   `tfsdk:"host_name"`
	ID            types.String   `tfsdk:"id"`
	Memory        types.Int64    `tfsdk:"memory"`
	Name          types.String   `tfsdk:"name"`
	NCPUs         types.Int64    `tfsdk:"ncpus"`
	ProjectID     types.String   `tfsdk:"project_id"`
	StartOnCreate types.Bool     `tfsdk:"start_on_create"`
	TimeCreated   types.String   `tfsdk:"time_created"`
	TimeModified  types.String   `tfsdk:"time_modified"`
	Timeouts      timeouts.Value `tfsdk:"timeouts"`
	UserData      types.String   `tfsdk:"user_data"`
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
			"attach_to_disks": schema.ListAttribute{
				Optional:    true,
				Description: "Disks to be attached to this instance.",
				ElementType: types.StringType,
				// TODO: Remove with https://github.com/oxidecomputer/terraform-provider-oxide/issues/82
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
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
			Start:       plan.StartOnCreate.ValueBool(),
			// Creating and attaching nics on instance create limits our
			// ability to perform CRUD actions reliably on them (e.g. no information
			// about the nics is returned on instance creation, so we'd have to
			// make additional API calls to retrieve that information based
			// on names instead of IDs) and increases complexity of the resource.
			// NICs have their own APIs, it makes sense that they have their own
			// resource at `oxide_instance_network_interface`
			NetworkInterfaces: oxideSDK.InstanceNetworkInterfaceAttachment{
				Type: oxideSDK.InstanceNetworkInterfaceAttachmentTypeNone,
			},
			UserData: plan.UserData.ValueString(),
		},
	}

	// TODO: Perhaps it makes sense to attach after the resource is created instead of
	// making it part of the create body.
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
			// We will only be attaching through terraform
			// any disk creation should be done through
			// the `oxide_disk` resource
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
	plan.TimeModified = types.StringValue(instance.TimeCreated.String())

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
	state.TimeModified = types.StringValue(instance.TimeCreated.String())

	//state.AttachToDisks = TODO

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

	// Detach any associated disks
	disks, err := r.client.InstanceDiskList(oxideSDK.InstanceDiskListParams{
		Instance: oxideSDK.NameOrId(state.ID.ValueString()),
		Limit:    1000000000,
	})
	if err != nil {
		if !is404(err) {
			resp.Diagnostics.AddError(
				"Unable to list attached disks:",
				"API error: "+err.Error(),
			)
			return
		}
	}
	tflog.Trace(ctx, fmt.Sprintf("listed all attached disks from instance with ID: %v", state.ID.ValueString()), map[string]any{"success": true})

	for _, disk := range disks.Items {
		_, err = r.client.InstanceDiskDetach(
			oxideSDK.InstanceDiskDetachParams{
				Instance: oxideSDK.NameOrId(state.ID.ValueString()),
				Body:     &oxideSDK.DiskPath{Disk: oxideSDK.NameOrId(disk.Id)},
			},
		)
		if err != nil {
			if !is404(err) {
				resp.Diagnostics.AddError(
					"Unable to detach disk:",
					"API error: "+err.Error(),
				)
				return
			}
		}
		tflog.Trace(ctx, fmt.Sprintf("detached disk with ID: %v", disk.Id), map[string]any{"success": true})
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

	// TODO: Double check if this is necessary
	// instance appears to stay in "stopping" state indefinitely
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

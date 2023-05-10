// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	oxideSDK "github.com/oxidecomputer/oxide.go/oxide"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = (*instanceDiskAttachmentResource)(nil)
	_ resource.ResourceWithConfigure = (*instanceDiskAttachmentResource)(nil)
)

// NewInstanceDiskAttachmentResource is a helper function to simplify the provider implementation.
func NewInstanceDiskAttachmentResource() resource.Resource {
	return &instanceDiskAttachmentResource{}
}

// instanceDiskAttachmentResource is the resource implementation.
type instanceDiskAttachmentResource struct {
	client *oxideSDK.Client
}

type instanceDiskAttachmentResourceModel struct {
	ID         types.String   `tfsdk:"id"`
	InstanceID types.String   `tfsdk:"instance_id"`
	DiskID     types.String   `tfsdk:"disk_id"`
	DiskName   types.String   `tfsdk:"disk_name"`
	Timeouts   timeouts.Value `tfsdk:"timeouts"`
}

// Metadata returns the resource type name.
func (r *instanceDiskAttachmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "oxide_instance_disk_attachment"
}

// Configure adds the provider configured client to the data source.
func (r *instanceDiskAttachmentResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*oxideSDK.Client)
}

// TODO: To handle imports these would have to be through disk ID
// func (r *instanceDiskAttachmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
// 	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
// }

// Schema defines the schema for the resource.
func (r *instanceDiskAttachmentResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"instance_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the instance the disk will be attached to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"disk_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the disk to be attached.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Delete: true,
			}),
			"disk_name": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the disk that is attached to the designated instance.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier of the terraform resource.",
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *instanceDiskAttachmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan instanceDiskAttachmentResourceModel

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

	params := oxideSDK.InstanceDiskAttachParams{
		Instance: oxideSDK.NameOrId(plan.InstanceID.ValueString()),
		Body: &oxideSDK.DiskPath{
			Disk: oxideSDK.NameOrId(plan.DiskID.ValueString()),
		},
	}
	disk, err := r.client.InstanceDiskAttach(params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error attaching disk",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(ctx, fmt.Sprintf(
		"attached disk with ID '%v' to instance with ID '%v'.",
		disk.Id, disk.State.Instance,
	),
		map[string]any{"success": true})

	// Map response body to schema and populate Computed attribute values
	plan.DiskName = types.StringValue(string(disk.Name))
	// Set a unique ID for the resource state
	plan.ID = types.StringValue(uuid.New().String())

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *instanceDiskAttachmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state instanceDiskAttachmentResourceModel

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

	disk, err := r.client.DiskView(oxideSDK.DiskViewParams{
		Disk: oxideSDK.NameOrId(state.DiskID.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read disk:",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("read disk with ID: %v", disk.Id), map[string]any{"success": true})

	state.DiskID = types.StringValue(disk.Id)
	state.InstanceID = types.StringValue(disk.State.Instance)
	state.DiskName = types.StringValue(string(disk.Name))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *instanceDiskAttachmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Error updating instance disk attachment",
		"the oxide API currently does not support updating instance disk attachments")
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *instanceDiskAttachmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state instanceDiskAttachmentResourceModel

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

	disk, err := r.client.InstanceDiskDetach(oxideSDK.InstanceDiskDetachParams{
		Instance: oxideSDK.NameOrId(state.InstanceID.ValueString()),
		Body:     &oxideSDK.DiskPath{Disk: oxideSDK.NameOrId(state.DiskID.ValueString())},
	})
	if err != nil {
		if !is404(err) {
			resp.Diagnostics.AddError(
				"Error detaching disk:",
				"API error: "+err.Error(),
			)
			return
		}
	}
	tflog.Trace(ctx, fmt.Sprintf(
		"detached disk with ID '%v' from instance with ID '%v'.",
		disk.Id, disk.State.Instance,
	),
		map[string]any{"success": true})
}

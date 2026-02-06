// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/oxidecomputer/oxide.go/oxide"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = (*externalSubnetAttachmentResource)(nil)
	_ resource.ResourceWithConfigure   = (*externalSubnetAttachmentResource)(nil)
	_ resource.ResourceWithImportState = (*externalSubnetAttachmentResource)(nil)
)

// NewExternalSubnetAttachmentResource is a helper function to simplify the provider implementation.
func NewExternalSubnetAttachmentResource() resource.Resource {
	return &externalSubnetAttachmentResource{}
}

// externalSubnetAttachmentResource is the resource implementation.
type externalSubnetAttachmentResource struct {
	client *oxide.Client
}

type externalSubnetAttachmentResourceModel struct {
	ID               types.String   `tfsdk:"id"`
	ExternalSubnetID types.String   `tfsdk:"external_subnet_id"`
	InstanceID       types.String   `tfsdk:"instance_id"`
	Timeouts         timeouts.Value `tfsdk:"timeouts"`
}

// Metadata returns the resource type name.
func (r *externalSubnetAttachmentResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = "oxide_external_subnet_attachment"
}

// Configure adds the provider configured client to the resource.
func (r *externalSubnetAttachmentResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*oxide.Client)
}

// ImportState imports an external subnet attachment using the external subnet ID.
func (r *externalSubnetAttachmentResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Schema defines the schema for the resource.
func (r *externalSubnetAttachmentResource) Schema(
	ctx context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource manages the attachment of an external subnet to an instance.",
		Attributes: map[string]schema.Attribute{
			// External subnet attachments don't have their own IDs, and a given external subnet can
			// be attached to at most one instance, so we use the instance ID as the ID of the
			// attachment.
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for the attachment. Set to the external subnet ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"external_subnet_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the external subnet to attach.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"instance_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the instance to attach the external subnet to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Delete: true,
			}),
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *externalSubnetAttachmentResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan externalSubnetAttachmentResourceModel

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

	params := oxide.ExternalSubnetAttachParams{
		ExternalSubnet: oxide.NameOrId(plan.ExternalSubnetID.ValueString()),
		Body: &oxide.ExternalSubnetAttach{
			Instance: oxide.NameOrId(plan.InstanceID.ValueString()),
		},
	}

	externalSubnet, err := r.client.ExternalSubnetAttach(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error attaching external subnet",
			"API error: "+err.Error(),
		)
		return
	}

	tflog.Trace(
		ctx,
		fmt.Sprintf(
			"attached external subnet %v to instance %v",
			externalSubnet.Id,
			externalSubnet.InstanceId,
		),
		map[string]any{"success": true},
	)

	plan.ID = types.StringValue(externalSubnet.Id)
	plan.ExternalSubnetID = types.StringValue(externalSubnet.Id)
	plan.InstanceID = types.StringValue(externalSubnet.InstanceId)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *externalSubnetAttachmentResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state externalSubnetAttachmentResourceModel

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

	params := oxide.ExternalSubnetViewParams{
		ExternalSubnet: oxide.NameOrId(state.ID.ValueString()),
	}

	externalSubnet, err := r.client.ExternalSubnetView(ctx, params)
	if err != nil {
		if is404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to read external subnet attachment:",
			"API error: "+err.Error(),
		)
		return
	}

	// If the subnet is no longer attached to any instance, remove from state.
	if externalSubnet.InstanceId == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	tflog.Trace(
		ctx,
		fmt.Sprintf("read external subnet attachment with ID: %v", externalSubnet.Id),
		map[string]any{"success": true},
	)

	state.ID = types.StringValue(externalSubnet.Id)
	state.ExternalSubnetID = types.StringValue(externalSubnet.Id)
	state.InstanceID = types.StringValue(externalSubnet.InstanceId)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
// Only timeouts can change in-place; both mutable attributes trigger replacement.
func (r *externalSubnetAttachmentResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan externalSubnetAttachmentResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *externalSubnetAttachmentResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state externalSubnetAttachmentResourceModel

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

	// Only detach if the subnet is still attached to the expected
	// instance. If it's gone, already detached, or re-attached to
	// a different instance out of band, there's nothing to do.
	viewParams := oxide.ExternalSubnetViewParams{
		ExternalSubnet: oxide.NameOrId(state.ID.ValueString()),
	}
	externalSubnet, err := r.client.ExternalSubnetView(
		ctx, viewParams,
	)
	if err != nil {
		if is404(err) {
			return
		}
		resp.Diagnostics.AddError(
			"Error reading external subnet during delete:",
			"API error: "+err.Error(),
		)
		return
	}

	if externalSubnet.InstanceId != state.InstanceID.ValueString() {
		return
	}

	detachParams := oxide.ExternalSubnetDetachParams{
		ExternalSubnet: oxide.NameOrId(state.ID.ValueString()),
	}
	if _, err := r.client.ExternalSubnetDetach(
		ctx, detachParams,
	); err != nil {
		if !is404(err) {
			resp.Diagnostics.AddError(
				"Error detaching external subnet:",
				"API error: "+err.Error(),
			)
			return
		}
	}

	tflog.Trace(
		ctx,
		fmt.Sprintf("detached external subnet with ID: %v", state.ID.ValueString()),
		map[string]any{"success": true},
	)
}

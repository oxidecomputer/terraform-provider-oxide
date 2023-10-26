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
	_ resource.Resource              = (*sshKeyResource)(nil)
	_ resource.ResourceWithConfigure = (*sshKeyResource)(nil)
)

// NewSSHKeyResource is a helper function to simplify the provider implementation.
func NewSSHKeyResource() resource.Resource {
	return &sshKeyResource{}
}

// sshKeyResource is the resource implementation.
type sshKeyResource struct {
	client *oxide.Client
}

// sshKeyResourceModel are the attributes that are supported on this resource.
type sshKeyResourceModel struct {
	ID           types.String   `tfsdk:"id"`
	Name         types.String   `tfsdk:"name"`
	Description  types.String   `tfsdk:"description"`
	PublicKey    types.String   `tfsdk:"public_key"`
	SiloUserID   types.String   `tfsdk:"silo_user_id"`
	TimeCreated  types.String   `tfsdk:"time_created"`
	TimeModified types.String   `tfsdk:"time_modified"`
	Timeouts     timeouts.Value `tfsdk:"timeouts"`
}

// Metadata sets the resource type name.
func (r *sshKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "oxide_ssh_key"
}

// Configure adds the provider configured client to the resource.
func (r *sshKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*oxide.Client)
}

// ImportState configures the resource to be imported by its ID.
func (r *sshKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Schema defines the schema for the resource.
func (r *sshKeyResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier of the SSH key.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the SSH key.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Required:    true,
				Description: "Description for the SSH key.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"public_key": schema.StringAttribute{
				Required:    true,
				Description: "Public SSH key.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"silo_user_id": schema.StringAttribute{
				Computed:    true,
				Description: "User ID that owns this SSH key.",
			},
			"time_created": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this SSH key was created.",
			},
			"time_modified": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this SSH key was last modified.",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *sshKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan sshKeyResourceModel

	// Read Terraform plan data into the model.
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

	params := oxide.CurrentUserSshKeyCreateParams{
		Body: &oxide.SshKeyCreate{
			Description: plan.Description.ValueString(),
			Name:        oxide.Name(plan.Name.ValueString()),
			PublicKey:   plan.PublicKey.ValueString(),
		},
	}

	sshKey, err := r.client.CurrentUserSshKeyCreate(params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating SSH key",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("created SSH key with ID: %v", sshKey.Id), map[string]any{"success": true})

	// Map response body to schema and populate computed attribute values.
	plan.ID = types.StringValue(sshKey.Id)
	plan.SiloUserID = types.StringValue(sshKey.SiloUserId)
	plan.TimeCreated = types.StringValue(sshKey.TimeCreated.String())
	plan.TimeModified = types.StringValue(sshKey.TimeModified.String())

	// Save plan into Terraform state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *sshKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state sshKeyResourceModel

	// Read Terraform prior state data into the model.
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

	sshKey, err := r.client.CurrentUserSshKeyView(oxide.CurrentUserSshKeyViewParams{
		SshKey: oxide.NameOrId(state.ID.ValueString()),
	})
	if err != nil {
		if is404(err) {
			// Remove resource from state during a refresh
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to read SSH key:",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("read SSH key with ID: %v", sshKey.Id), map[string]any{"success": true})

	state.Description = types.StringValue(sshKey.Description)
	state.ID = types.StringValue(sshKey.Id)
	state.Name = types.StringValue(string(sshKey.Name))
	state.PublicKey = types.StringValue(string(sshKey.PublicKey))
	state.SiloUserID = types.StringValue(string(sshKey.SiloUserId))
	state.TimeCreated = types.StringValue(sshKey.TimeCreated.String())
	state.TimeModified = types.StringValue(sshKey.TimeModified.String())

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update is intentionally unimplemented since SSH keys do not have an update
// API. All of its configurable attributes are marked as requiring replacement
// to tell Terraform to destroy and create this resource upon change to its
// attributes. If an update API is created in the future this method should be
// implemented.
func (r *sshKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *sshKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state sshKeyResourceModel

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

	if err := r.client.CurrentUserSshKeyDelete(oxide.CurrentUserSshKeyDeleteParams{
		SshKey: oxide.NameOrId(state.ID.ValueString()),
	}); err != nil {
		if !is404(err) {
			resp.Diagnostics.AddError(
				"Error deleting SSH key:",
				"API error: "+err.Error(),
			)
			return
		}
	}
	tflog.Trace(ctx, fmt.Sprintf("deleted SSH key with ID: %v", state.ID.ValueString()), map[string]any{"success": true})
}

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ippoolsilolink

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/oxidecomputer/oxide.go/oxide"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/shared"
	oxidevalidator "github.com/oxidecomputer/terraform-provider-oxide/internal/provider/validator"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = (*Resource)(nil)
	_ resource.ResourceWithConfigure = (*Resource)(nil)
)

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &Resource{}
}

// Resource is the resource implementation.
type Resource struct {
	client *oxide.Client
}

type ResourceModel struct {
	ID        types.String   `tfsdk:"id"`
	SiloID    types.String   `tfsdk:"silo_id"`
	IPPoolID  types.String   `tfsdk:"ip_pool_id"`
	IsDefault types.Bool     `tfsdk:"is_default"`
	Timeouts  timeouts.Value `tfsdk:"timeouts"`
}

// Metadata returns the resource type name.
func (r *Resource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = "oxide_ip_pool_silo_link"
}

// Configure adds the provider configured client to the data source.
func (r *Resource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*oxide.Client)
}

// ImportState imports an existing resource into Terraform state.
func (r *Resource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	idParts := strings.Split(req.ID, "/")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID format: ip_pool_id/silo_id, got: %s", req.ID),
		)
		return
	}

	// Use the import ID directly as the terraform ID (it's already in the correct format)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(
		resp.State.SetAttribute(ctx, path.Root("ip_pool_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("silo_id"), idParts[1])...)
}

// Schema defines the schema for the resource.
func (r *Resource) Schema(
	ctx context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
This resource manages IP pool to silo links.
`,
		Attributes: map[string]schema.Attribute{
			"silo_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the silo to link the IP pool to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					oxidevalidator.IsUUID(),
				},
			},
			"ip_pool_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the IP pool that will be linked to the silo.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					oxidevalidator.IsUUID(),
				},
			},
			"is_default": schema.BoolAttribute{
				Required:    true,
				Description: "Whether this is the default IP pool for a silo. Only a single IP pool silo link can be marked as default.",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Update: true,
				Delete: true,
			}),
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier of the IP pool silo link.",
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *Resource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan ResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := plan.Timeouts.Create(ctx, shared.DefaultTimeout())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	params := oxide.SystemIpPoolSiloLinkParams{
		Pool: oxide.NameOrId(plan.IPPoolID.ValueString()),
		Body: &oxide.IpPoolLinkSilo{
			IsDefault: plan.IsDefault.ValueBoolPointer(),
			Silo:      oxide.NameOrId(plan.SiloID.ValueString()),
		},
	}
	link, err := r.client.SystemIpPoolSiloLink(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating IP pool silo link",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(
		ctx,
		fmt.Sprintf("created IP pool silo link for IP pool: %v", link.IpPoolId),
		map[string]any{"success": true},
	)

	// Set a deterministic ID based on composite attributes.
	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", link.IpPoolId, link.SiloId))

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *Resource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state ResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	readTimeout, diags := state.Timeouts.Read(ctx, shared.DefaultTimeout())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()

	// This `/v1/system/silos/{silo}/ip-pools` API works with non-discoverable silos
	// whereas the `/v1/system/ip-pools/{pool}/silos` does not.
	pools, err := r.client.SiloIpPoolListAllPages(ctx, oxide.SiloIpPoolListParams{
		Silo: oxide.NameOrId(state.SiloID.ValueString()),
	})
	if err != nil {
		if shared.Is404(err) {
			// Remove resource from state during a refresh
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to read links:",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(
		ctx,
		fmt.Sprintf("read IP pool links for silo: %v", state.SiloID.ValueString()),
		map[string]any{"success": true},
	)

	ipPoolID := state.IPPoolID.ValueString()
	idx := slices.IndexFunc(
		pools,
		func(p oxide.SiloIpPool) bool {
			// We check for both ID and name equality to ensure resources that mistakenly
			// used the IP pool name aren't removed from state.
			return p.Id == ipPoolID || p.Name == oxide.Name(ipPoolID)
		},
	)
	if idx < 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	// Resolve the silo to its UUID so the composite ID is always IP_POOL_ID/SILO_ID
	// in UUID form, even when silo_id was previously configured by name.
	silo, err := r.client.SiloView(ctx, oxide.SiloViewParams{
		Silo: oxide.NameOrId(state.SiloID.ValueString()),
	})
	if err != nil {
		if shared.Is404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to read silo:",
			"API error: "+err.Error(),
		)
		return
	}

	// Set a deterministic ID based on composite attributes.
	state.ID = types.StringValue(
		fmt.Sprintf("%s/%s", pools[idx].Id, silo.Id),
	)

	state.SiloID = types.StringValue(silo.Id)
	state.IPPoolID = types.StringValue(pools[idx].Id)
	state.IsDefault = types.BoolPointerValue(pools[idx].IsDefault)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *Resource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan ResourceModel
	var state ResourceModel

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

	updateTimeout, diags := plan.Timeouts.Update(ctx, shared.DefaultTimeout())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	params := oxide.SystemIpPoolSiloUpdateParams{
		Pool: oxide.NameOrId(state.IPPoolID.ValueString()),
		Silo: oxide.NameOrId(state.SiloID.ValueString()),
		Body: &oxide.IpPoolSiloUpdate{
			IsDefault: plan.IsDefault.ValueBoolPointer(),
		},
	}
	link, err := r.client.SystemIpPoolSiloUpdate(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating link",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(
		ctx,
		fmt.Sprintf("updated IP pool silo link for IP pool: %v", link.IpPoolId),
		map[string]any{"success": true},
	)

	// Set a deterministic ID based on composite attributes.
	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", link.IpPoolId, link.SiloId))

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *Resource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state ResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := state.Timeouts.Delete(ctx, shared.DefaultTimeout())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	params := oxide.SystemIpPoolSiloUnlinkParams{
		Pool: oxide.NameOrId(state.IPPoolID.ValueString()),
		Silo: oxide.NameOrId(state.SiloID.ValueString()),
	}
	if err := r.client.SystemIpPoolSiloUnlink(ctx, params); err != nil {
		if !shared.Is404(err) {
			resp.Diagnostics.AddError(
				"Error deleting link:",
				"API error: "+err.Error(),
			)
			return
		}
	}
	tflog.Trace(
		ctx,
		fmt.Sprintf("deleted link with ID: %v", state.ID.ValueString()),
		map[string]any{"success": true},
	)
}

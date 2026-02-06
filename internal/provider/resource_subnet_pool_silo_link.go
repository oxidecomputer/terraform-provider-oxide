// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

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
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/oxidecomputer/oxide.go/oxide"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = (*subnetPoolSiloLinkResource)(nil)
	_ resource.ResourceWithConfigure = (*subnetPoolSiloLinkResource)(nil)
)

// NewSubnetPoolSiloLinkResource is a helper function to simplify the provider implementation.
func NewSubnetPoolSiloLinkResource() resource.Resource {
	return &subnetPoolSiloLinkResource{}
}

// subnetPoolSiloLinkResource is the resource implementation.
type subnetPoolSiloLinkResource struct {
	client *oxide.Client
}

type subnetPoolSiloLinkResourceModel struct {
	ID           types.String   `tfsdk:"id"`
	SiloID       types.String   `tfsdk:"silo_id"`
	SubnetPoolID types.String   `tfsdk:"subnet_pool_id"`
	IsDefault    types.Bool     `tfsdk:"is_default"`
	Timeouts     timeouts.Value `tfsdk:"timeouts"`
}

// Metadata returns the resource type name.
func (r *subnetPoolSiloLinkResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = "oxide_subnet_pool_silo_link"
}

// Configure adds the provider configured client to the data source.
func (r *subnetPoolSiloLinkResource) Configure(
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
func (r *subnetPoolSiloLinkResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	idParts := strings.Split(req.ID, "/")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID format: subnet_pool_id/silo_id, got: %s", req.ID),
		)
		return
	}

	// Use the import ID directly as the terraform ID (it's already in the correct format)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(
		resp.State.SetAttribute(ctx, path.Root("subnet_pool_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("silo_id"), idParts[1])...)
}

// Schema defines the schema for the resource.
func (r *subnetPoolSiloLinkResource) Schema(
	ctx context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource manages subnet pool to silo links.",
		Attributes: map[string]schema.Attribute{
			"silo_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the silo to link the subnet pool to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"subnet_pool_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the subnet pool that will be linked to the silo.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"is_default": schema.BoolAttribute{
				Required:    true,
				Description: "Whether this is the default subnet pool for the silo. When true, external subnet allocations that don't specify a pool use this one.",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Update: true,
				Delete: true,
			}),
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Identifier for this resource, formatted as `subnet_pool_id/silo_id`.",
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *subnetPoolSiloLinkResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan subnetPoolSiloLinkResourceModel

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

	params := oxide.SubnetPoolSiloLinkParams{
		Pool: oxide.NameOrId(plan.SubnetPoolID.ValueString()),
		Body: &oxide.SubnetPoolLinkSilo{
			IsDefault: plan.IsDefault.ValueBoolPointer(),
			Silo:      oxide.NameOrId(plan.SiloID.ValueString()),
		},
	}
	link, err := r.client.SubnetPoolSiloLink(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating subnet pool silo link",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(
		ctx,
		fmt.Sprintf("created subnet pool silo link for subnet pool: %v", link.SubnetPoolId),
		map[string]any{"success": true},
	)

	// Set a deterministic ID based on the composite key (pool_id/silo_id)
	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", link.SubnetPoolId, link.SiloId))

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *subnetPoolSiloLinkResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state subnetPoolSiloLinkResourceModel

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

	pools, err := r.client.SiloSubnetPoolListAllPages(ctx, oxide.SiloSubnetPoolListParams{
		Silo: oxide.NameOrId(state.SiloID.ValueString()),
	})
	if err != nil {
		if is404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to read subnet pool silo links:",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(
		ctx,
		fmt.Sprintf("read subnet pool silo links for pool: %v", state.SubnetPoolID.ValueString()),
		map[string]any{"success": true},
	)

	idx := slices.IndexFunc(
		pools,
		func(p oxide.SiloSubnetPool) bool { return p.Id == state.SubnetPoolID.ValueString() },
	)
	if idx < 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	state.SubnetPoolID = types.StringValue(pools[idx].Id)
	state.IsDefault = types.BoolPointerValue(pools[idx].IsDefault)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *subnetPoolSiloLinkResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan subnetPoolSiloLinkResourceModel
	var state subnetPoolSiloLinkResourceModel

	// Read Terraform plan data into the plan model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

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

	params := oxide.SubnetPoolSiloUpdateParams{
		Pool: oxide.NameOrId(state.SubnetPoolID.ValueString()),
		Silo: oxide.NameOrId(state.SiloID.ValueString()),
		Body: &oxide.SubnetPoolSiloUpdate{
			IsDefault: plan.IsDefault.ValueBoolPointer(),
		},
	}
	link, err := r.client.SubnetPoolSiloUpdate(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating subnet pool silo link",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(
		ctx,
		fmt.Sprintf("updated subnet pool silo link for subnet pool: %v", link.SubnetPoolId),
		map[string]any{"success": true},
	)

	// This is a terraform-specific ID. We just copy it from the state
	plan.ID = state.ID

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *subnetPoolSiloLinkResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state subnetPoolSiloLinkResourceModel

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
	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	params := oxide.SubnetPoolSiloUnlinkParams{
		Pool: oxide.NameOrId(state.SubnetPoolID.ValueString()),
		Silo: oxide.NameOrId(state.SiloID.ValueString()),
	}
	if err := r.client.SubnetPoolSiloUnlink(ctx, params); err != nil {
		if !is404(err) {
			resp.Diagnostics.AddError(
				"Error deleting subnet pool silo link:",
				"API error: "+err.Error(),
			)
			return
		}
	}
	tflog.Trace(
		ctx,
		fmt.Sprintf("deleted subnet pool silo link with ID: %v", state.ID.ValueString()),
		map[string]any{"success": true},
	)
}

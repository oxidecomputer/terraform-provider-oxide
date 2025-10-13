// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"
	"fmt"

	"github.com/google/uuid"
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
	_ resource.Resource              = (*ipPoolSiloLinkResource)(nil)
	_ resource.ResourceWithConfigure = (*ipPoolSiloLinkResource)(nil)
)

// NewIpPoolSiloLinkResource is a helper function to simplify the provider implementation.
func NewIpPoolSiloLinkResource() resource.Resource {
	return &ipPoolSiloLinkResource{}
}

// ipPoolSiloLinkResource is the resource implementation.
type ipPoolSiloLinkResource struct {
	client *oxide.Client
}

type ipPoolSiloLinkResourceModel struct {
	ID        types.String   `tfsdk:"id"`
	SiloID    types.String   `tfsdk:"silo_id"`
	IPPoolID  types.String   `tfsdk:"ip_pool_id"`
	IsDefault types.Bool     `tfsdk:"is_default"`
	Timeouts  timeouts.Value `tfsdk:"timeouts"`
}

// Metadata returns the resource type name.
func (r *ipPoolSiloLinkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "oxide_ip_pool_silo_link"
}

// Configure adds the provider configured client to the data source.
func (r *ipPoolSiloLinkResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*oxide.Client)
}

func (r *ipPoolSiloLinkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("ip_pool_id"), req, resp)
}

// Schema defines the schema for the resource.
func (r *ipPoolSiloLinkResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			},
			"ip_pool_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the IP pool that will be linked to the silo.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"is_default": schema.BoolAttribute{
				Required: true,
				Description: "Whether this is the default IP pool for a silo. " +
					"Only a single IP pool silo link can be marked as default",
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *ipPoolSiloLinkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ipPoolSiloLinkResourceModel

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

	params := oxide.IpPoolSiloLinkParams{
		Pool: oxide.NameOrId(plan.IPPoolID.ValueString()),
		Body: &oxide.IpPoolLinkSilo{
			IsDefault: plan.IsDefault.ValueBoolPointer(),
			Silo:      oxide.NameOrId(plan.SiloID.ValueString()),
		},
	}
	link, err := r.client.IpPoolSiloLink(ctx, params)
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

	// Set a unique ID for the resource payload
	plan.ID = types.StringValue(uuid.New().String())

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *ipPoolSiloLinkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ipPoolSiloLinkResourceModel

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

	params := oxide.IpPoolSiloListParams{
		Pool:   oxide.NameOrId(state.IPPoolID.ValueString()),
		Limit:  oxide.NewPointer(1000000000),
		SortBy: oxide.IdSortModeIdAscending,
	}

	links, err := r.client.IpPoolSiloList(ctx, params)
	if err != nil {
		if is404(err) {
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
	tflog.Trace(ctx, fmt.Sprintf("read IP pool links with ID: %v", state.IPPoolID), map[string]any{"success": true})

	link := findLinkinIPPoolLinks(state.SiloID.ValueString(), links.Items)
	if link == nil {
		resp.Diagnostics.AddError(
			"Missing resource",
			fmt.Sprintf("Unable to find requested link between IP pool %v and silo %v",
				state.IPPoolID.ValueString(), state.SiloID.ValueString()),
		)
		return
	}

	state.IPPoolID = types.StringValue(link.IpPoolId)
	state.IsDefault = types.BoolPointerValue(link.IsDefault)
	state.SiloID = types.StringValue(link.SiloId)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *ipPoolSiloLinkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ipPoolSiloLinkResourceModel
	var state ipPoolSiloLinkResourceModel

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

	params := oxide.IpPoolSiloUpdateParams{
		Pool: oxide.NameOrId(state.IPPoolID.ValueString()),
		Silo: oxide.NameOrId(state.SiloID.ValueString()),
		Body: &oxide.IpPoolSiloUpdate{
			IsDefault: plan.IsDefault.ValueBoolPointer(),
		},
	}
	link, err := r.client.IpPoolSiloUpdate(ctx, params)
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

	// This is a terraform-specific ID. We just copy it from the state
	plan.ID = state.ID

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *ipPoolSiloLinkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ipPoolSiloLinkResourceModel

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

	params := oxide.IpPoolSiloUnlinkParams{
		Pool: oxide.NameOrId(state.IPPoolID.ValueString()),
		Silo: oxide.NameOrId(state.SiloID.ValueString()),
	}
	if err := r.client.IpPoolSiloUnlink(ctx, params); err != nil {
		if !is404(err) {
			resp.Diagnostics.AddError(
				"Error deleting link:",
				"API error: "+err.Error(),
			)
			return
		}
	}
	tflog.Trace(ctx, fmt.Sprintf("deleted link with ID: %v", state.ID.ValueString()), map[string]any{"success": true})
}

func findLinkinIPPoolLinks(siloID string, links []oxide.IpPoolSiloLink) *oxide.IpPoolSiloLink {
	for _, link := range links {
		if siloID == link.SiloId {
			return &oxide.IpPoolSiloLink{
				IpPoolId:  link.IpPoolId,
				SiloId:    link.SiloId,
				IsDefault: link.IsDefault,
			}
		}
	}
	return nil
}

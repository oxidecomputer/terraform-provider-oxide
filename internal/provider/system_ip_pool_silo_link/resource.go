// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package systemippoolsilolink

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

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/shared"
)

var (
	_ resource.Resource                = (*Resource)(nil)
	_ resource.ResourceWithConfigure   = (*Resource)(nil)
	_ resource.ResourceWithImportState = (*Resource)(nil)
)

// NewResource initializes a system IP pool silo link resource.
func NewResource() resource.Resource {
	return &Resource{}
}

// Resource is the resource implementation.
type Resource struct {
	client *oxide.Client
}

type ResourceModel struct {
	ID       types.String   `tfsdk:"id"`
	Pool     types.String   `tfsdk:"pool"`
	Silo     types.String   `tfsdk:"silo"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

// Metadata returns the resource type name.
func (r *Resource) Metadata(
	_ context.Context,
	_ resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = "oxide_system_ip_pool_silo_link"
}

// Configure adds the provider configured client to the resource.
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

// ImportState imports an existing link using pool/silo.
func (r *Resource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	idParts := strings.Split(req.ID, "/")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID format: pool/silo, got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("pool"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("silo"), idParts[1])...)
}

// Schema defines the schema for the resource.
func (r *Resource) Schema(
	ctx context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource manages a system IP pool to silo link. The link is non-default; use the pool and silo names or IDs to identify its endpoints.",
		Attributes: map[string]schema.Attribute{
			"pool": schema.StringAttribute{
				Required:    true,
				Description: "Name or ID of the system IP pool to link to the silo.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"silo": schema.StringAttribute{
				Required:    true,
				Description: "Name or ID of the silo to link the system IP pool to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Delete: true,
			}),
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Identifier for this resource, formatted as `IP_POOL_ID/SILO_ID`.",
			},
		},
	}
}

// Create creates the link and sets the initial Terraform state.
func (r *Resource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan ResourceModel

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

	link, err := r.client.SystemIpPoolSiloLink(ctx, oxide.SystemIpPoolSiloLinkParams{
		Pool: oxide.NameOrId(plan.Pool.ValueString()),
		Body: &oxide.IpPoolLinkSilo{
			IsDefault: oxide.NewPointer(false),
			Silo:      oxide.NameOrId(plan.Silo.ValueString()),
		},
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating system IP pool silo link",
			"API error: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", link.IpPoolId, link.SiloId))
	tflog.Trace(ctx, "created system IP pool silo link", map[string]any{
		"ip_pool_id": link.IpPoolId,
		"silo_id":    link.SiloId,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *Resource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state ResourceModel

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

	// Link list results contain silo IDs only, so resolve a configured silo name
	// before searching the system IP pool's links.
	silo, err := r.client.SiloView(ctx, oxide.SiloViewParams{
		Silo: oxide.NameOrId(state.Silo.ValueString()),
	})
	if err != nil {
		if shared.Is404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to read silo",
			"API error: "+err.Error(),
		)
		return
	}

	links, err := r.client.SystemIpPoolSiloListAllPages(
		ctx,
		oxide.SystemIpPoolSiloListParams{
			Pool: oxide.NameOrId(state.Pool.ValueString()),
		},
	)
	if err != nil {
		if shared.Is404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to read system IP pool silo links",
			"API error: "+err.Error(),
		)
		return
	}

	idx := slices.IndexFunc(links, func(link oxide.IpPoolSiloLink) bool {
		return link.SiloId == silo.Id
	})
	if idx < 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	state.ID = types.StringValue(
		fmt.Sprintf("%s/%s", links[idx].IpPoolId, links[idx].SiloId),
	)
	tflog.Trace(ctx, "read system IP pool silo link", map[string]any{
		"ip_pool_id": links[idx].IpPoolId,
		"silo_id":    links[idx].SiloId,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update stores timeout changes. Pool and silo changes require replacement.
func (r *Resource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan ResourceModel
	var state ResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Timeouts = plan.Timeouts
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete unlinks the system IP pool from the silo.
func (r *Resource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state ResourceModel

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

	err := r.client.SystemIpPoolSiloUnlink(ctx, oxide.SystemIpPoolSiloUnlinkParams{
		Pool: oxide.NameOrId(state.Pool.ValueString()),
		Silo: oxide.NameOrId(state.Silo.ValueString()),
	})
	if err != nil && !shared.Is404(err) {
		resp.Diagnostics.AddError(
			"Error deleting system IP pool silo link",
			"API error: "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, "deleted system IP pool silo link", map[string]any{
		"id": state.ID.ValueString(),
	})
}

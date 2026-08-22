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
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/oxidecomputer/oxide.go/oxide"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/shared"
)

var (
	_ resource.Resource                = (*Resource)(nil)
	_ resource.ResourceWithConfigure   = (*Resource)(nil)
	_ resource.ResourceWithImportState = (*Resource)(nil)
	_ resource.ResourceWithModifyPlan  = (*Resource)(nil)
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
	ID        types.String   `tfsdk:"id"`
	IsDefault types.Bool     `tfsdk:"is_default"`
	Pool      types.String   `tfsdk:"pool"`
	Silo      types.String   `tfsdk:"silo"`
	Timeouts  timeouts.Value `tfsdk:"timeouts"`
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

// ImportState imports an existing link using pool/silo names or IDs.
func (r *Resource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	pool, silo, ok := linkIDs(req.ID)
	if !ok {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf(
				"Expected import ID format: IP_POOL_NAME_OR_ID/SILO_NAME_OR_ID, got: %s",
				req.ID,
			),
		)
		return
	}

	poolID, siloID, err := r.resolveLinkIDs(ctx, pool, silo)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to import system IP pool silo link",
			"API error: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx,
		path.Root("id"),
		fmt.Sprintf("%s/%s", poolID, siloID),
	)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("pool"), pool)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("silo"), silo)...)
}

// Schema defines the schema for the resource.
func (r *Resource) Schema(
	ctx context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource manages a system IP pool to silo link. Use the pool and silo names or IDs to identify its endpoints.",
		Attributes: map[string]schema.Attribute{
			"pool": schema.StringAttribute{
				Required:    true,
				Description: "Name or ID of the system IP pool to link to the silo.",
			},
			"silo": schema.StringAttribute{
				Required:    true,
				Description: "Name or ID of the silo to link the system IP pool to.",
			},
			"is_default": schema.BoolAttribute{
				Required:    true,
				Description: "Whether this is the default IP pool for the silo. Only a single IP pool silo link can be marked as default for a given IP version and pool type.",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Update: true,
				Delete: true,
			}),
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Identifier for this resource, formatted as `IP_POOL_ID/SILO_ID`.",
			},
		},
	}
}

// ModifyPlan compares pool and silo selectors by their API identities. This
// avoids replacing an imported link when configuration uses a name for an
// endpoint that import recorded by ID, or vice versa.
func (r *Resource) ModifyPlan(
	ctx context.Context,
	req resource.ModifyPlanRequest,
	resp *resource.ModifyPlanResponse,
) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	var plan ResourceModel
	var state ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if plan.Pool.IsUnknown() {
		resp.RequiresReplace = append(resp.RequiresReplace, path.Root("pool"))
	}
	if plan.Silo.IsUnknown() {
		resp.RequiresReplace = append(resp.RequiresReplace, path.Root("silo"))
	}
	if plan.Pool.IsUnknown() || plan.Pool.IsNull() ||
		plan.Silo.IsUnknown() || plan.Silo.IsNull() {
		return
	}

	poolChanged := !plan.Pool.Equal(state.Pool)
	siloChanged := !plan.Silo.Equal(state.Silo)
	if !poolChanged && !siloChanged {
		return
	}

	statePoolID, stateSiloID, ok := linkIDs(state.ID.ValueString())
	if !ok {
		resp.Diagnostics.AddError(
			"Unable to plan system IP pool silo link",
			fmt.Sprintf(
				"Invalid resource ID %q; expected IP_POOL_ID/SILO_ID",
				state.ID.ValueString(),
			),
		)
		return
	}

	if poolChanged {
		planPoolID, err := r.resolvePoolID(ctx, plan.Pool.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to plan system IP pool silo link",
				"API error: "+err.Error(),
			)
			return
		}
		if planPoolID != statePoolID {
			resp.RequiresReplace = append(resp.RequiresReplace, path.Root("pool"))
		}
	}
	if siloChanged {
		planSiloID, err := r.resolveSiloID(ctx, plan.Silo.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to plan system IP pool silo link",
				"API error: "+err.Error(),
			)
			return
		}
		if planSiloID != stateSiloID {
			resp.RequiresReplace = append(resp.RequiresReplace, path.Root("silo"))
		}
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
			IsDefault: plan.IsDefault.ValueBoolPointer(),
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
	plan.IsDefault = types.BoolPointerValue(link.IsDefault)
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

	poolID, siloID, ok := linkIDs(state.ID.ValueString())
	if !ok {
		resp.Diagnostics.AddError(
			"Unable to read system IP pool silo link",
			fmt.Sprintf(
				"Invalid resource ID %q; expected IP_POOL_ID/SILO_ID",
				state.ID.ValueString(),
			),
		)
		return
	}

	links, err := r.client.SystemIpPoolSiloListAllPages(
		ctx,
		oxide.SystemIpPoolSiloListParams{
			Pool: oxide.NameOrId(poolID),
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
		return link.SiloId == siloID
	})
	if idx < 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	state.ID = types.StringValue(
		fmt.Sprintf("%s/%s", links[idx].IpPoolId, links[idx].SiloId),
	)
	state.IsDefault = types.BoolPointerValue(links[idx].IsDefault)
	tflog.Trace(ctx, "read system IP pool silo link", map[string]any{
		"ip_pool_id": links[idx].IpPoolId,
		"silo_id":    links[idx].SiloId,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update changes the default status or stores timeout changes. Pool and silo
// changes require replacement.
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

	statePoolID, stateSiloID, ok := linkIDs(state.ID.ValueString())
	if !ok {
		resp.Diagnostics.AddError(
			"Error updating system IP pool silo link",
			fmt.Sprintf(
				"Invalid resource ID %q; expected IP_POOL_ID/SILO_ID",
				state.ID.ValueString(),
			),
		)
		return
	}

	if !plan.Pool.Equal(state.Pool) || !plan.Silo.Equal(state.Silo) {
		planPoolID, planSiloID, err := r.resolveLinkIDs(
			ctx,
			plan.Pool.ValueString(),
			plan.Silo.ValueString(),
		)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating system IP pool silo link",
				"API error: "+err.Error(),
			)
			return
		}
		if planPoolID != statePoolID || planSiloID != stateSiloID {
			resp.Diagnostics.AddError(
				"Unexpected Update",
				"Pool and silo changes require replacement.",
			)
			return
		}
	}

	if plan.IsDefault.ValueBool() != state.IsDefault.ValueBool() {
		updateTimeout, diags := plan.Timeouts.Update(ctx, shared.DefaultTimeout())
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		ctx, cancel := context.WithTimeout(ctx, updateTimeout)
		defer cancel()

		link, err := r.client.SystemIpPoolSiloUpdate(ctx, oxide.SystemIpPoolSiloUpdateParams{
			Pool: oxide.NameOrId(statePoolID),
			Silo: oxide.NameOrId(stateSiloID),
			Body: &oxide.IpPoolSiloUpdate{
				IsDefault: plan.IsDefault.ValueBoolPointer(),
			},
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating system IP pool silo link",
				"API error: "+err.Error(),
			)
			return
		}

		plan.ID = types.StringValue(fmt.Sprintf("%s/%s", link.IpPoolId, link.SiloId))
		plan.IsDefault = types.BoolPointerValue(link.IsDefault)
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	plan.ID = state.ID
	plan.IsDefault = state.IsDefault
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
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

	poolID, siloID, ok := linkIDs(state.ID.ValueString())
	if !ok {
		resp.Diagnostics.AddError(
			"Error deleting system IP pool silo link",
			fmt.Sprintf(
				"Invalid resource ID %q; expected IP_POOL_ID/SILO_ID",
				state.ID.ValueString(),
			),
		)
		return
	}

	err := r.client.SystemIpPoolSiloUnlink(ctx, oxide.SystemIpPoolSiloUnlinkParams{
		Pool: oxide.NameOrId(poolID),
		Silo: oxide.NameOrId(siloID),
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

func linkIDs(id string) (string, string, bool) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}
	return parts[0], parts[1], true
}

func (r *Resource) resolveLinkIDs(
	ctx context.Context,
	pool, silo string,
) (string, string, error) {
	poolID, err := r.resolvePoolID(ctx, pool)
	if err != nil {
		return "", "", err
	}
	siloID, err := r.resolveSiloID(ctx, silo)
	if err != nil {
		return "", "", err
	}
	return poolID, siloID, nil
}

func (r *Resource) resolvePoolID(ctx context.Context, pool string) (string, error) {
	ipPool, err := r.client.SystemIpPoolView(ctx, oxide.SystemIpPoolViewParams{
		Pool: oxide.NameOrId(pool),
	})
	if err != nil {
		return "", fmt.Errorf("resolving system IP pool %q: %w", pool, err)
	}
	return ipPool.Id, nil
}

func (r *Resource) resolveSiloID(ctx context.Context, silo string) (string, error) {
	resolvedSilo, err := r.client.SiloView(ctx, oxide.SiloViewParams{
		Silo: oxide.NameOrId(silo),
	})
	if err != nil {
		return "", fmt.Errorf("resolving silo %q: %w", silo, err)
	}
	return resolvedSilo.Id, nil
}

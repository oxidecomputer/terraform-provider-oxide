// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package systemsubnetpoolsilolink

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
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

func NewResource() resource.Resource {
	return &Resource{}
}

type Resource struct {
	client *oxide.Client
}

type ResourceModel struct {
	ID        types.String   `tfsdk:"id"`
	Pool      types.String   `tfsdk:"pool"`
	Silo      types.String   `tfsdk:"silo"`
	IsDefault types.Bool     `tfsdk:"is_default"`
	Timeouts  timeouts.Value `tfsdk:"timeouts"`
}

func (r *Resource) Metadata(
	_ context.Context,
	_ resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = "oxide_system_subnet_pool_silo_link"
}

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

func (r *Resource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	pool, silo, err := parseID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID format: pool/silo, got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("pool"), pool)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("silo"), silo)...)
}

func (r *Resource) Schema(
	ctx context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource manages a system subnet pool's link to a silo.",
		Attributes: map[string]schema.Attribute{
			"pool": schema.StringAttribute{
				Required:    true,
				Description: "Name or ID of the subnet pool to link to the silo.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"silo": schema.StringAttribute{
				Required:    true,
				Description: "Name or ID of the silo to link to the subnet pool.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"is_default": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
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

	link, err := r.client.SystemSubnetPoolSiloLink(ctx, oxide.SystemSubnetPoolSiloLinkParams{
		Pool: oxide.NameOrId(plan.Pool.ValueString()),
		Body: &oxide.SubnetPoolLinkSilo{
			IsDefault: plan.IsDefault.ValueBoolPointer(),
			Silo:      oxide.NameOrId(plan.Silo.ValueString()),
		},
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating system subnet pool silo link",
			"API error: "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, "created system subnet pool silo link", map[string]any{"success": true})
	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", link.SubnetPoolId, link.SiloId))
	plan.IsDefault = types.BoolPointerValue(link.IsDefault)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

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

	poolID, siloID, err := parseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid system subnet pool silo link ID", err.Error())
		return
	}

	links, err := r.client.SystemSubnetPoolSiloListAllPages(
		ctx,
		oxide.SystemSubnetPoolSiloListParams{
			Pool: oxide.NameOrId(poolID),
		},
	)
	if err != nil {
		if shared.Is404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to read system subnet pool silo links",
			"API error: "+err.Error(),
		)
		return
	}

	idx := slices.IndexFunc(links, func(link oxide.SubnetPoolSiloLink) bool {
		return link.SiloId == siloID
	})
	if idx < 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	link := links[idx]
	state.ID = types.StringValue(fmt.Sprintf("%s/%s", link.SubnetPoolId, link.SiloId))
	state.IsDefault = types.BoolPointerValue(link.IsDefault)
	tflog.Trace(ctx, "read system subnet pool silo link", map[string]any{"success": true})
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *Resource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan ResourceModel
	var state ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
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

	poolID, siloID, err := parseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid system subnet pool silo link ID", err.Error())
		return
	}

	link, err := r.client.SystemSubnetPoolSiloUpdate(ctx, oxide.SystemSubnetPoolSiloUpdateParams{
		Pool: oxide.NameOrId(poolID),
		Silo: oxide.NameOrId(siloID),
		Body: &oxide.SubnetPoolSiloUpdate{IsDefault: plan.IsDefault.ValueBoolPointer()},
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating system subnet pool silo link",
			"API error: "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, "updated system subnet pool silo link", map[string]any{"success": true})
	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", link.SubnetPoolId, link.SiloId))
	plan.IsDefault = types.BoolPointerValue(link.IsDefault)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

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

	poolID, siloID, err := parseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid system subnet pool silo link ID", err.Error())
		return
	}

	err = r.client.SystemSubnetPoolSiloUnlink(ctx, oxide.SystemSubnetPoolSiloUnlinkParams{
		Pool: oxide.NameOrId(poolID),
		Silo: oxide.NameOrId(siloID),
	})
	if err != nil && !shared.Is404(err) {
		resp.Diagnostics.AddError(
			"Error deleting system subnet pool silo link",
			"API error: "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, "deleted system subnet pool silo link", map[string]any{"success": true})
}

func parseID(id string) (string, string, error) {
	idParts := strings.Split(id, "/")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return "", "", fmt.Errorf("expected ID format pool/silo, got %q", id)
	}

	return idParts[0], idParts[1], nil
}

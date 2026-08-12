// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package systemsubnetpoolsilolink

import (
	"context"
	"errors"
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
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/oxidecomputer/oxide.go/oxide"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/shared"
	oxidevalidator "github.com/oxidecomputer/terraform-provider-oxide/internal/provider/validator"
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
	ID           types.String   `tfsdk:"id"`
	SubnetPoolID types.String   `tfsdk:"subnet_pool_id"`
	SiloID       types.String   `tfsdk:"silo_id"`
	IsDefault    types.Bool     `tfsdk:"is_default"`
	Timeouts     timeouts.Value `tfsdk:"timeouts"`
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

// ImportState imports this resource using a composite ID in the format
// `${SUBNET_POOL_ID}/${SILO_ID}`.
func (r *Resource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	poolID, siloID, err := parseID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf(
				"Expected import ID format: subnet_pool_id/silo_id, got: %s",
				req.ID,
			),
		)
		return
	}

	resp.Diagnostics.Append(
		resp.State.SetAttribute(ctx, path.Root("subnet_pool_id"), poolID)...,
	)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("silo_id"), siloID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

func (r *Resource) Schema(
	ctx context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource manages a system subnet pool's link to a silo.",
		Attributes: map[string]schema.Attribute{
			"subnet_pool_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the subnet pool to link to the silo.",
				Validators: []validator.String{
					oxidevalidator.IsUUID(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"silo_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the silo to link to the subnet pool.",
				Validators: []validator.String{
					oxidevalidator.IsUUID(),
				},
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
		Pool: oxide.NameOrId(plan.SubnetPoolID.ValueString()),
		Body: &oxide.SubnetPoolLinkSilo{
			IsDefault: plan.IsDefault.ValueBoolPointer(),
			Silo:      oxide.NameOrId(plan.SiloID.ValueString()),
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

	links, err := r.client.SystemSubnetPoolSiloListAllPages(
		ctx,
		oxide.SystemSubnetPoolSiloListParams{
			Pool: oxide.NameOrId(state.SubnetPoolID.ValueString()),
		},
	)
	if err != nil {
		if errors.Is(err, oxide.ErrHTTP404) {
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
		return link.SiloId == state.SiloID.ValueString()
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

	link, err := r.client.SystemSubnetPoolSiloUpdate(ctx, oxide.SystemSubnetPoolSiloUpdateParams{
		Pool: oxide.NameOrId(state.SubnetPoolID.ValueString()),
		Silo: oxide.NameOrId(state.SiloID.ValueString()),
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

	err := r.client.SystemSubnetPoolSiloUnlink(ctx, oxide.SystemSubnetPoolSiloUnlinkParams{
		Pool: oxide.NameOrId(state.SubnetPoolID.ValueString()),
		Silo: oxide.NameOrId(state.SiloID.ValueString()),
	})
	if err != nil && !errors.Is(err, oxide.ErrHTTP404) {
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
		return "", "", fmt.Errorf(
			"expected ID format subnet_pool_id/silo_id, got %q",
			id,
		)
	}

	return idParts[0], idParts[1], nil
}

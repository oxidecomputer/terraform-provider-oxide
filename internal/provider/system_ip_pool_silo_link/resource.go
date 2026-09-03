// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package systemippoolsilolink

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/google/uuid"
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

	legacyippoolsilolink "github.com/oxidecomputer/terraform-provider-oxide/internal/provider/ip_pool_silo_link"
	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/shared"
	oxidevalidator "github.com/oxidecomputer/terraform-provider-oxide/internal/provider/validator"
)

var (
	_ resource.Resource                = (*Resource)(nil)
	_ resource.ResourceWithConfigure   = (*Resource)(nil)
	_ resource.ResourceWithImportState = (*Resource)(nil)
	_ resource.ResourceWithMoveState   = (*Resource)(nil)
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
	IPPoolID  types.String   `tfsdk:"ip_pool_id"`
	SiloID    types.String   `tfsdk:"silo_id"`
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

// ImportState imports an existing link using its pool and silo IDs.
func (r *Resource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	poolID, siloID, ok := linkIDs(req.ID)
	if !ok {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf(
				"Expected import ID format: IP_POOL_ID/SILO_ID, got: %s",
				req.ID,
			),
		)
		return
	}
	if _, err := uuid.Parse(poolID); err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("IP pool ID must be a UUID, got: %s", poolID),
		)
		return
	}
	if _, err := uuid.Parse(siloID); err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("silo ID must be a UUID, got: %s", siloID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx,
		path.Root("id"),
		req.ID,
	)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ip_pool_id"), poolID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("silo_id"), siloID)...)
}

// MoveState moves a legacy IP pool silo link into a system IP pool silo link.
func (r *Resource) MoveState(ctx context.Context) []resource.StateMover {
	var sourceSchemaResponse resource.SchemaResponse
	legacyippoolsilolink.NewResource().Schema(
		ctx,
		resource.SchemaRequest{},
		&sourceSchemaResponse,
	)

	return []resource.StateMover{{
		SourceSchema: &sourceSchemaResponse.Schema,
		StateMover: func(
			ctx context.Context,
			req resource.MoveStateRequest,
			resp *resource.MoveStateResponse,
		) {
			if req.SourceTypeName != "oxide_ip_pool_silo_link" ||
				req.SourceSchemaVersion != 0 ||
				!strings.HasSuffix(
					req.SourceProviderAddress,
					"/oxidecomputer/oxide",
				) {
				return
			}
			if req.SourceState == nil {
				resp.Diagnostics.AddError(
					"Unable to Move Resource State",
					"The oxide_ip_pool_silo_link state could not be decoded.",
				)
				return
			}

			var source legacyippoolsilolink.ResourceModel
			resp.Diagnostics.Append(req.SourceState.Get(ctx, &source)...)
			if resp.Diagnostics.HasError() {
				return
			}
			if _, err := uuid.Parse(source.IPPoolID.ValueString()); err != nil {
				resp.Diagnostics.AddError(
					"Unable to Move Resource State",
					"The oxide_ip_pool_silo_link ip_pool_id state is not a UUID. Upgrade to v0.21.0 and refresh the resource state before upgrading to v0.22.0.",
				)
				return
			}
			if _, err := uuid.Parse(source.SiloID.ValueString()); err != nil {
				resp.Diagnostics.AddError(
					"Unable to Move Resource State",
					"The oxide_ip_pool_silo_link silo_id state is not a UUID. Upgrade to v0.21.0 and refresh the resource state before upgrading to v0.22.0.",
				)
				return
			}

			target := ResourceModel{
				ID: types.StringValue(fmt.Sprintf(
					"%s/%s",
					source.IPPoolID.ValueString(),
					source.SiloID.ValueString(),
				)),
				IsDefault: source.IsDefault,
				IPPoolID:  source.IPPoolID,
				SiloID:    source.SiloID,
				Timeouts:  source.Timeouts,
			}
			resp.Diagnostics.Append(resp.TargetState.Set(ctx, &target)...)
		},
	}}
}

// Schema defines the schema for the resource.
func (r *Resource) Schema(
	ctx context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource manages a system IP pool to silo link.",
		Attributes: map[string]schema.Attribute{
			"ip_pool_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the system IP pool to link to the silo.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					oxidevalidator.IsUUID(),
				},
			},
			"silo_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the silo to link the system IP pool to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					oxidevalidator.IsUUID(),
				},
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
		Pool: oxide.NameOrId(plan.IPPoolID.ValueString()),
		Body: &oxide.IpPoolLinkSilo{
			IsDefault: plan.IsDefault.ValueBoolPointer(),
			Silo:      oxide.NameOrId(plan.SiloID.ValueString()),
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
		if errors.Is(err, oxide.ErrHTTP404) {
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
	state.IPPoolID = types.StringValue(links[idx].IpPoolId)
	state.SiloID = types.StringValue(links[idx].SiloId)
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

	// Only timeouts and is_default are able to be updated. Since timeouts is
	// internal to Terraform we can return early if is_default is unchanged.
	if plan.IsDefault.ValueBool() == state.IsDefault.ValueBool() {
		plan.ID = state.ID
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	updateTimeout, diags := plan.Timeouts.Update(ctx, shared.DefaultTimeout())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	link, err := r.client.SystemIpPoolSiloUpdate(ctx, oxide.SystemIpPoolSiloUpdateParams{
		Pool: oxide.NameOrId(state.IPPoolID.ValueString()),
		Silo: oxide.NameOrId(state.SiloID.ValueString()),
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
	if err != nil && !errors.Is(err, oxide.ErrHTTP404) {
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

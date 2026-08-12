// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package systemippoolrange

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/iptypes"
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

// NewResource returns a new system IP pool range resource.
func NewResource() resource.Resource {
	return &Resource{}
}

// Resource manages an address range within a system IP pool.
type Resource struct {
	client *oxide.Client
}

type ResourceModel struct {
	ID       types.String      `tfsdk:"id"`
	Pool     types.String      `tfsdk:"pool"`
	PoolID   types.String      `tfsdk:"pool_id"`
	First    iptypes.IPAddress `tfsdk:"first"`
	Last     iptypes.IPAddress `tfsdk:"last"`
	Timeouts timeouts.Value    `tfsdk:"timeouts"`
}

// Metadata returns the resource type name.
func (r *Resource) Metadata(
	_ context.Context,
	_ resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = "oxide_system_ip_pool_range"
}

// Configure adds the provider-configured client to the resource.
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

// ImportState imports a range using pool/range_id.
func (r *Resource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	idParts := strings.Split(req.ID, "/")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID format: pool/range_id, got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("pool"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
}

// Schema defines the schema for the resource.
func (r *Resource) Schema(
	ctx context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	requiresReplace := []planmodifier.String{stringplanmodifier.RequiresReplace()}

	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource manages an address range within a system IP pool.",
		Attributes: map[string]schema.Attribute{
			"pool": schema.StringAttribute{
				Required:      true,
				Description:   "Name or ID of the system IP pool containing the range.",
				PlanModifiers: requiresReplace,
			},
			"first": schema.StringAttribute{
				Required:      true,
				CustomType:    iptypes.IPAddressType{},
				Description:   "First IP address in the range.",
				PlanModifiers: requiresReplace,
			},
			"last": schema.StringAttribute{
				Required:      true,
				CustomType:    iptypes.IPAddressType{},
				Description:   "Last IP address in the range.",
				PlanModifiers: requiresReplace,
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier of the system IP pool range.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"pool_id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier of the system IP pool containing the range.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
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

// Create adds the range to the system IP pool.
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

	ipRange, err := oxide.NewIpRange(plan.First.ValueString(), plan.Last.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating system IP pool range", err.Error())
		return
	}

	created, err := r.client.SystemIpPoolRangeAdd(ctx, oxide.SystemIpPoolRangeAddParams{
		Pool: oxide.NameOrId(plan.Pool.ValueString()),
		Body: &ipRange,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating system IP pool range", "API error: "+err.Error())
		return
	}

	plan.ID = types.StringValue(created.Id)
	plan.PoolID = types.StringValue(created.IpPoolId)
	if !setAddresses(&plan, created.Range) {
		resp.Diagnostics.AddError(
			"Error reading created system IP pool range",
			fmt.Sprintf("Unexpected IP range variant type %T", created.Range.Value),
		)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("created system IP pool range with ID: %s", created.Id))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the range from the system IP pool's range list.
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

	ranges, err := r.listRanges(ctx, state)
	if err != nil {
		if shared.Is404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read system IP pool ranges", "API error: "+err.Error())
		return
	}

	for _, ipRange := range ranges {
		if ipRange.Id != state.ID.ValueString() {
			continue
		}

		state.PoolID = types.StringValue(ipRange.IpPoolId)
		if !setAddresses(&state, ipRange.Range) {
			resp.Diagnostics.AddError(
				"Unable to read system IP pool range",
				fmt.Sprintf("Unexpected IP range variant type %T", ipRange.Range.Value),
			)
			return
		}
		tflog.Trace(ctx, fmt.Sprintf("read system IP pool range with ID: %s", ipRange.Id))
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		return
	}

	resp.State.RemoveResource(ctx)
}

// Update stores timeout changes. Pool and address changes require replacement.
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

	if plan.Pool.ValueString() != state.Pool.ValueString() ||
		plan.First.ValueString() != state.First.ValueString() ||
		plan.Last.ValueString() != state.Last.ValueString() {
		resp.Diagnostics.AddError(
			"Unexpected Update",
			"Pool and address changes require replacement.",
		)
		return
	}

	state.Timeouts = plan.Timeouts
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete removes the range from the system IP pool.
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

	ranges, err := r.listRanges(ctx, state)
	if err != nil {
		if shared.Is404(err) {
			return
		}
		resp.Diagnostics.AddError(
			"Error reading system IP pool ranges before removal",
			"API error: "+err.Error(),
		)
		return
	}

	var found *oxide.IpPoolRange
	for index := range ranges {
		if ranges[index].Id == state.ID.ValueString() {
			found = &ranges[index]
			break
		}
	}
	if found == nil {
		return
	}

	err = r.client.SystemIpPoolRangeRemove(ctx, oxide.SystemIpPoolRangeRemoveParams{
		Pool: oxide.NameOrId(found.IpPoolId),
		Body: &found.Range,
	})
	if err != nil {
		remaining, readErr := r.listRanges(ctx, state)
		if readErr == nil {
			for _, ipRange := range remaining {
				if ipRange.Id == state.ID.ValueString() {
					resp.Diagnostics.AddError(
						"Error removing system IP pool range",
						"API error: "+err.Error(),
					)
					return
				}
			}
			return
		}
		if shared.Is404(readErr) {
			return
		}
		resp.Diagnostics.AddError("Error removing system IP pool range", "API error: "+err.Error())
		return
	}

	tflog.Trace(
		ctx,
		fmt.Sprintf("removed system IP pool range with ID: %s", state.ID.ValueString()),
	)
}

func (r *Resource) listRanges(
	ctx context.Context,
	state ResourceModel,
) ([]oxide.IpPoolRange, error) {
	pool := state.Pool.ValueString()
	if !state.PoolID.IsNull() && !state.PoolID.IsUnknown() {
		pool = state.PoolID.ValueString()
	}

	return r.client.SystemIpPoolRangeListAllPages(ctx, oxide.SystemIpPoolRangeListParams{
		Pool: oxide.NameOrId(pool),
	})
}

func setAddresses(model *ResourceModel, ipRange oxide.IpRange) bool {
	switch value := ipRange.Value.(type) {
	case *oxide.Ipv4Range:
		model.First = iptypes.NewIPAddressValue(value.First)
		model.Last = iptypes.NewIPAddressValue(value.Last)
	case *oxide.Ipv6Range:
		model.First = iptypes.NewIPAddressValue(value.First)
		model.Last = iptypes.NewIPAddressValue(value.Last)
	default:
		return false
	}

	return true
}

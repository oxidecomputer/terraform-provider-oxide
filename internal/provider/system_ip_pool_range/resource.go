// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package systemippoolrange

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-nettypes/iptypes"
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
	ID           types.String      `tfsdk:"id"`
	IPPoolID     types.String      `tfsdk:"ip_pool_id"`
	FirstAddress iptypes.IPAddress `tfsdk:"first_address"`
	LastAddress  iptypes.IPAddress `tfsdk:"last_address"`
	Timeouts     timeouts.Value    `tfsdk:"timeouts"`
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

// ImportState imports a range using ip_pool_id/range_id or ip_pool_id/first_address/last_address.
func (r *Resource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	idParts := strings.Split(req.ID, "/")
	if (len(idParts) != 2 && len(idParts) != 3) ||
		slices.Contains(idParts, "") {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf(
				"Expected import ID format: ip_pool_id/range_id or ip_pool_id/first_address/last_address, got: %s",
				req.ID,
			),
		)
		return
	}
	if _, err := uuid.Parse(idParts[0]); err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("IP pool ID must be a UUID, got: %s", idParts[0]),
		)
		return
	}
	if len(idParts) == 2 {
		if _, err := uuid.Parse(idParts[1]); err != nil {
			resp.Diagnostics.AddError(
				"Invalid Import ID",
				fmt.Sprintf("range ID must be a UUID, got: %s", idParts[1]),
			)
			return
		}
	}

	rangeID := idParts[1]
	if len(idParts) == 3 {
		requested, err := oxide.NewIpRange(idParts[1], idParts[2])
		if err != nil {
			resp.Diagnostics.AddError("Invalid Import ID", err.Error())
			return
		}

		ranges, err := r.client.SystemIpPoolRangeListAllPages(
			ctx,
			oxide.SystemIpPoolRangeListParams{
				Pool: oxide.NameOrId(idParts[0]),
			},
		)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to import system IP pool range",
				"API error: "+err.Error(),
			)
			return
		}

		idx := slices.IndexFunc(ranges, func(ipRange oxide.IpPoolRange) bool {
			return ipRange.Range.String() == requested.String()
		})
		if idx < 0 {
			resp.Diagnostics.AddError(
				"Unable to import system IP pool range",
				fmt.Sprintf(
					"No range from %s to %s exists in system IP pool %s.",
					idParts[1],
					idParts[2],
					idParts[0],
				),
			)
			return
		}
		rangeID = ranges[idx].Id
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ip_pool_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), rangeID)...)
}

// Schema defines the schema for the resource.
func (r *Resource) Schema(
	ctx context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource manages an address range within a system IP pool.",
		Attributes: map[string]schema.Attribute{
			"ip_pool_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the system IP pool containing the range.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					oxidevalidator.IsUUID(),
				},
			},
			"first_address": schema.StringAttribute{
				Required:    true,
				CustomType:  iptypes.IPAddressType{},
				Description: "First IP address in the range.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"last_address": schema.StringAttribute{
				Required:    true,
				CustomType:  iptypes.IPAddressType{},
				Description: "Last IP address in the range.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier of the system IP pool range.",
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

	ipRange, err := oxide.NewIpRange(
		plan.FirstAddress.ValueString(),
		plan.LastAddress.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating system IP pool range", err.Error())
		return
	}

	created, err := r.client.SystemIpPoolRangeAdd(ctx, oxide.SystemIpPoolRangeAddParams{
		Pool: oxide.NameOrId(plan.IPPoolID.ValueString()),
		Body: &ipRange,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating system IP pool range", "API error: "+err.Error())
		return
	}

	plan.ID = types.StringValue(created.Id)
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

	ranges, err := r.client.SystemIpPoolRangeListAllPages(
		ctx,
		oxide.SystemIpPoolRangeListParams{
			Pool: oxide.NameOrId(state.IPPoolID.ValueString()),
		},
	)
	if err != nil {
		if errors.Is(err, oxide.ErrHTTP404) {
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

// Update stores timeout changes.
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

	ranges, err := r.client.SystemIpPoolRangeListAllPages(
		ctx,
		oxide.SystemIpPoolRangeListParams{
			Pool: oxide.NameOrId(state.IPPoolID.ValueString()),
		},
	)
	if err != nil {
		if errors.Is(err, oxide.ErrHTTP404) {
			return
		}
		resp.Diagnostics.AddError(
			"Error reading system IP pool ranges before removal",
			"API error: "+err.Error(),
		)
		return
	}

	index := slices.IndexFunc(ranges, func(ipRange oxide.IpPoolRange) bool {
		return ipRange.Id == state.ID.ValueString()
	})
	if index == -1 {
		return
	}
	found := ranges[index]

	err = r.client.SystemIpPoolRangeRemove(ctx, oxide.SystemIpPoolRangeRemoveParams{
		Pool: oxide.NameOrId(found.IpPoolId),
		Body: &found.Range,
	})
	if err != nil {
		// Verify whether removal succeeded despite an ambiguous API error.
		remaining, readErr := r.client.SystemIpPoolRangeListAllPages(
			ctx,
			oxide.SystemIpPoolRangeListParams{
				Pool: oxide.NameOrId(state.IPPoolID.ValueString()),
			},
		)
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
		if errors.Is(readErr, oxide.ErrHTTP404) {
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

func setAddresses(model *ResourceModel, ipRange oxide.IpRange) bool {
	switch value := ipRange.Value.(type) {
	case *oxide.Ipv4Range:
		model.FirstAddress = iptypes.NewIPAddressValue(value.First)
		model.LastAddress = iptypes.NewIPAddressValue(value.Last)
	case *oxide.Ipv6Range:
		model.FirstAddress = iptypes.NewIPAddressValue(value.First)
		model.LastAddress = iptypes.NewIPAddressValue(value.Last)
	default:
		return false
	}

	return true
}

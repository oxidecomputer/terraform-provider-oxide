// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package systemsubnetpoolmember

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
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
	ID              types.String       `tfsdk:"id"`
	SubnetPoolID    types.String       `tfsdk:"subnet_pool_id"`
	Subnet          cidrtypes.IPPrefix `tfsdk:"subnet"`
	MaxPrefixLength types.Int64        `tfsdk:"max_prefix_length"`
	MinPrefixLength types.Int64        `tfsdk:"min_prefix_length"`
	TimeCreated     types.String       `tfsdk:"time_created"`
	Timeouts        timeouts.Value     `tfsdk:"timeouts"`
}

func (r *Resource) Metadata(
	_ context.Context,
	_ resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = "oxide_system_subnet_pool_member"
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
// `${SUBNET_POOL_ID}/${ID}`. The `id` attribute alone is not enough to qualify
// this resource.
func (r *Resource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	idParts := strings.Split(req.ID, "/")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf(
				"Expected import ID format: subnet_pool_id/member_id, got: %s",
				req.ID,
			),
		)
		return
	}

	resp.Diagnostics.Append(
		resp.State.SetAttribute(ctx, path.Root("subnet_pool_id"), idParts[0])...,
	)
	resp.Diagnostics.Append(
		resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...,
	)
}

func (r *Resource) Schema(
	ctx context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource manages a subnet pool member using the system API.",
		Attributes: map[string]schema.Attribute{
			"subnet_pool_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the subnet pool.",
				Validators: []validator.String{
					oxidevalidator.IsUUID(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"subnet": schema.StringAttribute{
				Required:    true,
				CustomType:  cidrtypes.IPPrefixType{},
				Description: "Subnet CIDR to add to the pool.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"max_prefix_length": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Maximum prefix length for allocations from this subnet. Defaults to 32 for IPv4 and 128 for IPv6.",
				Validators: []validator.Int64{
					int64validator.Between(0, 128),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"min_prefix_length": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Minimum prefix length for allocations from this subnet. Defaults to the subnet's prefix length.",
				Validators: []validator.Int64{
					int64validator.Between(0, 128),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Delete: true,
			}),
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier of the subnet pool member.",
			},
			"time_created": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this subnet pool member was created.",
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

	subnet, err := oxide.NewIpNet(plan.Subnet.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error parsing subnet CIDR", err.Error())
		return
	}

	body := &oxide.SubnetPoolMemberAdd{Subnet: subnet}
	if !plan.MaxPrefixLength.IsNull() && !plan.MaxPrefixLength.IsUnknown() {
		body.MaxPrefixLength = oxide.NewPointer(int(plan.MaxPrefixLength.ValueInt64()))
	}
	if !plan.MinPrefixLength.IsNull() && !plan.MinPrefixLength.IsUnknown() {
		body.MinPrefixLength = oxide.NewPointer(int(plan.MinPrefixLength.ValueInt64()))
	}

	member, err := r.client.SystemSubnetPoolMemberAdd(
		ctx,
		oxide.SystemSubnetPoolMemberAddParams{
			Pool: oxide.NameOrId(plan.SubnetPoolID.ValueString()),
			Body: body,
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating system subnet pool member",
			"API error: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(member.Id)
	plan.MaxPrefixLength = types.Int64Value(int64(*member.MaxPrefixLength))
	plan.MinPrefixLength = types.Int64Value(int64(*member.MinPrefixLength))
	plan.TimeCreated = types.StringValue(member.TimeCreated.String())
	tflog.Trace(ctx, "created system subnet pool member", map[string]any{
		"id":     member.Id,
		"pool":   plan.SubnetPoolID.ValueString(),
		"subnet": plan.Subnet.ValueString(),
	})

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

	members, err := r.client.SystemSubnetPoolMemberListAllPages(
		ctx,
		oxide.SystemSubnetPoolMemberListParams{
			Pool: oxide.NameOrId(state.SubnetPoolID.ValueString()),
		},
	)
	if errors.Is(err, oxide.ErrHTTP404) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read system subnet pool members",
			"API error: "+err.Error(),
		)
		return
	}

	var foundMember *oxide.SubnetPoolMember
	for i := range members {
		if members[i].Id == state.ID.ValueString() {
			foundMember = &members[i]
			break
		}
	}

	if foundMember == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Subnet = cidrtypes.NewIPPrefixValue(foundMember.Subnet.String())
	state.ID = types.StringValue(foundMember.Id)
	state.SubnetPoolID = types.StringValue(foundMember.SubnetPoolId)
	state.MaxPrefixLength = types.Int64Value(int64(*foundMember.MaxPrefixLength))
	state.MinPrefixLength = types.Int64Value(int64(*foundMember.MinPrefixLength))
	state.TimeCreated = types.StringValue(foundMember.TimeCreated.String())
	tflog.Trace(ctx, "read system subnet pool member", map[string]any{
		"id":     foundMember.Id,
		"pool":   state.SubnetPoolID.ValueString(),
		"subnet": foundMember.Subnet.String(),
	})
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *Resource) Update(
	_ context.Context,
	_ resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	resp.Diagnostics.AddError(
		"Unexpected Update",
		"This resource does not support in-place updates. All changes require replacement.",
	)
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

	subnet, err := oxide.NewIpNet(state.Subnet.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error parsing subnet CIDR", err.Error())
		return
	}

	err = r.client.SystemSubnetPoolMemberRemove(
		ctx,
		oxide.SystemSubnetPoolMemberRemoveParams{
			Pool: oxide.NameOrId(state.SubnetPoolID.ValueString()),
			Body: &oxide.SubnetPoolMemberRemove{Subnet: subnet},
		},
	)
	if err != nil && !isSubnetPoolMemberNotFound(err, state.Subnet.ValueString()) {
		resp.Diagnostics.AddError(
			"Error deleting system subnet pool member",
			"API error: "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, "deleted system subnet pool member", map[string]any{
		"pool":   state.SubnetPoolID.ValueString(),
		"subnet": state.Subnet.ValueString(),
	})
}

func isSubnetPoolMemberNotFound(err error, subnet string) bool {
	if errors.Is(err, oxide.ErrHTTP404) {
		return true
	}

	var httpErr *oxide.HTTPError
	return errors.Is(err, oxide.ErrInvalidRequest) &&
		errors.As(err, &httpErr) &&
		httpErr.ErrorResponse != nil &&
		httpErr.ErrorResponse.Message == fmt.Sprintf(
			"A provided subnet pool member with subnet %s does not exist",
			subnet,
		)
}

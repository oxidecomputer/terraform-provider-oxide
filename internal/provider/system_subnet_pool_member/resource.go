// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package systemsubnetpoolmember

import (
	"context"
	"fmt"
	"net/netip"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
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

func NewResource() resource.Resource {
	return &Resource{}
}

type Resource struct {
	client *oxide.Client
}

type ResourceModel struct {
	ID       types.String       `tfsdk:"id"`
	Pool     types.String       `tfsdk:"pool"`
	Subnet   cidrtypes.IPPrefix `tfsdk:"subnet"`
	Timeouts timeouts.Value     `tfsdk:"timeouts"`
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

func (r *Resource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	pool, subnet, err := parseImportID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", err.Error())
		return
	}

	id := resourceID(pool, subnet)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("pool"), pool)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("subnet"), subnet)...)
}

func (r *Resource) Schema(
	ctx context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource manages a subnet pool member using the system API.",
		Attributes: map[string]schema.Attribute{
			"pool": schema.StringAttribute{
				Required:    true,
				Description: "Name or ID of the subnet pool.",
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
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Delete: true,
			}),
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Identifier for this resource, formatted as `pool/subnet`.",
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

	member, err := r.client.SystemSubnetPoolMemberAdd(
		ctx,
		oxide.SystemSubnetPoolMemberAddParams{
			Pool: oxide.NameOrId(plan.Pool.ValueString()),
			Body: &oxide.SubnetPoolMemberAdd{Subnet: subnet},
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating system subnet pool member",
			"API error: "+err.Error(),
		)
		return
	}

	memberSubnet, err := canonicalSubnet(member.Subnet.String())
	if err != nil {
		resp.Diagnostics.AddError("Invalid subnet returned by API", err.Error())
		return
	}
	plan.Subnet = cidrtypes.NewIPPrefixValue(memberSubnet)
	plan.ID = types.StringValue(resourceID(plan.Pool.ValueString(), memberSubnet))
	tflog.Trace(ctx, "created system subnet pool member", map[string]any{
		"member_id": member.Id,
		"pool":      plan.Pool.ValueString(),
		"subnet":    memberSubnet,
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

	stateSubnet, err := netip.ParsePrefix(state.Subnet.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid subnet in state", err.Error())
		return
	}

	members, err := r.client.SystemSubnetPoolMemberListAllPages(
		ctx,
		oxide.SystemSubnetPoolMemberListParams{
			Pool: oxide.NameOrId(state.Pool.ValueString()),
		},
	)
	if err != nil {
		if shared.Is404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to read system subnet pool members",
			"API error: "+err.Error(),
		)
		return
	}

	for _, member := range members {
		memberSubnet, err := netip.ParsePrefix(member.Subnet.String())
		if err != nil {
			resp.Diagnostics.AddError("Invalid subnet returned by API", err.Error())
			return
		}
		if memberSubnet != stateSubnet {
			continue
		}

		canonicalSubnet := memberSubnet.String()
		state.Subnet = cidrtypes.NewIPPrefixValue(canonicalSubnet)
		state.ID = types.StringValue(resourceID(state.Pool.ValueString(), canonicalSubnet))
		tflog.Trace(ctx, "read system subnet pool member", map[string]any{
			"member_id": member.Id,
			"pool":      state.Pool.ValueString(),
			"subnet":    canonicalSubnet,
		})
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		return
	}

	resp.State.RemoveResource(ctx)
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
			Pool: oxide.NameOrId(state.Pool.ValueString()),
			Body: &oxide.SubnetPoolMemberRemove{Subnet: subnet},
		},
	)
	if err != nil && !shared.Is404(err) && !strings.Contains(err.Error(), "does not exist") {
		resp.Diagnostics.AddError(
			"Error deleting system subnet pool member",
			"API error: "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, "deleted system subnet pool member", map[string]any{
		"pool":   state.Pool.ValueString(),
		"subnet": state.Subnet.ValueString(),
	})
}

func resourceID(pool, subnet string) string {
	return pool + "/" + subnet
}

func parseImportID(id string) (string, string, error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return "", "", fmt.Errorf("expected import ID format: pool/subnet, got: %s", id)
	}

	subnet, err := canonicalSubnet(idParts[1])
	if err != nil {
		return "", "", fmt.Errorf("invalid subnet: %w", err)
	}
	return idParts[0], subnet, nil
}

func canonicalSubnet(subnet string) (string, error) {
	prefix, err := netip.ParsePrefix(subnet)
	if err != nil {
		return "", err
	}
	return prefix.String(), nil
}

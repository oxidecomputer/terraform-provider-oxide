// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/oxidecomputer/oxide.go/oxide"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = (*subnetPoolMemberResource)(nil)
	_ resource.ResourceWithConfigure = (*subnetPoolMemberResource)(nil)
)

// NewSubnetPoolMemberResource is a helper function to simplify the provider implementation.
func NewSubnetPoolMemberResource() resource.Resource {
	return &subnetPoolMemberResource{}
}

// subnetPoolMemberResource is the resource implementation.
type subnetPoolMemberResource struct {
	client *oxide.Client
}

type subnetPoolMemberResourceModel struct {
	ID              types.String   `tfsdk:"id"`
	SubnetPoolID    types.String   `tfsdk:"subnet_pool_id"`
	Subnet          types.String   `tfsdk:"subnet"`
	MinPrefixLength types.Int64    `tfsdk:"min_prefix_length"`
	MaxPrefixLength types.Int64    `tfsdk:"max_prefix_length"`
	TimeCreated     types.String   `tfsdk:"time_created"`
	Timeouts        timeouts.Value `tfsdk:"timeouts"`
}

// Metadata returns the resource type name.
func (r *subnetPoolMemberResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = "oxide_subnet_pool_member"
}

// Configure adds the provider configured client to the data source.
func (r *subnetPoolMemberResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*oxide.Client)
}

func (r *subnetPoolMemberResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	idParts := strings.Split(req.ID, "/")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID format: subnet_pool_id/member_id, got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(
		resp.State.SetAttribute(ctx, path.Root("subnet_pool_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
}

// Schema defines the schema for the resource.
func (r *subnetPoolMemberResource) Schema(
	ctx context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource manages a member (subnet) within a subnet pool.",
		Attributes: map[string]schema.Attribute{
			"subnet_pool_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the subnet pool this member belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"subnet": schema.StringAttribute{
				Required:    true,
				Description: "The subnet CIDR to add to the pool (e.g., '10.0.0.0/16').",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"min_prefix_length": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Minimum prefix length for allocations from this subnet. Defaults to the subnet's prefix length.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"max_prefix_length": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Maximum prefix length for allocations from this subnet. Defaults to 32 for IPv4 and 128 for IPv6.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
					int64planmodifier.UseStateForUnknown(),
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"time_created": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this subnet pool member was created.",
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *subnetPoolMemberResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan subnetPoolMemberResourceModel

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

	subnet, err := oxide.NewIpNet(plan.Subnet.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error parsing subnet CIDR",
			err.Error(),
		)
		return
	}

	body := &oxide.SubnetPoolMemberAdd{
		Subnet: subnet,
	}

	if !plan.MinPrefixLength.IsNull() && !plan.MinPrefixLength.IsUnknown() {
		minPrefixLen := int(plan.MinPrefixLength.ValueInt64())
		body.MinPrefixLength = &minPrefixLen
	}

	if !plan.MaxPrefixLength.IsNull() && !plan.MaxPrefixLength.IsUnknown() {
		maxPrefixLen := int(plan.MaxPrefixLength.ValueInt64())
		body.MaxPrefixLength = &maxPrefixLen
	}

	params := oxide.SubnetPoolMemberAddParams{
		Pool: oxide.NameOrId(plan.SubnetPoolID.ValueString()),
		Body: body,
	}

	member, err := r.client.SubnetPoolMemberAdd(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating subnet pool member",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(
		ctx,
		fmt.Sprintf("created subnet pool member with ID: %v", member.Id),
		map[string]any{"success": true},
	)

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(member.Id)
	plan.MinPrefixLength = types.Int64Value(int64(*member.MinPrefixLength))
	plan.MaxPrefixLength = types.Int64Value(int64(*member.MaxPrefixLength))
	plan.TimeCreated = types.StringValue(member.TimeCreated.String())

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *subnetPoolMemberResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state subnetPoolMemberResourceModel

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

	// The API doesn't have a direct "view" endpoint for a single member by ID.
	// We need to list all members and find the one matching our ID.
	members, err := r.client.SubnetPoolMemberListAllPages(
		ctx,
		oxide.SubnetPoolMemberListParams{
			Pool: oxide.NameOrId(state.SubnetPoolID.ValueString()),
		},
	)
	if err != nil {
		if is404(err) {
			// Pool doesn't exist, remove resource from state
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to read subnet pool members:",
			"API error: "+err.Error(),
		)
		return
	}

	// Find the member with matching ID
	var foundMember *oxide.SubnetPoolMember
	for i := range members {
		if members[i].Id == state.ID.ValueString() {
			foundMember = &members[i]
			break
		}
	}

	if foundMember == nil {
		// Member not found, remove resource from state
		resp.State.RemoveResource(ctx)
		return
	}

	tflog.Trace(
		ctx,
		fmt.Sprintf("read subnet pool member with ID: %v", foundMember.Id),
		map[string]any{"success": true},
	)

	state.ID = types.StringValue(foundMember.Id)
	state.Subnet = types.StringValue(foundMember.Subnet.String())
	state.MinPrefixLength = types.Int64Value(int64(*foundMember.MinPrefixLength))
	state.MaxPrefixLength = types.Int64Value(int64(*foundMember.MaxPrefixLength))
	state.TimeCreated = types.StringValue(foundMember.TimeCreated.String())

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
// Note: All attributes require replacement, so this should never be called.
func (r *subnetPoolMemberResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	// All attributes either require replacement or are computed.
	// This method should never be called.
	resp.Diagnostics.AddError(
		"Unexpected Update",
		"This resource does not support in-place updates. All changes require replacement.",
	)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *subnetPoolMemberResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state subnetPoolMemberResourceModel

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

	subnet, err := oxide.NewIpNet(state.Subnet.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error parsing subnet CIDR",
			err.Error(),
		)
		return
	}

	params := oxide.SubnetPoolMemberRemoveParams{
		Pool: oxide.NameOrId(state.SubnetPoolID.ValueString()),
		Body: &oxide.SubnetPoolMemberRemove{
			Subnet: subnet,
		},
	}

	if err := r.client.SubnetPoolMemberRemove(ctx, params); err != nil {
		// The API returns 400 with "does not exist" if the member is already gone,
		// rather than 404. Handle both cases for idempotent deletes.
		//
		// TODO: Switch to a 404 in omicron.
		if !is404(err) && !strings.Contains(err.Error(), "does not exist") {
			resp.Diagnostics.AddError(
				"Error deleting subnet pool member:",
				"API error: "+err.Error(),
			)
			return
		}
	}
	tflog.Trace(
		ctx,
		fmt.Sprintf("deleted subnet pool member with ID: %v", state.ID.ValueString()),
		map[string]any{"success": true},
	)
}

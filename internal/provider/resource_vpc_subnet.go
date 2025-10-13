// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/oxidecomputer/oxide.go/oxide"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = (*vpcSubnetResource)(nil)
	_ resource.ResourceWithConfigure = (*vpcSubnetResource)(nil)
)

// NewVPCSubnetResource is a helper function to simplify the provider implementation.
func NewVPCSubnetResource() resource.Resource {
	return &vpcSubnetResource{}
}

// vpcSubnetResource is the resource implementation.
type vpcSubnetResource struct {
	client *oxide.Client
}

type vpcSubnetResourceModel struct {
	Description  types.String   `tfsdk:"description"`
	ID           types.String   `tfsdk:"id"`
	IPV4Block    types.String   `tfsdk:"ipv4_block"`
	IPV6Block    types.String   `tfsdk:"ipv6_block"`
	Name         types.String   `tfsdk:"name"`
	VPCID        types.String   `tfsdk:"vpc_id"`
	TimeCreated  types.String   `tfsdk:"time_created"`
	TimeModified types.String   `tfsdk:"time_modified"`
	Timeouts     timeouts.Value `tfsdk:"timeouts"`
}

// Metadata returns the resource type name.
func (r *vpcSubnetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "oxide_vpc_subnet"
}

// Configure adds the provider configured client to the data source.
func (r *vpcSubnetResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*oxide.Client)
}

func (r *vpcSubnetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Schema defines the schema for the resource.
func (r *vpcSubnetResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
This resource manages VPC subnets.
`,
		Attributes: map[string]schema.Attribute{
			"vpc_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the VPC that will contain the subnet.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the VPC subnet.",
			},
			"description": schema.StringAttribute{
				Required:    true,
				Description: "Description for the VPC subnet.",
			},
			"ipv4_block": schema.StringAttribute{
				Required: true,
				Description: "IPv4 address range for this VPC subnet. " +
					"It must be allocated from an RFC 1918 private address range, " +
					"and must not overlap with any other existing subnet in the VPC.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ipv6_block": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Description: "IPv6 address range for this VPC subnet. " +
					"It must be allocated from the RFC 4193 Unique Local Address range, " +
					"with the prefix equal to the parent VPC's prefix. " +
					"A random `/64` block will be assigned if one is not provided. " +
					"It must not overlap with any existing subnet in the VPC.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Update: true,
				Delete: true,
			}),
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier of the VPC.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"time_created": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this VPC was created.",
			},
			"time_modified": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this VPC was last modified.",
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *vpcSubnetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan vpcSubnetResourceModel

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

	params := oxide.VpcSubnetCreateParams{
		Vpc: oxide.NameOrId(plan.VPCID.ValueString()),
		Body: &oxide.VpcSubnetCreate{
			Description: plan.Description.ValueString(),
			Name:        oxide.Name(plan.Name.ValueString()),
			Ipv4Block:   oxide.Ipv4Net(plan.IPV4Block.ValueString()),
			Ipv6Block:   oxide.Ipv6Net(plan.IPV6Block.ValueString()),
		},
	}
	subnet, err := r.client.VpcSubnetCreate(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating VPC subnet",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("created VPC subnet with ID: %v", subnet.Id), map[string]any{"success": true})

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(subnet.Id)
	plan.TimeCreated = types.StringValue(subnet.TimeCreated.String())
	plan.TimeModified = types.StringValue(subnet.TimeModified.String())
	// IPV6Block is added as well as it is Optional/Computed
	plan.IPV6Block = types.StringValue(string(subnet.Ipv6Block))

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *vpcSubnetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state vpcSubnetResourceModel

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

	params := oxide.VpcSubnetViewParams{
		Subnet: oxide.NameOrId(state.ID.ValueString()),
	}
	subnet, err := r.client.VpcSubnetView(ctx, params)
	if err != nil {
		if is404(err) {
			// Remove resource from state during a refresh
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to read VPC subnet:",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("read VPC subnet with ID: %v", subnet.Id), map[string]any{"success": true})

	state.Description = types.StringValue(subnet.Description)
	state.ID = types.StringValue(subnet.Id)
	state.IPV4Block = types.StringValue(string(subnet.Ipv4Block))
	state.IPV6Block = types.StringValue(string(subnet.Ipv6Block))
	state.Name = types.StringValue(string(subnet.Name))
	state.VPCID = types.StringValue(subnet.VpcId)
	state.TimeCreated = types.StringValue(subnet.TimeCreated.String())
	state.TimeModified = types.StringValue(subnet.TimeModified.String())

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *vpcSubnetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan vpcSubnetResourceModel
	var state vpcSubnetResourceModel

	// Read Terraform plan data into the plan model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read Terraform prior state data into the state model to retrieve ID
	// which is a computed attribute, so it won't show up in the plan.
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout, diags := plan.Timeouts.Update(ctx, defaultTimeout())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	params := oxide.VpcSubnetUpdateParams{
		Subnet: oxide.NameOrId(state.ID.ValueString()),
		Body: &oxide.VpcSubnetUpdate{
			Description: plan.Description.ValueString(),
			Name:        oxide.Name(plan.Name.ValueString()),
		},
	}
	subnet, err := r.client.VpcSubnetUpdate(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating VPC subnet",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("updated VPC subnet with ID: %v", subnet.Id), map[string]any{"success": true})

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(subnet.Id)
	plan.TimeCreated = types.StringValue(subnet.TimeCreated.String())
	plan.TimeModified = types.StringValue(subnet.TimeModified.String())
	plan.IPV6Block = types.StringValue(string(subnet.Ipv6Block))

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *vpcSubnetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state vpcSubnetResourceModel

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

	params := oxide.VpcSubnetDeleteParams{
		Subnet: oxide.NameOrId(state.ID.ValueString()),
	}
	if err := r.client.VpcSubnetDelete(ctx, params); err != nil {
		if !is404(err) {
			resp.Diagnostics.AddError(
				"Error deleting VPC subnet:",
				"API error: "+err.Error(),
			)
			return
		}
	}
	tflog.Trace(ctx, fmt.Sprintf("deleted VPC subnet with ID: %v", state.ID.ValueString()), map[string]any{"success": true})
}

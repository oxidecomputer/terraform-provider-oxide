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
	_ resource.Resource              = (*vpcResource)(nil)
	_ resource.ResourceWithConfigure = (*vpcResource)(nil)
)

// NewVPCResource is a helper function to simplify the provider implementation.
func NewVPCResource() resource.Resource {
	return &vpcResource{}
}

// vpcResource is the resource implementation.
type vpcResource struct {
	client *oxide.Client
}

type vpcResourceModel struct {
	Description    types.String   `tfsdk:"description"`
	DNSName        types.String   `tfsdk:"dns_name"`
	ID             types.String   `tfsdk:"id"`
	IPV6Prefix     types.String   `tfsdk:"ipv6_prefix"`
	Name           types.String   `tfsdk:"name"`
	ProjectID      types.String   `tfsdk:"project_id"`
	SystemRouterID types.String   `tfsdk:"system_router_id"`
	TimeCreated    types.String   `tfsdk:"time_created"`
	TimeModified   types.String   `tfsdk:"time_modified"`
	Timeouts       timeouts.Value `tfsdk:"timeouts"`
}

// Metadata returns the resource type name.
func (r *vpcResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "oxide_vpc"
}

// Configure adds the provider configured client to the data source.
func (r *vpcResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*oxide.Client)
}

func (r *vpcResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Schema defines the schema for the resource.
func (r *vpcResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the project that will contain the VPC.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the VPC.",
			},
			"description": schema.StringAttribute{
				Required:    true,
				Description: "Description for the VPC.",
			},
			"dns_name": schema.StringAttribute{
				Required:    true,
				Description: "DNS Name of the VPC.",
			},
			"ipv6_prefix": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "DNS Name of the VPC.",
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
			},
			"system_router_id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier of the system router.",
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
func (r *vpcResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan vpcResourceModel

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

	params := oxide.VpcCreateParams{
		Project: oxide.NameOrId(plan.ProjectID.ValueString()),
		Body: &oxide.VpcCreate{
			Description: plan.Description.ValueString(),
			Name:        oxide.Name(plan.Name.ValueString()),
			DnsName:     oxide.Name(plan.DNSName.ValueString()),
			Ipv6Prefix:  oxide.Ipv6Net(plan.IPV6Prefix.ValueString()),
		},
	}
	vpc, err := r.client.VpcCreate(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating VPC",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("created VPC with ID: %v", vpc.Id), map[string]any{"success": true})

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(vpc.Id)
	plan.SystemRouterID = types.StringValue(vpc.SystemRouterId)
	plan.TimeCreated = types.StringValue(vpc.TimeCreated.String())
	plan.TimeModified = types.StringValue(vpc.TimeModified.String())
	// IPV6Prefix is added as well as it is Optional/Computed
	plan.IPV6Prefix = types.StringValue(string(vpc.Ipv6Prefix))

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *vpcResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state vpcResourceModel

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

	params := oxide.VpcViewParams{
		Vpc: oxide.NameOrId(state.ID.ValueString()),
	}
	vpc, err := r.client.VpcView(ctx, params)
	if err != nil {
		if is404(err) {
			// Remove resource from state during a refresh
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to read VPC:",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("read VPC with ID: %v", vpc.Id), map[string]any{"success": true})

	state.Description = types.StringValue(vpc.Description)
	state.DNSName = types.StringValue(string(vpc.DnsName))
	state.ID = types.StringValue(vpc.Id)
	state.IPV6Prefix = types.StringValue(string(vpc.Ipv6Prefix))
	state.Name = types.StringValue(string(vpc.Name))
	state.ProjectID = types.StringValue(vpc.ProjectId)
	state.SystemRouterID = types.StringValue(vpc.SystemRouterId)
	state.TimeCreated = types.StringValue(vpc.TimeCreated.String())
	state.TimeModified = types.StringValue(vpc.TimeModified.String())

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *vpcResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan vpcResourceModel
	var state vpcResourceModel

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

	params := oxide.VpcUpdateParams{
		Vpc: oxide.NameOrId(state.ID.ValueString()),
		Body: &oxide.VpcUpdate{
			Description: plan.Description.ValueString(),
			Name:        oxide.Name(plan.Name.ValueString()),
			DnsName:     oxide.Name(plan.DNSName.ValueString()),
		},
	}
	vpc, err := r.client.VpcUpdate(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating vpc",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("updated VPC with ID: %v", vpc.Id), map[string]any{"success": true})

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(vpc.Id)
	plan.SystemRouterID = types.StringValue(vpc.SystemRouterId)
	plan.TimeCreated = types.StringValue(vpc.TimeCreated.String())
	plan.TimeModified = types.StringValue(vpc.TimeModified.String())
	plan.IPV6Prefix = types.StringValue(string(vpc.Ipv6Prefix))

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *vpcResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state vpcResourceModel

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
	_, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	params := oxide.VpcDeleteParams{
		Vpc: oxide.NameOrId(state.ID.ValueString()),
	}
	if err := r.client.VpcDelete(ctx, params); err != nil {
		if !is404(err) {
			resp.Diagnostics.AddError(
				"Error deleting VPC:",
				"API error: "+err.Error(),
			)
			return
		}
	}
	tflog.Trace(ctx, fmt.Sprintf("deleted VPC with ID: %v", state.ID.ValueString()), map[string]any{"success": true})
}

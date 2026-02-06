// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/oxidecomputer/oxide.go/oxide"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = (*subnetPoolResource)(nil)
	_ resource.ResourceWithConfigure = (*subnetPoolResource)(nil)
)

// NewSubnetPoolResource is a helper function to simplify the provider implementation.
func NewSubnetPoolResource() resource.Resource {
	return &subnetPoolResource{}
}

// subnetPoolResource is the resource implementation.
type subnetPoolResource struct {
	client *oxide.Client
}

type subnetPoolResourceModel struct {
	Description  types.String   `tfsdk:"description"`
	ID           types.String   `tfsdk:"id"`
	IpVersion    types.String   `tfsdk:"ip_version"`
	Name         types.String   `tfsdk:"name"`
	TimeCreated  types.String   `tfsdk:"time_created"`
	TimeModified types.String   `tfsdk:"time_modified"`
	Timeouts     timeouts.Value `tfsdk:"timeouts"`
}

// Metadata returns the resource type name.
func (r *subnetPoolResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = "oxide_subnet_pool"
}

// Configure adds the provider configured client to the data source.
func (r *subnetPoolResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*oxide.Client)
}

func (r *subnetPoolResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Schema defines the schema for the resource.
func (r *subnetPoolResource) Schema(
	ctx context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource manages subnet pools. Use `oxide_subnet_pool_member` to add members to the pool.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the subnet pool.",
			},
			"description": schema.StringAttribute{
				Required:    true,
				Description: "Description for the subnet pool.",
			},
			"ip_version": schema.StringAttribute{
				Required:    true,
				Description: "The IP version for this pool. All subnets in the pool must match this version.",
				Validators: []validator.String{
					stringvalidator.OneOf("v4", "v6"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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
				Description: "Unique, immutable, system-controlled identifier of the subnet pool.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"time_created": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this subnet pool was created.",
			},
			"time_modified": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this subnet pool was last modified.",
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *subnetPoolResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan subnetPoolResourceModel

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

	params := oxide.SubnetPoolCreateParams{
		Body: &oxide.SubnetPoolCreate{
			Description: plan.Description.ValueString(),
			Name:        oxide.Name(plan.Name.ValueString()),
			IpVersion:   oxide.IpVersion(plan.IpVersion.ValueString()),
		},
	}
	pool, err := r.client.SubnetPoolCreate(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating subnet pool",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(
		ctx,
		fmt.Sprintf("created subnet pool with ID: %v", pool.Id),
		map[string]any{"success": true},
	)

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(pool.Id)
	plan.TimeCreated = types.StringValue(pool.TimeCreated.String())
	plan.TimeModified = types.StringValue(pool.TimeModified.String())

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *subnetPoolResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state subnetPoolResourceModel

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

	pool, err := r.client.SubnetPoolView(ctx, oxide.SubnetPoolViewParams{
		Pool: oxide.NameOrId(state.ID.ValueString()),
	})
	if err != nil {
		if is404(err) {
			// Remove resource from state during a refresh
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to read subnet pool:",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(
		ctx,
		fmt.Sprintf("read subnet pool with ID: %v", pool.Id),
		map[string]any{"success": true},
	)

	state.Description = types.StringValue(pool.Description)
	state.ID = types.StringValue(pool.Id)
	state.IpVersion = types.StringValue(string(pool.IpVersion))
	state.Name = types.StringValue(string(pool.Name))
	state.TimeCreated = types.StringValue(pool.TimeCreated.String())
	state.TimeModified = types.StringValue(pool.TimeModified.String())

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *subnetPoolResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan subnetPoolResourceModel
	var state subnetPoolResourceModel

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

	params := oxide.SubnetPoolUpdateParams{
		Pool: oxide.NameOrId(state.ID.ValueString()),
		Body: &oxide.SubnetPoolUpdate{
			Description: plan.Description.ValueString(),
			Name:        oxide.Name(plan.Name.ValueString()),
		},
	}

	pool, err := r.client.SubnetPoolUpdate(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating subnet pool",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(
		ctx,
		fmt.Sprintf("updated subnet pool with ID: %v", pool.Id),
		map[string]any{"success": true},
	)

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(pool.Id)
	plan.TimeCreated = types.StringValue(pool.TimeCreated.String())
	plan.TimeModified = types.StringValue(pool.TimeModified.String())

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *subnetPoolResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state subnetPoolResourceModel

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

	if err := r.client.SubnetPoolDelete(
		ctx,
		oxide.SubnetPoolDeleteParams{
			Pool: oxide.NameOrId(state.ID.ValueString()),
		}); err != nil {
		if !is404(err) {
			resp.Diagnostics.AddError(
				"Error deleting subnet pool:",
				"API error: "+err.Error(),
			)
			return
		}
	}
	tflog.Trace(
		ctx,
		fmt.Sprintf("deleted subnet pool with ID: %v", state.ID.ValueString()),
		map[string]any{"success": true},
	)
}

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
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
	_ resource.Resource              = (*vpcRouterResource)(nil)
	_ resource.ResourceWithConfigure = (*vpcRouterResource)(nil)
)

// NewVPCRouterResource is a helper function to simplify the provider implementation.
func NewVPCRouterResource() resource.Resource {
	return &vpcRouterResource{}
}

// vpcRouterResource is the resource implementation.
type vpcRouterResource struct {
	client *oxide.Client
}

type vpcRouterResourceModel struct {
	Description  types.String      `tfsdk:"description"`
	ID           types.String      `tfsdk:"id"`
	Kind         types.String      `tfsdk:"kind"`
	Name         types.String      `tfsdk:"name"`
	VPCID        types.String      `tfsdk:"vpc_id"`
	TimeCreated  timetypes.RFC3339 `tfsdk:"time_created"`
	TimeModified timetypes.RFC3339 `tfsdk:"time_modified"`
	Timeouts     timeouts.Value    `tfsdk:"timeouts"`
}

// Metadata returns the resource type name.
func (r *vpcRouterResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = "oxide_vpc_router"
}

// Configure adds the provider configured client to the data source.
func (r *vpcRouterResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*oxide.Client)
}

func (r *vpcRouterResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Schema defines the schema for the resource.
func (r *vpcRouterResource) Schema(
	ctx context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
This resource manages VPC routers.
`,
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the VPC router.",
			},
			"description": schema.StringAttribute{
				Required:    true,
				Description: "Description for the VPC Router.",
			},
			"vpc_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the VPC that will contain the VPC router.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier of the VPC router.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"kind": schema.StringAttribute{
				Computed:    true,
				Description: "Whether the VPC router is custom or system created.",
			},
			"time_created": schema.StringAttribute{
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
				Description: "Timestamp of when this VPC router was created.",
			},
			"time_modified": schema.StringAttribute{
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
				Description: "Timestamp of when this VPC router was last modified.",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *vpcRouterResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan vpcRouterResourceModel

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

	params := oxide.VpcRouterCreateParams{
		Vpc: oxide.NameOrId(plan.VPCID.ValueString()),
		Body: &oxide.VpcRouterCreate{
			Description: plan.Description.ValueString(),
			Name:        oxide.Name(plan.Name.ValueString()),
		},
	}

	vpcRouter, err := r.client.VpcRouterCreate(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating vpcRouter",
			"API error: "+err.Error(),
		)
		return
	}

	tflog.Trace(
		ctx,
		fmt.Sprintf("created VPC router with ID: %v", vpcRouter.Id),
		map[string]any{"success": true},
	)

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(vpcRouter.Id)
	plan.Kind = types.StringValue(string(vpcRouter.Kind))
	plan.TimeCreated = timetypes.NewRFC3339TimeValue(vpcRouter.TimeCreated.UTC())
	plan.TimeModified = timetypes.NewRFC3339TimeValue(vpcRouter.TimeModified.UTC())

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *vpcRouterResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state vpcRouterResourceModel

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

	params := oxide.VpcRouterViewParams{
		Router: oxide.NameOrId(state.ID.ValueString()),
	}
	vpcRouter, err := r.client.VpcRouterView(ctx, params)
	if err != nil {
		if is404(err) {
			// Remove resource from state during a refresh
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to read vpcRouter:",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(
		ctx,
		fmt.Sprintf("read VPC Router with ID: %v", vpcRouter.Id),
		map[string]any{"success": true},
	)

	state.Description = types.StringValue(vpcRouter.Description)
	state.ID = types.StringValue(vpcRouter.Id)
	state.Kind = types.StringValue(string(vpcRouter.Kind))
	state.Name = types.StringValue(string(vpcRouter.Name))
	state.VPCID = types.StringValue(string(vpcRouter.VpcId))
	state.TimeCreated = timetypes.NewRFC3339TimeValue(vpcRouter.TimeCreated.UTC())
	state.TimeModified = timetypes.NewRFC3339TimeValue(vpcRouter.TimeModified.UTC())

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *vpcRouterResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan vpcRouterResourceModel
	var state vpcRouterResourceModel

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

	params := oxide.VpcRouterUpdateParams{
		Router: oxide.NameOrId(state.ID.ValueString()),
		Body: &oxide.VpcRouterUpdate{
			Description: plan.Description.ValueString(),
			Name:        oxide.Name(plan.Name.ValueString()),
		},
	}
	vpcRouter, err := r.client.VpcRouterUpdate(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating VPC router",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(
		ctx,
		fmt.Sprintf("updated VPC router with ID: %v", vpcRouter.Id),
		map[string]any{"success": true},
	)

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(vpcRouter.Id)
	plan.Kind = types.StringValue(string(vpcRouter.Kind))
	plan.TimeCreated = timetypes.NewRFC3339TimeValue(vpcRouter.TimeCreated.UTC())
	plan.TimeModified = timetypes.NewRFC3339TimeValue(vpcRouter.TimeModified.UTC())

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *vpcRouterResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state vpcRouterResourceModel

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

	params := oxide.VpcRouterDeleteParams{
		Router: oxide.NameOrId(state.ID.ValueString()),
	}
	if err := r.client.VpcRouterDelete(ctx, params); err != nil {
		if !is404(err) {
			resp.Diagnostics.AddError(
				"Unable to delete VPC router:",
				"API error: "+err.Error(),
			)
			return
		}
	}

	tflog.Trace(
		ctx,
		fmt.Sprintf("deleted VPC router with ID: %v", state.ID.ValueString()),
		map[string]any{"success": true},
	)
}

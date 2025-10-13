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
	_ resource.Resource              = (*vpcRouterRouteResource)(nil)
	_ resource.ResourceWithConfigure = (*vpcRouterRouteResource)(nil)
)

// NewVPCRouterRouteResource is a helper function to simplify the provider implementation.
func NewVPCRouterRouteResource() resource.Resource {
	return &vpcRouterRouteResource{}
}

// vpcRouterRouteResource is the resource implementation.
type vpcRouterRouteResource struct {
	client *oxide.Client
}

type vpcRouterRouteResourceModel struct {
	Description  types.String                    `tfsdk:"description"`
	Destination  *vpcRouterRouteDestinationModel `tfsdk:"destination"`
	ID           types.String                    `tfsdk:"id"`
	Kind         types.String                    `tfsdk:"kind"`
	Name         types.String                    `tfsdk:"name"`
	Target       *vpcRouterRouteTargetModel      `tfsdk:"target"`
	TimeCreated  types.String                    `tfsdk:"time_created"`
	TimeModified types.String                    `tfsdk:"time_modified"`
	VPCRouterID  types.String                    `tfsdk:"vpc_router_id"`
	Timeouts     timeouts.Value                  `tfsdk:"timeouts"`
}

type vpcRouterRouteDestinationModel struct {
	Type  types.String `tfsdk:"type"`
	Value types.String `tfsdk:"value"`
}

type vpcRouterRouteTargetModel struct {
	Type  types.String `tfsdk:"type"`
	Value types.String `tfsdk:"value"`
}

// Metadata returns the resource type name.
func (r *vpcRouterRouteResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "oxide_vpc_router_route"
}

// Configure adds the provider configured client to the data source.
func (r *vpcRouterRouteResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*oxide.Client)
}

func (r *vpcRouterRouteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Schema defines the schema for the resource.
func (r *vpcRouterRouteResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
This resource manages VPC router routes.
`,
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the VPC router route.",
			},
			"description": schema.StringAttribute{
				Required:    true,
				Description: "Description for the VPC Router Route.",
			},
			"destination": schema.SingleNestedAttribute{
				Required:    true,
				Description: "Selects which traffic this routing rule will apply to",
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: "Route destination type. Possible values: `vpc`, `subnet`, `ip`, `ip_net`.",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.OneOf(
								string(oxide.RouteDestinationTypeIp),
								string(oxide.RouteDestinationTypeIpNet),
								string(oxide.RouteDestinationTypeSubnet),
								string(oxide.RouteDestinationTypeVpc),
							),
						},
					},
					"value": schema.StringAttribute{
						MarkdownDescription: replaceBackticks(`
Depending on the type, it will be one of the following:
  - ''vpc'': Name of the VPC
  - ''subnet'': Name of the VPC subnet
  - ''ip'': IP address
  - ''ip_net'': IPv4 or IPv6 subnet
`),
						Required: true,
					},
				},
			},
			"target": schema.SingleNestedAttribute{
				Required:    true,
				Description: "Location that matched packets should be forwarded to.",
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: "Route destination type. Possible values: `vpc`, `subnet`, `instance`, `ip`, `internet_gateway`, `drop`.",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.OneOf(
								string(oxide.RouteTargetTypeDrop),
								string(oxide.RouteTargetTypeIp),
								string(oxide.RouteTargetTypeInstance),
								string(oxide.RouteTargetTypeInternetGateway),
								string(oxide.RouteTargetTypeSubnet),
								string(oxide.RouteTargetTypeVpc),
							),
						},
					},
					"value": schema.StringAttribute{
						Description: replaceBackticks(`
Depending on the type, it will be one of the following:
  - ''vpc'': Name of the VPC
  - ''subnet'': Name of the VPC subnet
  - ''instance'': Name of the instance
  - ''ip'': IP address
  - ''internet_gateway'': Name of the internet gateway
`),
						Optional: true,
					},
				},
			},
			"vpc_router_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the VPC router route that will contain the route.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier of the VPC router route.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"kind": schema.StringAttribute{
				Computed:    true,
				Description: "Whether the VPC router route is custom or system created.",
			},
			"time_created": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this VPC router route was created.",
			},
			"time_modified": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this VPC router route was last modified.",
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
func (r *vpcRouterRouteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan vpcRouterRouteResourceModel

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

	params := oxide.VpcRouterRouteCreateParams{
		Router: oxide.NameOrId(plan.VPCRouterID.ValueString()),
		Body: &oxide.RouterRouteCreate{
			Description: plan.Description.ValueString(),
			Destination: oxide.RouteDestination{
				Type:  oxide.RouteDestinationType(plan.Destination.Type.ValueString()),
				Value: plan.Destination.Value.ValueString(),
			},
			Name: oxide.Name(plan.Name.ValueString()),
			Target: oxide.RouteTarget{
				Type: oxide.RouteTargetType(plan.Target.Type.ValueString()),
			},
		},
	}

	// When the target type is set to "drop" the value will be nil
	if !plan.Target.Value.IsNull() {
		params.Body.Target.Value = plan.Target.Value.ValueString()
	}

	vpcRouterRoute, err := r.client.VpcRouterRouteCreate(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating vpcRouterRoute",
			"API error: "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("created VPC router route with ID: %v", vpcRouterRoute.Id), map[string]any{"success": true})

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(vpcRouterRoute.Id)
	plan.Kind = types.StringValue(string(vpcRouterRoute.Kind))
	plan.TimeCreated = types.StringValue(vpcRouterRoute.TimeCreated.String())
	plan.TimeModified = types.StringValue(vpcRouterRoute.TimeModified.String())

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *vpcRouterRouteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state vpcRouterRouteResourceModel

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

	params := oxide.VpcRouterRouteViewParams{
		Route: oxide.NameOrId(state.ID.ValueString()),
	}
	vpcRouterRoute, err := r.client.VpcRouterRouteView(ctx, params)
	if err != nil {
		if is404(err) {
			// Remove resource from state during a refresh
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to read vpcRouterRoute:",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("read VPC RouterRoute with ID: %v", vpcRouterRoute.Id), map[string]any{"success": true})

	dm := vpcRouterRouteDestinationModel{
		Type:  types.StringValue(string(vpcRouterRoute.Destination.Type)),
		Value: types.StringValue(vpcRouterRoute.Destination.Value.(string)),
	}

	tm := vpcRouterRouteTargetModel{
		Type: types.StringValue(string(vpcRouterRoute.Target.Type)),
	}

	// When the target type is set to "drop" the value will be nil
	if vpcRouterRoute.Target.Value != nil {
		tm.Value = types.StringValue(vpcRouterRoute.Target.Value.(string))
	}

	state.Description = types.StringValue(vpcRouterRoute.Description)
	state.Destination = &dm
	state.ID = types.StringValue(vpcRouterRoute.Id)
	state.Kind = types.StringValue(string(vpcRouterRoute.Kind))
	state.Name = types.StringValue(string(vpcRouterRoute.Name))
	state.Target = &tm
	state.VPCRouterID = types.StringValue(vpcRouterRoute.VpcRouterId)
	state.TimeCreated = types.StringValue(vpcRouterRoute.TimeCreated.String())
	state.TimeModified = types.StringValue(vpcRouterRoute.TimeModified.String())

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *vpcRouterRouteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan vpcRouterRouteResourceModel
	var state vpcRouterRouteResourceModel

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

	params := oxide.VpcRouterRouteUpdateParams{
		Route: oxide.NameOrId(state.ID.ValueString()),
		Body: &oxide.RouterRouteUpdate{
			Description: plan.Description.ValueString(),
			Name:        oxide.Name(plan.Name.ValueString()),
			Destination: oxide.RouteDestination{
				Type:  oxide.RouteDestinationType(plan.Destination.Type.ValueString()),
				Value: plan.Destination.Value.ValueString(),
			},
			Target: oxide.RouteTarget{
				Type: oxide.RouteTargetType(plan.Target.Type.ValueString()),
			},
		},
	}

	// When the target type is set to "drop" the value will be nil
	if !plan.Target.Value.IsNull() {
		params.Body.Target.Value = plan.Target.Value.ValueString()
	}

	vpcRouterRoute, err := r.client.VpcRouterRouteUpdate(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating VPC router route",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("updated VPC router route with ID: %v", vpcRouterRoute.Id), map[string]any{"success": true})

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(vpcRouterRoute.Id)
	plan.Kind = types.StringValue(string(vpcRouterRoute.Kind))
	plan.TimeCreated = types.StringValue(vpcRouterRoute.TimeCreated.String())
	plan.TimeModified = types.StringValue(vpcRouterRoute.TimeModified.String())

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *vpcRouterRouteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state vpcRouterRouteResourceModel

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

	params := oxide.VpcRouterRouteDeleteParams{
		Route: oxide.NameOrId(state.ID.ValueString()),
	}
	if err := r.client.VpcRouterRouteDelete(ctx, params); err != nil {
		if !is404(err) {
			resp.Diagnostics.AddError(
				"Unable to delete VPC router route:",
				"API error: "+err.Error(),
			)
			return
		}
	}

	tflog.Trace(ctx, fmt.Sprintf("deleted VPC router route with ID: %v", state.ID.ValueString()), map[string]any{"success": true})
}

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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/oxidecomputer/oxide.go/oxide"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = (*vpcInternetGatewayResource)(nil)
	_ resource.ResourceWithConfigure = (*vpcInternetGatewayResource)(nil)
)

// NewVPCInternetGatewayResource is a helper function to simplify the provider implementation.
func NewVPCInternetGatewayResource() resource.Resource {
	return &vpcInternetGatewayResource{}
}

// vpcInternetGatewayResource is the resource implementation.
type vpcInternetGatewayResource struct {
	client *oxide.Client
}

type vpcInternetGatewayResourceModel struct {
	CascadeDelete types.Bool        `tfsdk:"cascade_delete"`
	Description   types.String      `tfsdk:"description"`
	ID            types.String      `tfsdk:"id"`
	Name          types.String      `tfsdk:"name"`
	VPCID         types.String      `tfsdk:"vpc_id"`
	TimeCreated   timetypes.RFC3339 `tfsdk:"time_created"`
	TimeModified  timetypes.RFC3339 `tfsdk:"time_modified"`
	Timeouts      timeouts.Value    `tfsdk:"timeouts"`
}

// Metadata returns the resource type name.
func (r *vpcInternetGatewayResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = "oxide_vpc_internet_gateway"
}

// Configure adds the provider configured client to the data source.
func (r *vpcInternetGatewayResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*oxide.Client)
}

// ImportState imports the resource state from Terraform.
func (r *vpcInternetGatewayResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Schema defines the schema for the resource.
func (r *vpcInternetGatewayResource) Schema(
	ctx context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
This resource manages VPC internet gateways.
`,
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the VPC internet gateway.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Required:    true,
				Description: "Description for the VPC internet gateway.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vpc_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the VPC that will contain the VPC internet gateway.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cascade_delete": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Description: "Whether to also delete routes targeting the VPC internet gateway " +
					"when deleting the VPC internet gateway.",
				Default: booldefault.StaticBool(false),
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier of the VPC internet gateway.",
			},
			"time_created": schema.StringAttribute{
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
				Description: "Timestamp of when this VPC internet gateway was created.",
			},
			"time_modified": schema.StringAttribute{
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
				Description: "Timestamp of when this VPC internet gateway was last modified.",
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
func (r *vpcInternetGatewayResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan vpcInternetGatewayResourceModel

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

	params := oxide.InternetGatewayCreateParams{
		Vpc: oxide.NameOrId(plan.VPCID.ValueString()),
		Body: &oxide.InternetGatewayCreate{
			Description: plan.Description.ValueString(),
			Name:        oxide.Name(plan.Name.ValueString()),
		},
	}

	vpcInternetGateway, err := r.client.InternetGatewayCreate(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating VPC internet gateway",
			"API error: "+err.Error(),
		)
		return
	}

	tflog.Trace(
		ctx,
		fmt.Sprintf("created VPC internet gateway with ID: %v", vpcInternetGateway.Id),
		map[string]any{"success": true},
	)

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(vpcInternetGateway.Id)
	plan.TimeCreated = timetypes.NewRFC3339TimeValue(vpcInternetGateway.TimeCreated.UTC())
	plan.TimeModified = timetypes.NewRFC3339TimeValue(vpcInternetGateway.TimeModified.UTC())

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *vpcInternetGatewayResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state vpcInternetGatewayResourceModel

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

	params := oxide.InternetGatewayViewParams{
		Gateway: oxide.NameOrId(state.ID.ValueString()),
	}
	vpcInternetGateway, err := r.client.InternetGatewayView(ctx, params)
	if err != nil {
		if is404(err) {
			// Remove resource from state during a refresh
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to read VPC internet gateway:",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(
		ctx,
		fmt.Sprintf("read VPC internet gateway with ID: %v", vpcInternetGateway.Id),
		map[string]any{"success": true},
	)

	state.Description = types.StringValue(vpcInternetGateway.Description)
	state.ID = types.StringValue(vpcInternetGateway.Id)
	state.Name = types.StringValue(string(vpcInternetGateway.Name))
	state.VPCID = types.StringValue(string(vpcInternetGateway.VpcId))
	state.TimeCreated = timetypes.NewRFC3339TimeValue(vpcInternetGateway.TimeCreated.UTC())
	state.TimeModified = timetypes.NewRFC3339TimeValue(vpcInternetGateway.TimeModified.UTC())

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *vpcInternetGatewayResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	// Internet gateways are currently not updateable.
	// We only update whether the user performs a cascade delete or not,
	// which is not part of the InternetGateway object, but rather a query
	// parameter which you only use during delete. This means we only want
	// to save it as part of the state.

	var plan vpcInternetGatewayResourceModel
	var state vpcInternetGatewayResourceModel

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

	params := oxide.InternetGatewayViewParams{
		Gateway: oxide.NameOrId(state.ID.ValueString()),
	}
	vpcInternetGateway, err := r.client.InternetGatewayView(ctx, params)
	if err != nil {
		if is404(err) {
			// Remove resource from state during a refresh
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to read VPC internet gateway:",
			"API error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(vpcInternetGateway.Id)
	plan.TimeCreated = timetypes.NewRFC3339TimeValue(vpcInternetGateway.TimeCreated.UTC())
	plan.TimeModified = timetypes.NewRFC3339TimeValue(vpcInternetGateway.TimeModified.UTC())

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *vpcInternetGatewayResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state vpcInternetGatewayResourceModel

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

	params := oxide.InternetGatewayDeleteParams{
		Gateway: oxide.NameOrId(state.ID.ValueString()),
		// We expect all routes to be managed by terraform so we shouldn't
		// use a cascade delete. If there are any dangling routes that
		// were manually created by the user, terraform shouldn't attemt
		// to handle them, and the user should manage them manually as well.
		Cascade: state.CascadeDelete.ValueBoolPointer(),
	}
	if err := r.client.InternetGatewayDelete(ctx, params); err != nil {
		if !is404(err) {
			resp.Diagnostics.AddError(
				"Unable to delete VPC internet gateway:",
				"API error: "+err.Error(),
			)
			return
		}
	}

	tflog.Trace(
		ctx,
		fmt.Sprintf("deleted VPC internet gateway with ID: %v", state.ID.ValueString()),
		map[string]any{"success": true},
	)
}

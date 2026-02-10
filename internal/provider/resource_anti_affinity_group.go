// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
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
	_ resource.Resource              = (*antiAffinityGroupResource)(nil)
	_ resource.ResourceWithConfigure = (*antiAffinityGroupResource)(nil)
)

// NewAntiAffinityGroupResource is a helper function to simplify the provider implementation.
func NewAntiAffinityGroupResource() resource.Resource {
	return &antiAffinityGroupResource{}
}

// antiAffinityGroupResource is the resource implementation.
type antiAffinityGroupResource struct {
	client *oxide.Client
}

type antiAffinityGroupResourceModel struct {
	Description   types.String      `tfsdk:"description"`
	FailureDomain types.String      `tfsdk:"failure_domain"`
	ID            types.String      `tfsdk:"id"`
	Name          types.String      `tfsdk:"name"`
	Policy        types.String      `tfsdk:"policy"`
	ProjectID     types.String      `tfsdk:"project_id"`
	TimeCreated   timetypes.RFC3339 `tfsdk:"time_created"`
	TimeModified  timetypes.RFC3339 `tfsdk:"time_modified"`
	Timeouts      timeouts.Value    `tfsdk:"timeouts"`
}

// Metadata returns the resource type name.
func (r *antiAffinityGroupResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = "oxide_anti_affinity_group"
}

// Configure adds the provider configured client to the data source.
func (r *antiAffinityGroupResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*oxide.Client)
}

// ImportState imports an existing anti-affinity group into Terraform state.
func (r *antiAffinityGroupResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Schema defines the schema for the resource.
func (r *antiAffinityGroupResource) Schema(
	ctx context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
This resource manages anti-affinity groups.
`,
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the project that will contain the anti-affinity group.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the anti-affinity group.",
			},
			"description": schema.StringAttribute{
				Required:    true,
				Description: "Description for the anti-affinity group.",
			},
			"policy": schema.StringAttribute{
				Required:    true,
				Description: "Affinity policy used to describe what to do when a request cannot be satisfied.",
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(oxide.AffinityPolicyAllow),
						string(oxide.AffinityPolicyFail),
					),
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
				Description: "Unique, immutable, system-controlled identifier of the anti-affinity group.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"failure_domain": schema.StringAttribute{
				// For now this will remain as a computed attribute as there is
				// only a single option: "sled".
				Computed:    true,
				Description: "Describes the scope of affinity for the purposes of co-location.",
			},
			"time_created": schema.StringAttribute{
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
				Description: "Timestamp of when this anti-affinity group was created.",
			},
			"time_modified": schema.StringAttribute{
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
				Description: "Timestamp of when this anti-affinity group was last modified.",
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *antiAffinityGroupResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan antiAffinityGroupResourceModel

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

	params := oxide.AntiAffinityGroupCreateParams{
		Project: oxide.NameOrId(plan.ProjectID.ValueString()),
		Body: &oxide.AntiAffinityGroupCreate{
			Description: plan.Description.ValueString(),
			Name:        oxide.Name(plan.Name.ValueString()),
			// TODO: For now, the only option is "sled", change this into
			// an attribute when there are more options.
			FailureDomain: oxide.FailureDomainSled,
			Policy:        oxide.AffinityPolicy(plan.Policy.ValueString()),
		},
	}
	antiAffinityGroup, err := r.client.AntiAffinityGroupCreate(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating AntiAffinityGroup",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(
		ctx,
		fmt.Sprintf("created AntiAffinityGroup with ID: %v", antiAffinityGroup.Id),
		map[string]any{"success": true},
	)

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(antiAffinityGroup.Id)
	plan.FailureDomain = types.StringValue(string(antiAffinityGroup.FailureDomain))
	plan.TimeCreated = timetypes.NewRFC3339TimeValue(antiAffinityGroup.TimeCreated.UTC())
	plan.TimeModified = timetypes.NewRFC3339TimeValue(antiAffinityGroup.TimeModified.UTC())

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *antiAffinityGroupResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state antiAffinityGroupResourceModel

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

	params := oxide.AntiAffinityGroupViewParams{
		AntiAffinityGroup: oxide.NameOrId(state.ID.ValueString()),
	}
	antiAffinityGroup, err := r.client.AntiAffinityGroupView(ctx, params)
	if err != nil {
		if is404(err) {
			// Remove resource from state during a refresh
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to read anti-affinity group:",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(
		ctx,
		fmt.Sprintf("read anti-affinity group with ID: %v", antiAffinityGroup.Id),
		map[string]any{"success": true},
	)

	state.Description = types.StringValue(antiAffinityGroup.Description)
	state.FailureDomain = types.StringValue(string(antiAffinityGroup.FailureDomain))
	state.ID = types.StringValue(antiAffinityGroup.Id)
	state.Policy = types.StringValue(string(antiAffinityGroup.Policy))
	state.Name = types.StringValue(string(antiAffinityGroup.Name))
	state.ProjectID = types.StringValue(antiAffinityGroup.ProjectId)
	state.TimeCreated = timetypes.NewRFC3339TimeValue(antiAffinityGroup.TimeCreated.UTC())
	state.TimeModified = timetypes.NewRFC3339TimeValue(antiAffinityGroup.TimeModified.UTC())

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *antiAffinityGroupResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan antiAffinityGroupResourceModel
	var state antiAffinityGroupResourceModel

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

	params := oxide.AntiAffinityGroupUpdateParams{
		AntiAffinityGroup: oxide.NameOrId(state.ID.ValueString()),
		Body: &oxide.AntiAffinityGroupUpdate{
			Description: plan.Description.ValueString(),
			Name:        oxide.Name(plan.Name.ValueString()),
		},
	}
	antiAffinityGroup, err := r.client.AntiAffinityGroupUpdate(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating anti-affinity group",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(
		ctx,
		fmt.Sprintf("updated anti-affinity group with ID: %v", antiAffinityGroup.Id),
		map[string]any{"success": true},
	)

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(antiAffinityGroup.Id)
	plan.FailureDomain = types.StringValue(string(antiAffinityGroup.FailureDomain))
	plan.TimeCreated = timetypes.NewRFC3339TimeValue(antiAffinityGroup.TimeCreated.UTC())
	plan.TimeModified = timetypes.NewRFC3339TimeValue(antiAffinityGroup.TimeModified.UTC())

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *antiAffinityGroupResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state antiAffinityGroupResourceModel

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

	params := oxide.AntiAffinityGroupDeleteParams{
		AntiAffinityGroup: oxide.NameOrId(state.ID.ValueString()),
	}
	if err := r.client.AntiAffinityGroupDelete(ctx, params); err != nil {
		if !is404(err) {
			resp.Diagnostics.AddError(
				"Error deleting AntiAffinityGroup:",
				"API error: "+err.Error(),
			)
			return
		}
	}
	tflog.Trace(
		ctx,
		fmt.Sprintf("deleted AntiAffinityGroup with ID: %v", state.ID.ValueString()),
		map[string]any{"success": true},
	)
}

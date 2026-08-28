// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package systemippool

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/oxidecomputer/oxide.go/oxide"

	legacyippool "github.com/oxidecomputer/terraform-provider-oxide/internal/provider/ip_pool"
	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/shared"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = (*Resource)(nil)
	_ resource.ResourceWithConfigure   = (*Resource)(nil)
	_ resource.ResourceWithImportState = (*Resource)(nil)
	_ resource.ResourceWithMoveState   = (*Resource)(nil)
)

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &Resource{}
}

// Resource is the resource implementation.
type Resource struct {
	client *oxide.Client
}

type ResourceModel struct {
	Description  types.String   `tfsdk:"description"`
	ID           types.String   `tfsdk:"id"`
	Name         types.String   `tfsdk:"name"`
	TimeCreated  types.String   `tfsdk:"time_created"`
	TimeModified types.String   `tfsdk:"time_modified"`
	Timeouts     timeouts.Value `tfsdk:"timeouts"`
}

// Metadata returns the resource type name.
func (r *Resource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = "oxide_system_ip_pool"
}

// Configure adds the provider configured client to the data source.
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
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// MoveState moves a legacy IP pool into a system IP pool. Ranges are omitted
// because each range must be imported into an oxide_system_ip_pool_range.
func (r *Resource) MoveState(ctx context.Context) []resource.StateMover {
	var sourceSchemaResponse resource.SchemaResponse
	legacyippool.NewResource().Schema(
		ctx,
		resource.SchemaRequest{},
		&sourceSchemaResponse,
	)

	return []resource.StateMover{{
		SourceSchema: &sourceSchemaResponse.Schema,
		StateMover: func(
			ctx context.Context,
			req resource.MoveStateRequest,
			resp *resource.MoveStateResponse,
		) {
			if req.SourceTypeName != "oxide_ip_pool" ||
				req.SourceSchemaVersion != 0 ||
				!strings.HasSuffix(
					req.SourceProviderAddress,
					"/oxidecomputer/oxide",
				) {
				return
			}
			if req.SourceState == nil {
				resp.Diagnostics.AddError(
					"Unable to Move Resource State",
					"The oxide_ip_pool state could not be decoded.",
				)
				return
			}

			var source legacyippool.ResourceModel
			resp.Diagnostics.Append(req.SourceState.Get(ctx, &source)...)
			if resp.Diagnostics.HasError() {
				return
			}

			target := ResourceModel{
				Description:  source.Description,
				ID:           source.ID,
				Name:         source.Name,
				TimeCreated:  source.TimeCreated,
				TimeModified: source.TimeModified,
				Timeouts:     source.Timeouts,
			}
			resp.Diagnostics.Append(resp.TargetState.Set(ctx, &target)...)
		},
	}}
}

// Schema defines the schema for the resource.
func (r *Resource) Schema(
	ctx context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
This resource manages system IP pools.
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier of the IP pool.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the IP pool.",
			},
			"description": schema.StringAttribute{
				Description: "Description for the IP pool.",
				Required:    true,
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Update: true,
				Delete: true,
			}),
			"time_created": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this IP pool was created.",
			},
			"time_modified": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this IP pool was last modified.",
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *Resource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan ResourceModel

	// Read Terraform plan data into the model
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

	params := oxide.SystemIpPoolCreateParams{
		Body: &oxide.IpPoolCreate{
			Name:        oxide.Name(plan.Name.ValueString()),
			Description: plan.Description.ValueString(),
		},
	}
	ipPool, err := r.client.SystemIpPoolCreate(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating IP Pool",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(
		ctx,
		fmt.Sprintf("created IP Pool with ID: %v", ipPool.Id),
		map[string]any{"success": true},
	)

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(ipPool.Id)
	plan.TimeCreated = types.StringValue(ipPool.TimeCreated.String())
	plan.TimeModified = types.StringValue(ipPool.TimeModified.String())

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *Resource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state ResourceModel

	// Read Terraform prior state data into the model
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

	ipPool, err := r.client.SystemIpPoolView(ctx, oxide.SystemIpPoolViewParams{
		Pool: oxide.NameOrId(state.ID.ValueString()),
	})
	if err != nil {
		if errors.Is(err, oxide.ErrHTTP404) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to read IP Pool:",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(
		ctx,
		fmt.Sprintf("read IP Pool with ID: %v", ipPool.Id),
		map[string]any{"success": true},
	)

	state.ID = types.StringValue(ipPool.Id)
	state.Name = types.StringValue(string(ipPool.Name))
	state.Description = types.StringValue(ipPool.Description)
	state.TimeCreated = types.StringValue(ipPool.TimeCreated.String())
	state.TimeModified = types.StringValue(ipPool.TimeModified.String())

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *Resource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan ResourceModel
	var state ResourceModel

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

	updateTimeout, diags := plan.Timeouts.Update(ctx, shared.DefaultTimeout())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	params := oxide.SystemIpPoolUpdateParams{
		Pool: oxide.NameOrId(state.ID.ValueString()),
		Body: &oxide.IpPoolUpdate{
			Name:        oxide.Name(plan.Name.ValueString()),
			Description: plan.Description.ValueString(),
		},
	}

	ipPool, err := r.client.SystemIpPoolUpdate(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating IP Pool",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(
		ctx,
		fmt.Sprintf("updated IP Pool with ID: %v", ipPool.Id),
		map[string]any{"success": true},
	)

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(ipPool.Id)
	plan.TimeCreated = types.StringValue(ipPool.TimeCreated.String())
	plan.TimeModified = types.StringValue(ipPool.TimeModified.String())

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *Resource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state ResourceModel

	// Read Terraform prior state data into the model
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

	if err := r.client.SystemIpPoolDelete(
		ctx,
		oxide.SystemIpPoolDeleteParams{
			Pool: oxide.NameOrId(state.ID.ValueString()),
		}); err != nil {
		if !errors.Is(err, oxide.ErrHTTP404) {
			resp.Diagnostics.AddError(
				"Error deleting IP Pool:",
				"API error: "+err.Error(),
			)
			return
		}
	}
	tflog.Trace(
		ctx,
		fmt.Sprintf("deleted IP pool with ID: %v", state.ID.ValueString()),
		map[string]any{"success": true},
	)
}

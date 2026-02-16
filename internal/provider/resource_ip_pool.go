// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	_ resource.Resource              = (*ipPoolResource)(nil)
	_ resource.ResourceWithConfigure = (*ipPoolResource)(nil)
)

// NewIPPoolResource is a helper function to simplify the provider implementation.
func NewIPPoolResource() resource.Resource {
	return &ipPoolResource{}
}

// ipPoolResource is the resource implementation.
type ipPoolResource struct {
	client *oxide.Client
}

type ipPoolResourceModel struct {
	Description  types.String               `tfsdk:"description"`
	ID           types.String               `tfsdk:"id"`
	Name         types.String               `tfsdk:"name"`
	Ranges       []ipPoolResourceRangeModel `tfsdk:"ranges"`
	TimeCreated  types.String               `tfsdk:"time_created"`
	TimeModified types.String               `tfsdk:"time_modified"`
	Timeouts     timeouts.Value             `tfsdk:"timeouts"`
}

type ipPoolResourceRangeModel struct {
	FirstAddress types.String `tfsdk:"first_address"`
	LastAddress  types.String `tfsdk:"last_address"`
}

// Metadata returns the resource type name.
func (r *ipPoolResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = "oxide_ip_pool"
}

// Configure adds the provider configured client to the data source.
func (r *ipPoolResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*oxide.Client)
}

func (r *ipPoolResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Schema defines the schema for the resource.
func (r *ipPoolResource) Schema(
	ctx context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
This resource manages IP pools.
`,
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the IP pool.",
			},
			"description": schema.StringAttribute{
				Required:    true,
				Description: "Description for the IP pool.",
			},
			"ranges": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"first_address": schema.StringAttribute{
							Description: "First address in the range.",
							Required:    true,
						},
						"last_address": schema.StringAttribute{
							Description: "Last address in the range.",
							Required:    true,
						},
					},
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
				Description: "Unique, immutable, system-controlled identifier of the IP pool.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
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
func (r *ipPoolResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan ipPoolResourceModel

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

	params := oxide.SystemIpPoolCreateParams{
		Body: &oxide.IpPoolCreate{
			Description: plan.Description.ValueString(),
			Name:        oxide.Name(plan.Name.ValueString()),
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

	resp.Diagnostics.Append(addRanges(ctx, r.client, plan.Ranges, plan.ID.ValueString())...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *ipPoolResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state ipPoolResourceModel

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

	ipPool, err := r.client.SystemIpPoolView(ctx, oxide.SystemIpPoolViewParams{
		Pool: oxide.NameOrId(state.ID.ValueString()),
	})
	if err != nil {
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

	state.Description = types.StringValue(ipPool.Description)
	state.ID = types.StringValue(ipPool.Id)
	state.Name = types.StringValue(string(ipPool.Name))
	state.TimeCreated = types.StringValue(ipPool.TimeCreated.String())
	state.TimeModified = types.StringValue(ipPool.TimeModified.String())

	// Append information about IP Pool ranges
	listParams := oxide.SystemIpPoolRangeListParams{
		Pool:  oxide.NameOrId(ipPool.Id),
		Limit: oxide.NewPointer(1000000000),
	}
	ipPoolRanges, err := r.client.SystemIpPoolRangeList(ctx, listParams)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read IP Pool ranges:",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(
		ctx,
		fmt.Sprintf("read all IP pool ranges from IP pool with ID: %v", ipPool.Id),
		map[string]any{"success": true},
	)

	// Set the size of the slice to avoid a panic when importing
	if len(state.Ranges) == 0 && len(ipPoolRanges.Items) != 0 {
		state.Ranges = make([]ipPoolResourceRangeModel, len(ipPoolRanges.Items))
	}

	for index, item := range ipPoolRanges.Items {
		ipPoolRange := ipPoolResourceRangeModel{}

		// Extract first/last addresses from the IpRange variant
		switch v := item.Range.Value.(type) {
		case *oxide.Ipv4Range:
			ipPoolRange.FirstAddress = types.StringValue(v.First)
			ipPoolRange.LastAddress = types.StringValue(v.Last)
		case *oxide.Ipv6Range:
			ipPoolRange.FirstAddress = types.StringValue(v.First)
			ipPoolRange.LastAddress = types.StringValue(v.Last)
		default:
			resp.Diagnostics.AddError(
				"Unable to read IP Pool ranges:",
				fmt.Sprintf(
					"internal error: unexpected IpRange variant type %T. If you hit this bug, please contact support",
					item.Range.Value,
				),
			)
			return
		}

		state.Ranges[index] = ipPoolRange
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *ipPoolResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan ipPoolResourceModel
	var state ipPoolResourceModel

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

	planRanges := plan.Ranges
	stateRanges := state.Ranges

	// Check plan and if it has a range that the state doesn't then attach it
	rangesToAdd := sliceDiff(planRanges, stateRanges)
	resp.Diagnostics.Append(addRanges(ctx, r.client, rangesToAdd, state.ID.ValueString())...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check state and if it has a range that the plan doesn't then detach it
	rangesToDetach := sliceDiff(stateRanges, planRanges)
	resp.Diagnostics.Append(removeRanges(ctx, r.client, rangesToDetach, state.ID.ValueString())...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := oxide.SystemIpPoolUpdateParams{
		Pool: oxide.NameOrId(state.ID.ValueString()),
		Body: &oxide.IpPoolUpdate{
			Description: plan.Description.ValueString(),
			Name:        oxide.Name(plan.Name.ValueString()),
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
func (r *ipPoolResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state ipPoolResourceModel

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

	// Remove all IP pool ranges first
	ranges, err := r.client.SystemIpPoolRangeList(
		ctx,
		oxide.SystemIpPoolRangeListParams{
			Pool:  oxide.NameOrId(state.ID.ValueString()),
			Limit: oxide.NewPointer(1000000000),
		},
	)
	if err != nil {
		if !is404(err) {
			resp.Diagnostics.AddError(
				"Error retrieving IP Pool ranges:",
				"API error: "+err.Error(),
			)
			return
		}
	}
	tflog.Trace(
		ctx,
		fmt.Sprintf("read all IP pool ranges from IP pool with ID: %v", state.ID.ValueString()),
		map[string]any{"success": true},
	)

	for _, item := range ranges.Items {
		// item.Range is now a struct with a Value field containing the variant
		params := oxide.SystemIpPoolRangeRemoveParams{
			Pool: oxide.NameOrId(state.ID.ValueString()),
			Body: &item.Range,
		}
		if err := r.client.SystemIpPoolRangeRemove(ctx, params); err != nil {
			if !is404(err) {
				resp.Diagnostics.AddError(
					"Error deleting IP Pool range:",
					"API error: "+err.Error(),
				)
				return
			}
		}
		tflog.Trace(ctx, fmt.Sprintf(
			"removed IP pool range %v from IP pool with ID: %v",
			item.Range.String(),
			state.ID.ValueString(),
		), map[string]any{"success": true})
	}

	if err := r.client.SystemIpPoolDelete(
		ctx,
		oxide.SystemIpPoolDeleteParams{
			Pool: oxide.NameOrId(state.ID.ValueString()),
		}); err != nil {
		if !is404(err) {
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

func addRanges(
	ctx context.Context,
	client *oxide.Client,
	ranges []ipPoolResourceRangeModel,
	poolID string,
) diag.Diagnostics {
	var diags diag.Diagnostics

	for _, ipPoolRange := range ranges {
		firstAddress := ipPoolRange.FirstAddress.ValueString()
		lastAddress := ipPoolRange.LastAddress.ValueString()

		body, err := oxide.NewIpRange(firstAddress, lastAddress)
		if err != nil {
			diags.AddError(
				"Error creating range within IP Pool",
				err.Error(),
			)
			return diags
		}

		params := oxide.SystemIpPoolRangeAddParams{
			Pool: oxide.NameOrId(poolID),
			Body: &body,
		}

		ipR, err := client.SystemIpPoolRangeAdd(ctx, params)
		if err != nil {
			diags.AddError(
				"Error creating range within IP Pool",
				"API error: "+err.Error(),
			)
			return diags
		}
		tflog.Trace(
			ctx,
			fmt.Sprintf(
				"added IP Pool range with ID: %v, from: %v to: %v",
				ipR.Id,
				firstAddress,
				lastAddress,
			),
			map[string]any{"success": true},
		)
	}

	return nil
}

func removeRanges(
	ctx context.Context,
	client *oxide.Client,
	ranges []ipPoolResourceRangeModel,
	poolID string,
) diag.Diagnostics {
	var diags diag.Diagnostics

	for _, ipPoolRange := range ranges {
		firstAddress := ipPoolRange.FirstAddress.ValueString()
		lastAddress := ipPoolRange.LastAddress.ValueString()

		body, err := oxide.NewIpRange(firstAddress, lastAddress)
		if err != nil {
			diags.AddError(
				"Error removing range within IP Pool",
				err.Error(),
			)
			return diags
		}

		params := oxide.SystemIpPoolRangeRemoveParams{
			Pool: oxide.NameOrId(poolID),
			Body: &body,
		}

		err = client.SystemIpPoolRangeRemove(ctx, params)
		if err != nil {
			diags.AddError(
				"Error removing range within IP Pool",
				"API error: "+err.Error(),
			)
			return diags
		}
		tflog.Trace(
			ctx,
			fmt.Sprintf("removed IP Pool range from: %v to: %v", firstAddress, lastAddress),
			map[string]any{"success": true},
		)
	}

	return nil
}

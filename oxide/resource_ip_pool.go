// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"context"
	"fmt"
	"reflect"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	oxideSDK "github.com/oxidecomputer/oxide.go/oxide"
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
	client *oxideSDK.Client
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
	ID           types.String `tfsdk:"id"`
	TimeCreated  types.String `tfsdk:"time_created"`
}

// Metadata returns the resource type name.
func (r *ipPoolResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "oxide_ip_pool"
}

// Configure adds the provider configured client to the data source.
func (r *ipPoolResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*oxideSDK.Client)
}

// Schema defines the schema for the resource.
func (r *ipPoolResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the IP Pool.",
			},
			"description": schema.StringAttribute{
				Required:    true,
				Description: "Description for the IP Pool.",
			},
			"ranges": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"first_address": schema.StringAttribute{
							Description: "First address in the range",
							Required:    true,
						},
						"last_address": schema.StringAttribute{
							Description: "Last address in the range",
							Required:    true,
						},
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "Unique, immutable, system-controlled identifier of the range.",
						},
						"time_created": schema.StringAttribute{
							Computed:    true,
							Description: "Timestamp of when this range was created.",
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
				Description: "Unique, immutable, system-controlled identifier of the IP Pool.",
			},
			"time_created": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this IP Pool was created.",
			},
			"time_modified": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this IP Pool was last modified.",
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *ipPoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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

	params := oxideSDK.IpPoolCreateParams{
		Body: &oxideSDK.IpPoolCreate{
			Description: plan.Description.ValueString(),
			Name:        oxideSDK.Name(plan.Name.ValueString()),
		},
	}
	ipPool, err := r.client.IpPoolCreate(params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating IP Pool",
			"API error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(ipPool.Id)
	plan.TimeCreated = types.StringValue(ipPool.TimeCreated.String())
	plan.TimeModified = types.StringValue(ipPool.TimeCreated.String())

	for index, ipPoolRange := range plan.Ranges {
		var body oxideSDK.IpRange

		// TODO: Error checking here can be improved by checking both addresses
		// TODO: Check if I really need the unquote if I use ValueString() instead
		firstAddress, err := strconv.Unquote(ipPoolRange.FirstAddress.String())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating range within IP Pool",
				err.Error(),
			)
			return
		}
		// TODO: Check if I really need the unquote if I use ValueString() instead
		lastAddress, err := strconv.Unquote(ipPoolRange.LastAddress.String())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating range within IP Pool",
				err.Error(),
			)
			return
		}
		if isIPv4(firstAddress) {
			body = oxideSDK.Ipv4Range{
				First: firstAddress,
				Last:  lastAddress,
			}
		} else if isIPv6(firstAddress) {
			body = oxideSDK.Ipv6Range{
				First: firstAddress,
				Last:  lastAddress,
			}
		} else {
			resp.Diagnostics.AddError(
				"Error creating range within IP Pool",
				fmt.Errorf("%s is neither a valid IPv4 or IPv6",
					firstAddress).Error(),
			)
			return
		}

		params := oxideSDK.IpPoolRangeAddParams{
			Pool: oxideSDK.NameOrId(plan.ID.ValueString()),
			Body: &body,
		}

		ipR, err := r.client.IpPoolRangeAdd(params)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating range within IP Pool",
				"API error: "+err.Error(),
			)
			return
		}

		ipPoolRange.ID = types.StringValue(ipR.Id)
		ipPoolRange.TimeCreated = types.StringValue(ipR.TimeCreated.String())

		plan.Ranges[index] = ipPoolRange
	}

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *ipPoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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

	ipPool, err := r.client.IpPoolView(oxideSDK.IpPoolViewParams{
		Pool: oxideSDK.NameOrId(state.ID.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read IP Pool:",
			"API error: "+err.Error(),
		)
		return
	}

	state.Description = types.StringValue(ipPool.Description)
	state.ID = types.StringValue(ipPool.Id)
	state.Name = types.StringValue(string(ipPool.Name))
	state.TimeCreated = types.StringValue(ipPool.TimeCreated.String())
	state.TimeModified = types.StringValue(ipPool.TimeCreated.String())

	// Append information about IP Pool ranges
	listParams := oxideSDK.IpPoolRangeListParams{
		Pool:  oxideSDK.NameOrId(ipPool.Id),
		Limit: 1000000000,
	}
	ipPoolRanges, err := r.client.IpPoolRangeList(listParams)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read IP Pool ranges:",
			"API error: "+err.Error(),
		)
		return
	}

	for index, item := range ipPoolRanges.Items {
		ipPoolRange := ipPoolResourceRangeModel{
			ID:          types.StringValue(item.Id),
			TimeCreated: types.StringValue(item.TimeCreated.String()),
		}

		// TODO: For the time being we are using interfaces for nested allOf within oneOf objects in
		// the OpenAPI spec. When we come up with a better approach this should be edited to reflect that.
		switch item.Range.(type) {
		case map[string]interface{}:
			rs := item.Range.(map[string]interface{})
			ipPoolRange.FirstAddress = types.StringValue(rs["first"].(string))
			ipPoolRange.LastAddress = types.StringValue(rs["last"].(string))
		default:
			// Theoretically this should never happen. Just in case though!
			resp.Diagnostics.AddError(
				"Unable to read IP Pool ranges:",
				fmt.Sprintf(
					"internal error: %v is not map[string]interface{}. Debugging content: %+v. If you hit this bug, please contact support",
					reflect.TypeOf(item.Range),
					item.Range,
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
func (r *ipPoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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

	// TODO: Support updates here
	if !reflect.DeepEqual(plan.Ranges, state.Ranges) {
		resp.Diagnostics.AddError(
			"Error updating IP Pool",
			"IP pool ranges cannot be updated; please revert to previous configuration",
		)
		return
	}

	params := oxideSDK.IpPoolUpdateParams{
		Pool: oxideSDK.NameOrId(state.ID.ValueString()),
		Body: &oxideSDK.IpPoolUpdate{
			Description: plan.Description.ValueString(),
			Name:        oxideSDK.Name(plan.Name.ValueString()),
		},
	}

	ipPool, err := r.client.IpPoolUpdate(params)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating IP Pool",
			"API error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(ipPool.Id)
	plan.TimeCreated = types.StringValue(ipPool.TimeCreated.String())
	plan.TimeModified = types.StringValue(ipPool.TimeCreated.String())

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *ipPoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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
	ranges, err := r.client.IpPoolRangeList(
		oxideSDK.IpPoolRangeListParams{
			Pool:  oxideSDK.NameOrId(state.ID.ValueString()),
			Limit: 1000000000,
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

	for _, item := range ranges.Items {
		var ipRange oxideSDK.IpRange
		rs := item.Range.(map[string]interface{})
		if isIPv4(rs["first"].(string)) {
			ipRange = oxideSDK.Ipv4Range{
				First: rs["first"].(string),
				Last:  rs["last"].(string),
			}
		} else if isIPv6(rs["first"].(string)) {
			ipRange = oxideSDK.Ipv6Range{
				First: rs["first"].(string),
				Last:  rs["last"].(string),
			}
		} else {
			// This should never happen as we are retrieving information from Nexus. If we do encounter
			// this error we have a huge problem.
			resp.Diagnostics.AddError(
				"Unable to read IP Pool ranges:",
				fmt.Sprintf(
					"internal error: %v is not map[string]interface{}. Debugging content: %+v. If you hit this bug, please contact support",
					reflect.TypeOf(item.Range),
					item.Range,
				),
			)
			return
		}

		params := oxideSDK.IpPoolRangeRemoveParams{
			Pool: oxideSDK.NameOrId(state.ID.ValueString()),
			Body: &ipRange,
		}
		if err := r.client.IpPoolRangeRemove(params); err != nil {
			if !is404(err) {
				resp.Diagnostics.AddError(
					"Error deleting IP Pool range:",
					"API error: "+err.Error(),
				)
				return
			}
		}
	}

	if err := r.client.IpPoolDelete(oxideSDK.IpPoolDeleteParams{
		Pool: oxideSDK.NameOrId(state.ID.ValueString()),
	}); err != nil {
		if !is404(err) {
			resp.Diagnostics.AddError(
				"Error deleting IP Pool:",
				"API error: "+err.Error(),
			)
			return
		}
	}
}

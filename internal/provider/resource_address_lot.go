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
	_ resource.Resource              = (*addressLotResource)(nil)
	_ resource.ResourceWithConfigure = (*addressLotResource)(nil)
)

// NewAddressLotResource is a helper function to simplify the provider implementation.
func NewAddressLotResource() resource.Resource {
	return &addressLotResource{}
}

// addressLotResource is the resource implementation.
type addressLotResource struct {
	client *oxide.Client
}

type addressLotResourceModel struct {
	Blocks       []addressLotResourceBlockModel `tfsdk:"blocks"`
	Description  types.String                   `tfsdk:"description"`
	Kind         types.String                   `tfsdk:"kind"`
	Name         types.String                   `tfsdk:"name"`
	ID           types.String                   `tfsdk:"id"`
	TimeCreated  types.String                   `tfsdk:"time_created"`
	TimeModified types.String                   `tfsdk:"time_modified"`
	Timeouts     timeouts.Value                 `tfsdk:"timeouts"`
}

type addressLotResourceBlockModel struct {
	ID           types.String `tfsdk:"id"`
	FirstAddress types.String `tfsdk:"first_address"`
	LastAddress  types.String `tfsdk:"last_address"`
}

// Metadata returns the resource type name.
func (r *addressLotResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "oxide_address_lot"
}

// Configure adds the provider configured client to the data source.
func (r *addressLotResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*oxide.Client)
}

func (r *addressLotResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Schema defines the schema for the resource.
func (r *addressLotResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier of the address lot.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the address lot.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Required:    true,
				Description: "Description for the address lot.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"kind": schema.StringAttribute{
				Required:    true,
				Description: `Kind for the address lot. Must be one of "infra" or "pool".`,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(oxide.AddressLotKindInfra),
						string(oxide.AddressLotKindPool),
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"blocks": schema.SetNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "ID of the address lot block.",
							Computed:    true,
						},
						"first_address": schema.StringAttribute{
							Description: "First address in the lot.",
							Required:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"last_address": schema.StringAttribute{
							Description: "Last address in the lot.",
							Required:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
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
			"time_created": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this address lot was created.",
			},
			"time_modified": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this address lot was last modified.",
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *addressLotResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan addressLotResourceModel

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

	blocks := make([]oxide.AddressLotBlockCreate, len(plan.Blocks))
	for i, block := range plan.Blocks {
		blocks[i] = oxide.AddressLotBlockCreate{
			FirstAddress: block.FirstAddress.ValueString(),
			LastAddress:  block.LastAddress.ValueString(),
		}
	}
	params := oxide.NetworkingAddressLotCreateParams{
		Body: &oxide.AddressLotCreate{
			Description: plan.Description.ValueString(),
			Name:        oxide.Name(plan.Name.ValueString()),
			Kind:        oxide.AddressLotKind(plan.Kind.ValueString()),
			Blocks:      blocks,
		},
	}
	lot, err := r.client.NetworkingAddressLotCreate(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating address lot",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("created address lot with ID: %v", lot.Lot.Id), map[string]any{"success": true})

	// Map response body to schema and populate computed attribute values.
	plan.ID = types.StringValue(lot.Lot.Id)
	plan.TimeCreated = types.StringValue(lot.Lot.TimeCreated.String())
	plan.TimeModified = types.StringValue(lot.Lot.TimeCreated.String())

	// Populate blocks with computed values.
	blockModels := make([]addressLotResourceBlockModel, len(lot.Blocks))
	for index, item := range lot.Blocks {
		blockModels[index] = addressLotResourceBlockModel{
			ID:           types.StringValue(item.Id),
			FirstAddress: types.StringValue(item.FirstAddress),
			LastAddress:  types.StringValue(item.LastAddress),
		}
	}
	plan.Blocks = blockModels

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *addressLotResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state addressLotResourceModel

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

	addressLot, err := r.client.NetworkingAddressLotView(ctx, oxide.NetworkingAddressLotViewParams{
		AddressLot: oxide.NameOrId(state.ID.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read address lot:",
			"API error: "+err.Error(),
		)
		return
	}
	lot := addressLot.Lot
	tflog.Trace(ctx, fmt.Sprintf("read address lot with ID: %v", lot.Id), map[string]any{"success": true})

	state.ID = types.StringValue(lot.Id)
	state.Name = types.StringValue(string(lot.Name))
	state.Kind = types.StringValue(string(lot.Kind))
	state.Description = types.StringValue(lot.Description)
	state.TimeCreated = types.StringValue(lot.TimeCreated.String())
	state.TimeModified = types.StringValue(lot.TimeCreated.String())

	blockModels := make([]addressLotResourceBlockModel, len(addressLot.Blocks))
	for index, item := range addressLot.Blocks {
		blockModels[index] = addressLotResourceBlockModel{
			ID:           types.StringValue(item.Id),
			FirstAddress: types.StringValue(item.FirstAddress),
			LastAddress:  types.StringValue(item.LastAddress),
		}
	}
	state.Blocks = blockModels

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
// Note: the API doesn't currently support updating an Address Lot in place, so we leave this implementation blank and mark all attributes with RequiresReplace.
// TODO: support in-place updates.
func (r *addressLotResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Error updating address lot",
		"the oxide API currently does not support updating address lots",
	)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *addressLotResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state addressLotResourceModel

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

	if err := r.client.NetworkingAddressLotDelete(
		ctx,
		oxide.NetworkingAddressLotDeleteParams{
			AddressLot: oxide.NameOrId(state.ID.ValueString()),
		}); err != nil {
		if !is404(err) {
			resp.Diagnostics.AddError(
				"Error deleting Address Lot:",
				"API error: "+err.Error(),
			)
			return
		}
	}
	tflog.Trace(ctx, fmt.Sprintf("deleted Address Lot with ID: %v", state.ID.ValueString()), map[string]any{"success": true})
}

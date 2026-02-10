// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/oxidecomputer/oxide.go/oxide"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                     = (*externalSubnetResource)(nil)
	_ resource.ResourceWithConfigure        = (*externalSubnetResource)(nil)
	_ resource.ResourceWithImportState      = (*externalSubnetResource)(nil)
	_ resource.ResourceWithConfigValidators = (*externalSubnetResource)(nil)
)

// NewExternalSubnetResource is a helper function to simplify the provider implementation.
func NewExternalSubnetResource() resource.Resource {
	return &externalSubnetResource{}
}

// externalSubnetResource is the resource implementation.
type externalSubnetResource struct {
	client *oxide.Client
}

type externalSubnetResourceModel struct {
	ID                 types.String       `tfsdk:"id"`
	Name               types.String       `tfsdk:"name"`
	Description        types.String       `tfsdk:"description"`
	ProjectID          types.String       `tfsdk:"project_id"`
	Subnet             cidrtypes.IPPrefix `tfsdk:"subnet"`
	PrefixLen          types.Int64        `tfsdk:"prefix_len"`
	SubnetPoolID       types.String       `tfsdk:"subnet_pool_id"`
	IPVersion          types.String       `tfsdk:"ip_version"`
	SubnetPoolMemberID types.String       `tfsdk:"subnet_pool_member_id"`
	InstanceID         types.String       `tfsdk:"instance_id"`
	TimeCreated        timetypes.RFC3339  `tfsdk:"time_created"`
	TimeModified       timetypes.RFC3339  `tfsdk:"time_modified"`
	Timeouts           timeouts.Value     `tfsdk:"timeouts"`
}

// Metadata returns the resource type name.
func (r *externalSubnetResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = "oxide_external_subnet"
}

// Configure adds the provider configured client to the data source.
func (r *externalSubnetResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*oxide.Client)
}

// ImportState imports an external subnet using its ID.
func (r *externalSubnetResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// ConfigValidators returns the config validators for the resource.
func (r *externalSubnetResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("subnet"),
			path.MatchRoot("prefix_len"),
		),
	}
}

// Schema defines the schema for the resource.
func (r *externalSubnetResource) Schema(
	ctx context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource manages external subnets allocated from subnet pools.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier of the external subnet.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Unique, mutable, user-controlled identifier for the external subnet.",
			},
			"description": schema.StringAttribute{
				Required:    true,
				Description: "Human-readable free-form text about the external subnet.",
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "Project ID where this external subnet is located.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"subnet": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				CustomType:          cidrtypes.IPPrefixType{},
				MarkdownDescription: "The subnet CIDR to reserve. Must be available in the pool. Conflicts with `prefix_len`. If unset, a subnet will be automatically allocated with the specified `prefix_len`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("prefix_len")),
					stringvalidator.ConflictsWith(path.MatchRoot("subnet_pool_id")),
					stringvalidator.ConflictsWith(path.MatchRoot("ip_version")),
				},
			},
			"prefix_len": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "The prefix length for automatic subnet allocation (e.g., 24 for a /24). Conflicts with `subnet`. Required when using automatic allocation.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.ConflictsWith(path.MatchRoot("subnet")),
					int64validator.Between(1, 128),
				},
			},
			"subnet_pool_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Subnet pool ID to allocate from. If unset when using automatic allocation (`prefix_len`), the silo's default subnet pool is used. Conflicts with `subnet`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("subnet")),
				},
			},
			"ip_version": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "IP version to use when multiple default pools exist. Required if both IPv4 and IPv6 default subnet pools are configured for the silo. Possible values: `v4`, `v6`. Conflicts with `subnet`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("subnet")),
					stringvalidator.OneOf(
						string(oxide.IpVersionV4),
						string(oxide.IpVersionV6),
					),
				},
			},
			"subnet_pool_member_id": schema.StringAttribute{
				Computed:    true,
				Description: "The subnet pool member this subnet was allocated from.",
			},
			"instance_id": schema.StringAttribute{
				Computed:    true,
				Description: "Instance ID this external subnet is attached to, if any.",
			},
			"time_created": schema.StringAttribute{
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
				Description: "Timestamp when this external subnet was created.",
			},
			"time_modified": schema.StringAttribute{
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
				Description: "Timestamp when this external subnet was last modified.",
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
func (r *externalSubnetResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan externalSubnetResourceModel

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

	params := oxide.ExternalSubnetCreateParams{
		Project: oxide.NameOrId(plan.ProjectID.ValueString()),
		Body: &oxide.ExternalSubnetCreate{
			Name:        oxide.Name(plan.Name.ValueString()),
			Description: plan.Description.ValueString(),
		},
	}

	if subnet := plan.Subnet.ValueString(); subnet != "" {
		// Explicit subnet allocation with pool inferred from CIDR.
		ipNet, err := oxide.NewIpNet(subnet)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error parsing subnet CIDR",
				err.Error(),
			)
			return
		}
		params.Body.Allocator = oxide.ExternalSubnetAllocator{
			Type:   oxide.ExternalSubnetAllocatorTypeExplicit,
			Subnet: ipNet,
		}
	} else {
		// Automatic allocation. The `prefix_len` attribute will always be defined if we reach this
		// code due to validation in ConfigValidators.
		prefixLen := int(plan.PrefixLen.ValueInt64())
		params.Body.Allocator = oxide.ExternalSubnetAllocator{
			Type:      oxide.ExternalSubnetAllocatorTypeAuto,
			PrefixLen: &prefixLen,
		}

		// Set pool selector
		if pool := plan.SubnetPoolID.ValueString(); pool != "" {
			// Auto subnet allocation from explicit pool.
			params.Body.Allocator.PoolSelector = oxide.PoolSelector{
				Type: oxide.PoolSelectorTypeExplicit,
				Pool: oxide.NameOrId(pool),
			}
		} else {
			// Auto subnet allocation from default pool. If there are multiple default pools IP
			// version is required.
			params.Body.Allocator.PoolSelector = oxide.PoolSelector{
				Type:      oxide.PoolSelectorTypeAuto,
				IpVersion: oxide.IpVersion(plan.IPVersion.ValueString()),
			}
		}
	}

	externalSubnet, err := r.client.ExternalSubnetCreate(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating external subnet",
			"API error: "+err.Error(),
		)
		return
	}

	tflog.Trace(
		ctx,
		fmt.Sprintf("created external subnet with ID: %v", externalSubnet.Id),
		map[string]any{"success": true},
	)

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(externalSubnet.Id)
	plan.Name = types.StringValue(string(externalSubnet.Name))
	plan.Description = types.StringValue(externalSubnet.Description)
	plan.ProjectID = types.StringValue(externalSubnet.ProjectId)
	plan.Subnet = cidrtypes.NewIPPrefixValue(externalSubnet.Subnet.String())
	plan.SubnetPoolID = types.StringValue(externalSubnet.SubnetPoolId)
	plan.SubnetPoolMemberID = types.StringValue(externalSubnet.SubnetPoolMemberId)
	plan.InstanceID = types.StringValue(externalSubnet.InstanceId)
	plan.TimeCreated = timetypes.NewRFC3339TimeValue(externalSubnet.TimeCreated.UTC())
	plan.TimeModified = timetypes.NewRFC3339TimeValue(externalSubnet.TimeModified.UTC())

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *externalSubnetResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state externalSubnetResourceModel

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

	params := oxide.ExternalSubnetViewParams{
		ExternalSubnet: oxide.NameOrId(state.ID.ValueString()),
	}

	externalSubnet, err := r.client.ExternalSubnetView(ctx, params)
	if err != nil {
		if is404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to read external subnet:",
			"API error: "+err.Error(),
		)
		return
	}

	tflog.Trace(
		ctx,
		fmt.Sprintf("read external subnet with ID: %v", externalSubnet.Id),
		map[string]any{"success": true},
	)

	state.ID = types.StringValue(externalSubnet.Id)
	state.Name = types.StringValue(string(externalSubnet.Name))
	state.Description = types.StringValue(externalSubnet.Description)
	state.ProjectID = types.StringValue(externalSubnet.ProjectId)
	state.Subnet = cidrtypes.NewIPPrefixValue(externalSubnet.Subnet.String())
	state.SubnetPoolID = types.StringValue(externalSubnet.SubnetPoolId)
	state.SubnetPoolMemberID = types.StringValue(externalSubnet.SubnetPoolMemberId)
	state.InstanceID = types.StringValue(externalSubnet.InstanceId)
	state.TimeCreated = timetypes.NewRFC3339TimeValue(externalSubnet.TimeCreated.UTC())
	state.TimeModified = timetypes.NewRFC3339TimeValue(externalSubnet.TimeModified.UTC())

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *externalSubnetResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan externalSubnetResourceModel
	var state externalSubnetResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

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

	params := oxide.ExternalSubnetUpdateParams{
		ExternalSubnet: oxide.NameOrId(state.ID.ValueString()),
		Body: &oxide.ExternalSubnetUpdate{
			Name:        oxide.Name(plan.Name.ValueString()),
			Description: plan.Description.ValueString(),
		},
	}

	externalSubnet, err := r.client.ExternalSubnetUpdate(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update external subnet:",
			"API error: "+err.Error(),
		)
		return
	}

	tflog.Trace(
		ctx,
		fmt.Sprintf("updated external subnet with ID: %v", externalSubnet.Id),
		map[string]any{"success": true},
	)

	plan.ID = types.StringValue(externalSubnet.Id)
	plan.Name = types.StringValue(string(externalSubnet.Name))
	plan.Description = types.StringValue(externalSubnet.Description)
	plan.ProjectID = types.StringValue(externalSubnet.ProjectId)
	plan.Subnet = cidrtypes.NewIPPrefixValue(externalSubnet.Subnet.String())
	plan.SubnetPoolID = types.StringValue(externalSubnet.SubnetPoolId)
	plan.SubnetPoolMemberID = types.StringValue(externalSubnet.SubnetPoolMemberId)
	plan.InstanceID = types.StringValue(externalSubnet.InstanceId)
	plan.TimeCreated = timetypes.NewRFC3339TimeValue(externalSubnet.TimeCreated.UTC())
	plan.TimeModified = timetypes.NewRFC3339TimeValue(externalSubnet.TimeModified.UTC())

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *externalSubnetResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state externalSubnetResourceModel

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

	params := oxide.ExternalSubnetDeleteParams{
		ExternalSubnet: oxide.NameOrId(state.ID.ValueString()),
	}

	if err := r.client.ExternalSubnetDelete(ctx, params); err != nil {
		if !is404(err) {
			resp.Diagnostics.AddError(
				"Error deleting external subnet:",
				"API error: "+err.Error(),
			)
			return
		}
	}

	tflog.Trace(
		ctx,
		fmt.Sprintf("deleted external subnet with ID: %v", state.ID.ValueString()),
		map[string]any{"success": true},
	)
}

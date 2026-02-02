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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/oxidecomputer/oxide.go/oxide"
)

// floatingIPResourceModel represents the Terraform configuration and state for
// the Oxide floating IP resource.
type floatingIPResourceModel struct {
	ID           types.String   `tfsdk:"id"`
	Name         types.String   `tfsdk:"name"`
	Description  types.String   `tfsdk:"description"`
	IP           types.String   `tfsdk:"ip"`
	InstanceID   types.String   `tfsdk:"instance_id"`
	IPPoolID     types.String   `tfsdk:"ip_pool_id"`
	IPVersion    types.String   `tfsdk:"ip_version"`
	ProjectID    types.String   `tfsdk:"project_id"`
	TimeCreated  types.String   `tfsdk:"time_created"`
	TimeModified types.String   `tfsdk:"time_modified"`
	Timeouts     timeouts.Value `tfsdk:"timeouts"`
}

// Compile-time assertions to check that the floatingIPResource implements the
// necessary Terraform resource interfaces.
var (
	_ resource.Resource                = (*floatingIPResource)(nil)
	_ resource.ResourceWithConfigure   = (*floatingIPResource)(nil)
	_ resource.ResourceWithImportState = (*floatingIPResource)(nil)
)

// floatingIPResource is the concrete type that implements the necessary
// Terraform resource interfaces. It holds state to interact with the Oxide API.
type floatingIPResource struct {
	client *oxide.Client
}

// NewFloatingIPResource is a helper to easily construct a floatingIPResource as
// a type that implements the Terraform resource interface.
func NewFloatingIPResource() resource.Resource {
	return &floatingIPResource{}
}

// Metadata configures the Terraform resource name for the Oxide floating IP
// resource.
func (f *floatingIPResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = "oxide_floating_ip"
}

// Configure sets up necessary data or clients needed by the floatingIPResource.
func (f *floatingIPResource) Configure(
	ctx context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	f.client = req.ProviderData.(*oxide.Client)
}

// ImportState imports an Oxide floating IP using its ID.
func (f *floatingIPResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Schema defines the attributes for this Oxide floating IP resource.
func (f *floatingIPResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
This resource manages Oxide floating IPs.
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier for the floating IP.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Unique, mutable, user-controlled identifier for the floating IP.",
			},
			"description": schema.StringAttribute{
				Required:    true,
				Description: "Human-readable free-form text about the floating IP.",
			},
			"instance_id": schema.StringAttribute{
				Computed:    true,
				Description: "Instance ID that this floating IP is attached to, if presently attached.",
			},
			"ip": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "IP address for this floating IP. If unset an IP address will be chosen from the given `ip_pool_id`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("ip_pool_id")),
					stringvalidator.ConflictsWith(path.MatchRoot("ip_version")),
				},
			},
			"ip_pool_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "IP pool ID to allocate this floating IP from. If unset the silo's default IP pool is used.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("ip")),
					stringvalidator.ConflictsWith(path.MatchRoot("ip_version")),
				},
			},
			"ip_version": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "IP version to use when multiple default pools exist. Required if both IPv4 and IPv6 default pools are configured. Possible values: `v4`, `v6`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("ip_pool_id")),
					stringvalidator.ConflictsWith(path.MatchRoot("ip")),
					stringvalidator.OneOf(
						string(oxide.IpVersionV4),
						string(oxide.IpVersionV6),
					),
				},
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "Project ID where this floating IP is located.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"time_created": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when this floating IP was created.",
			},
			"time_modified": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when this floating IP was last modified.",
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

// Create creates an Oxide floating IP using the Oxide API.
func (f *floatingIPResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan floatingIPResourceModel

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

	params := oxide.FloatingIpCreateParams{
		Project: oxide.NameOrId(plan.ProjectID.ValueString()),
		Body: &oxide.FloatingIpCreate{
			Name:        oxide.Name(plan.Name.ValueString()),
			Description: plan.Description.ValueString(),
		},
	}

	if ip := plan.IP.ValueString(); ip != "" {
		// Explicit IP with the pool inferred from IP address.
		params.Body.AddressAllocator = oxide.AddressAllocator{
			Type: oxide.AddressAllocatorTypeExplicit,
			Ip:   ip,
		}
	} else if pool := plan.IPPoolID.ValueString(); pool != "" {
		// Auto IP from explicit pool.
		params.Body.AddressAllocator = oxide.AddressAllocator{
			Type: oxide.AddressAllocatorTypeAuto,
			PoolSelector: oxide.PoolSelector{
				Type: oxide.PoolSelectorTypeExplicit,
				Pool: oxide.NameOrId(pool),
			},
		}
	} else {
		// Auto IP from default pool. If there are multiple default pools IP
		// version is required.
		params.Body.AddressAllocator = oxide.AddressAllocator{
			Type: oxide.AddressAllocatorTypeAuto,
			PoolSelector: oxide.PoolSelector{
				Type:      oxide.PoolSelectorTypeAuto,
				IpVersion: oxide.IpVersion(plan.IPVersion.ValueString()),
			},
		}
	}

	floatingIP, err := f.client.FloatingIpCreate(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating floating IP:",
			"API error: "+err.Error(),
		)
		return
	}

	tflog.Trace(
		ctx,
		fmt.Sprintf("created floating IP with ID: %v", floatingIP.Id),
		map[string]any{"success": true},
	)

	plan.ID = types.StringValue(floatingIP.Id)
	plan.Name = types.StringValue(string(floatingIP.Name))
	plan.Description = types.StringValue(floatingIP.Description)
	plan.IP = types.StringValue(floatingIP.Ip)
	plan.InstanceID = types.StringValue(floatingIP.InstanceId)
	plan.IPPoolID = types.StringValue(floatingIP.IpPoolId)
	plan.ProjectID = types.StringValue(floatingIP.ProjectId)
	plan.TimeCreated = types.StringValue(floatingIP.TimeCreated.String())
	plan.TimeModified = types.StringValue(floatingIP.TimeModified.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read fetches an Oxide floating IP from the Oxide API.
func (f *floatingIPResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state floatingIPResourceModel

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

	params := oxide.FloatingIpViewParams{
		FloatingIp: oxide.NameOrId(state.ID.ValueString()),
	}

	floatingIP, err := f.client.FloatingIpView(ctx, params)
	if err != nil {
		if is404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to read floating IP:",
			"API error: "+err.Error(),
		)
		return
	}

	tflog.Trace(
		ctx,
		fmt.Sprintf("read floating IP with ID: %v", floatingIP.Id),
		map[string]any{"success": true},
	)

	state.ID = types.StringValue(floatingIP.Id)
	state.Name = types.StringValue(string(floatingIP.Name))
	state.Description = types.StringValue(floatingIP.Description)
	state.IP = types.StringValue(floatingIP.Ip)
	state.InstanceID = types.StringValue(floatingIP.InstanceId)
	state.IPPoolID = types.StringValue(floatingIP.IpPoolId)
	state.ProjectID = types.StringValue(floatingIP.ProjectId)
	state.TimeCreated = types.StringValue(floatingIP.TimeCreated.String())
	state.TimeModified = types.StringValue(floatingIP.TimeModified.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates an Oxide floating IP using the Oxide API. Not all attributes
// can be updated. Refer to [Schema] and the floating_ip_update Oxide API
// documentation for more information.
func (f *floatingIPResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan floatingIPResourceModel
	var state floatingIPResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout, diags := state.Timeouts.Update(ctx, defaultTimeout())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	params := oxide.FloatingIpUpdateParams{
		FloatingIp: oxide.NameOrId(state.ID.ValueString()),
		Body: &oxide.FloatingIpUpdate{
			Description: plan.Description.ValueString(),
			Name:        oxide.Name(plan.Name.ValueString()),
		},
	}

	floatingIP, err := f.client.FloatingIpUpdate(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update floating IP:",
			"API error: "+err.Error(),
		)
		return
	}

	tflog.Trace(
		ctx,
		fmt.Sprintf("updated floating IP with ID: %v", floatingIP.Id),
		map[string]any{"success": true},
	)

	plan.ID = types.StringValue(floatingIP.Id)
	plan.Name = types.StringValue(string(floatingIP.Name))
	plan.Description = types.StringValue(floatingIP.Description)
	plan.IP = types.StringValue(floatingIP.Ip)
	plan.InstanceID = types.StringValue(floatingIP.InstanceId)
	plan.IPPoolID = types.StringValue(floatingIP.IpPoolId)
	plan.ProjectID = types.StringValue(floatingIP.ProjectId)
	plan.TimeCreated = types.StringValue(floatingIP.TimeCreated.String())
	plan.TimeModified = types.StringValue(floatingIP.TimeModified.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes an Oxide floating IP using the Oxide API.
func (f *floatingIPResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state floatingIPResourceModel

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

	params := oxide.FloatingIpDeleteParams{
		FloatingIp: oxide.NameOrId(state.ID.ValueString()),
	}

	if err := f.client.FloatingIpDelete(ctx, params); err != nil {
		if !is404(err) {
			resp.Diagnostics.AddError(
				"Unable to delete floating IP:",
				"API error: "+err.Error(),
			)
			return
		}
	}

	tflog.Trace(
		ctx,
		fmt.Sprintf("deleted floating IP with ID: %v", state.ID.ValueString()),
		map[string]any{"success": true},
	)
}

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	_ resource.Resource                 = (*vpcFirewallRulesResource)(nil)
	_ resource.ResourceWithConfigure    = (*vpcFirewallRulesResource)(nil)
	_ resource.ResourceWithUpgradeState = (*vpcFirewallRulesResource)(nil)
)

var (
	vpcFirewallRuleNameRegexp = regexp.MustCompile(`^[a-z][a-zA-Z0-9-]{0,61}[a-zA-Z0-9]$`)
)

// NewVPCFirewallRulesResource is a helper function to simplify the provider implementation.
func NewVPCFirewallRulesResource() resource.Resource {
	return &vpcFirewallRulesResource{}
}

// vpcFirewallRulesResource is the resource implementation.
type vpcFirewallRulesResource struct {
	client *oxide.Client
}

type vpcFirewallRulesResourceModel struct {
	// This ID is specific to Terraform only
	ID       types.String                                 `tfsdk:"id"`
	Rules    map[string]vpcFirewallRulesResourceRuleModel `tfsdk:"rules"`
	Timeouts timeouts.Value                               `tfsdk:"timeouts"`
	VPCID    types.String                                 `tfsdk:"vpc_id"`

	// Populated from the same fields within [vpcFirewallRulesResourceRuleModel].
	TimeCreated  types.String `tfsdk:"time_created"`
	TimeModified types.String `tfsdk:"time_modified"`
}

type vpcFirewallRulesResourceRuleModel struct {
	Action      types.String                              `tfsdk:"action"`
	Description types.String                              `tfsdk:"description"`
	Direction   types.String                              `tfsdk:"direction"`
	Filters     *vpcFirewallRulesResourceRuleFiltersModel `tfsdk:"filters"`
	Name        types.String                              `tfsdk:"name"`
	Priority    types.Int64                               `tfsdk:"priority"`
	Status      types.String                              `tfsdk:"status"`
	Targets     []vpcFirewallRulesResourceRuleTargetModel `tfsdk:"targets"`

	// Used to retrieve the timestamps from the API and populate the same fields
	// within [vpcFirewallRulesResourceModel]. The `tfsdk:"-"` struct field tag is used
	// to tell Terraform not to populate these values in the schema.
	TimeCreated  types.String `tfsdk:"-"`
	TimeModified types.String `tfsdk:"-"`
}

type vpcFirewallRulesResourceRuleTargetModel struct {
	Type  types.String `tfsdk:"type"`
	Value types.String `tfsdk:"value"`
}

type vpcFirewallRulesResourceRuleFiltersModel struct {
	Hosts     []vpcFirewallRuleHostFilterModel     `tfsdk:"hosts"`
	Ports     types.Set                            `tfsdk:"ports"`
	Protocols []vpcFirewallRuleProtocolFilterModel `tfsdk:"protocols"`
}

type vpcFirewallRuleHostFilterModel struct {
	Type  types.String `tfsdk:"type"`
	Value types.String `tfsdk:"value"`
}

type vpcFirewallRuleProtocolFilterModel struct {
	Type     types.String `tfsdk:"type"`
	IcmpType types.Int32  `tfsdk:"icmp_type"`
	IcmpCode types.String `tfsdk:"icmp_code"`
}

// Metadata returns the resource type name.
func (r *vpcFirewallRulesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "oxide_vpc_firewall_rules"
}

// Configure adds the provider configured client to the resource.
func (r *vpcFirewallRulesResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*oxide.Client)
}

// ImportState imports an existing VPC firewall rules resource into Terraform state.
func (r *vpcFirewallRulesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("vpc_id"), req, resp)
}

// UpgradeState upgrades the Terraform state for the oxide_vpc_firewall_rules
// resource from a previous schema version to the current version.
//
// Schema upgrades are not expected to be applied sequentially, since users are
// allowed to jump to whatever new version they choose. When adding a new
// version, you must ensure that each of the existing StateUpgrader functions
// are also updated to handle the new schema.
func (r *vpcFirewallRulesResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {StateUpgrader: r.stateUpgraderV01},
		1: {StateUpgrader: r.stateUpgraderV01},
	}
}

// Schema defines the schema for the resource.
func (r *vpcFirewallRulesResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	// TODO: Make sure users can define a single block per VPC ID, not many, is this even possible?
	resp.Schema = schema.Schema{
		Version: 2,
		MarkdownDescription: replaceBackticks(`
This resource manages VPC firewall rules.

!> Firewall rules defined by this resource are considered exhaustive and will
overwrite any other firewall rules for the VPC once applied.

!> Setting the ''rules'' attribute to ''{}'' will delete all firewall rules for the
VPC which may cause undesired network traffic. Please double check the firewall
rules when updating this resource.

### Migrating ''rules''

Previous versions of this resource stored firewall rules in a set. This
resulted in slow plans in environments with a significant number of rules.

Newer versions store the rules in a map for better performance, but this change
requires you to update your configuration files to:

1. Update the ''rules'' attribute from a set to a map.
2. Define the ''rules'' map keys as the VPC firewall rule name. Note that this
   key must then comply with the [Oxide
   API](https://docs.oxide.computer/api/vpc_firewall_rules_update) requirements
   for VPC firewall rule names.
3. Remove the ''name'' attribute from all entries of the ''rules'' map.

Previous ''rules'' schema:

''''''terraform
resource "oxide_vpc_firewall_rules" "example" {
  vpc_id = "6556fc6a-63c0-420b-bb23-c3205410f5cc"
  rules = [
    {
      name        = "allow-https"
      action      = "allow"
      description = "Allow HTTPS."
      # ...
    }
  ]
}
''''''

New ''rules'' schema:

''''''terraform
resource "oxide_vpc_firewall_rules" "example" {
  vpc_id = "6556fc6a-63c0-420b-bb23-c3205410f5cc"
  rules = {
    allow-https = {
      action      = "allow"
      description = "Allow HTTPS."
      # ...
    }
  }
}
''''''

You can use the ''provider::oxide::to_vpc_firewall_rules_map'' provider
function to help you convert existing rules, but note that this function is
provided as a temporary solution. You should update your configuration files
to use the new schema as soon as possible.
`),
		Attributes: map[string]schema.Attribute{
			"vpc_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the VPC that will have the firewall rules applied to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			// The rules attribute cannot contain computed attributes since the upstream API
			// returns updated attributes for every rule, irrespective of which rules actually
			// change. See https://github.com/oxidecomputer/terraform-provider-oxide/issues/453
			// for more information.
			"rules": schema.MapNestedAttribute{
				Required:    true,
				Description: "Associated firewall rules. The map key defines the rule name and must follow the API requirements for VPC firewall rule name.",
				Validators: []validator.Map{
					mapvalidator.KeysAre(
						stringvalidator.RegexMatches(
							vpcFirewallRuleNameRegexp,
							`Names must begin with a lower case ASCII letter, be composed exclusively of lowercase ASCII, uppercase ASCII, numbers, and '-', and may not end with a '-'. They can be at most 63 characters long.`,
						),
					),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"action": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Whether traffic matching the rule should be allowed or dropped. Possible values are: `allow` or `deny`.",
							Validators: []validator.String{
								stringvalidator.OneOf(
									string(oxide.VpcFirewallRuleActionAllow),
									string(oxide.VpcFirewallRuleActionDeny),
								),
							},
						},
						"description": schema.StringAttribute{
							Required:    true,
							Description: "Description for the VPC firewall rule.",
						},
						"direction": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Whether this rule is for incoming or outgoing traffic. Possible values are: `inbound` or `outbound`.",
							Validators: []validator.String{
								stringvalidator.OneOf(
									string(oxide.VpcFirewallRuleDirectionInbound),
									string(oxide.VpcFirewallRuleDirectionOutbound),
								),
							},
						},
						"filters": schema.SingleNestedAttribute{
							Required:    true,
							Description: "Reductions on the scope of the rule.",
							Attributes: map[string]schema.Attribute{
								"hosts": schema.SetNestedAttribute{
									Optional:    true,
									Description: "If present, the sources (if incoming) or destinations (if outgoing) this rule applies to.",
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"type": schema.StringAttribute{
												MarkdownDescription: "The rule applies to a single or all instances of this type, or specific IPs. Possible values: `vpc`, `subnet`, `instance`, `ip`, `ip_net`.",
												Required:            true,
												Validators: []validator.String{
													stringvalidator.OneOf(
														string(oxide.VpcFirewallRuleHostFilterTypeInstance),
														string(oxide.VpcFirewallRuleHostFilterTypeIp),
														string(oxide.VpcFirewallRuleHostFilterTypeIpNet),
														string(oxide.VpcFirewallRuleHostFilterTypeSubnet),
														string(oxide.VpcFirewallRuleHostFilterTypeVpc),
													),
												},
											},
											"value": schema.StringAttribute{
												// Important, if the name of the associated instance is changed Terraform will not be able to sync
												MarkdownDescription: replaceBackticks(`
Depending on the type, it will be one of the following:
  - ''vpc'': Name of the VPC.
  - ''subnet'': Name of the VPC subnet.
  - ''instance'': Name of the instance.
  - ''ip'': IP address.
  - ''ip_net'': IPv4 or IPv6 subnet.
 `),
												Required: true,
											},
										},
									},
									Validators: []validator.Set{
										setvalidator.SizeAtLeast(1),
									},
								},
								"protocols": schema.SetNestedAttribute{
									Description: "The protocols in a firewall rule's filter.",
									Optional:    true,
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"type": schema.StringAttribute{
												Required:    true,
												Description: "The protocol type. Must be one of `tcp`, `udp`, or `icmp`.",
												Validators: []validator.String{
													stringvalidator.OneOf(
														string(oxide.VpcFirewallRuleProtocolTypeTcp),
														string(oxide.VpcFirewallRuleProtocolTypeUdp),
														string(oxide.VpcFirewallRuleProtocolTypeIcmp),
													),
												},
											},
											"icmp_type": schema.Int32Attribute{
												Optional:    true,
												Description: "ICMP type. Only valid when type is `icmp`.",
												Validators: []validator.Int32{
													int32validator.Between(0, 255),
												},
											},
											"icmp_code": schema.StringAttribute{
												Optional:    true,
												Description: "ICMP code (e.g., 0) or range (e.g., 1-3). Omit to filter all traffic of the specified `icmp_type`. Only valid when type is `icmp` and `icmp_type` is provided.",
												Validators: []validator.String{
													stringvalidator.AlsoRequires(path.Expressions{
														path.MatchRelative().AtParent().AtName("icmp_type"),
													}...),
												},
											},
										},
									},
									Validators: []validator.Set{
										setvalidator.SizeAtLeast(1),
									},
								},
								"ports": schema.SetAttribute{
									Description: "If present, the destination ports this rule applies to.",
									Optional:    true,
									ElementType: types.StringType,
									Validators: []validator.Set{
										setvalidator.SizeAtLeast(1),
									},
								},
							},
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Name of the VPC firewall rule.",
						},
						"priority": schema.Int64Attribute{
							Required:    true,
							Description: "The relative priority of this rule.",
						},
						"status": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Whether this rule is in effect. Possible values are: `enabled` or `disabled`.",
							Validators: []validator.String{
								stringvalidator.OneOf(
									string(oxide.VpcFirewallRuleStatusDisabled),
									string(oxide.VpcFirewallRuleStatusEnabled),
								),
							},
						},
						"targets": schema.SetNestedAttribute{
							Required:    true,
							Description: "Sets of instances that the rule applies to.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"type": schema.StringAttribute{
										MarkdownDescription: "The rule applies to a single or all instances of this type, or specific IPs. Possible values: `vpc`, `subnet`, `instance`, `ip`, `ip_net`.",
										Required:            true,
										Validators: []validator.String{
											stringvalidator.OneOf(
												string(oxide.VpcFirewallRuleTargetTypeInstance),
												string(oxide.VpcFirewallRuleTargetTypeIp),
												string(oxide.VpcFirewallRuleTargetTypeIpNet),
												string(oxide.VpcFirewallRuleTargetTypeSubnet),
												string(oxide.VpcFirewallRuleTargetTypeVpc),
											),
										},
									},
									"value": schema.StringAttribute{
										// Important, if the name of the associated instance is changed Terraform will not be able to sync
										MarkdownDescription: replaceBackticks(`
Depending on the type, it will be one of the following:
  - ''vpc'': Name of the VPC.
  - ''subnet'': Name of the VPC subnet.
  - ''instance'': Name of the instance.
  - ''ip'': IP address.
  - ''ip_net'': IPv4 or IPv6 subnet.
`),
										Required: true,
									},
								},
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
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier of the firewall rules. Specific only to Terraform.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"time_created": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when the VPC firewall rules were last created.",
			},
			"time_modified": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when the VPC firewall rules were last modified.",
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *vpcFirewallRulesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan vpcFirewallRulesResourceModel

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

	params := oxide.VpcFirewallRulesUpdateParams{
		Vpc:  oxide.NameOrId(plan.VPCID.ValueString()),
		Body: newVPCFirewallRulesUpdateBody(plan.Rules),
	}

	firewallRules, err := r.client.VpcFirewallRulesUpdate(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating firewall rules",
			"API error: "+err.Error(),
		)
		return
	}

	if firewallRules != nil && len(firewallRules.Rules) > 0 {
		tflog.Trace(ctx, fmt.Sprintf("created firewall rules for VPC with ID: %v", firewallRules.Rules[0].VpcId), map[string]any{"success": true})
	}

	// Response does not include single ID for the set of rules.
	// This means we'll set it here solely for Terraform.
	plan.ID = types.StringValue(uuid.New().String())

	// The order of the response is not guaranteed to be the same as the one set
	// by the tf files. We will be populating all values, not just computed ones
	plan.Rules, diags = newVPCFirewallRulesModel(firewallRules.Rules)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.TimeCreated = types.StringNull()
	plan.TimeModified = types.StringNull()
	if len(plan.Rules) > 0 {
		for _, rule := range plan.Rules {
			plan.TimeCreated = rule.TimeCreated
			plan.TimeModified = rule.TimeModified
			break
		}
	}

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *vpcFirewallRulesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state vpcFirewallRulesResourceModel

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

	params := oxide.VpcFirewallRulesViewParams{
		Vpc: oxide.NameOrId(state.VPCID.ValueString()),
	}
	firewallRules, err := r.client.VpcFirewallRulesView(ctx, params)
	if err != nil {
		if is404(err) {
			// Remove resource from state during a refresh
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to read firewall rules:",
			"API error: "+err.Error(),
		)
		return
	}

	if firewallRules != nil && len(firewallRules.Rules) > 0 {
		tflog.Trace(ctx, fmt.Sprintf("read firewall rules for VPC with ID: %v", firewallRules.Rules[0].VpcId), map[string]any{"success": true})

		// We do not set ID as this was created solely for Terraform
		state.VPCID = types.StringValue(firewallRules.Rules[0].VpcId)
	}

	rules, diags := newVPCFirewallRulesModel(firewallRules.Rules)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Rules = rules

	state.TimeCreated = types.StringNull()
	state.TimeModified = types.StringNull()
	if len(state.Rules) > 0 {
		for _, rule := range state.Rules {
			state.TimeCreated = rule.TimeCreated
			state.TimeModified = rule.TimeModified
			break
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *vpcFirewallRulesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan vpcFirewallRulesResourceModel
	var state vpcFirewallRulesResourceModel

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

	params := oxide.VpcFirewallRulesUpdateParams{
		Vpc:  oxide.NameOrId(plan.VPCID.ValueString()),
		Body: newVPCFirewallRulesUpdateBody(plan.Rules),
	}
	firewallRules, err := r.client.VpcFirewallRulesUpdate(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating VPC firewall rules",
			"API error: "+err.Error(),
		)
		return
	}

	if firewallRules != nil && len(firewallRules.Rules) > 0 {
		tflog.Trace(ctx, fmt.Sprintf("updated firewall rules for VPC with ID: %v", firewallRules.Rules[0].VpcId), map[string]any{"success": true})
	}

	// Map response body to schema and populate Computed attribute values

	// We do not set ID from the response as this was created solely for Terraform
	plan.ID = state.ID

	// The order of the response is not guaranteed to be the same as the one set
	// by the tf files. We will be populating all values, not just computed ones
	plan.Rules, diags = newVPCFirewallRulesModel(firewallRules.Rules)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.TimeCreated = types.StringNull()
	plan.TimeModified = types.StringNull()
	if len(plan.Rules) > 0 {
		for _, rule := range plan.Rules {
			plan.TimeCreated = rule.TimeCreated
			plan.TimeModified = rule.TimeModified
			break
		}
	}

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *vpcFirewallRulesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state vpcFirewallRulesResourceModel

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

	// There is no delete endpoint; to delete we pass an empty body to the update endpoint
	params := oxide.VpcFirewallRulesUpdateParams{
		Vpc: oxide.NameOrId(state.VPCID.ValueString()),
		Body: &oxide.VpcFirewallRuleUpdateParams{
			Rules: []oxide.VpcFirewallRuleUpdate{},
		},
	}
	_, err := r.client.VpcFirewallRulesUpdate(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting VPC firewall rules",
			"API error: "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("deleted firewall rules for VPC with ID: %v", state.VPCID.ValueString()), map[string]any{"success": true})
}

// newVPCFirewallRulesUpdateBody builds the parameters required by the Oxide
// vpc_firewall_rules_update API using the specified rules.
func newVPCFirewallRulesUpdateBody(rules map[string]vpcFirewallRulesResourceRuleModel) *oxide.VpcFirewallRuleUpdateParams {
	// The make builtin is used to explicitly get an empty slice rather than a zero
	// value slice for the use case of removing all the firewall rules from a VPC.
	//
	// This is necessary because of the following.
	// * The vpc_firewall_rules_update API requires `{"rules": []}` to remove all rules.
	// * [oxide.VpcFirewallRuleUpdateParams] uses `omitzero` on its Rules field.
	updateRules := make([]oxide.VpcFirewallRuleUpdate, 0)
	body := new(oxide.VpcFirewallRuleUpdateParams)

	for ruleName, rule := range rules {
		r := oxide.VpcFirewallRuleUpdate{
			Action:      oxide.VpcFirewallRuleAction(rule.Action.ValueString()),
			Description: rule.Description.ValueString(),
			Direction:   oxide.VpcFirewallRuleDirection(rule.Direction.ValueString()),
			Name:        oxide.Name(ruleName),
			// We can safely dereference rule.Priority as it's a required field
			Priority: oxide.NewPointer(int(*rule.Priority.ValueInt64Pointer())),
			Status:   oxide.VpcFirewallRuleStatus(rule.Status.ValueString()),
			Filters:  newFilterTypeFromModel(rule.Filters),
			Targets:  newTargetTypeFromModel(rule.Targets),
		}

		updateRules = append(updateRules, r)
	}

	body.Rules = updateRules
	return body
}

// newVPCFirewallRulesModel translates a slice of [oxide.VpcFirewallRule] into a
// slice of [vpcFirewallRulesResourceRuleModel].
func newVPCFirewallRulesModel(rules []oxide.VpcFirewallRule) (map[string]vpcFirewallRulesResourceRuleModel, diag.Diagnostics) {
	// The make builtin is used to explicitly get an empty slice rather than a zero
	// value slice for the use case of removing all the firewall rules from a VPC.
	// See the comment within [newVPCFirewallRulesUpdateBody] for more information.
	model := make(map[string]vpcFirewallRulesResourceRuleModel)

	for _, rule := range rules {
		m := vpcFirewallRulesResourceRuleModel{
			Action:      types.StringValue(string(rule.Action)),
			Description: types.StringValue(rule.Description),
			Direction:   types.StringValue(string(rule.Direction)),
			Name:        types.StringValue(string(rule.Name)),
			// We can safely dereference rule.Priority as it's a required field
			Priority:     types.Int64Value(int64(*rule.Priority)),
			Status:       types.StringValue(string(rule.Status)),
			Targets:      newTargetsModelFromResponse(rule.Targets),
			TimeCreated:  types.StringValue(rule.TimeCreated.String()),
			TimeModified: types.StringValue(rule.TimeModified.String()),
		}

		filters, diags := newFiltersModelFromResponse(rule.Filters)
		diags.Append(diags...)
		if diags.HasError() {
			return nil, diags
		}
		m.Filters = filters

		model[string(rule.Name)] = m
	}

	return model, nil
}

func newFiltersModelFromResponse(filter oxide.VpcFirewallRuleFilter) (*vpcFirewallRulesResourceRuleFiltersModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	var hostsModel = []vpcFirewallRuleHostFilterModel{}
	for _, h := range filter.Hosts {
		m := vpcFirewallRuleHostFilterModel{
			Type:  types.StringValue(string(h.Type)),
			Value: types.StringValue(h.Value.(string)),
		}

		hostsModel = append(hostsModel, m)
	}

	var ports = []attr.Value{}
	for _, port := range filter.Ports {
		ports = append(ports, types.StringValue(string(port)))
	}
	portSet, diags := types.SetValue(types.StringType, ports)
	diags.Append(diags...)
	if diags.HasError() {
		return nil, diags
	}

	var protocolModels = []vpcFirewallRuleProtocolFilterModel{}
	for _, protocol := range filter.Protocols {
		protocolModel := vpcFirewallRuleProtocolFilterModel{
			Type: types.StringValue(string(protocol.Type)),
			IcmpCode: func() types.String {
				if protocol.Value.Code == "" {
					return types.StringNull()
				}
				return types.StringValue(string(protocol.Value.Code))
			}(),
			IcmpType: func() types.Int32 {
				if protocol.Value.IcmpType == nil {
					return types.Int32Null()
				}
				return types.Int32Value(int32(*protocol.Value.IcmpType))
			}(),
		}

		protocolModels = append(protocolModels, protocolModel)
	}

	model := vpcFirewallRulesResourceRuleFiltersModel{}

	if len(hostsModel) > 0 {
		model.Hosts = hostsModel
	}

	if len(portSet.Elements()) > 0 {
		model.Ports = portSet
	} else {
		model.Ports = types.SetNull(types.StringType)
	}

	if len(protocolModels) > 0 {
		model.Protocols = protocolModels
	} else {
		model.Protocols = nil
	}

	return &model, nil
}

func newTargetsModelFromResponse(target []oxide.VpcFirewallRuleTarget) []vpcFirewallRulesResourceRuleTargetModel {
	var model []vpcFirewallRulesResourceRuleTargetModel

	for _, t := range target {
		m := vpcFirewallRulesResourceRuleTargetModel{
			Type:  types.StringValue(string(t.Type)),
			Value: types.StringValue(t.Value.(string)),
		}

		model = append(model, m)
	}

	return model
}

func newFilterTypeFromModel(model *vpcFirewallRulesResourceRuleFiltersModel) oxide.VpcFirewallRuleFilter {
	var hosts []oxide.VpcFirewallRuleHostFilter
	for _, host := range model.Hosts {
		h := oxide.VpcFirewallRuleHostFilter{
			Type: oxide.VpcFirewallRuleHostFilterType(host.Type.ValueString()),
			// Note: This `Name` is a quirk from the SDK which should be fixed
			Value: oxide.Name(host.Value.ValueString()),
		}

		hosts = append(hosts, h)
	}

	ports := []oxide.L4PortRange{}
	for _, port := range model.Ports.Elements() {
		p, _ := strconv.Unquote(port.String())
		ports = append(ports, oxide.L4PortRange(p))
	}

	protocols := []oxide.VpcFirewallRuleProtocol{}
	for _, protocolModel := range model.Protocols {
		protocol := oxide.VpcFirewallRuleProtocol{
			Type: oxide.VpcFirewallRuleProtocolType(protocolModel.Type.ValueString()),
			Value: oxide.VpcFirewallIcmpFilter{
				Code: oxide.IcmpParamRange(protocolModel.IcmpCode.ValueString()),
				IcmpType: func() *int {
					if protocolModel.IcmpType.IsNull() {
						return nil
					}

					return oxide.NewPointer(int(protocolModel.IcmpType.ValueInt32()))
				}(),
			},
		}

		protocols = append(protocols, protocol)
	}

	return oxide.VpcFirewallRuleFilter{
		Hosts:     hosts,
		Ports:     ports,
		Protocols: protocols,
	}
}

func newTargetTypeFromModel(model []vpcFirewallRulesResourceRuleTargetModel) []oxide.VpcFirewallRuleTarget {
	var target []oxide.VpcFirewallRuleTarget

	for _, m := range model {
		t := oxide.VpcFirewallRuleTarget{
			Type:  oxide.VpcFirewallRuleTargetType(m.Type.ValueString()),
			Value: oxide.Name(m.Value.ValueString()),
		}
		target = append(target, t)
	}

	return target
}

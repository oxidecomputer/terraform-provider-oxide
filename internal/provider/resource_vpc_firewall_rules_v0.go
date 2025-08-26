// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/oxidecomputer/oxide.go/oxide"
)

// Schema and model definitions for upgrading the oxide_vpc_firewall_rules
// resource from version 0.
// https://github.com/oxidecomputer/terraform-provider-oxide/blob/v0.12.0/internal/provider/resource_vpc_firewall_rules.go

type vpcFirewallRulesResourceModelV0 struct {
	// This ID is specific to Terraform only
	ID       types.String                          `tfsdk:"id"`
	Rules    []vpcFirewallRulesResourceRuleModelV0 `tfsdk:"rules"`
	Timeouts timeouts.Value                        `tfsdk:"timeouts"`
	VPCID    types.String                          `tfsdk:"vpc_id"`

	// Populated from the same fields within [vpcFirewallRulesResourceRuleModel].
	TimeCreated  types.String `tfsdk:"time_created"`
	TimeModified types.String `tfsdk:"time_modified"`
}

func (m vpcFirewallRulesResourceModelV0) upgrade(ctx context.Context) (vpcFirewallRulesResourceModel, diag.Diagnostics) {
	res := vpcFirewallRulesResourceModel{
		ID:           m.ID,
		Timeouts:     m.Timeouts,
		VPCID:        m.VPCID,
		TimeCreated:  m.TimeCreated,
		TimeModified: m.TimeModified,

		Rules: make([]vpcFirewallRulesResourceRuleModel, len(m.Rules)),
	}

	var rulesDiags diag.Diagnostics
	for i, r := range m.Rules {
		rules, diags := r.upgrade(ctx)
		if diags.HasError() {
			rulesDiags.Append(diags...)
			continue
		}
		res.Rules[i] = rules
	}
	if rulesDiags.HasError() {
		return vpcFirewallRulesResourceModel{}, rulesDiags
	}

	return res, nil
}

type vpcFirewallRulesResourceRuleModelV0 struct {
	Action      types.String                                `tfsdk:"action"`
	Description types.String                                `tfsdk:"description"`
	Direction   types.String                                `tfsdk:"direction"`
	Filters     *vpcFirewallRulesResourceRuleFiltersModelV0 `tfsdk:"filters"`
	Name        types.String                                `tfsdk:"name"`
	Priority    types.Int64                                 `tfsdk:"priority"`
	Status      types.String                                `tfsdk:"status"`
	Targets     []vpcFirewallRulesResourceRuleTargetModel   `tfsdk:"targets"`

	// Used to retrieve the timestamps from the API and populate the same fields
	// within [vpcFirewallRulesResourceModel]. The `tfsdk:"-"` struct field tag is used
	// to tell Terraform not to populate these values in the schema.
	TimeCreated  types.String `tfsdk:"-"`
	TimeModified types.String `tfsdk:"-"`
}

func (r vpcFirewallRulesResourceRuleModelV0) upgrade(ctx context.Context) (vpcFirewallRulesResourceRuleModel, diag.Diagnostics) {
	filters, diags := r.Filters.upgrade(ctx)
	if diags.HasError() {
		return vpcFirewallRulesResourceRuleModel{}, diags
	}

	res := vpcFirewallRulesResourceRuleModel{
		Action:       r.Action,
		Description:  r.Description,
		Direction:    r.Direction,
		Filters:      filters,
		Name:         r.Name,
		Priority:     r.Priority,
		Status:       r.Status,
		Targets:      r.Targets,
		TimeCreated:  r.TimeCreated,
		TimeModified: r.TimeModified,
	}
	return res, nil
}

type vpcFirewallRulesResourceRuleFiltersModelV0 struct {
	Hosts     []vpcFirewallRuleHostFilterModel `tfsdk:"hosts"`
	Ports     types.Set                        `tfsdk:"ports"`
	Protocols types.Set                        `tfsdk:"protocols"`
}

func (f vpcFirewallRulesResourceRuleFiltersModelV0) upgrade(ctx context.Context) (*vpcFirewallRulesResourceRuleFiltersModel, diag.Diagnostics) {
	var protocols []types.String
	diag := f.Protocols.ElementsAs(ctx, &protocols, false)
	if diag.HasError() {
		return nil, diag
	}

	res := &vpcFirewallRulesResourceRuleFiltersModel{
		Hosts:     f.Hosts,
		Ports:     f.Ports,
		Protocols: make([]vpcFirewallRuleProtocolFilterModel, len(protocols)),
	}
	for i, p := range protocols {
		res.Protocols[i] = vpcFirewallRuleProtocolFilterModel{
			Type: types.StringValue(strings.ToLower(p.ValueString())),
		}
	}

	return res, nil
}

func (r *vpcFirewallRulesResource) schemaV0(ctx context.Context) *schema.Schema {
	return &schema.Schema{
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
			"rules": schema.SetNestedAttribute{
				Required:    true,
				Description: "Associated firewall rules.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"action": schema.StringAttribute{
							Required:    true,
							Description: "Whether traffic matching the rule should be allowed or dropped. Possible values are: allow or deny",
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
							Required:    true,
							Description: "Whether this rule is for incoming or outgoing traffic. Possible values are: inbound or outbound",
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
												Description: "The rule applies to a single or all instances of this type, or specific IPs. Possible values: vpc, subnet, instance, ip, ip_net",
												Required:    true,
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
												Description: "Depending on the type, it will be one of the following:" +
													"- `vpc`: Name of the VPC " +
													"- `subnet`: Name of the VPC subnet " +
													"- `instance`: Name of the instance " +
													"- `ip`: IP address " +
													"- `ip_net`: IPv4 or IPv6 subnet",
												Required: true,
											},
										},
									},
									Validators: []validator.Set{
										setvalidator.SizeAtLeast(1),
									},
								},
								"protocols": schema.SetAttribute{
									Description: "If present, the networking protocols this rule applies to. Possible values are: TCP, UDP and ICMP.",
									Optional:    true,
									ElementType: types.StringType,
									Validators: []validator.Set{
										setvalidator.ValueStringsAre(stringvalidator.Any(
											stringvalidator.OneOf(
												string("TCP"),
												string("UDP"),
												string("ICMP"),
											),
										)),
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
							Required:    true,
							Description: "Name of the VPC firewall rule.",
						},
						"priority": schema.Int64Attribute{
							Required:    true,
							Description: "The relative priority of this rule.",
						},
						"status": schema.StringAttribute{
							Required:    true,
							Description: "Whether this rule is in effect. Possible values are: enabled or disabled",
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
										Description: "The rule applies to a single or all instances of this type, or specific IPs. Possible values: vpc, subnet, instance, ip, ip_net",
										Required:    true,
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
										Description: "Depending on the type, it will be one of the following:" +
											"- `vpc`: Name of the VPC " +
											"- `subnet`: Name of the VPC subnet " +
											"- `instance`: Name of the instance " +
											"- `ip`: IP address " +
											"- `ip_net`: IPv4 or IPv6 subnet",
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

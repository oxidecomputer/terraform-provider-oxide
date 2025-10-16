// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/oxidecomputer/oxide.go/oxide"
)

// Schemas and model definitions for upgrading the oxide_vpc_firewall_rules
// resource from version 0.
//
// Version 0.13.0 of the provider had breaking changes to the schema without
// a version bump, so there are two different schemas versioned as 0. The
// original schema (present in provider versions <0.13.0) is referred to here
// as v0.0, and the schema with breaking changes (present in provider version
// =0.13.0) is referred to as v0.1
//
// Schemas and structs are mostly copied from their original sources.
// https://github.com/oxidecomputer/terraform-provider-oxide/blob/v0.12.0/internal/provider/resource_vpc_firewall_rules.go
// https://github.com/oxidecomputer/terraform-provider-oxide/blob/v0.13.0/internal/provider/resource_vpc_firewall_rules.go

// stateUpgraderV0 is a StateUpgrader function to upgrades an
// oxide_vpc_firewall_rules resource from schema version 0 to latest.
//
// It must be able to handle both versions of schema v0. States in schema v0.0
// are first upgraded to v0.1, and then the v0.1 state is upgraded to the
// latest schema version. This way future upgrades only need to handle
// upgrades from v0.1.
func (r *vpcFirewallRulesResource) stateUpgraderV0(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	var modelV01 vpcFirewallRulesResourceModelV01

	// Check if we need to upgrade from v0.0 to v0.1 first.
	schemaV00 := r.schemaV00(ctx)
	rawStateValue, err := req.RawState.Unmarshal(schemaV00.Type().TerraformType(ctx))
	if err == nil {
		stateV00 := &tfsdk.State{
			Raw:    rawStateValue,
			Schema: schemaV00,
		}

		var modelV00 vpcFirewallRulesResourceModelV00
		resp.Diagnostics.Append(stateV00.Get(ctx, &modelV00)...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Upgrade from schema v0.0 to v0.1.
		var diags diag.Diagnostics
		modelV01, diags = modelV00.upgrade(ctx)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		tflog.Info(ctx, "failed to unmarshal state using schema v0.0, trying v0.1", map[string]any{"err": err})

		// Unmarshalling to schema v0.0 failed, try with schema v0.1.
		schemaV01 := r.schemaV01(ctx)
		rawStateValue, err := req.RawState.Unmarshal(schemaV01.Type().TerraformType(ctx))
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Upgraded State From Version 0",
				fmt.Sprintf("failed to convert state to schema version 0.1: %v", err),
			)
			return
		}

		stateV01 := &tfsdk.State{
			Raw:    rawStateValue,
			Schema: schemaV01,
		}
		resp.Diagnostics.Append(stateV01.Get(ctx, &modelV01)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Upgrade from schema v0.1 to latest.
	newState := modelV01.upgrade()
	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

func (r *vpcFirewallRulesResource) schemaV00(ctx context.Context) schema.Schema {
	return schema.Schema{
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
												"TCP",
												"UDP",
												"ICMP",
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

type vpcFirewallRulesResourceModelV00 struct {
	// This ID is specific to Terraform only
	ID       types.String                           `tfsdk:"id"`
	Rules    []vpcFirewallRulesResourceRuleModelV00 `tfsdk:"rules"`
	Timeouts timeouts.Value                         `tfsdk:"timeouts"`
	VPCID    types.String                           `tfsdk:"vpc_id"`

	// Populated from the same fields within [vpcFirewallRulesResourceRuleModel].
	TimeCreated  types.String `tfsdk:"time_created"`
	TimeModified types.String `tfsdk:"time_modified"`
}

func (m vpcFirewallRulesResourceModelV00) upgrade(ctx context.Context) (vpcFirewallRulesResourceModelV01, diag.Diagnostics) {
	var diags diag.Diagnostics

	rules := make([]vpcFirewallRulesResourceRuleModelV01, len(m.Rules))
	for i, r := range m.Rules {
		rule, ruleDiags := r.upgrade(ctx)
		if ruleDiags.HasError() {
			diags.Append(ruleDiags...)
			continue
		}
		rules[i] = rule
	}

	if diags.HasError() {
		return vpcFirewallRulesResourceModelV01{}, diags
	}

	return vpcFirewallRulesResourceModelV01{
		ID:           m.ID,
		Rules:        rules,
		Timeouts:     m.Timeouts,
		VPCID:        m.VPCID,
		TimeCreated:  m.TimeCreated,
		TimeModified: m.TimeModified,
	}, nil
}

type vpcFirewallRulesResourceRuleModelV00 struct {
	Action      types.String                                 `tfsdk:"action"`
	Description types.String                                 `tfsdk:"description"`
	Direction   types.String                                 `tfsdk:"direction"`
	Filters     *vpcFirewallRulesResourceRuleFiltersModelV00 `tfsdk:"filters"`
	Name        types.String                                 `tfsdk:"name"`
	Priority    types.Int64                                  `tfsdk:"priority"`
	Status      types.String                                 `tfsdk:"status"`
	Targets     []vpcFirewallRulesResourceRuleTargetModelV00 `tfsdk:"targets"`

	// Used to retrieve the timestamps from the API and populate the same fields
	// within [vpcFirewallRulesResourceModel]. The `tfsdk:"-"` struct field tag is used
	// to tell Terraform not to populate these values in the schema.
	TimeCreated  types.String `tfsdk:"-"`
	TimeModified types.String `tfsdk:"-"`
}

func (r vpcFirewallRulesResourceRuleModelV00) upgrade(ctx context.Context) (vpcFirewallRulesResourceRuleModelV01, diag.Diagnostics) {
	var diags diag.Diagnostics

	targets := make([]vpcFirewallRulesResourceRuleTargetModelV01, len(r.Targets))
	for i, t := range r.Targets {
		targets[i] = vpcFirewallRulesResourceRuleTargetModelV01(t)
	}

	filters, filtersDiags := r.Filters.upgrade(ctx)
	diags.Append(filtersDiags...)

	if diags.HasError() {
		return vpcFirewallRulesResourceRuleModelV01{}, diags
	}

	return vpcFirewallRulesResourceRuleModelV01{
		Action:       r.Action,
		Description:  r.Description,
		Direction:    r.Direction,
		Filters:      filters,
		Name:         r.Name,
		Priority:     r.Priority,
		Status:       r.Status,
		Targets:      targets,
		TimeCreated:  r.TimeCreated,
		TimeModified: r.TimeModified,
	}, nil
}

type vpcFirewallRulesResourceRuleFiltersModelV00 struct {
	Hosts     []vpcFirewallRuleHostFilterModelV00 `tfsdk:"hosts"`
	Ports     types.Set                           `tfsdk:"ports"`
	Protocols types.Set                           `tfsdk:"protocols"`
}

func (f *vpcFirewallRulesResourceRuleFiltersModelV00) upgrade(ctx context.Context) (*vpcFirewallRulesResourceRuleFiltersModelV01, diag.Diagnostics) {
	if f == nil {
		return nil, nil
	}

	hosts := make([]vpcFirewallRuleHostFilterModelV01, len(f.Hosts))
	for i, h := range f.Hosts {
		hosts[i] = vpcFirewallRuleHostFilterModelV01(h)
	}

	var protocols []types.String
	diags := f.Protocols.ElementsAs(ctx, &protocols, false)
	if diags.HasError() {
		return nil, diags
	}

	upgradedProtocols := make([]vpcFirewallRuleProtocolFilterModelV01, len(protocols))
	for i, p := range protocols {
		upgradedProtocols[i] = vpcFirewallRuleProtocolFilterModelV01{
			Type: types.StringValue(strings.ToLower(p.ValueString())),
		}
	}

	return &vpcFirewallRulesResourceRuleFiltersModelV01{
		Hosts:     hosts,
		Ports:     f.Ports,
		Protocols: upgradedProtocols,
	}, nil
}

type vpcFirewallRuleHostFilterModelV00 struct {
	Type  types.String `tfsdk:"type"`
	Value types.String `tfsdk:"value"`
}

type vpcFirewallRulesResourceRuleTargetModelV00 struct {
	Type  types.String `tfsdk:"type"`
	Value types.String `tfsdk:"value"`
}

func (r *vpcFirewallRulesResource) schemaV01(ctx context.Context) schema.Schema {
	return schema.Schema{
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

type vpcFirewallRulesResourceModelV01 struct {
	// This ID is specific to Terraform only
	ID       types.String                           `tfsdk:"id"`
	Rules    []vpcFirewallRulesResourceRuleModelV01 `tfsdk:"rules"`
	Timeouts timeouts.Value                         `tfsdk:"timeouts"`
	VPCID    types.String                           `tfsdk:"vpc_id"`

	// Populated from the same fields within [vpcFirewallRulesResourceRuleModel].
	TimeCreated  types.String `tfsdk:"time_created"`
	TimeModified types.String `tfsdk:"time_modified"`
}

func (m vpcFirewallRulesResourceModelV01) upgrade() vpcFirewallRulesResourceModel {
	rules := make([]vpcFirewallRulesResourceRuleModel, len(m.Rules))
	for i, r := range m.Rules {
		rules[i] = r.upgrade()
	}

	return vpcFirewallRulesResourceModel{
		ID: m.ID,
		// Rules:        rules,
		Timeouts:     m.Timeouts,
		VPCID:        m.VPCID,
		TimeCreated:  m.TimeCreated,
		TimeModified: m.TimeModified,
	}
}

type vpcFirewallRulesResourceRuleModelV01 struct {
	Action      types.String                                 `tfsdk:"action"`
	Description types.String                                 `tfsdk:"description"`
	Direction   types.String                                 `tfsdk:"direction"`
	Filters     *vpcFirewallRulesResourceRuleFiltersModelV01 `tfsdk:"filters"`
	Name        types.String                                 `tfsdk:"name"`
	Priority    types.Int64                                  `tfsdk:"priority"`
	Status      types.String                                 `tfsdk:"status"`
	Targets     []vpcFirewallRulesResourceRuleTargetModelV01 `tfsdk:"targets"`

	// Used to retrieve the timestamps from the API and populate the same fields
	// within [vpcFirewallRulesResourceModel]. The `tfsdk:"-"` struct field tag is used
	// to tell Terraform not to populate these values in the schema.
	TimeCreated  types.String `tfsdk:"-"`
	TimeModified types.String `tfsdk:"-"`
}

func (r vpcFirewallRulesResourceRuleModelV01) upgrade() vpcFirewallRulesResourceRuleModel {
	targets := make([]vpcFirewallRulesResourceRuleTargetModel, len(r.Targets))
	for i, t := range r.Targets {
		targets[i] = vpcFirewallRulesResourceRuleTargetModel(t)
	}

	return vpcFirewallRulesResourceRuleModel{
		Action:      r.Action,
		Description: r.Description,
		Direction:   r.Direction,
		Filters:     r.Filters.upgrade(),
		// Name:         r.Name,
		Priority:     r.Priority,
		Status:       r.Status,
		Targets:      targets,
		TimeCreated:  r.TimeCreated,
		TimeModified: r.TimeModified,
	}
}

type vpcFirewallRulesResourceRuleFiltersModelV01 struct {
	Hosts     []vpcFirewallRuleHostFilterModelV01     `tfsdk:"hosts"`
	Ports     types.Set                               `tfsdk:"ports"`
	Protocols []vpcFirewallRuleProtocolFilterModelV01 `tfsdk:"protocols"`
}

func (f *vpcFirewallRulesResourceRuleFiltersModelV01) upgrade() *vpcFirewallRulesResourceRuleFiltersModel {
	if f == nil {
		return nil
	}

	hosts := make([]vpcFirewallRuleHostFilterModel, len(f.Hosts))
	for i, h := range f.Hosts {
		hosts[i] = vpcFirewallRuleHostFilterModel(h)
	}

	protocols := make([]vpcFirewallRuleProtocolFilterModel, len(f.Protocols))
	for i, p := range f.Protocols {
		protocols[i] = vpcFirewallRuleProtocolFilterModel(p)
	}

	return &vpcFirewallRulesResourceRuleFiltersModel{
		Hosts:     hosts,
		Ports:     f.Ports,
		Protocols: protocols,
	}
}

type vpcFirewallRuleHostFilterModelV01 struct {
	Type  types.String `tfsdk:"type"`
	Value types.String `tfsdk:"value"`
}

type vpcFirewallRuleProtocolFilterModelV01 struct {
	Type     types.String `tfsdk:"type"`
	IcmpType types.Int32  `tfsdk:"icmp_type"`
	IcmpCode types.String `tfsdk:"icmp_code"`
}

type vpcFirewallRulesResourceRuleTargetModelV01 struct {
	Type  types.String `tfsdk:"type"`
	Value types.String `tfsdk:"value"`
}

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ function.Function = (*toVPCFirewallRulesMap)(nil)

// vpcFirewallRuleReturnObjectType represents the type structure of the object
// type returned by the function.
var vpcFirewallRuleReturnObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"action":      types.StringType,
		"description": types.StringType,
		"name":        types.StringType,
		"direction":   types.StringType,
		"filters": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"hosts": types.SetType{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"type":  types.StringType,
							"value": types.StringType,
						},
					},
				},
				"protocols": types.SetType{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"type":      types.StringType,
							"icmp_type": types.Int32Type,
							"icmp_code": types.StringType,
						},
					},
				},
				"ports": types.SetType{
					ElemType: types.StringType,
				},
			},
		},
		"priority": types.Int64Type,
		"status":   types.StringType,
		"targets": types.SetType{
			ElemType: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"type":  types.StringType,
					"value": types.StringType,
				},
			},
		},
	},
}

// NewToVPCFirewallRulesMapFunction returns a new function.Function for the
// to_vpc_firewall_rules_map provider function.
func NewToVPCFirewallRulesMapFunction() function.Function {
	return &toVPCFirewallRulesMap{}
}

type toVPCFirewallRulesMap struct{}

func (f *toVPCFirewallRulesMap) Metadata(
	_ context.Context,
	_ function.MetadataRequest,
	resp *function.MetadataResponse,
) {
	resp.Name = "to_vpc_firewall_rules_map"
}

func (f *toVPCFirewallRulesMap) Definition(
	ctx context.Context,
	_ function.DefinitionRequest,
	resp *function.DefinitionResponse,
) {
	resp.Definition = function.Definition{
		Summary: "Converts a VPC firewall rule set to the updated map schema.",
		MarkdownDescription: replaceBackticks(`
The ''provider::oxide::to_vpc_firewall_rules_map'' function converts the
''rules'' attribute of an
[''oxide_vpc_firewall_rules''](https://registry.terraform.io/providers/oxidecomputer/oxide/latest/docs/resources/oxide_vpc_firewall_rules)
resource from the old set schema to the new map value.

It is intended to help reduce the amount of configuration changes required when
updating the provider to a new version, but it should not be used long term.

This function will be removed in a future release. Migrate to the new map
''rules'' schema as soon as possible.
`),
		DeprecationMessage: "This function is only intended to be used to help upgrade the provider version and will be removed in a future release. Migrate to the map rules schema.",
		Parameters: []function.Parameter{
			// It is not possible to have optional attributes in object types,
			// so we "hack" around it by receiving a JSON encoded string of the
			// HCL configuration.
			function.StringParameter{
				Name: "rules_json",
				MarkdownDescription: replaceBackticks(`
JSON encoded string of the ''rules'' set. Use the
[''jsonencode''](https://developer.hashicorp.com/terraform/language/functions/jsonencode)
function to convert the existing ''rules'' set to a JSON string.
`),
			},
		},
		Return: function.MapReturn{
			ElementType: vpcFirewallRuleReturnObjectType,
		},
	}
}

func (f *toVPCFirewallRulesMap) Run(
	ctx context.Context,
	req function.RunRequest,
	resp *function.RunResponse,
) {
	var rulesJSON string
	resp.Error = req.Arguments.Get(ctx, &rulesJSON)
	if resp.Error != nil {
		return
	}

	var rulesSet []vpcFirewallRuleV1Json
	err := json.Unmarshal([]byte(rulesJSON), &rulesSet)
	if err != nil {
		resp.Error = function.NewArgumentFuncError(0, err.Error())
		return
	}

	rulesMap := make(map[string]vpcFirewallRulesResourceRuleModel)
	for _, r := range rulesSet {
		rule, diags := r.toModel()
		if diags.HasError() {
			resp.Error = function.ConcatFuncErrors(resp.Error,
				function.FuncErrorFromDiags(ctx, diags))
			continue
		}
		rulesMap[r.Name] = rule
	}
	if resp.Error != nil {
		return
	}

	resp.Error = resp.Result.Set(ctx, &rulesMap)
}

type vpcFirewallRuleV1Json struct {
	Action      string                        `json:"action"`
	Description string                        `json:"description"`
	Direction   string                        `json:"direction"`
	Filters     *vpcFirewallRuleFiltersV1Json `json:"filters"`
	Name        string                        `json:"name"`
	Priority    int64                         `json:"priority"`
	Status      string                        `json:"status"`
	Targets     []vpcFirewallRuleTargetV1Json `json:"targets"`
}

func (j vpcFirewallRuleV1Json) toModel() (vpcFirewallRulesResourceRuleModel, diag.Diagnostics) {
	rule := vpcFirewallRulesResourceRuleModel{
		Action:      types.StringValue(j.Action),
		Description: types.StringValue(j.Description),
		Direction:   types.StringValue(j.Direction),
		Priority:    types.Int64Value(j.Priority),
		Status:      types.StringValue(j.Status),
	}

	if j.Filters != nil {
		filters, diags := j.Filters.toModel()
		if diags.HasError() {
			return vpcFirewallRulesResourceRuleModel{}, diags
		}
		rule.Filters = &filters
	}

	if j.Targets != nil {
		rule.Targets = make([]vpcFirewallRulesResourceRuleTargetModel, len(j.Targets))
		for i, t := range j.Targets {
			rule.Targets[i] = vpcFirewallRulesResourceRuleTargetModel{
				Type:  types.StringValue(t.Type),
				Value: types.StringValue(t.Value),
			}
		}
	}

	return rule, nil
}

type vpcFirewallRuleFiltersV1Json struct {
	Hosts     []vpcFirewallRuleHostV1Json     `json:"hosts"`
	Protocols []vpcFirewallRuleProtocolV1Json `json:"protocols"`

	// Terraform allows the ports set to be defined as either a string or
	// number, so we must support both.
	Ports []any `json:"ports"`
}

func (j vpcFirewallRuleFiltersV1Json) toModel() (vpcFirewallRulesResourceRuleFiltersModel, diag.Diagnostics) {
	var hosts []vpcFirewallRuleHostFilterModel
	if j.Hosts != nil {
		hosts = make([]vpcFirewallRuleHostFilterModel, len(j.Hosts))
		for i, h := range j.Hosts {
			hosts[i] = vpcFirewallRuleHostFilterModel{
				Type:  types.StringValue(h.Type),
				Value: types.StringValue(h.Value),
			}
		}
	}

	ports := types.SetNull(types.StringType)
	if j.Ports != nil {
		portsValues := make([]attr.Value, len(j.Ports))
		for i, p := range j.Ports {
			portsValues[i] = types.StringValue(fmt.Sprintf("%v", p))
		}

		var diags diag.Diagnostics
		ports, diags = types.SetValue(types.StringType, portsValues)
		if diags.HasError() {
			return vpcFirewallRulesResourceRuleFiltersModel{}, diags
		}
	}

	var protocols []vpcFirewallRuleProtocolFilterModel
	if j.Protocols != nil {
		protocols = make([]vpcFirewallRuleProtocolFilterModel, len(j.Protocols))
		for i, p := range j.Protocols {
			protocols[i] = p.toModel()
		}
	}

	return vpcFirewallRulesResourceRuleFiltersModel{
		Hosts:     hosts,
		Ports:     ports,
		Protocols: protocols,
	}, nil
}

type vpcFirewallRuleHostV1Json struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type vpcFirewallRuleProtocolV1Json struct {
	Type     string  `json:"type"`
	IcmpType *int32  `json:"icmp_type"`
	IcmpCode *string `json:"icmp_code"`
}

func (j vpcFirewallRuleProtocolV1Json) toModel() vpcFirewallRuleProtocolFilterModel {
	model := vpcFirewallRuleProtocolFilterModel{
		Type: types.StringValue(j.Type),
	}

	if j.IcmpType != nil {
		model.IcmpType = types.Int32Value(*j.IcmpType)
	}

	if j.IcmpCode != nil {
		model.IcmpCode = types.StringValue(*j.IcmpCode)
	}

	return model
}

type vpcFirewallRuleTargetV1Json struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

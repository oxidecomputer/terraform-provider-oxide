// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package vpcfirewallrules

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/oxidecomputer/oxide.go/oxide"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/shared"
)

var (
	_ datasource.DataSource              = (*DataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*DataSource)(nil)
)

// NewDataSource returns a new VPC firewall rules data source.
func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

// DataSource is the VPC firewall rules data source implementation.
type DataSource struct {
	client *oxide.Client
}

type DataSourceModel struct {
	Project  types.String          `tfsdk:"project"`
	Rules    []RuleDataSourceModel `tfsdk:"rules"`
	Timeouts timeouts.Value        `tfsdk:"timeouts"`
	VPC      types.String          `tfsdk:"vpc"`
}

type RuleDataSourceModel struct {
	Action       types.String              `tfsdk:"action"`
	Description  types.String              `tfsdk:"description"`
	Direction    types.String              `tfsdk:"direction"`
	Filters      *RuleFiltersResourceModel `tfsdk:"filters"`
	ID           types.String              `tfsdk:"id"`
	Name         types.String              `tfsdk:"name"`
	Priority     types.Int64               `tfsdk:"priority"`
	Status       types.String              `tfsdk:"status"`
	Targets      []RuleTargetResourceModel `tfsdk:"targets"`
	TimeCreated  types.String              `tfsdk:"time_created"`
	TimeModified types.String              `tfsdk:"time_modified"`
	VPCID        types.String              `tfsdk:"vpc_id"`
}

// Metadata returns the data source type name.
func (d *DataSource) Metadata(
	_ context.Context,
	_ datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = "oxide_vpc_firewall_rules"
}

// Configure adds the provider configured client to the data source.
func (d *DataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	_ *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*oxide.Client)
}

// Schema defines the schema for the data source.
func (d *DataSource) Schema(
	ctx context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieve the firewall rules associated with a VPC.",
		Attributes: map[string]schema.Attribute{
			"project": schema.StringAttribute{
				Required:    true,
				Description: "Name or ID of the project that contains the VPC.",
			},
			"vpc": schema.StringAttribute{
				Required:    true,
				Description: "Name or ID of the VPC.",
			},
			"rules": schema.SetNestedAttribute{
				Computed:    true,
				Description: "Firewall rules associated with the VPC.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"action": schema.StringAttribute{
							Computed:    true,
							Description: "Whether traffic matching the rule is allowed or denied.",
						},
						"description": schema.StringAttribute{
							Computed:    true,
							Description: "Description for the VPC firewall rule.",
						},
						"direction": schema.StringAttribute{
							Computed:    true,
							Description: "Whether the rule applies to inbound or outbound traffic.",
						},
						"filters": schema.SingleNestedAttribute{
							Computed:    true,
							Description: "Reductions on the scope of the rule.",
							Attributes: map[string]schema.Attribute{
								"hosts": schema.SetNestedAttribute{
									Computed:    true,
									Description: "Sources or destinations to which the rule applies.",
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"type": schema.StringAttribute{
												Computed:    true,
												Description: "Host filter type.",
											},
											"value": schema.StringAttribute{
												Computed:    true,
												Description: "Host filter value.",
											},
										},
									},
								},
								"ports": schema.SetAttribute{
									Computed:    true,
									Description: "Destination ports or port ranges to which the rule applies.",
									ElementType: types.StringType,
								},
								"protocols": schema.SetNestedAttribute{
									Computed:    true,
									Description: "Networking protocols to which the rule applies.",
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"type": schema.StringAttribute{
												Computed:    true,
												Description: "Protocol type.",
											},
											"icmp_type": schema.Int32Attribute{
												Computed:    true,
												Description: "ICMP type.",
											},
											"icmp_code": schema.StringAttribute{
												Computed:    true,
												Description: "ICMP code or code range.",
											},
										},
									},
								},
							},
						},
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "Unique, immutable, system-controlled identifier of the firewall rule.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Name of the VPC firewall rule.",
						},
						"priority": schema.Int64Attribute{
							Computed:    true,
							Description: "Relative priority of the rule.",
						},
						"status": schema.StringAttribute{
							Computed:    true,
							Description: "Whether the rule is enabled or disabled.",
						},
						"targets": schema.SetNestedAttribute{
							Computed:    true,
							Description: "Sets of instances to which the rule applies.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"type": schema.StringAttribute{
										Computed:    true,
										Description: "Target type.",
									},
									"value": schema.StringAttribute{
										Computed:    true,
										Description: "Target value.",
									},
								},
							},
						},
						"time_created": schema.StringAttribute{
							Computed:    true,
							Description: "Timestamp of when the firewall rule was created.",
						},
						"time_modified": schema.StringAttribute{
							Computed:    true,
							Description: "Timestamp of when the firewall rule was last modified.",
						},
						"vpc_id": schema.StringAttribute{
							Computed:    true,
							Description: "ID of the VPC to which the rule belongs.",
						},
					},
				},
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

// Read retrieves the VPC firewall rules and saves them in Terraform state.
func (d *DataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state DataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
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

	params := oxide.VpcFirewallRulesViewParams{
		Project: oxide.NameOrId(state.Project.ValueString()),
		Vpc:     oxide.NameOrId(state.VPC.ValueString()),
	}
	firewallRules, err := d.client.VpcFirewallRulesView(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read VPC firewall rules",
			"API error: "+err.Error(),
		)
		return
	}

	tflog.Trace(
		ctx,
		fmt.Sprintf("read firewall rules for VPC: %v", state.VPC.ValueString()),
		map[string]any{"success": true},
	)

	state.Rules = make([]RuleDataSourceModel, 0, len(firewallRules.Rules))
	for _, rule := range firewallRules.Rules {
		filters, filterDiags := newFiltersModel(rule.Filters)
		resp.Diagnostics.Append(filterDiags...)
		if resp.Diagnostics.HasError() {
			return
		}

		state.Rules = append(state.Rules, RuleDataSourceModel{
			Action:       types.StringValue(string(rule.Action)),
			Description:  types.StringValue(rule.Description),
			Direction:    types.StringValue(string(rule.Direction)),
			Filters:      filters,
			ID:           types.StringValue(rule.Id),
			Name:         types.StringValue(string(rule.Name)),
			Priority:     types.Int64Value(int64(*rule.Priority)),
			Status:       types.StringValue(string(rule.Status)),
			Targets:      newTargetsModelFromResponse(rule.Targets),
			TimeCreated:  types.StringValue(rule.TimeCreated.String()),
			TimeModified: types.StringValue(rule.TimeModified.String()),
			VPCID:        types.StringValue(rule.VpcId),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

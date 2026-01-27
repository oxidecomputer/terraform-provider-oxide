// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/oxidecomputer/oxide.go/oxide"
)

var (
	_ datasource.DataSource              = (*antiAffinityGroupDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*antiAffinityGroupDataSource)(nil)
)

// NewAntiAffinityGroupDataSource initialises an anti-affinity group datasource
func NewAntiAffinityGroupDataSource() datasource.DataSource {
	return &antiAffinityGroupDataSource{}
}

type antiAffinityGroupDataSource struct {
	client *oxide.Client
}

type antiAffinityGroupDataSourceModel struct {
	Description   types.String   `tfsdk:"description"`
	FailureDomain types.String   `tfsdk:"failure_domain"`
	ID            types.String   `tfsdk:"id"`
	Name          types.String   `tfsdk:"name"`
	Policy        types.String   `tfsdk:"policy"`
	ProjectID     types.String   `tfsdk:"project_id"`
	TimeCreated   types.String   `tfsdk:"time_created"`
	TimeModified  types.String   `tfsdk:"time_modified"`
	Timeouts      timeouts.Value `tfsdk:"timeouts"`
	ProjectName   types.String   `tfsdk:"project_name"`
}

func (d *antiAffinityGroupDataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = "oxide_anti_affinity_group"
}

// Configure adds the provider configured client to the data source.
func (d *antiAffinityGroupDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	_ *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*oxide.Client)
}

func (d *antiAffinityGroupDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
Retrieve information about a specified anti-affinity group.
`,
		Attributes: map[string]schema.Attribute{
			"project_name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the project that contains the anti-affinity group.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the anti-affinity group.",
			},
			"project_id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the project that contains the anti-affinity group.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description for the anti-affinity group.",
			},
			"policy": schema.StringAttribute{
				Computed:    true,
				Description: "Affinity policy used to describe what to do when a request cannot be satisfied.",
			},
			"failure_domain": schema.StringAttribute{
				Computed:    true,
				Description: "Describes the scope of affinity for the purposes of co-location.",
			},
			"timeouts": timeouts.Attributes(ctx),
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier of the anti-affinity group.",
			},
			"time_created": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this anti-affinity group was created.",
			},
			"time_modified": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this anti-affinity group was last modified.",
			},
		},
	}
}

func (d *antiAffinityGroupDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state antiAffinityGroupDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
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

	params := oxide.AntiAffinityGroupViewParams{
		AntiAffinityGroup: oxide.NameOrId(state.Name.ValueString()),
		Project:           oxide.NameOrId(state.ProjectName.ValueString()),
	}
	antiAffinityGroup, err := d.client.AntiAffinityGroupView(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read anti-affinity group:",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(
		ctx,
		fmt.Sprintf("read anti-affinity group with ID: %v", antiAffinityGroup.Id),
		map[string]any{"success": true},
	)

	state.Description = types.StringValue(antiAffinityGroup.Description)
	state.FailureDomain = types.StringValue(string(antiAffinityGroup.FailureDomain))
	state.ID = types.StringValue(antiAffinityGroup.Id)
	state.Name = types.StringValue(string(antiAffinityGroup.Name))
	state.Policy = types.StringValue(string(antiAffinityGroup.Policy))
	state.ProjectID = types.StringValue(antiAffinityGroup.ProjectId)
	state.TimeCreated = types.StringValue(antiAffinityGroup.TimeCreated.String())
	state.TimeModified = types.StringValue(antiAffinityGroup.TimeModified.String())

	// Save state into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package antiaffinitygroupmembers

import (
	"context"
	"fmt"

	"github.com/google/uuid"
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

// NewDataSource initialises an anti-affinity group members datasource
func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

type DataSource struct {
	client *oxide.Client
}

type DataSourceModel struct {
	ID          types.String            `tfsdk:"id"`
	Members     []MemberDataSourceModel `tfsdk:"members"`
	Name        types.String            `tfsdk:"name"`
	ProjectName types.String            `tfsdk:"project_name"`
	Timeouts    timeouts.Value          `tfsdk:"timeouts"`
}

type MemberDataSourceModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	RunState types.String `tfsdk:"run_state"`
	Type     types.String `tfsdk:"type"`
}

func (d *DataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = "oxide_anti_affinity_group_members"
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

func (d *DataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
Retrieve all members of a specified anti-affinity group.
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
			"id": schema.StringAttribute{
				Computed: true,
			},
			"timeouts": timeouts.Attributes(ctx),
			"members": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Members of the anti-affinity group.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "Unique, immutable, system-controlled identifier of the member.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Name of the member.",
						},
						"run_state": schema.StringAttribute{
							Computed:    true,
							Description: "Running state of the member.",
						},
						"type": schema.StringAttribute{
							Computed:    true,
							Description: "Type of the member (e.g. instance).",
						},
					},
				},
			},
		},
	}
}

func (d *DataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state DataSourceModel

	// Read Terraform configuration data into the model
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

	params := oxide.AntiAffinityGroupMemberListParams{
		AntiAffinityGroup: oxide.NameOrId(state.Name.ValueString()),
		Project:           oxide.NameOrId(state.ProjectName.ValueString()),
		SortBy:            oxide.NameOrIdSortModeNameAscending,
	}
	members, err := d.client.AntiAffinityGroupMemberListAllPages(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read anti-affinity group members:",
			"API error: "+err.Error(),
		)
		return
	}

	tflog.Trace(
		ctx,
		fmt.Sprintf("read all members of anti-affinity group: %v", state.Name.ValueString()),
		map[string]any{"success": true},
	)

	// Set a unique ID for the datasource payload
	state.ID = types.StringValue(uuid.New().String())

	state.Members = make([]MemberDataSourceModel, len(members))
	for i, member := range members {
		instance, ok := member.AsInstance()
		if !ok {
			resp.Diagnostics.AddError(
				"Unexpected anti-affinity group member type",
				fmt.Sprintf(
					"Encountered unexpected anti-affinity group member type: %T",
					member.Value,
				),
			)
			return
		}
		state.Members[i] = MemberDataSourceModel{
			ID:       types.StringValue(instance.Value.Id),
			Name:     types.StringValue(string(instance.Value.Name)),
			RunState: types.StringValue(string(instance.Value.RunState)),
			Type:     types.StringValue(string(member.Type())),
		}
	}

	// Save state into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

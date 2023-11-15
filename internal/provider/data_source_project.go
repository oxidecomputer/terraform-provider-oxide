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

var _ datasource.DataSource = (*projectDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*projectDataSource)(nil)

type projectDataSource struct {
	client *oxide.Client
}

type projectDataSourceModel struct {
	Description  types.String   `tfsdk:"description"`
	ID           types.String   `tfsdk:"id"`
	Timeouts     timeouts.Value `tfsdk:"timeouts"`
	Name         types.String   `tfsdk:"name"`
	TimeCreated  types.String   `tfsdk:"time_created"`
	TimeModified types.String   `tfsdk:"time_modified"`
}

// NewProjectDataSource initialises a project datasource
func NewProjectDataSource() datasource.DataSource {
	return &projectDataSource{}
}

func (d *projectDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "oxide_project"
}

// Configure adds the provider configured client to the data source.
func (d *projectDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*oxide.Client)
}

func (d *projectDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the project.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description for the project.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier of the project.",
			},
			"time_created": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this project was created.",
			},
			"time_modified": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this project was last modified.",
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

func (d *projectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state projectDataSourceModel

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

	params := oxide.ProjectViewParams{
		Project: oxide.NameOrId(state.Name.ValueString()),
	}
	project, err := d.client.ProjectView(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read project:",
			err.Error(),
		)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("read project with ID: %v", project.Id), map[string]any{"success": true})

	// Map response body to model
	state.Description = types.StringValue(project.Description)
	state.ID = types.StringValue(project.Id)
	state.Name = types.StringValue(string(project.Name))
	state.TimeCreated = types.StringValue(project.TimeCreated.String())
	state.TimeModified = types.StringValue(project.TimeCreated.String())

	// Save state into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

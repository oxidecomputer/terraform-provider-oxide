// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	//	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	oxideSDK "github.com/oxidecomputer/oxide.go/oxide"
)

var _ datasource.DataSource = &ProjectsDataSource{}
var _ datasource.DataSourceWithConfigure = &ProjectsDataSource{}

type ProjectsDataSource struct {
	client *oxideSDK.Client
}

// projectsDataSourceModel maps the data source schema data.
type projectsDataSourceModel struct {
	Projects []projectModel `tfsdk:"projects"`
}

type projectModel struct {
	Description  types.String `tfsdk:"description"`
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	TimeCreated  types.String `tfsdk:"time_created"`
	TimeModified types.String `tfsdk:"time_modified"`
}

func (d *ProjectsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "oxide_projects"
}

// Configure adds the provider configured client to the data source.
func (d *ProjectsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*oxideSDK.Client)
}

func (d *ProjectsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"projects": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"description": schema.StringAttribute{
							Computed:    true,
							Description: "Description for the project.",
						},
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "Unique, immutable, system-controlled identifier of the project.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Name of the project.",
						},
						"time_created": schema.StringAttribute{
							Computed:    true,
							Description: "Timestamp of when this project was created.",
						},
						"time_modified": schema.StringAttribute{
							Computed:    true,
							Description: "Timestamp of when this project was last modified.",
						},
					},
				},
			},
		},
	}
}

func (d *ProjectsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state projectsDataSourceModel

	// Read Terraform configuration data into the model
	//resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	// TODO: It would be preferable to us the client.Projects.ListAllPages method instead.
	// Unfortunately, currently that method has a bug where it returns twice as many results
	// as there are in reality. For now I'll use the List method with a limit of 1,000,000,000 results.
	// Seems unlikely anyone will have more than one billion projects.
	params := oxideSDK.ProjectListParams{
		Limit:  1000000000,
		SortBy: oxideSDK.NameOrIdSortModeIdAscending,
	}
	projects, err := d.client.ProjectList(params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read projects:",
			err.Error(),
		)
		return
	}

	// Map response body to model
	for _, project := range projects.Items {
		projectState := projectModel{
			Description:  types.StringValue(project.Description),
			ID:           types.StringValue(project.Id),
			Name:         types.StringValue(string(project.Name)),
			TimeCreated:  types.StringValue(project.TimeCreated.String()),
			TimeModified: types.StringValue(project.TimeCreated.String()),
		}

		state.Projects = append(state.Projects, projectState)
	}

	// Save state into Terraform state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

//func projectsDataSource() *schema.Resource {
//	return &schema.Resource{
//		ReadContext: projectsDataSourceRead,
//		Schema:      newProjectsDataSourceSchema(),
//		Timeouts: &schema.ResourceTimeout{
//			Default: schema.DefaultTimeout(5 * time.Minute),
//		},
//	}
//}

//func newProjectsDataSourceSchema() map[string]*schema.Schema {
//	return map[string]*schema.Schema{
//		"projects": {
//			Computed:    true,
//			Type:        schema.TypeList,
//			Description: "A list of all projects",
//			Elem: &schema.Resource{
//				Schema: map[string]*schema.Schema{
//					"description": {
//						Type:        schema.TypeString,
//					Description: "Description for the project.",
//					Computed:    true,
//				},
//				"id": {
//						Type:        schema.TypeString,
//						Description: "Unique, immutable, system-controlled identifier of the project.",
//						Computed:    true,
//					},
//					"name": {
//						Type:        schema.TypeString,
//						Description: "Name of the project.",
//						Computed:    true,
//					},
//					"time_created": {
//						Type:        schema.TypeString,
//						Description: "Timestamp of when this project was created.",
//						Computed:    true,
//					},
//					"time_modified": {
//						Type:        schema.TypeString,
//						Description: "Timestamp of when this project was last modified.",
//						Computed:    true,
//					},
//				},
//			},
//		},
//	}
//}

//func projectsDataSourceRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
//	client := meta.(*oxideSDK.Client)
//
//	// TODO: It would be preferable to us the client.Projectss.ListAllPages method instead.
//	// Unfortunately, currently that method has a bug where it returns twice as many results
//	// as there are in reality. For now I'll use the List method with a limit of 1,000,000 results.
//	// Seems unlikely anyone will have more than one billion projects.
//	params := oxideSDK.ProjectListParams{
//		Limit:  1000000000,
//		SortBy: oxideSDK.NameOrIdSortModeIdAscending,
//	}
//	result, err := client.ProjectList(params)
//	if err != nil {
//		return diag.FromErr(err)
//	}
//
//	d.SetId(strconv.Itoa(schema.HashString(time.Now().String())))
//
//	if err := projectsToState(d, result); err != nil {
//		return diag.FromErr(err)
//	}
//
//	return nil
//}

//func projectsToState(d *schema.ResourceData, projects *oxideSDK.ProjectResultsPage) error {
//	if projects == nil {
//		return nil
//	}
//
//	var result = make([]interface{}, 0, len(projects.Items))
//	for _, project := range projects.Items {
//		var m = make(map[string]interface{})
//
//		m["description"] = project.Description
//		m["id"] = project.Id
//		m["name"] = project.Name
//		m["time_created"] = project.TimeCreated.String()
//		m["time_modified"] = project.TimeModified.String()
//
//		result = append(result, m)
//
//		if len(result) > 0 {
//			if err := d.Set("projects", result); err != nil {
//				return err
//			}
//		}
//	}
//
//	return nil
//}

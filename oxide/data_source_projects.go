// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	oxideSDK "github.com/oxidecomputer/oxide.go"
)

func projectsDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: projectsDataSourceRead,
		Schema:      newProjectsDataSourceSchema(),
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(5 * time.Minute),
		},
	}
}

func newProjectsDataSourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"organization_name": {
			Type:        schema.TypeString,
			Description: "Name of the organization.",
			Required:    true,
		},
		"projects": {
			Computed:    true,
			Type:        schema.TypeList,
			Description: "A list of all projects",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"description": {
						Type:        schema.TypeString,
						Description: "Description for the project.",
						Computed:    true,
					},
					"id": {
						Type:        schema.TypeString,
						Description: "Unique, immutable, system-controlled identifier of the project.",
						Computed:    true,
					},
					"name": {
						Type:        schema.TypeString,
						Description: "Name of the project.",
						Computed:    true,
					},
					"organization_id": {
						Type:        schema.TypeString,
						Description: "Unique, immutable, system-controlled identifier of the organization.",
						Computed:    true,
					},
					"time_created": {
						Type:        schema.TypeString,
						Description: "Timestamp of when this project was created.",
						Computed:    true,
					},
					"time_modified": {
						Type:        schema.TypeString,
						Description: "Timestamp of when this project was last modified.",
						Computed:    true,
					},
				},
			},
		},
	}
}

func projectsDataSourceRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*oxideSDK.Client)
	orgName := d.Get("organization_name").(string)

	// TODO: It would be preferable to us the client.Projectss.ListAllPages method instead.
	// Unfortunately, currently that method has a bug where it returns twice as many results
	// as there are in reality. For now I'll use the List method with a limit of 1,000,000 results.
	// Seems unlikely anyone will have more than one million projects.
	result, err := client.ProjectList(1000000, "", oxideSDK.NameOrIdSortModeIdAscending, orgName)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(schema.HashString(time.Now().String())))

	if err := projectsToState(d, result); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func projectsToState(d *schema.ResourceData, projects *oxideSDK.ProjectResultsPage) error {
	if projects == nil {
		return nil
	}

	var result = make([]interface{}, 0, len(projects.Items))
	for _, project := range projects.Items {
		var m = make(map[string]interface{})

		m["description"] = project.Description
		m["id"] = project.Id
		m["name"] = project.Name
		m["organization_id"] = project.OrganizationId
		m["time_created"] = project.TimeCreated.String()
		m["time_modified"] = project.TimeModified.String()

		result = append(result, m)

		if len(result) > 0 {
			if err := d.Set("projects", result); err != nil {
				return err
			}
		}
	}

	return nil
}

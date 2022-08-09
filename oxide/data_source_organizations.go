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

func organizationsDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: organizationsDataSourceRead,
		Schema:      newOrganizationsDataSourceSchema(),
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(5 * time.Minute),
		},
	}
}

func newOrganizationsDataSourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"organizations": {
			Computed:    true,
			Type:        schema.TypeList,
			Description: "A list of all organizations",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"description": {
						Type:        schema.TypeString,
						Description: "Description of the organization.",
						Computed:    true,
					},
					"id": {
						Type:        schema.TypeString,
						Description: "Unique, immutable, system-controlled identifier of the organization.",
						Computed:    true,
					},
					"name": {
						Type:        schema.TypeString,
						Description: "Name of the organization.",
						Computed:    true,
					},
					"time_created": {
						Type:        schema.TypeString,
						Description: "Timestamp of when this organization was created.",
						Computed:    true,
					},
					"time_modified": {
						Type:        schema.TypeString,
						Description: "Timestamp of when this organization was last modified.",
						Computed:    true,
					},
				},
			},
		},
	}
}

func organizationsDataSourceRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*oxideSDK.Client)

	// TODO: It would be preferable to us the client.Organizations.ListAllPages method instead.
	// Unfortunately, currently that method has a bug where it returns twice as many results
	// as there are in reality. For now I'll use the List method with a limit of 1,000,000 results.
	// Seems unlikely anyone will have more than one million organizations.
	result, err := client.OrganizationList(1000000, "", oxideSDK.NameOrIdSortModeIdAscending)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(schema.HashString(time.Now().String())))

	if err := organizationsToState(d, result); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func organizationsToState(d *schema.ResourceData, orgs *oxideSDK.OrganizationResultsPage) error {
	if orgs == nil {
		return nil
	}

	var result = make([]interface{}, 0, len(orgs.Items))
	for _, org := range orgs.Items {
		var m = make(map[string]interface{})

		m["description"] = org.Description
		m["id"] = org.Id
		m["name"] = org.Name
		m["time_created"] = org.TimeCreated.String()
		m["time_modified"] = org.TimeModified.String()

		result = append(result, m)

		if len(result) > 0 {
			if err := d.Set("organizations", result); err != nil {
				return err
			}
		}
	}

	return nil
}

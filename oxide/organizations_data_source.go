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
		Schema:      newOrganizationDataSourceSchema(),
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(5 * time.Minute),
		},
	}
}

func newOrganizationDataSourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"organizations": &schema.Schema{
			Computed: true,
			Type:     schema.TypeList,
			// Description: "Some description",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"description": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"id": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"name": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"time_created": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"time_modified": {
						Type:     schema.TypeString,
						Computed: true,
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
	result, err := client.Organizations.List(1000000, "", oxideSDK.NameOrIdSortModeIdAscending)
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
		m["id"] = org.ID
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

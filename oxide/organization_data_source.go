// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"context"
	"fmt"
	"time"

	//	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	oxideSDK "github.com/oxidecomputer/oxide.go"
)

func organizationDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: organizationDataSourceRead,
		Schema:      newOrganizationDataSourceSchema(),
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(5 * time.Minute),
		},
	}
}

func newOrganizationDataSourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"organization": &schema.Schema{
			Computed: true,
			Type:     schema.TypeList,
			// Description: "Some description",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"description": {
						Type:     schema.TypeString,
						Computed: true,
					},
					//	"id": {
					//		Type:     schema.TypeString,
					//		Computed: true,
					//	},
					//	"name": {
					//		Type:     schema.TypeString,
					//		Computed: true,
					//	},
					//	"time_created": {
					//		Type:     schema.TypeString,
					//		Computed: true,
					//	},
					//	"time_modified": {
					//		Type:     schema.TypeString,
					//		Computed: true,
					//	},
				},
			},
		},
	}
}

func organizationDataSourceRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*oxideSDK.Client)

	result, err := client.Organizations.ListAllPages(oxideSDK.NameOrIdSortModeIdAscending)
	if err != nil {
		return diag.FromErr(
			fmt.Errorf("Error reading organizations"),
		)
	}

	// TODO: Set a realistic ID, leaving this for now
	if d.Id() == "" {
		// d.SetId(strconv.Itoa(schema.HashString(Version)))
		d.SetId("v0.1-dev")
	}

	if err := organizationsToState(d, result); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func organizationsToState(d *schema.ResourceData, orgs *[]oxideSDK.Organization) error {
	if orgs == nil {
		return nil
	}

	var result = make([]interface{}, 0, len(*orgs))

	for _, org := range *orgs {
		var m = make(map[string]interface{})

		m["description"] = org.Description

		//		if err := d.Set("description", orgs.Description); err != nil {
		//			return err
		//		}

		result = append(result, m)

		if len(result) > 0 {
			if err := d.Set("orgs", result); err != nil {
				return err
			}
		}
	}
	return nil

}

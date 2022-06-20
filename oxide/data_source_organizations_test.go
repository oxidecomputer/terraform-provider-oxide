// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceOrganizations(t *testing.T) {
	datasourceName := "data.oxide_organizations.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: testDataSourceOrganizationsConfig,
				// PreventDiskCleanup: true,
				Check: checkDataSourceOrganizations(datasourceName),
				//resource.TestCheckResourceAttr(datasourceName, "version_regex", "latest"),
				//resource.TestCheckResourceAttr(datasourceName, "lock", "true"),
				//resource.TestCheckResourceAttr(datasourceName, "region", getRegion()),

			},
		},
	})
}

var testDataSourceOrganizationsConfig = `
data "oxide_organizations" "test" {}
`

func checkDataSourceOrganizations(dataName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(dataName, "id"),
		resource.TestCheckResourceAttrSet(dataName, "organizations.0.description"),
		resource.TestCheckResourceAttrSet(dataName, "organizations.0.id"),
		// Ideally we would like to test that the organization has the name we want set with:
		// resource.TestCheckResourceAttr(dataName, "organizations.0.name", "corp"),
		// Unfortunately, for now we can't guarantee that the organizations will be in the
		// same order for everyone who runs the tests. This means we'll only check that it's set.
		resource.TestCheckResourceAttrSet(dataName, "organizations.0.name"),
		resource.TestCheckResourceAttrSet(dataName, "organizations.0.time_created"),
		resource.TestCheckResourceAttrSet(dataName, "organizations.0.time_modified"),
	}...)
}

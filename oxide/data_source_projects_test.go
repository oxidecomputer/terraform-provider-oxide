// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceProjects(t *testing.T) {
	datasourceName := "data.oxide_projects.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testDataSourceProjectsConfig,
				Check:  checkDataSourceProjects(datasourceName),
			},
		},
	})
}

var testDataSourceProjectsConfig = `
data "oxide_projects" "test" {}
`

func checkDataSourceProjects(dataName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(dataName, "id"),
		resource.TestCheckResourceAttrSet(dataName, "projects.0.description"),
		resource.TestCheckResourceAttrSet(dataName, "projects.0.id"),
		// Ideally we would like to test that a project has the name we want set with:
		// resource.TestCheckResourceAttr(dataName, "projects.0.name", "test"),
		// Unfortunately, for now we can't guarantee that the projects will be in the
		// same order for everyone who runs the tests. This means we'll only check that it's set.
		resource.TestCheckResourceAttrSet(dataName, "projects.0.name"),
		resource.TestCheckResourceAttrSet(dataName, "projects.0.time_created"),
		resource.TestCheckResourceAttrSet(dataName, "projects.0.time_modified"),
	}...)
}

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// NB: The project must be populated with at least one image for this test to pass
func TestAccDataSourceImages(t *testing.T) {
	datasourceName := "data.oxide_images.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testDataSourceImagesConfig,
				Check:  checkDataSourceImages(datasourceName),
			},
		},
	})
}

var testDataSourceImagesConfig = `
data "oxide_projects" "project_list" {}

data "oxide_images" "test" {
  project_id = element(tolist(data.oxide_projects.project_list.projects[*].id), 0)
}
`

func checkDataSourceImages(dataName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(dataName, "id"),
		resource.TestCheckResourceAttrSet(dataName, "images.0.block_size"),
		resource.TestCheckResourceAttrSet(dataName, "images.0.description"),
		resource.TestCheckResourceAttrSet(dataName, "images.0.os"),
		resource.TestCheckResourceAttrSet(dataName, "images.0.id"),
		// Ideally we would like to test that a global image has the name we want set with:
		// resource.TestCheckResourceAttr(dataName, "images.0.name", "alpine"),
		// Unfortunately, for now we can't guarantee that the global images will be in the
		// same order for everyone who runs the tests. This means we'll only check that it's set.
		resource.TestCheckResourceAttrSet(dataName, "images.0.name"),
		resource.TestCheckResourceAttrSet(dataName, "images.0.size"),
		resource.TestCheckResourceAttrSet(dataName, "images.0.time_created"),
		resource.TestCheckResourceAttrSet(dataName, "images.0.time_modified"),
		resource.TestCheckResourceAttrSet(dataName, "images.0.version"),
	}...)
}

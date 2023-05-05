// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// NB: The project must be populated with at least one image for this test to pass
func TestAccDataSourceImage_full(t *testing.T) {
	datasourceName := "data.oxide_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testDataSourceImageConfig,
				Check:  checkDataSourceImage(datasourceName),
			},
		},
	})
}

type dataSourceImageConfig struct {
	BlockName  string
	BlockName2 string
	BlockName3 string
}

var testDataSourceImageConfig = `
data "oxide_projects" "project_list" {}

data "oxide_images" "image_list" {
  project_id = element(tolist(data.oxide_projects.project_list.projects[*].id), 0)
}

data "oxide_image" "test" {
  project_name = element(tolist(data.oxide_projects.project_list.projects[*].name), 0)
  name = element(tolist(data.oxide_images.image_list.images[*].name), 0)
  timeouts = {
    read = "1m"
  }
}
`

func checkDataSourceImage(dataName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(dataName, "id"),
		resource.TestCheckResourceAttr(dataName, "timeouts.read", "1m"),
		resource.TestCheckResourceAttrSet(dataName, "block_size"),
		resource.TestCheckResourceAttrSet(dataName, "description"),
		resource.TestCheckResourceAttrSet(dataName, "os"),
		resource.TestCheckResourceAttrSet(dataName, "project_id"),
		resource.TestCheckResourceAttrSet(dataName, "project_name"),
		resource.TestCheckResourceAttrSet(dataName, "name"),
		resource.TestCheckResourceAttrSet(dataName, "size"),
		resource.TestCheckResourceAttrSet(dataName, "time_created"),
		resource.TestCheckResourceAttrSet(dataName, "time_modified"),
		resource.TestCheckResourceAttrSet(dataName, "version"),
	}...)
}

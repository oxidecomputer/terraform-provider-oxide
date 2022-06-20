// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceGlobalImages(t *testing.T) {
	datasourceName := "data.oxide_global_images.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: testDataSourceGlobalImagesConfig,
				Check:  checkDataSourceGlobalImages(datasourceName),
			},
		},
	})
}

var testDataSourceGlobalImagesConfig = `
data "oxide_global_images" "test" {}
`

func checkDataSourceGlobalImages(dataName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(dataName, "id"),
		resource.TestCheckResourceAttrSet(dataName, "global_images.0.block_size"),
		resource.TestCheckResourceAttrSet(dataName, "global_images.0.description"),
		resource.TestCheckResourceAttrSet(dataName, "global_images.0.distribution"),
		resource.TestCheckResourceAttrSet(dataName, "global_images.0.id"),
		// Ideally we would like to test that a global image has the name we want set with:
		// resource.TestCheckResourceAttr(dataName, "global_images.0.name", "alpine"),
		// Unfortunately, for now we can't guarantee that the global images will be in the
		// same order for everyone who runs the tests. This means we'll only check that it's set.
		resource.TestCheckResourceAttrSet(dataName, "global_images.0.name"),
		resource.TestCheckResourceAttrSet(dataName, "global_images.0.size"),
		resource.TestCheckResourceAttrSet(dataName, "global_images.0.time_created"),
		resource.TestCheckResourceAttrSet(dataName, "global_images.0.time_modified"),
		resource.TestCheckResourceAttrSet(dataName, "global_images.0.version"),
		// TODO: When the global images resource is developed we should also check that
		// digest and url are set.
	}...)
}

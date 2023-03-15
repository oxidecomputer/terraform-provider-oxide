// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// TODO: Restore test when it is possible to delete a global image

func TestAccResourceImage(t *testing.T) {
	resourceName := "oxide_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactory,
		CheckDestroy:      testAccImageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testResourceImageConfig,
				Check:  checkResourceImage(resourceName),
			},
		},
	})
}

var testResourceImageConfig = `
data "oxide_organizations" "org_list" {}

data "oxide_projects" "project_list" {
  organization_name = data.oxide_organizations.org_list.organizations.0.name
}

 resource "oxide_image" "test" {
   project_id   = data.oxide_projects.project_list.projects.0.id
   description  = "a test image"
   name         = "terraform-acc-myglobalimage"
   image_source = { you_can_boot_anything_as_long_as_its_alpine = "noop" }
   block_size   = 512
   os           = "alpine"
   version      = "propolis_blob"
 }
 `

func checkResourceImage(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test image"),
		resource.TestCheckResourceAttr(resourceName, "name", "terraform-acc-myglobalimage"),
		resource.TestCheckResourceAttr(resourceName, "block_size", "512"),
		resource.TestCheckResourceAttrSet(resourceName, "image_source.you_can_boot_anything_as_long_as_its_alpine"),
		resource.TestCheckResourceAttrSet(resourceName, "size"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		// TODO: Eventually we'll want to test creating a image from URL and snapshot
	}...)
}

func testAccImageDestroy(s *terraform.State) error {
	client, err := newTestClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "oxide_image" {
			continue
		}

		res, err := client.ImageViewV1("corp", "test", "terraform-acc-myglobalimage")

		if err != nil && is404(err) {
			continue
		}

		return fmt.Errorf("image (%v) still exists", &res.Name)
	}

	return nil
}

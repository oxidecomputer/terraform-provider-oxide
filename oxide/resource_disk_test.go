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

func TestAccResourceDisk(t *testing.T) {
	resourceName := "oxide_disk.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactory,
		CheckDestroy:      testAccDiskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testResourceDiskConfig,
				Check:  checkResourceDisk(resourceName),
			},
		},
	})
}

var testResourceDiskConfig = `
data "oxide_projects" "project_list" {}

resource "oxide_disk" "test" {
  project_id        = data.oxide_projects.project_list.projects.0.id
  description       = "a test disk"
  name              = "terraform-acc-mydisk"
  size              = 1073741824
  disk_source       = { blank = 512 }
}
`

func checkResourceDisk(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test disk"),
		resource.TestCheckResourceAttr(resourceName, "name", "terraform-acc-mydisk"),
		resource.TestCheckResourceAttr(resourceName, "size", "1073741824"),
		resource.TestCheckResourceAttr(resourceName, "device_path", "/mnt/terraform-acc-mydisk"),
		resource.TestCheckResourceAttr(resourceName, "block_size", "512"),
		resource.TestCheckResourceAttr(resourceName, "disk_source.blank", "512"),
		resource.TestCheckResourceAttr(resourceName, "state.0.state", "detached"),
		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		// TODO: Eventually we'll want to test creating a disk from images and snapshot
	}...)
}

func testAccDiskDestroy(s *terraform.State) error {
	client, err := newTestClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "oxide_disk" {
			continue
		}

		res, err := client.DiskView("terraform-acc-mydisk", "test")

		if err != nil && is404(err) {
			continue
		}

		return fmt.Errorf("disk (%v) still exists", &res.Name)
	}

	return nil
}

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

func TestAccResourceInstance(t *testing.T) {
	resourceName := "oxide_instance.test"
	secondResourceName := "oxide_instance.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactory,
		CheckDestroy:      testAccInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testResourceInstanceConfig,
				Check:  checkResourceInstance(resourceName),
			},
			{
				Config: testResourceInstanceDiskConfig,
				Check:  checkResourceInstanceDisk(secondResourceName),
			},
		},
	})
}

var testResourceInstanceConfig = `
resource "oxide_instance" "test" {
  organization_name = "corp"
  project_name      = "test"
  description       = "a test instance"
  name              = "terraform-acc-myinstance"
  host_name         = "terraform-acc-myhost"
  memory            = 512
  ncpus             = 1
}
`

func checkResourceInstance(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "organization_name", "corp"),
		resource.TestCheckResourceAttr(resourceName, "project_name", "test"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test instance"),
		resource.TestCheckResourceAttr(resourceName, "name", "terraform-acc-myinstance"),
		resource.TestCheckResourceAttr(resourceName, "host_name", "terraform-acc-myhost"),
		resource.TestCheckResourceAttr(resourceName, "memory", "512"),
		resource.TestCheckResourceAttr(resourceName, "ncpus", "1"),
		resource.TestCheckResourceAttrSet(resourceName, "run_state"),
		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		resource.TestCheckResourceAttrSet(resourceName, "time_run_state_updated"),
	}...)
}

var testResourceInstanceDiskConfig = `
resource "oxide_disk" "test-instance" {
  organization_name = "corp"
  project_name      = "test"
  description       = "a test disk"
  name              = "terraform-acc-mydisk1"
  size              = 1024
  disk_source       = { blank = 512 }
}

resource "oxide_disk" "test-instance2" {
  organization_name = "corp"
  project_name      = "test"
  description       = "a test disk"
  name              = "terraform-acc-mydisk2"
  size              = 1024
  disk_source       = { blank = 512 }
}

resource "oxide_instance" "test2" {
  organization_name = "corp"
  project_name      = "test"
  description       = "a test instance"
  name              = "terraform-acc-myinstance2"
  host_name         = "terraform-acc-myhost"
  memory            = 512
  ncpus             = 1
  disks { name = "terraform-acc-mydisk1" }
  disks { name = "terraform-acc-mydisk2" }
}
`

func checkResourceInstanceDisk(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "organization_name", "corp"),
		resource.TestCheckResourceAttr(resourceName, "project_name", "test"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test instance"),
		resource.TestCheckResourceAttr(resourceName, "name", "terraform-acc-myinstance2"),
		resource.TestCheckResourceAttr(resourceName, "host_name", "terraform-acc-myhost"),
		resource.TestCheckResourceAttr(resourceName, "memory", "512"),
		resource.TestCheckResourceAttr(resourceName, "ncpus", "1"),
		resource.TestCheckResourceAttr(resourceName, "disks.0.name", "terraform-acc-mydisk1"),
		resource.TestCheckResourceAttr(resourceName, "disks.1.name", "terraform-acc-mydisk2"),
		resource.TestCheckResourceAttrSet(resourceName, "run_state"),
		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		resource.TestCheckResourceAttrSet(resourceName, "time_run_state_updated"),
	}...)
}

func testAccInstanceDestroy(s *terraform.State) error {
	client, err := newTestClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "oxide_instance" {
			continue
		}

		res, err := client.Instances.Get("terraform-acc-myinstance", "corp", "test")
		if err != nil && is404(err) {
			continue
		}

		return fmt.Errorf("instance (%v) still exists", &res.Name)
	}

	return nil
}

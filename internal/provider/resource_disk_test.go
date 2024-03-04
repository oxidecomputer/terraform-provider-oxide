// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/oxidecomputer/oxide.go/oxide"
)

type resourceDiskConfig struct {
	BlockName        string
	DiskName         string
	SupportBlockName string
}

var resourceDiskConfigTpl = `
data "oxide_project" "{{.SupportBlockName}}" {
	name = "tf-acc-test"
}

resource "oxide_disk" "{{.BlockName}}" {
  project_id  = data.oxide_project.{{.SupportBlockName}}.id
  description = "a test disk"
  name        = "{{.DiskName}}"
  size        = 1073741824
  block_size  = 512
  timeouts = {
    read   = "1m"
	create = "3m"
	delete = "2m"
  }
}
`

func TestAccCloudResourceDisk_full(t *testing.T) {
	diskName := newResourceName()
	blockName := newBlockName("disk")
	resourceName := fmt.Sprintf("oxide_disk.%s", blockName)
	config, err := parsedAccConfig(
		resourceDiskConfig{
			BlockName:        blockName,
			DiskName:         diskName,
			SupportBlockName: newBlockName("support"),
		},
		resourceDiskConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccDiskDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  checkResourceDisk(resourceName, diskName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func checkResourceDisk(resourceName, diskName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test disk"),
		resource.TestCheckResourceAttr(resourceName, "name", diskName),
		resource.TestCheckResourceAttr(resourceName, "size", "1073741824"),
		resource.TestCheckResourceAttr(resourceName, "device_path", "/mnt/"+diskName),
		resource.TestCheckResourceAttr(resourceName, "block_size", "512"),
		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.read", "1m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.delete", "2m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.create", "3m"),
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

		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()

		params := oxide.DiskViewParams{
			Disk: oxide.NameOrId(rs.Primary.Attributes["id"]),
		}
		res, err := client.DiskView(ctx, params)

		if err != nil && is404(err) {
			continue
		}

		return fmt.Errorf("disk (%v) still exists", &res.Name)
	}

	return nil
}

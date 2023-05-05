// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	oxideSDK "github.com/oxidecomputer/oxide.go/oxide"
)

type resourceDiskConfig struct {
	BlockName        string
	DiskName         string
	SupportBlockName string
}

var resourceDiskConfigTpl = `
data "oxide_projects" "{{.SupportBlockName}}" {}

resource "oxide_disk" "{{.BlockName}}" {
  project_id        = element(tolist(data.oxide_projects.{{.SupportBlockName}}.projects[*].id), 0)
  description       = "a test disk"
  name              = "{{.DiskName}}"
  size              = 1073741824
  disk_source       = { blank = 512 }
  timeouts = {
    read   = "1m"
	create = "3m"
	delete = "2m"
  }
}
`

func TestAccResourceDisk_full(t *testing.T) {
	diskName := fmt.Sprintf("acc-terraform-%s", uuid.New())
	blockName := fmt.Sprintf("acc-resource-disk-%s", uuid.New())
	resourceName := fmt.Sprintf("oxide_disk.%s", blockName)
	config, err := parsedAccConfig(
		resourceDiskConfig{
			BlockName:        blockName,
			DiskName:         diskName,
			SupportBlockName: fmt.Sprintf("acc-support-%s", uuid.New()),
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
				// TODO: Remove once https://github.com/oxidecomputer/terraform-provider-oxide/issues/101
				// has been worked on.
				ImportStateVerifyIgnore: []string{"disk_source"},
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
		resource.TestCheckResourceAttr(resourceName, "disk_source.blank", "512"),
		resource.TestCheckResourceAttrSet(resourceName, "state.state"),
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

		params := oxideSDK.DiskViewParams{
			Disk: oxideSDK.NameOrId(rs.Primary.Attributes["id"]),
		}
		res, err := client.DiskView(params)

		if err != nil && is404(err) {
			continue
		}

		return fmt.Errorf("disk (%v) still exists", &res.Name)
	}

	return nil
}

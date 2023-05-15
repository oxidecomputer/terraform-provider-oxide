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

type resourceSnapshotConfig struct {
	BlockName        string
	SnapshotName     string
	DiskBlockName    string
	DiskName         string
	SupportBlockName string
}

var resourceSnapshotConfigTpl = `
data "oxide_projects" "{{.SupportBlockName}}" {}

resource "oxide_disk" "{{.DiskBlockName}}" {
  project_id  = element(tolist(data.oxide_projects.{{.SupportBlockName}}.projects[*].id), 0)
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

resource "oxide_snapshot" "{{.BlockName}}" {
  project_id  = element(tolist(data.oxide_projects.{{.SupportBlockName}}.projects[*].id), 0)
  description = "a test snapshot"
  name        = "{{.SnapshotName}}"
  disk_id     = oxide_disk.{{.DiskBlockName}}.id
  timeouts = {
    read   = "1m"
    create = "3m"
    delete = "2m"
  }
}
`

func TestAccResourceSnapshot_full(t *testing.T) {
	diskName := fmt.Sprintf("acc-terraform-%s", uuid.New())
	snapshotName := fmt.Sprintf("acc-terraform-%s", uuid.New())
	blockName := fmt.Sprintf("acc-resource-snapshot-%s", uuid.New())
	resourceName := fmt.Sprintf("oxide_snapshot.%s", blockName)
	config, err := parsedAccConfig(
		resourceSnapshotConfig{
			BlockName:        blockName,
			SnapshotName:     snapshotName,
			DiskName:         diskName,
			SupportBlockName: fmt.Sprintf("acc-support-%s", uuid.New()),
			DiskBlockName:    fmt.Sprintf("acc-resource-disk-%s", uuid.New()),
		},
		resourceSnapshotConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  checkResourceSnapshot(resourceName, snapshotName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func checkResourceSnapshot(resourceName, snapshotName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test snapshot"),
		resource.TestCheckResourceAttr(resourceName, "name", snapshotName),
		resource.TestCheckResourceAttrSet(resourceName, "size"),
		resource.TestCheckResourceAttrSet(resourceName, "disk_id"),
		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.read", "1m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.delete", "2m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.create", "3m"),
		// TODO: Eventually we'll want to test creating a disk from images and snapshot
	}...)
}

func testAccSnapshotDestroy(s *terraform.State) error {
	client, err := newTestClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "oxide_snapshot" {
			continue
		}

		params := oxideSDK.SnapshotViewParams{
			Snapshot: oxideSDK.NameOrId(rs.Primary.Attributes["id"]),
		}
		res, err := client.SnapshotView(params)

		if err != nil && is404(err) {
			continue
		}

		return fmt.Errorf("snapshot (%v) still exists", &res.Name)
	}

	return nil
}

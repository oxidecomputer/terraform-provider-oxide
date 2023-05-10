// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type resourceInstanceDiskAttachmentConfig struct {
	BlockName        string
	InstanceName     string
	SupportBlockName string
	DiskBlockName    string
	DiskName         string
}

var resourceInstanceDiskAttachmentConfigTpl = `
 data "oxide_projects" "{{.SupportBlockName}}" {}

resource "oxide_disk" "{{.DiskBlockName}}" {
  project_id  = element(tolist(data.oxide_projects.{{.SupportBlockName}}.projects[*].id), 0)
  description = "a test disk"
  name        = "{{.DiskName}}"
  size        = 1073741824
  disk_source = { blank = 512 }
}

resource "oxide_instance" "{{.InstanceBlockName}}" {
  project_id      = element(tolist(data.oxide_projects.{{.SupportBlockName}}.projects[*].id), 0)
  description     = "a test instance"
  name            = "{{.InstanceName}}"
  host_name       = "terraform-acc-myhost"
  memory          = 1073741824
  ncpus           = 1
  start_on_create = false
}

resource "oxide_instance_disk_attachment" "{{.BlockName}}" {
  disk_id = oxide_disk.{{.DiskBlockName}}.id
  instance_id = oxide_instance.{{.InstanceBlockName}}.id
  timeouts = {
    read   = "1m"
    create = "3m"
    delete = "2m"
  }
}
`

func TestAccResourceInstanceDiskAttachment_full(t *testing.T) {
	blockName := fmt.Sprintf("acc-resource-instance-disk-attachment-%s", uuid.New())
	resourceName := fmt.Sprintf("oxide_instance_disk_attachment.%s", blockName)
	diskName := fmt.Sprintf("acc-terraform-%s", uuid.New())
	config, err := parsedAccConfig(
		resourceInstanceDiskAttachmentConfig{
			BlockName:        blockName,
			InstanceName:     fmt.Sprintf("acc-terraform-%s", uuid.New()),
			DiskName:         diskName,
			DiskBlockName:    fmt.Sprintf("acc-resource-instance-disk-attachment-%s", uuid.New()),
			SupportBlockName: fmt.Sprintf("acc-support-%s", uuid.New()),
		},
		resourceInstanceDiskAttachmentConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  checkResourceInstanceDiskAttachment(resourceName, diskName),
			},
			// TODO: Implement imports
			//	{
			//		ResourceName:      resourceName,
			//		ImportState:       true,
			//		ImportStateVerify: true,
			//	},
		},
	})
}

func checkResourceInstanceDiskAttachment(resourceName, diskName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttrSet(resourceName, "instance_id"),
		resource.TestCheckResourceAttrSet(resourceName, "disk_id"),
		resource.TestCheckResourceAttr(resourceName, "disk_name", diskName),
		resource.TestCheckResourceAttr(resourceName, "timeouts.read", "1m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.delete", "2m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.create", "3m"),
	}...)
}

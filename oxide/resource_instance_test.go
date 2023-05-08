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

type resourceInstanceConfig struct {
	BlockName        string
	InstanceName     string
	SupportBlockName string
	DiskBlockName    string
	DiskName         string
	DiskBlockName2   string
	DiskName2        string
}

var resourceInstanceConfigTpl = `
data "oxide_projects" "{{.SupportBlockName}}" {}

resource "oxide_instance" "{{.BlockName}}" {
  project_id      = element(tolist(data.oxide_projects.{{.SupportBlockName}}.projects[*].id), 0)
  description     = "a test instance"
  name            = "{{.InstanceName}}"
  host_name       = "terraform-acc-myhost"
  memory          = 1073741824
  ncpus           = 1
  start_on_create = false
  timeouts = {
    read   = "1m"
	create = "3m"
	delete = "2m"
  }
}
`

var resourceInstanceFullConfigTpl = `
data "oxide_projects" "{{.SupportBlockName}}" {}

resource "oxide_disk" "{{.DiskBlockName}}" {
  project_id  = element(tolist(data.oxide_projects.{{.SupportBlockName}}.projects[*].id), 0)
  description = "a test disk"
  name        = "{{.DiskName}}"
  size        = 1073741824
  disk_source = { blank = 512 }
}

resource "oxide_disk" "{{.DiskBlockName2}}" {
  project_id  = element(tolist(data.oxide_projects.{{.SupportBlockName}}.projects[*].id), 0)
  description = "a test disk"
  name        = "{{.DiskName2}}"
  size        = 1073741824
  disk_source = { blank = 512 }
}

resource "oxide_instance" "{{.BlockName}}" {
  project_id      = element(tolist(data.oxide_projects.{{.SupportBlockName}}.projects[*].id), 0)
  description     = "a test instance"
  name            = "{{.InstanceName}}"
  host_name       = "terraform-acc-myhost"
  memory          = 1073741824
  ncpus           = 1
  external_ips    = ["default"]
  attach_to_disks = ["{{.DiskName}}", "{{.DiskName2}}"]
}
`

func TestAccResourceInstance_full(t *testing.T) {
	instanceName := fmt.Sprintf("acc-terraform-%s", uuid.New())
	blockName := fmt.Sprintf("acc-resource-instance-%s", uuid.New())
	supportBlockName := fmt.Sprintf("acc-support-%s", uuid.New())
	resourceName := fmt.Sprintf("oxide_instance.%s", blockName)
	config, err := parsedAccConfig(
		resourceInstanceConfig{
			BlockName:        blockName,
			InstanceName:     instanceName,
			SupportBlockName: supportBlockName,
		},
		resourceInstanceConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	blockName2 := fmt.Sprintf("acc-resource-instance-%s", uuid.New())
	diskName := fmt.Sprintf("acc-terraform-%s", uuid.New())
	diskName2 := diskName + "-2"
	instanceName2 := instanceName + "-2"
	resourceName2 := fmt.Sprintf("oxide_instance.%s", blockName2)
	config2, err := parsedAccConfig(
		resourceInstanceConfig{
			BlockName:        blockName2,
			InstanceName:     instanceName2,
			DiskName:         diskName,
			DiskBlockName:    fmt.Sprintf("acc-resource-instance-%s", uuid.New()),
			DiskName2:        diskName2,
			DiskBlockName2:   fmt.Sprintf("acc-resource-instance-%s", uuid.New()),
			SupportBlockName: supportBlockName,
		},
		resourceInstanceFullConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  checkResourceInstance(resourceName, instanceName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// This option is only relevant for create, this means that it will
				// never be imported
				ImportStateVerifyIgnore: []string{"start_on_create"},
			},
			{
				Config: config2,
				Check:  checkResourceInstanceFull(resourceName2, instanceName2, diskName, diskName2),
			},
			{
				ResourceName:            resourceName2,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"start_on_create"},
			},
		},
	})
}

func checkResourceInstance(resourceName, instanceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test instance"),
		resource.TestCheckResourceAttr(resourceName, "name", instanceName),
		resource.TestCheckResourceAttr(resourceName, "host_name", "terraform-acc-myhost"),
		resource.TestCheckResourceAttr(resourceName, "memory", "1073741824"),
		resource.TestCheckResourceAttr(resourceName, "ncpus", "1"),
		resource.TestCheckResourceAttr(resourceName, "start_on_create", "false"),
		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.read", "1m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.delete", "2m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.create", "3m"),
	}...)
}

func checkResourceInstanceFull(resourceName, instanceName, diskName, diskName2 string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test instance"),
		resource.TestCheckResourceAttr(resourceName, "name", instanceName),
		resource.TestCheckResourceAttr(resourceName, "host_name", "terraform-acc-myhost"),
		resource.TestCheckResourceAttr(resourceName, "memory", "1073741824"),
		resource.TestCheckResourceAttr(resourceName, "ncpus", "1"),
		resource.TestCheckResourceAttr(resourceName, "attach_to_disks.0", diskName),
		resource.TestCheckResourceAttr(resourceName, "attach_to_disks.1", diskName2),
		resource.TestCheckResourceAttr(resourceName, "external_ips.0", "default"),
		resource.TestCheckResourceAttr(resourceName, "start_on_create", "true"),
		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
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

		// TODO: check for block name

		params := oxideSDK.InstanceViewParams{
			Instance: oxideSDK.NameOrId(rs.Primary.Attributes["id"]),
		}
		res, err := client.InstanceView(params)
		if err != nil && is404(err) {
			continue
		}

		return fmt.Errorf("instance (%v) still exists", &res.Name)
	}

	return nil
}

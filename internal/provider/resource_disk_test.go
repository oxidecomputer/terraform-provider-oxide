// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/oxidecomputer/oxide.go/oxide"
)

type resourceDiskConfig struct {
	DiskName string
}

var resourceDiskConfigTpl = `
data "oxide_project" "test" {
	name = "tf-acc-test"
}

resource "oxide_disk" "test" {
  project_id  = data.oxide_project.test.id
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
	config, err := parsedAccConfig(
		resourceDiskConfig{
			DiskName: diskName,
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
				Check:  checkResourceDisk("oxide_disk.test", diskName),
			},
			{
				ResourceName:      "oxide_disk.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

var resourceDiskLocalConfigTpl = `
data "oxide_project" "test" {
	name = "tf-acc-test"
}

resource "oxide_disk" "test" {
  project_id  = data.oxide_project.test.id
  description = "a test disk"
  name        = "{{.DiskName}}"
  size        = 1073741824
  disk_type   = "local"
}
`

func TestAccCloudResourceDisk_local(t *testing.T) {
	diskName := newResourceName()

	config, err := parsedAccConfig(
		resourceDiskConfig{
			DiskName: diskName,
		},
		resourceDiskLocalConfigTpl,
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
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("oxide_disk.test", "id"),
					resource.TestCheckResourceAttr("oxide_disk.test", "disk_type", "local"),
				),
			},
		},
	})
}

var resourceDiskLocalInvalidConfigTpl = `
data "oxide_project" "test" {
	name = "tf-acc-test"
}

resource "oxide_disk" "test" {
  project_id      = data.oxide_project.test.id
  description     = "a test disk"
  name            = "{{.DiskName}}"
  size            = 1073741824
  source_image_id = "00000000-0000-0000-0000-000000000000"
  disk_type       = "local"
}
`

func TestAccCloudResourceDisk_localSourceValidation(t *testing.T) {
	config, err := parsedAccConfig(
		resourceDiskConfig{
			DiskName: newResourceName(),
		},
		resourceDiskLocalInvalidConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`cannot be set when disk_type is "local"`),
			},
		},
	})
}

type resourceDiskReadOnlyConfig struct {
	DiskName string
	ReadOnly bool
}

var resourceDiskReadOnlyConfigTpl = `
data "oxide_project" "test" {
	name = "tf-acc-test"
}

data "oxide_images" "test" {
  project_id = data.oxide_project.test.id
}

resource "oxide_disk" "test" {
  project_id      = data.oxide_project.test.id
  description     = "a test read-only disk"
  name            = "{{.DiskName}}"
  size            = 1073741824
  source_image_id = element(tolist(data.oxide_images.test.images[*].id), 0)
  read_only       = {{.ReadOnly}}
}
`

func TestAccCloudResourceDisk_readOnly(t *testing.T) {
	diskName := newResourceName()

	configReadOnly, err := parsedAccConfig(
		resourceDiskReadOnlyConfig{
			DiskName: diskName,
			ReadOnly: true,
		},
		resourceDiskReadOnlyConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	configReadWrite, err := parsedAccConfig(
		resourceDiskReadOnlyConfig{
			DiskName: diskName,
			ReadOnly: false,
		},
		resourceDiskReadOnlyConfigTpl,
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
				Config: configReadOnly,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("oxide_disk.test", "id"),
					resource.TestCheckResourceAttr("oxide_disk.test", "read_only", "true"),
				),
			},
			{
				ResourceName:      "oxide_disk.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: configReadWrite,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							"oxide_disk.test",
							plancheck.ResourceActionReplace,
						),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("oxide_disk.test", "id"),
					resource.TestCheckResourceAttr("oxide_disk.test", "read_only", "false"),
				),
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
		resource.TestCheckResourceAttr(resourceName, "disk_type", "distributed"),
		resource.TestCheckResourceAttr(resourceName, "read_only", "false"),
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

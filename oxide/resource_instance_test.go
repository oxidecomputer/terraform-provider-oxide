// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	oxideSDK "github.com/oxidecomputer/oxide.go/oxide"
)

func TestAccResourceInstance_full(t *testing.T) {
	resourceName := "oxide_instance.test"
	secondResourceName := "oxide_instance.test2"
	// thirdResourceName := "oxide_instance.test3"
	fourthResourceName := "oxide_instance.test4"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testResourceInstanceConfig,
				Check:  checkResourceInstance(resourceName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testResourceInstanceDiskConfig,
				Check:  checkResourceInstanceDisk(secondResourceName),
			},
			{
				ResourceName:      secondResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// TODO: Restore when Network interfaces work on simulated omicron again
			//	{
			//		Config: testResourceInstanceNetworkInterfaceConfig,
			//		Check:  checkResourceInstanceNetworkInterface(thirdResourceName),
			//	},
			// {
			// 	ResourceName:      thirdResourceName,
			// 	ImportState:       true,
			// 	ImportStateVerify: true,
			// },
			{
				Config: testResourceInstanceExternalIpsConfig,
				Check:  checkResourceInstanceExternalIps(fourthResourceName),
			},
			{
				ResourceName:      fourthResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

var testResourceInstanceConfig = `
data "oxide_projects" "project_list" {}

resource "oxide_instance" "test" {
  project_id        = element(tolist(data.oxide_projects.project_list.projects[*].id), 0)
  description       = "a test instance"
  name              = "terraform-acc-myinstance"
  host_name         = "terraform-acc-myhost"
  memory            = 1073741824
  ncpus             = 1
  timeouts = {
    read   = "1m"
	create = "3m"
	delete = "2m"
  }
}
`

func checkResourceInstance(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test instance"),
		resource.TestCheckResourceAttr(resourceName, "name", "terraform-acc-myinstance"),
		resource.TestCheckResourceAttr(resourceName, "host_name", "terraform-acc-myhost"),
		resource.TestCheckResourceAttr(resourceName, "memory", "1073741824"),
		resource.TestCheckResourceAttr(resourceName, "ncpus", "1"),
		resource.TestCheckResourceAttrSet(resourceName, "run_state"),
		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		resource.TestCheckResourceAttrSet(resourceName, "time_run_state_updated"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.read", "1m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.delete", "2m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.create", "3m"),
	}...)
}

var testResourceInstanceDiskConfig = `
data "oxide_projects" "project_list" {}

resource "oxide_disk" "test-instance" {
  project_id        = element(tolist(data.oxide_projects.project_list.projects[*].id), 0)
  description       = "a test disk"
  name              = "terraform-acc-mydisk1"
  size              = 1073741824
  disk_source       = { blank = 512 }
}

resource "oxide_disk" "test-instance2" {
  project_id        = element(tolist(data.oxide_projects.project_list.projects[*].id), 0)
  description       = "a test disk"
  name              = "terraform-acc-mydisk2"
  size              = 1073741824
  disk_source       = { blank = 512 }
}

resource "oxide_instance" "test2" {
  project_id        = element(tolist(data.oxide_projects.project_list.projects[*].id), 0)
  description       = "a test instance"
  name              = "terraform-acc-myinstance2"
  host_name         = "terraform-acc-myhost"
  memory            = 1073741824
  ncpus             = 1
  attach_to_disks   = ["terraform-acc-mydisk1", "terraform-acc-mydisk2"]
}
`

func checkResourceInstanceDisk(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test instance"),
		resource.TestCheckResourceAttr(resourceName, "name", "terraform-acc-myinstance2"),
		resource.TestCheckResourceAttr(resourceName, "host_name", "terraform-acc-myhost"),
		resource.TestCheckResourceAttr(resourceName, "memory", "1073741824"),
		resource.TestCheckResourceAttr(resourceName, "ncpus", "1"),
		resource.TestCheckResourceAttr(resourceName, "attach_to_disks.0", "terraform-acc-mydisk1"),
		resource.TestCheckResourceAttr(resourceName, "attach_to_disks.1", "terraform-acc-mydisk2"),
		resource.TestCheckResourceAttrSet(resourceName, "run_state"),
		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		resource.TestCheckResourceAttrSet(resourceName, "time_run_state_updated"),
	}...)
}

// TODO: Restore when Network interfaces work on simulated omicron again
//
// var testResourceInstanceNetworkInterfaceConfig = `
// data "oxide_projects" "project_list" {}
//
// resource "oxide_instance" "test3" {
//   project_id        = element(tolist(data.oxide_projects.project_list.projects[*].id), 0)
//   description       = "a test instance"
//   name              = "terraform-acc-myinstance3"
//   host_name         = "terraform-acc-myhost"
//   memory            = 1073741824
//   ncpus             = 1
//   network_interface = [
//     {
//       description = "a network interface"
//       name        = "terraform-acc-mynetworkinterface"
//       subnet_name = "default"
//       vpc_name    = "default"
//     }
//   ]
// }
// `
//
// func checkResourceInstanceNetworkInterface(resourceName string) resource.TestCheckFunc {
// 	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
// 		resource.TestCheckResourceAttrSet(resourceName, "id"),
// 		resource.TestCheckResourceAttr(resourceName, "description", "a test instance"),
// 		resource.TestCheckResourceAttr(resourceName, "name", "terraform-acc-myinstance3"),
// 		resource.TestCheckResourceAttr(resourceName, "host_name", "terraform-acc-myhost"),
// 		resource.TestCheckResourceAttr(resourceName, "memory", "1073741824"),
// 		resource.TestCheckResourceAttr(resourceName, "ncpus", "1"),
// 		resource.TestCheckResourceAttr(resourceName, "network_interface.0.description", "a network interface"),
// 		resource.TestCheckResourceAttr(resourceName, "network_interface.0.name", "terraform-acc-mynetworkinterface"),
// 		resource.TestCheckResourceAttrSet(resourceName, "network_interface.0.ip"),
// 		resource.TestCheckResourceAttrSet(resourceName, "network_interface.0.subnet_id"),
// 		resource.TestCheckResourceAttrSet(resourceName, "network_interface.0.vpc_id"),
// 		resource.TestCheckResourceAttrSet(resourceName, "run_state"),
// 		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
// 		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
// 		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
// 		resource.TestCheckResourceAttrSet(resourceName, "time_run_state_updated"),
// 	}...)
// }

var testResourceInstanceExternalIpsConfig = `
data "oxide_projects" "project_list" {}

resource "oxide_instance" "test4" {
  project_id        = element(tolist(data.oxide_projects.project_list.projects[*].id), 0)
  description       = "a test instance"
  name              = "terraform-acc-myinstance4"
  host_name         = "terraform-acc-myhost"
  memory            = 1073741824
  ncpus             = 1
  external_ips   = ["default"]
}
`

func checkResourceInstanceExternalIps(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test instance"),
		resource.TestCheckResourceAttr(resourceName, "name", "terraform-acc-myinstance4"),
		resource.TestCheckResourceAttr(resourceName, "host_name", "terraform-acc-myhost"),
		resource.TestCheckResourceAttr(resourceName, "memory", "1073741824"),
		resource.TestCheckResourceAttr(resourceName, "ncpus", "1"),
		resource.TestCheckResourceAttr(resourceName, "external_ips.0", "default"),
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

		params := oxideSDK.InstanceViewParams{
			Instance: "terraform-acc-myinstance",
			Project:  "test",
		}
		res, err := client.InstanceView(params)
		if err != nil && is404(err) {
			continue
		}

		return fmt.Errorf("instance (%v) still exists", &res.Name)
	}

	return nil
}

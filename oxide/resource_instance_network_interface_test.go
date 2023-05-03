// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

// TODO: Currently the simulated omicron never reports a "stopped" state for an instance.
// This means that the following test never runs successfully. Find out what is happening
// and restore this test.

// func TestAccResourceInstanceNIC_full(t *testing.T) {
// 	resourceName := "oxide_instance_network_interface.test"
//
// 	resource.ParallelTest(t, resource.TestCase{
// 		PreCheck:                 func() { testAccPreCheck(t) },
// 		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
// 		CheckDestroy:             testAccInstanceNICDestroy,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testResourceInstanceNICConfig,
// 				Check:  checkResourceInstanceNIC(resourceName),
// 			},
// 			{
// 				ResourceName:      resourceName,
// 				ImportState:       true,
// 				ImportStateVerify: true,
// 			},
// 		},
// 	})
// }
//
// var testResourceInstanceNICConfig = `
// data "oxide_projects" "project_list" {}
//
// resource "oxide_vpc" "test_nic" {
// 	project_id  = element(tolist(data.oxide_projects.project_list.projects[*].id), 0)
// 	description = "a test vpc"
// 	name        = "terraform-acc-myvpc-nic"
// 	dns_name    = "my-vpc-dns"
// }
//
// resource "oxide_vpc_subnet" "test_nic" {
// 	vpc_id      = oxide_vpc.test_nic.id
// 	description = "a test vpc subnet"
// 	name        = "terraform-acc-mysubnet-nic"
// 	ipv4_block  = "192.168.1.0/24"
// }
//
// resource "oxide_instance" "test_nic" {
//   project_id  = element(tolist(data.oxide_projects.project_list.projects[*].id), 0)
//   description = "a test instance"
//   name        = "terraform-acc-myinstance-nic"
//   host_name   = "terraform-acc-myhost"
//   memory      = 1073741824
//   ncpus       = 1
// }
//
// resource "oxide_instance_network_interface" "test" {
//   instance_id = oxide_instance.test_nic.id
//   subnet_id   = oxide_vpc_subnet.test_nic.id
//   vpc_id      = oxide_vpc.test_nic.id
//   description = "a test nic"
//   name        = "terraform-acc-myinic"
// }
// `
//
// func checkResourceInstanceNIC(resourceName string) resource.TestCheckFunc {
// 	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
// 		resource.TestCheckResourceAttrSet(resourceName, "id"),
// 		resource.TestCheckResourceAttrSet(resourceName, "instance_id"),
// 		resource.TestCheckResourceAttrSet(resourceName, "subnet_id"),
// 		resource.TestCheckResourceAttr(resourceName, "description", "a test nic"),
// 		resource.TestCheckResourceAttr(resourceName, "name", "terraform-acc-mynic"),
// 		resource.TestCheckResourceAttrSet(resourceName, "ip_address"),
// 		resource.TestCheckResourceAttrSet(resourceName, "mac_address"),
// 		resource.TestCheckResourceAttrSet(resourceName, "primary"),
// 		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
// 		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
// 		resource.TestCheckResourceAttr(resourceName, "timeouts.read", "1m"),
// 		resource.TestCheckResourceAttr(resourceName, "timeouts.delete", "2m"),
// 		resource.TestCheckResourceAttr(resourceName, "timeouts.create", "3m"),
// 	}...)
// }
//
// func testAccInstanceNICDestroy(s *terraform.State) error {
// 	client, err := newTestClient()
// 	if err != nil {
// 		return err
// 	}
//
// 	for _, rs := range s.RootModule().Resources {
// 		if rs.Type != "oxide_instance_network_interface" {
// 			continue
// 		}
//
// 		params := oxideSDK.InstanceNetworkInterfaceViewParams{
// 			Interface: "terraform-acc-mynic",
// 			Instance:  "terraform-acc-myinstance-nic",
// 			Project:   "test",
// 		}
// 		res, err := client.InstanceNetworkInterfaceView(params)
// 		if err != nil && is404(err) {
// 			continue
// 		}
//
// 		return fmt.Errorf("instance NIC (%v) still exists", &res.Name)
// 	}
//
// 	return nil
// }

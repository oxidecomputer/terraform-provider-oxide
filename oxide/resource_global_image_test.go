// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

// TODO: Restore test when it is possible to delete a global image
//
// func TestAccResourceGlobalImage(t *testing.T) {
// 	resourceName := "oxide_global_image.test"
//
// 	resource.ParallelTest(t, resource.TestCase{
// 		PreCheck:          func() { testAccPreCheck(t) },
// 		ProviderFactories: testAccProviderFactory,
// 		CheckDestroy:      testAccGlobalImageDestroy,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testResourceGlobalImageConfig,
// 				Check:  checkResourceGlobalImage(resourceName),
// 			},
// 		},
// 	})
// }
//
// var testResourceGlobalImageConfig = `
// resource "oxide_global_image" "test" {
//   description          = "a test global_image"
//   name                 = "terraform-acc-myglobalimage"
//   image_source         = { you_can_boot_anything_as_long_as_its_alpine = "noop" }
//   block_size           = 512
//   distribution         = "alpine"
//   distribution_version = "propolis_blob"
// }
// `
//
// func checkResourceGlobalImage(resourceName string) resource.TestCheckFunc {
// 	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
// 		resource.TestCheckResourceAttrSet(resourceName, "id"),
// 		resource.TestCheckResourceAttr(resourceName, "description", "a test global_image"),
// 		resource.TestCheckResourceAttr(resourceName, "name", "terraform-acc-myglobalimage"),
// 		resource.TestCheckResourceAttr(resourceName, "block_size", "512"),
// 		resource.TestCheckResourceAttr(resourceName, "device_path", "/mnt/terraform-acc-myglobalimage"),
// 		resource.TestCheckResourceAttr(resourceName, "block_size", "512"),
// 		resource.TestCheckResourceAttrSet(resourceName, "image_source.you_can_boot_anything_as_long_as_its_alpine"),
// 		resource.TestCheckResourceAttrSet(resourceName, "size"),
// 		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
// 		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
// 		// TODO: Eventually we'll want to test creating a global_image from URL and snapshot
// 	}...)
// }
//
// func testAccGlobalImageDestroy(s *terraform.State) error {
// 	client, err := newTestClient()
// 	if err != nil {
// 		return err
// 	}
//
// 	for _, rs := range s.RootModule().Resources {
// 		if rs.Type != "oxide_global_image" {
// 			continue
// 		}
//
// 		res, err := client.ImageGlobalView("terraform-acc-myglobalimage")
//
// 		if err != nil && is404(err) {
// 			continue
// 		}
//
// 		return fmt.Errorf("global_image (%v) still exists", &res.Name)
// 	}
//
// 	return nil
// }
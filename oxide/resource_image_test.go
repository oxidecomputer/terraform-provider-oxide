// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

// TODO: Restore test when it is possible to delete a global image

// func TestAccResourceImage_full(t *testing.T) {
// 	resourceName := "oxide_image.test"
//
// 	resource.ParallelTest(t, resource.TestCase{
// 		PreCheck:                 func() { testAccPreCheck(t) },
// 		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
// 		CheckDestroy:             testAccImageDestroy,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testResourceImageConfig,
// 				Check:  checkResourceImage(resourceName),
// 			},
// 		},
// 	})
// }
//
// var testResourceImageConfig = `
// data "oxide_projects" "project_list" {}
//
// resource "oxide_image" "test" {
//   project_id   = element(tolist(data.oxide_projects.project_list.projects[*].id), 0)
//   description  = "a test image"
//   name         = "terraform-acc-myglobalimage"
//   image_source = { you_can_boot_anything_as_long_as_its_alpine = "noop" }
//   block_size   = 512
//   os           = "alpine"
//   version      = "propolis-blob"
// }
// `
//
// func checkResourceImage(resourceName string) resource.TestCheckFunc {
// 	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
// 		resource.TestCheckResourceAttrSet(resourceName, "id"),
// 		resource.TestCheckResourceAttr(resourceName, "description", "a test image"),
// 		resource.TestCheckResourceAttr(resourceName, "name", "terraform-acc-myglobalimage"),
// 		resource.TestCheckResourceAttr(resourceName, "block_size", "512"),
// 		resource.TestCheckResourceAttrSet(resourceName, "image_source.you_can_boot_anything_as_long_as_its_alpine"),
// 		resource.TestCheckResourceAttrSet(resourceName, "size"),
// 		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
// 		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
// 		// TODO: Eventually we'll want to test creating a image from URL and snapshot
// 	}...)
// }
//
// func testAccImageDestroy(s *terraform.State) error {
// 	client, err := newTestClient()
// 	if err != nil {
// 		return err
// 	}
//
// 	for _, rs := range s.RootModule().Resources {
//      if rs.Type != "oxide_image" {
//	        continue
//      }
//
//  	res, err := client.ImageView(oxide.ImageViewParams{
//			Image:   "terraform-acc-myglobalimage",
//			Project: "test",
//		})
//
//		if err != nil && is404(err) {
//			continue
//		}
//
//		return fmt.Errorf("image (%v) still exists", &res.Name)
//	}
//
//	return nil
// }

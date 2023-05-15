// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

// TODO: Restore test when it is possible to delete images, otherwise tests always fail

// type resourceImageConfig struct {
// 	BlockName        string
// 	ImageName        string
// 	SupportBlockName string
// }
//
// var resourceImageConfigTpl = `
//   data "oxide_projects" "{{.SupportBlockName}}" {}
//
//   resource "oxide_image" "{{.BlockName}}" {
//     project_id  = element(tolist(data.oxide_projects.{{.SupportBlockName}}.projects[*].id), 0)
//     description = "a test image"
//     name        = "{{.ImageName}}"
//     source_url  = "you_can_boot_anything_as_long_as_its_alpine"
//     block_size  = 512
//     os          = "alpine"
//     version     = "propolis-blob"
//     timeouts = {
//      read   = "1m"
//      create = "3m"
//     }
//   }
// `
//
// var resourceImageUpdateConfigTpl = `
//   data "oxide_projects" "{{.SupportBlockName}}" {}
//
//   resource "oxide_image" "{{.BlockName}}" {
//     description = "a test image"
//     name        = "{{.ImageName}}"
//     source_url  = "you_can_boot_anything_as_long_as_its_alpine"
//     block_size  = 512
//     os          = "alpine"
//     version     = "propolis-blob"
//   }
// `
//
// func TestAccResourceImage_full(t *testing.T) {
// 	imageName := newResourceName()
// 	blockName := newBlockName("image")
// 	supportBlockName := newBlockName("support")
// 	resourceName := fmt.Sprintf("oxide_image.%s", blockName)
// 	config, err := parsedAccConfig(
// 		resourceImageConfig{
// 			BlockName:        blockName,
// 			ImageName:        imageName,
// 			SupportBlockName: supportBlockName,
// 		},
// 		resourceImageConfigTpl,
// 	)
// 	if err != nil {
// 		t.Errorf("error parsing config template data: %e", err)
// 	}
//
// 	configUpdate, err := parsedAccConfig(
// 		resourceImageConfig{
// 			BlockName:        blockName,
// 			ImageName:        imageName,
// 			SupportBlockName: supportBlockName,
// 		},
// 		resourceImageUpdateConfigTpl,
// 	)
// 	if err != nil {
// 		t.Errorf("error parsing config template data: %e", err)
// 	}
//
// 	resource.ParallelTest(t, resource.TestCase{
// 		PreCheck:                 func() { testAccPreCheck(t) },
// 		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
// 		CheckDestroy:             testAccImageDestroy,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: config,
// 				Check:  checkResourceImage(resourceName, imageName),
// 			},
// 			{
// 				ResourceName:      resourceName,
// 				ImportState:       true,
// 				ImportStateVerify: true,
// 				// TODO: Remove once 'you_can_boot_anything_as_long_as_its_alpine'
// 				// is removed
// 				ImportStateVerifyIgnore: []string{"source_url"},
// 			},
// 			{
// 				Config: configUpdate,
// 				Check:  checkResourceImageUpdate(resourceName, imageName),
// 			},
// 			{
// 				Config: config,
// 				Check:  checkResourceImage(resourceName, imageName),
// 			},
// 		},
// 	})
// }
//
// func checkResourceImage(resourceName, imageName string) resource.TestCheckFunc {
// 	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
// 		resource.TestCheckResourceAttrSet(resourceName, "id"),
// 		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
// 		resource.TestCheckResourceAttr(resourceName, "description", "a test image"),
// 		resource.TestCheckResourceAttr(resourceName, "name", imageName),
// 		resource.TestCheckResourceAttr(resourceName, "block_size", "512"),
// 		resource.TestCheckResourceAttr(resourceName, "source_url", "you_can_boot_anything_as_long_as_its_alpine"),
// 		resource.TestCheckResourceAttrSet(resourceName, "size"),
// 		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
// 		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
// 		resource.TestCheckResourceAttr(resourceName, "timeouts.read", "1m"),
// 		resource.TestCheckResourceAttr(resourceName, "timeouts.create", "3m"),
// 		// TODO: Eventually we'll want to test creating a image from URL and snapshot
// 	}...)
// }
//
// func checkResourceImageUpdate(resourceName, imageName string) resource.TestCheckFunc {
// 	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
// 		resource.TestCheckResourceAttrSet(resourceName, "id"),
// 		resource.TestCheckNoResourceAttr(resourceName, "project_id"),
// 		resource.TestCheckResourceAttr(resourceName, "description", "a test image"),
// 		resource.TestCheckResourceAttr(resourceName, "name", imageName),
// 		resource.TestCheckResourceAttr(resourceName, "block_size", "512"),
// 		resource.TestCheckResourceAttr(resourceName, "source_url", "you_can_boot_anything_as_long_as_its_alpine"),
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
// 		if rs.Type != "oxide_image" {
// 			continue
// 		}
//
// 		res, err := client.ImageView(oxideSDK.ImageViewParams{
// 			Image: oxideSDK.NameOrId(rs.Primary.Attributes["id"]),
// 		})
//
// 		if err != nil && is404(err) {
// 			continue
// 		}
//
// 		return fmt.Errorf("image (%v) still exists", &res.Name)
// 	}
//
// 	return nil
// }
//

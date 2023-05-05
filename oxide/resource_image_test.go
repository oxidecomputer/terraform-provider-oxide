// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

// TODO: Restore test when it is possible to delete images

// type resourceImageConfig struct {
// 	BlockName        string
// 	ImageName        string
// 	SupportBlockName string
// }
//
// var resourceImageConfigTpl = `
// data "oxide_projects" "{{.SupportBlockName}}" {}
//
// resource "oxide_image" "{{.BlockName}}" {
//   project_id   = element(tolist(data.oxide_projects.{{.SupportBlockName}}.projects[*].id), 0)
//   description  = "a test image"
//   name         = "{{.ImageName}}"
//   image_source = { you_can_boot_anything_as_long_as_its_alpine = "noop" }
//   block_size   = 512
//   os           = "alpine"
//   version      = "propolis-blob"
//   timeouts = {
//    read   = "1m"
//    create = "3m"
//   }
// }
// `
//
// func TestAccResourceImage_full(t *testing.T) {
// 	imageName := fmt.Sprintf("acc-terraform-%s", uuid.New())
// 	blockName := fmt.Sprintf("acc-resource-image-%s", uuid.New())
// 	resourceName := fmt.Sprintf("oxide_image.%s", blockName)
// 	config, err := parsedAccConfig(
// 		resourceImageConfig{
// 			BlockName:        blockName,
// 			ImageName:        imageName,
// 			SupportBlockName: fmt.Sprintf("acc-support-%s", uuid.New()),
// 		},
// 		resourceImageConfigTpl,
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
// 				// TODO: Remove once https://github.com/oxidecomputer/terraform-provider-oxide/issues/102
// 				// has been worked on.
// 				ImportStateVerifyIgnore: []string{"image_source"},
// 			},
// 		},
// 	})
// }
//
// func checkResourceImage(resourceName, imageName string) resource.TestCheckFunc {
// 	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
// 		resource.TestCheckResourceAttrSet(resourceName, "id"),
// 		resource.TestCheckResourceAttr(resourceName, "description", "a test image"),
// 		resource.TestCheckResourceAttr(resourceName, "name", imageName),
// 		resource.TestCheckResourceAttr(resourceName, "block_size", "512"),
// 		resource.TestCheckResourceAttrSet(resourceName, "image_source.you_can_boot_anything_as_long_as_its_alpine"),
// 		resource.TestCheckResourceAttrSet(resourceName, "size"),
// 		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
// 		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
// 		resource.TestCheckResourceAttr(resourceName, "timeouts.read", "1m"),
// 		resource.TestCheckResourceAttr(resourceName, "timeouts.create", "3m"),
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

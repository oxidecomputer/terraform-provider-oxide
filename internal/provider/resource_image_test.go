// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/oxidecomputer/oxide.go/oxide"
)

type resourceImageConfig struct {
	BlockName        string
	ImageName        string
	SupportBlockName string
}

// TODO: Use a fetched snapshot ID when the snapshot data source is implemented
var resourceImageConfigTpl = `
 data "oxide_project" "{{.SupportBlockName}}" {
 	name = "tf-acc-test"
 }

 resource "oxide_image" "{{.BlockName}}" {
   project_id         = data.oxide_project.{{.SupportBlockName}}.id
   description        = "a test image"
   name               = "{{.ImageName}}"
   source_snapshot_id = "ffecbbfd-bd42-42ce-b023-e33f4020a858"
   os                 = "alpine"
   version            = "propolis-blob"
   timeouts = {
    read   = "1m"
    create = "3m"
   }
 }
 `

func TestAccResourceImage_full(t *testing.T) {
	imageName := newResourceName()
	blockName := newBlockName("image")
	supportBlockName := newBlockName("support")
	resourceName := fmt.Sprintf("oxide_image.%s", blockName)
	config, err := parsedAccConfig(
		resourceImageConfig{
			BlockName:        blockName,
			ImageName:        imageName,
			SupportBlockName: supportBlockName,
		},
		resourceImageConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccImageDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  checkResourceImage(resourceName, imageName),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"source_snapshot_id"},
			},
		},
	})
}

func checkResourceImage(resourceName, imageName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test image"),
		resource.TestCheckResourceAttr(resourceName, "name", imageName),
		resource.TestCheckResourceAttrSet(resourceName, "block_size"),
		resource.TestCheckResourceAttr(resourceName, "source_snapshot_id", "ffecbbfd-bd42-42ce-b023-e33f4020a858"),
		resource.TestCheckResourceAttrSet(resourceName, "size"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.read", "1m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.create", "3m"),
	}...)
}

func testAccImageDestroy(s *terraform.State) error {
	client, err := newTestClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "oxide_image" {
			continue
		}

		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()

		params := oxide.ImageViewParams{
			Image: oxide.NameOrId(rs.Primary.Attributes["id"]),
		}
		res, err := client.ImageView(ctx, params)

		if err != nil && is404(err) {
			continue
		}

		return fmt.Errorf("image (%v) still exists", &res.Name)
	}

	return nil
}

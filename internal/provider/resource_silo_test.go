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

type resourceSiloConfig struct {
	BlockName string
	SiloName  string
}

var resourceSiloConfigTpl = `
resource "oxide_silo" "{{.BlockName}}" {
	description       		= "a test silo"
	name              		= "{{.SiloName}}"
	admin_group_name 		= "test_admin"
	identity_mode 			= "saml_jit"
	discoverable 			= true
	mapped_fleet_roles 		= {
		admin  = ["admin", "collaborator"]
		viewer = ["viewer"]
	}
	quotas 					= {
		cpus 				= 8
		memory 				= 32
		storage 			= 100
	}
	tls_certificates 		= [
		{
			cert 			= "PUBLIC_KEY"
			description 	= "test cert 1"
			key 			= "PRIVATE_KEY"
			name 			= "silo_cert_1"
			service 		= "service1"
		},
	]
	timeouts = {
		read   = "1m"
		create = "3m"
		delete = "2m"
		update = "4m"
	}
  }
`

func TestAccCloudResourceSilo_full(t *testing.T) {
	siloName := newResourceName()
	blockName := newBlockName("silo")
	resourceName := fmt.Sprintf("oxide_silo.%s", blockName)

	config, err := parsedAccConfig(
		resourceSiloConfig{
			BlockName: blockName,
			SiloName:  siloName,
		},
		resourceSiloConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccSiloDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  checkResourceSilo(resourceName, siloName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func checkResourceSilo(resourceName, siloName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test silo"),
		resource.TestCheckResourceAttr(resourceName, "name", siloName),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.read", "1m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.delete", "2m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.create", "3m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.update", "4m"),
	}...)
}

func testAccSiloDestroy(s *terraform.State) error {
	client, err := newTestClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "oxide_silo" {
			continue
		}

		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()

		params := oxide.SiloViewParams{
			Silo: oxide.NameOrId(rs.Primary.Attributes["id"]),
		}
		res, err := client.SiloView(ctx, params)
		if err != nil && is404(err) {
			continue
		}
		return fmt.Errorf("silo (%v) still exists", &res.Name)
	}

	return nil
}

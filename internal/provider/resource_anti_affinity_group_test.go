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

type resourceAntiAffinityGroupConfig struct {
	BlockName             string
	SupportBlockName      string
	AntiAffinityGroupName string
}

var resourceAntiAffinityGroupConfigTpl = `
data "oxide_project" "{{.SupportBlockName}}" {
	name = "tf-acc-test"
}

resource "oxide_anti_affinity_group" "{{.BlockName}}" {
	project_id  = data.oxide_project.{{.SupportBlockName}}.id
	description = "a test anti-affinity group"
	name        = "{{.AntiAffinityGroupName}}"
	policy      = "allow"
	timeouts = {
		read   = "1m"
		create = "3m"
		delete = "2m"
		update = "4m"
	}
  }
`

var resourceAntiAffinityGroupUpdateConfigTpl = `
data "oxide_project" "{{.SupportBlockName}}" {
	name = "tf-acc-test"
}

resource "oxide_anti_affinity_group" "{{.BlockName}}" {
	project_id  = data.oxide_project.{{.SupportBlockName}}.id
	description = "a test updated"
	name        = "{{.AntiAffinityGroupName}}"
	policy      = "allow"
  }
`

func TestAccCloudResourceAntiAffinityGroup_full(t *testing.T) {
	antiAffinityGroupName := newResourceName()
	blockName := newBlockName("anti_affinity_group")
	resourceName := fmt.Sprintf("oxide_anti_affinity_group.%s", blockName)
	supportBlockName := newBlockName("support")
	config, err := parsedAccConfig(
		resourceAntiAffinityGroupConfig{
			BlockName:             blockName,
			AntiAffinityGroupName: antiAffinityGroupName,
			SupportBlockName:      supportBlockName,
		},
		resourceAntiAffinityGroupConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	antiAffinityGroupNameUpdated := antiAffinityGroupName + "-updated"
	configUpdate, err := parsedAccConfig(
		resourceAntiAffinityGroupConfig{
			BlockName:             blockName,
			AntiAffinityGroupName: antiAffinityGroupNameUpdated,
			SupportBlockName:      supportBlockName,
		},
		resourceAntiAffinityGroupUpdateConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccAntiAffinityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  checkResourceAntiAffinityGroup(resourceName, antiAffinityGroupName),
			},
			{
				Config: configUpdate,
				Check:  checkResourceAntiAffinityGroupUpdate(resourceName, antiAffinityGroupNameUpdated),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func checkResourceAntiAffinityGroup(resourceName, antiAffinityGroupName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test anti-affinity group"),
		resource.TestCheckResourceAttr(resourceName, "name", antiAffinityGroupName),
		resource.TestCheckResourceAttr(resourceName, "failure_domain", "sled"),
		resource.TestCheckResourceAttr(resourceName, "policy", "allow"),
		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.read", "1m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.delete", "2m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.create", "3m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.update", "4m"),
	}...)
}

func checkResourceAntiAffinityGroupUpdate(resourceName, antiAffinityGroupName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test updated"),
		resource.TestCheckResourceAttr(resourceName, "name", antiAffinityGroupName),
		resource.TestCheckResourceAttr(resourceName, "failure_domain", "sled"),
		resource.TestCheckResourceAttr(resourceName, "policy", "allow"),
		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
	}...)
}

func testAccAntiAffinityGroupDestroy(s *terraform.State) error {
	client, err := newTestClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "oxide_anti_affinity_group" {
			continue
		}

		params := oxide.AntiAffinityGroupViewParams{
			AntiAffinityGroup: oxide.NameOrId(rs.Primary.Attributes["id"]),
		}

		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()

		res, err := client.AntiAffinityGroupView(ctx, params)
		if err != nil && is404(err) {
			continue
		}

		return fmt.Errorf("anti-affinity group (%v) still exists", &res.Name)
	}

	return nil
}

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package project_test

import (
	"context"
	"fmt"
	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/sharedtest"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/oxidecomputer/oxide.go/oxide"
	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/shared"
)

type resourceProjectConfig struct {
	BlockName   string
	ProjectName string
}

var resourceProjectConfigTpl = `
resource "oxide_project" "{{.BlockName}}" {
	description       = "a test project"
	name              = "{{.ProjectName}}"
	timeouts = {
		read   = "1m"
		create = "3m"
		delete = "2m"
		update = "4m"
	}
  }
`

var resourceProjectUpdateConfigTpl = `
resource "oxide_project" "{{.BlockName}}" {
	description       = "a new description for project"
	name              = "{{.ProjectName}}"
  }
`

func TestAccCloudResourceProject_full(t *testing.T) {
	projectName := sharedtest.NewResourceName()
	blockName := sharedtest.NewBlockName("project")
	resourceName := fmt.Sprintf("oxide_project.%s", blockName)
	config, err := sharedtest.ParsedAccConfig(
		resourceProjectConfig{
			BlockName:   blockName,
			ProjectName: projectName,
		},
		resourceProjectConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	projectNameUpdated := projectName + "-updated"
	configUpdate, err := sharedtest.ParsedAccConfig(
		resourceProjectConfig{
			BlockName:   blockName,
			ProjectName: projectNameUpdated,
		},
		resourceProjectUpdateConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { sharedtest.PreCheck(t) },
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		CheckDestroy:             testAccProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  checkResourceProject(resourceName, projectName),
			},
			{
				Config: configUpdate,
				Check:  checkResourceProjectUpdate(resourceName, projectNameUpdated),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func checkResourceProject(resourceName, projectName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test project"),
		resource.TestCheckResourceAttr(resourceName, "name", projectName),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.read", "1m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.delete", "2m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.create", "3m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.update", "4m"),
	}...)
}

func checkResourceProjectUpdate(resourceName, projectName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(
			resourceName,
			"description",
			"a new description for project",
		),
		resource.TestCheckResourceAttr(resourceName, "name", projectName),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
	}...)
}

func testAccProjectDestroy(s *terraform.State) error {
	client, err := sharedtest.NewTestClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "oxide_project" {
			continue
		}

		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()

		params := oxide.ProjectViewParams{
			Project: oxide.NameOrId(rs.Primary.Attributes["id"]),
		}
		res, err := client.ProjectView(ctx, params)
		if err != nil && shared.Is404(err) {
			continue
		}
		return fmt.Errorf("project (%v) still exists", &res.Name)
	}

	return nil
}

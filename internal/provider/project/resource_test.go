// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package project_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/sharedtest"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/oxidecomputer/oxide.go/oxide"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/shared"
)

type resourceConfig struct {
	BlockName   string
	ProjectName string
}

var resourceConfigTpl = `
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

var resourceUpdateConfigTpl = `
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
		resourceConfig{
			BlockName:   blockName,
			ProjectName: projectName,
		},
		resourceConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	projectNameUpdated := projectName + "-updated"
	configUpdate, err := sharedtest.ParsedAccConfig(
		resourceConfig{
			BlockName:   blockName,
			ProjectName: projectNameUpdated,
		},
		resourceUpdateConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { sharedtest.PreCheck(t) },
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		CheckDestroy:             testAccResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  checkResource(resourceName, projectName),
			},
			{
				Config: configUpdate,
				Check:  checkResourceUpdate(resourceName, projectNameUpdated),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func checkResource(resourceName, projectName string) resource.TestCheckFunc {
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

func checkResourceUpdate(resourceName, projectName string) resource.TestCheckFunc {
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

func testAccResourceDestroy(s *terraform.State) error {
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

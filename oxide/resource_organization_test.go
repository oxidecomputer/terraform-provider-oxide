// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceOrganization(t *testing.T) {
	resourceName := "oxide_organization.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactory,
		CheckDestroy:      testAccOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testResourceOrganizationConfig,
				Check:  checkResourceOrganization(resourceName),
			},
			{
				Config: testResourceOrganizationUpdateConfig,
				Check:  checkResourceOrganizationUpdate(resourceName),
			},
		},
	})
}

var testResourceOrganizationConfig = `
resource "oxide_organization" "test" {
	description       = "a test organization"
	name              = "terraform-acc-myorg"
  }
`

func checkResourceOrganization(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test organization"),
		resource.TestCheckResourceAttr(resourceName, "name", "terraform-acc-myorg"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
	}...)
}

var testResourceOrganizationUpdateConfig = `
resource "oxide_organization" "test" {
	description       = "a new description for organization"
	name              = "terraform-acc-myorg"
  }
`

func checkResourceOrganizationUpdate(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a new description for organization"),
		resource.TestCheckResourceAttr(resourceName, "name", "terraform-acc-myorg"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
	}...)
}

func testAccOrganizationDestroy(s *terraform.State) error {
	client, err := newTestClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "oxide_organization" {
			continue
		}

		res, err := client.OrganizationView("terraform-acc-myorg")
		if err != nil && is404(err) {
			continue
		}
		return fmt.Errorf("organization (%v) still exists", &res.Name)
	}

	return nil
}

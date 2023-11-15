// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type dataSourceSSHKeyConfig struct {
	BlockName         string
	SupportBlockName  string
	SupportBlockName2 string
}

var dataSourceSSHKeyConfigTpl = `
data "oxide_project" "{{.SupportBlockName}}" {
	name = "tf-acc-test"
}

data "oxide_ssh_keys" "{{.SupportBlockName2}}" {
  project_id = data.oxide_project.{{.SupportBlockName}}.id
}

data "oxide_ssh_key" "{{.BlockName}}" {
  name = "{{.Name}}"
  timeouts = {
    read = "1m"
  }
}
`

// NB: This test is ignored as it won't pass until we implement the
// `oxide_ssh_keys` data source
//
// NB: The project must be populated with at least one SSH Key for this test to pass
func TestAccDataSourceSSHKey_full(t *testing.T) {
	t.Skip("skipping test until `oxide_ssh_keys` data source is implemented.")

	blockName := newBlockName("datasource-ssh-key")
	config, err := parsedAccConfig(
		dataSourceSSHKeyConfig{
			BlockName:         blockName,
			SupportBlockName:  newBlockName("support"),
			SupportBlockName2: newBlockName("support"),
		},
		dataSourceSSHKeyConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: checkDataSourceSSHKey(
					fmt.Sprintf("data.oxide_ssh_key.%s", blockName),
				),
			},
		},
	})
}

func checkDataSourceSSHKey(dataName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(dataName, "id"),
		resource.TestCheckResourceAttrSet(dataName, "name"),
		resource.TestCheckResourceAttr(dataName, "description", ""),
		resource.TestCheckResourceAttrSet(dataName, "public_key"),
		resource.TestCheckResourceAttrSet(dataName, "silo_user_id"),
		resource.TestCheckResourceAttrSet(dataName, "time_created"),
		resource.TestCheckResourceAttrSet(dataName, "time_modified"),
		resource.TestCheckResourceAttr(dataName, "timeouts.read", "1m"),
	}...)
}

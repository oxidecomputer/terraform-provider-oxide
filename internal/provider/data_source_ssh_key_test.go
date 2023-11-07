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
	BlockName string
	Name      string
}

var dataSourceSSHKeyConfigTpl = `
data "oxide_ssh_key" "{{.BlockName}}" {
  name = "{{.Name}}"
  timeouts = {
    read = "1m"
  }
}
`

func TestAccDataSourceSSHKey_full(t *testing.T) {
	blockName := newBlockName("datasource-ssh-key")
	sshKeyName := newResourceName()
	config, err := parsedAccConfig(
		dataSourceSSHKeyConfig{
			BlockName: blockName,
			Name:      sshKeyName,
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
					sshKeyName,
				),
			},
		},
	})
}

func checkDataSourceSSHKey(dataName string, resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(dataName, "id"),
		resource.TestCheckResourceAttr(dataName, "name", resourceName),
		resource.TestCheckResourceAttr(dataName, "description", ""),
		resource.TestCheckResourceAttrSet(dataName, "public_key"),
		resource.TestCheckResourceAttrSet(dataName, "silo_user_id"),
		resource.TestCheckResourceAttrSet(dataName, "time_created"),
		resource.TestCheckResourceAttrSet(dataName, "time_modified"),
		resource.TestCheckResourceAttr(dataName, "timeouts.read", "1m"),
	}...)
}

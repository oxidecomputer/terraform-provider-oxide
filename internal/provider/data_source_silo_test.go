// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type dataSourceSiloConfig struct {
	BlockName string
}

var dataSourceSiloConfigTpl = `
data "oxide_silo" "{{.BlockName}}" {
  name = "default"
  timeouts = {
    read = "1m"
  }
}
`

func TestAccSiloDataSourceSilo_full(t *testing.T) {
	blockName := newBlockName("datasource-silo")
	config, err := parsedAccConfig(
		dataSourceSiloConfig{
			BlockName: blockName,
		},
		dataSourceSiloConfigTpl,
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
				Check: checkDataSourceSilo(
					fmt.Sprintf("data.oxide_silo.%s", blockName),
				),
			},
		},
	})
}

func checkDataSourceSilo(dataName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(dataName, "name"),
		resource.TestCheckResourceAttr(dataName, "timeouts.read", "1m"),
		resource.TestCheckResourceAttrSet(dataName, "id"),
		resource.TestCheckResourceAttrSet(dataName, "discoverable"),
		resource.TestCheckResourceAttrSet(dataName, "identity_mode"),
		resource.TestCheckResourceAttrSet(dataName, "description"),
		resource.TestCheckResourceAttrSet(dataName, "time_created"),
		resource.TestCheckResourceAttrSet(dataName, "time_modified"),
	}...)
}

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type dataSourceFloatingIPConfig struct {
	BlockName         string
	Name              string
	SupportBlockName  string
	SupportBlockName2 string
}

var dataSourceFloatingIPConfigTpl = `
data "oxide_project" "{{.SupportBlockName}}" {
	name = "tf-acc-test"
}

resource "oxide_floating_ip" "{{.SupportBlockName2}}" {
  project_id  = data.oxide_project.{{.SupportBlockName}}.id
  name        = "{{.Name}}"
  description = "Floating IP."
  ip_version  = "v4"
}

data "oxide_floating_ip" "{{.BlockName}}" {
  project_name = data.oxide_project.{{.SupportBlockName}}.name
  name         = oxide_floating_ip.{{.SupportBlockName2}}.name
  timeouts = {
    read = "1m"
  }
}
`

func TestAccCloudDataSourceFloatingIP_full(t *testing.T) {
	blockName := newBlockName("datasource-floating-ip")
	resourceName := newResourceName()
	config, err := parsedAccConfig(
		dataSourceFloatingIPConfig{
			BlockName:         blockName,
			SupportBlockName:  newBlockName("support"),
			SupportBlockName2: newBlockName("support"),
			Name:              resourceName,
		},
		dataSourceFloatingIPConfigTpl,
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
				Check: checkDataSourceFloatingIP(
					fmt.Sprintf("data.oxide_floating_ip.%s", blockName),
					resourceName,
				),
			},
		},
	})
}

func checkDataSourceFloatingIP(dataName, keyName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(dataName, "id"),
		resource.TestCheckResourceAttr(dataName, "name", keyName),
		resource.TestCheckResourceAttr(dataName, "description", "Floating IP."),
		resource.TestCheckResourceAttrSet(dataName, "project_id"),
		resource.TestCheckResourceAttrSet(dataName, "project_name"),
		resource.TestCheckResourceAttrSet(dataName, "ip"),
		resource.TestCheckResourceAttrSet(dataName, "ip_pool_id"),
		resource.TestCheckResourceAttr(dataName, "instance_id", ""),
		resource.TestCheckResourceAttrSet(dataName, "time_created"),
		resource.TestCheckResourceAttrSet(dataName, "time_modified"),
		resource.TestCheckResourceAttr(dataName, "timeouts.read", "1m"),
	}...)
}

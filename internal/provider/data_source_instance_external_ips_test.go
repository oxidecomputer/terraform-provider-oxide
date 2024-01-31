// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type dataSourceInstanceExternalIPConfig struct {
	BlockName         string
	InstanceName      string
	InstanceBlockName string
	SupportBlockName  string
}

var datasourceInstanceExternalIPsConfigTpl = `
data "oxide_project" "{{.SupportBlockName}}" {
	name = "tf-acc-test"
}

resource "oxide_instance" "{{.InstanceBlockName}}" {
  project_id      = data.oxide_project.{{.SupportBlockName}}.id
  description     = "a test instance"
  name            = "{{.InstanceName}}"
  host_name       = "terraform-acc-myhost"
  memory          = 1073741824
  ncpus           = 1
  start_on_create = false
  external_ips = [
	{
	  type = "ephemeral"
	}
  ]
}

data "oxide_instance_external_ips" "{{.BlockName}}" {
	instance_id = oxide_instance.{{.InstanceBlockName}}.id
	timeouts = {
	  read = "1m"
	}
  }
`

func TestAccDataSourceInstanceExternalIPs_full(t *testing.T) {
	blockName := newBlockName("datasource-instance-external-ips")
	config, err := parsedAccConfig(
		dataSourceInstanceExternalIPConfig{
			BlockName:         blockName,
			SupportBlockName:  newBlockName("support"),
			InstanceName:      newResourceName(),
			InstanceBlockName: newBlockName("instance"),
		},
		datasourceInstanceExternalIPsConfigTpl,
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
				Check: checkDataSourceInstanceExternalIPs(
					fmt.Sprintf("data.oxide_instance_external_ips.%s", blockName),
				),
			},
		},
	})
}

func checkDataSourceInstanceExternalIPs(dataName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(dataName, "id"),
		resource.TestCheckResourceAttr(dataName, "timeouts.read", "1m"),
		resource.TestCheckResourceAttrSet(dataName, "external_ips.0.ip"),
		resource.TestCheckResourceAttrSet(dataName, "external_ips.0.kind"),
	}...)
}

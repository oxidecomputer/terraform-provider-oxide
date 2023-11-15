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

type resourceSSHKeyConfig struct {
	BlockName   string
	Name        string
	Description string
	PublicKey   string
}

var resourceSSHKeyConfigTpl = `
resource "oxide_ssh_key" "{{.BlockName}}" {
  name        = "{{.Name}}"
  description = "{{.Description}}"
  public_key  = "{{.PublicKey}}"
  timeouts = {
    read   = "1m"
    create = "3m"
    delete = "2m"
    update = "4m"
  }
}
`

func TestAccResourceSSHKey_full(t *testing.T) {
	sshKeyName := newResourceName()
	description := "An SSH key."
	publicKey := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIE1clIQrzlQNqxgvpCCUFFOcTTFDOaqV+aocfsDZvxqB"
	blockName := newBlockName("ssh_key")
	resourceName := fmt.Sprintf("oxide_ssh_key.%s", blockName)
	config, err := parsedAccConfig(
		resourceSSHKeyConfig{
			BlockName:   blockName,
			Name:        sshKeyName,
			Description: description,
			PublicKey:   publicKey,
		},
		resourceSSHKeyConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccSSHKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  checkResourceSSHKey(resourceName, sshKeyName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func checkResourceSSHKey(resourceName, sshKeyName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "name", sshKeyName),
		resource.TestCheckResourceAttr(resourceName, "description", "An SSH Key."),
		resource.TestCheckResourceAttr(resourceName, "public_key", "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIE1clIQrzlQNqxgvpCCUFFOcTTFDOaqV+aocfsDZvxqB"),
		resource.TestCheckResourceAttrSet(resourceName, "silo_user_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.read", "1m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.delete", "2m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.create", "3m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.update", "4m"),
	}...)
}

func testAccSSHKeyDestroy(s *terraform.State) error {
	client, err := newTestClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "oxide_ssh_key" {
			continue
		}

		params := oxide.CurrentUserSshKeyViewParams{
			SshKey: oxide.NameOrId(rs.Primary.Attributes["id"]),
		}

		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()

		res, err := client.CurrentUserSshKeyView(ctx, params)
		if err != nil && is404(err) {
			continue
		}

		return fmt.Errorf("ssh key (%v) still exists", &res.Name)
	}

	return nil
}

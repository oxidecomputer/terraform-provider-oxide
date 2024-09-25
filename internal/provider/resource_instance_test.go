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

func TestAccCloudResourceInstance_full(t *testing.T) {
	type resourceInstanceConfig struct {
		BlockName        string
		InstanceName     string
		SupportBlockName string
	}

	type resourceInstanceFullConfig struct {
		BlockName         string
		InstanceName      string
		DiskBlockName     string
		DiskName          string
		SSHKeyName        string
		SupportBlockName  string
		SupportBlockName2 string
		SSHBlockName      string
		NicName           string
	}

	resourceInstanceConfigTpl := `
data "oxide_project" "{{.SupportBlockName}}" {
	name = "tf-acc-test"
}

resource "oxide_instance" "{{.BlockName}}" {
  project_id      = data.oxide_project.{{.SupportBlockName}}.id
  description     = "a test instance"
  name            = "{{.InstanceName}}"
  host_name       = "terraform-acc-myhost"
  memory          = 1073741824
  ncpus           = 1
  start_on_create = false
}
`

	resourceInstanceFullConfigTpl := `
data "oxide_project" "{{.SupportBlockName}}" {
  name = "tf-acc-test"
}

data "oxide_vpc_subnet" "{{.SupportBlockName2}}" {
  project_name = data.oxide_project.{{.SupportBlockName}}.name
  vpc_name     = "default"
  name         = "default"
}

resource "oxide_disk" "{{.DiskBlockName}}" {
  project_id  = data.oxide_project.{{.SupportBlockName}}.id
  description = "a test disk"
  name        = "{{.DiskName}}"
  size        = 1073741824
  block_size  = 512
}

resource "oxide_ssh_key" "{{.SSHBlockName}}" {
  name        = "{{.SSHKeyName}}"
  description = "A test key"
  public_key  = "ssh-ed25519 AAAA"
}

resource "oxide_instance" "{{.BlockName}}" {
  project_id       = data.oxide_project.{{.SupportBlockName}}.id
  boot_disk_id     = oxide_disk.{{.DiskBlockName}}.id
  description      = "a test instance"
  name             = "{{.InstanceName}}"
  host_name        = "terraform-acc-myhost"
  memory           = 1073741824
  ncpus            = 1
  start_on_create  = true
  disk_attachments = [oxide_disk.{{.DiskBlockName}}.id]
  ssh_public_keys  = [oxide_ssh_key.{{.SSHBlockName}}.id]
  external_ips = [
	{
	  type = "ephemeral"
	}
  ]
  network_interfaces = [
    {
      subnet_id   = data.oxide_vpc_subnet.{{.SupportBlockName2}}.id
      vpc_id      = data.oxide_vpc_subnet.{{.SupportBlockName2}}.vpc_id
      description = "a sample nic"
      name        = "{{.NicName}}"
    }
  ]
  timeouts = {
	read   = "1m"
	create = "3m"
	delete = "2m"
  }
}
`

	instanceName := newResourceName()
	blockName := newBlockName("instance")
	supportBlockName := newBlockName("support")
	resourceName := fmt.Sprintf("oxide_instance.%s", blockName)
	config, err := parsedAccConfig(
		resourceInstanceConfig{
			BlockName:        blockName,
			InstanceName:     instanceName,
			SupportBlockName: supportBlockName,
		},
		resourceInstanceConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	instanceName2 := newResourceName()
	instanceDiskName := newResourceName()
	instanceNicName := newResourceName()
	instanceSshKeyName := newResourceName()
	blockName2 := newBlockName("instance")
	diskBlockName := newBlockName("disk")
	supportBlockName3 := newBlockName("support")
	supportBlockName2 := newBlockName("support")
	supportBlockNameSSHKeys := newBlockName("support-instance-ssh-keys")
	resourceName2 := fmt.Sprintf("oxide_instance.%s", blockName2)
	config2, err := parsedAccConfig(
		resourceInstanceFullConfig{
			BlockName:         blockName2,
			InstanceName:      instanceName2,
			DiskBlockName:     diskBlockName,
			DiskName:          instanceDiskName,
			SSHKeyName:        instanceSshKeyName,
			SupportBlockName:  supportBlockName3,
			SupportBlockName2: supportBlockName2,
			NicName:           instanceNicName,
			SSHBlockName:      supportBlockNameSSHKeys,
		},
		resourceInstanceFullConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  checkResourceInstance(resourceName, instanceName),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"start_on_create"},
			},
			{
				Config: config2,
				Check:  checkResourceInstanceFull(resourceName2, instanceName2, instanceNicName),
			},
			{
				ResourceName:      resourceName2,
				ImportState:       true,
				ImportStateVerify: true,
				// External IPs cannot be imported as they are only present at create time
				ImportStateVerifyIgnore: []string{"start_on_create", "external_ips"},
			},
		},
	})
}

func TestAccCloudResourceInstance_extIPs(t *testing.T) {
	type resourceInstanceConfig struct {
		BlockName        string
		InstanceName     string
		SupportBlockName string
	}

	resourceInstanceExternalIPConfigTpl := `
data "oxide_project" "{{.SupportBlockName}}" {
	name = "tf-acc-test"
}

resource "oxide_instance" "{{.BlockName}}" {
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
`

	instanceName := newResourceName()
	blockName := newBlockName("instance")
	supportBlockName := newBlockName("support")
	resourceName := fmt.Sprintf("oxide_instance.%s", blockName)
	config, err := parsedAccConfig(
		resourceInstanceConfig{
			BlockName:        blockName,
			InstanceName:     instanceName,
			SupportBlockName: supportBlockName,
		},
		resourceInstanceExternalIPConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  checkResourceInstanceIP(resourceName, instanceName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// External IPs cannot be imported as they are only present at create time
				ImportStateVerifyIgnore: []string{"start_on_create", "external_ips"},
			},
		},
	})
}

func TestAccCloudResourceInstance_sshKeys(t *testing.T) {
	type resourceInstanceSshKeyConfig struct {
		BlockName         string
		SshKeyName        string
		InstanceName      string
		SupportBlockName  string
		SupportBlockName2 string
	}

	resourceInstanceSSHKeysConfigTpl := `
data "oxide_project" "{{.SupportBlockName}}" {
	name = "tf-acc-test"
}

resource "oxide_ssh_key" "{{.SupportBlockName2}}" {
  name        = "{{.SshKeyName}}"
  description = "A test key"
  public_key  = "ssh-ed25519 AAAA"
}

resource "oxide_instance" "{{.BlockName}}" {
  project_id      = data.oxide_project.{{.SupportBlockName}}.id
  description     = "a test instance"
  name            = "{{.InstanceName}}"
  host_name       = "terraform-acc-myhost"
  memory          = 1073741824
  ncpus           = 1
  ssh_public_keys = [oxide_ssh_key.{{.SupportBlockName2}}.id]
  start_on_create = false
}
`

	instanceSshKeysName := newResourceName()
	instanceSshKeysName2 := newResourceName()
	blockNameSshKeys := newBlockName("instance-ssh-keys")
	supportBlockNameSshKeys := newBlockName("support-instance-ssh-keys")
	supportBlockNameSshKeys2 := newBlockName("support-instance-ssh-keys-2")
	resourceName := fmt.Sprintf("oxide_instance.%s", blockNameSshKeys)
	configSshKeys, err := parsedAccConfig(
		resourceInstanceSshKeyConfig{
			BlockName:         blockNameSshKeys,
			SshKeyName:        instanceSshKeysName2,
			InstanceName:      instanceSshKeysName,
			SupportBlockName:  supportBlockNameSshKeys,
			SupportBlockName2: supportBlockNameSshKeys2,
		},
		resourceInstanceSSHKeysConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: configSshKeys,
				Check:  checkResourceInstanceSSHKeys(resourceName, instanceSshKeysName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// SSH Keys cannot be imported as they are only present at create time
				ImportStateVerifyIgnore: []string{"start_on_create"},
			},
		},
	})
}

func TestAccCloudResourceInstance_nic(t *testing.T) {
	type resourceInstanceNicConfig struct {
		BlockName        string
		SubnetBlockName  string
		NicName          string
		InstanceName     string
		SupportBlockName string
	}

	resourceInstanceNicConfigTpl := `
data "oxide_project" "{{.SupportBlockName}}" {
	name = "tf-acc-test"
}
  
data "oxide_vpc_subnet" "{{.SubnetBlockName}}" {
  project_name = data.oxide_project.{{.SupportBlockName}}.name
  vpc_name     = "default"
  name         = "default"
}
  
resource "oxide_instance" "{{.BlockName}}" {
  project_id       = data.oxide_project.{{.SupportBlockName}}.id
  description      = "a test instance"
  name             = "{{.InstanceName}}"
  host_name        = "terraform-acc-myhost"
  memory           = 1073741824
  ncpus            = 1
  start_on_create  = false
  network_interfaces = [
    {
      subnet_id   = data.oxide_vpc_subnet.{{.SubnetBlockName}}.id
      vpc_id      = data.oxide_vpc_subnet.{{.SubnetBlockName}}.vpc_id
      description = "a sample nic"
      name        = "{{.NicName}}"
    },
  ]
}
`

	resourceInstanceNicConfigUpdateTpl := `
data "oxide_project" "{{.SupportBlockName}}" {
	name = "tf-acc-test"
}
  
resource "oxide_instance" "{{.BlockName}}" {
  project_id       = data.oxide_project.{{.SupportBlockName}}.id
  description      = "a test instance"
  name             = "{{.InstanceName}}"
  host_name        = "terraform-acc-myhost"
  memory           = 1073741824
  ncpus            = 1
  start_on_create  = false
}
`

	instanceNicName := newResourceName()
	nicName := newResourceName()
	blockNameSubnet := newBlockName("instance-nic-subnet")
	blockNameInstanceNic := newBlockName("instance-nic")
	supportBlockName := newBlockName("support")
	supportBlockName2 := newBlockName("support")
	resourceNameInstanceNic := fmt.Sprintf("oxide_instance.%s", blockNameInstanceNic)
	configNic, err := parsedAccConfig(
		resourceInstanceNicConfig{
			BlockName:        blockNameInstanceNic,
			SubnetBlockName:  blockNameSubnet,
			InstanceName:     instanceNicName,
			NicName:          nicName,
			SupportBlockName: supportBlockName,
		},
		resourceInstanceNicConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}
	configNicUpdate, err := parsedAccConfig(
		resourceInstanceNicConfig{
			BlockName:        blockNameInstanceNic,
			SubnetBlockName:  blockNameSubnet,
			InstanceName:     instanceNicName,
			NicName:          nicName,
			SupportBlockName: supportBlockName2,
		},
		resourceInstanceNicConfigUpdateTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: configNic,
				Check:  checkResourceInstanceNic(resourceNameInstanceNic, instanceNicName, nicName),
			},
			{
				// Delete a nic
				Config: configNicUpdate,
				Check:  checkResourceInstanceNicUpdate(resourceNameInstanceNic, instanceNicName),
			},
			{
				// Recreate a nic
				Config: configNic,
				Check:  checkResourceInstanceNic(resourceNameInstanceNic, instanceNicName, nicName),
			},
			{
				ResourceName:      resourceNameInstanceNic,
				ImportState:       true,
				ImportStateVerify: true,
				// This option is only relevant for create, this means that it will
				// never be imported
				ImportStateVerifyIgnore: []string{"start_on_create"},
			},
		},
	})
}

func TestAccCloudResourceInstance_disk(t *testing.T) {
	type resourceInstanceDiskConfig struct {
		BlockName        string
		DiskBlockName    string
		DiskBlockName2   string
		DiskName         string
		DiskName2        string
		InstanceName     string
		SupportBlockName string
	}

	resourceInstanceDiskConfigTpl := `
data "oxide_project" "{{.SupportBlockName}}" {
	name = "tf-acc-test"
}

resource "oxide_disk" "{{.DiskBlockName}}" {
  project_id  = data.oxide_project.{{.SupportBlockName}}.id
  description = "a test disk"
  name        = "{{.DiskName}}"
  size        = 1073741824
  block_size  = 512
}

resource "oxide_disk" "{{.DiskBlockName2}}" {
  project_id  = data.oxide_project.{{.SupportBlockName}}.id
  description = "a test disk"
  name        = "{{.DiskName2}}"
  size        = 1073741824
  block_size  = 512
}

resource "oxide_instance" "{{.BlockName}}" {
  project_id      = data.oxide_project.{{.SupportBlockName}}.id
  boot_disk_id    = oxide_disk.{{.DiskBlockName2}}.id
  description     = "a test instance"
  name            = "{{.InstanceName}}"
  host_name       = "terraform-acc-myhost"
  memory          = 1073741824
  ncpus           = 1
  start_on_create = false
  disk_attachments = [oxide_disk.{{.DiskBlockName}}.id, oxide_disk.{{.DiskBlockName2}}.id]
}
`

	resourceInstanceDiskConfigUpdateTpl := `
data "oxide_project" "{{.SupportBlockName}}" {
	name = "tf-acc-test"
}

resource "oxide_disk" "{{.DiskBlockName}}" {
  project_id  = data.oxide_project.{{.SupportBlockName}}.id
  description = "a test disk"
  name        = "{{.DiskName}}"
  size        = 1073741824
  block_size  = 512
}

resource "oxide_disk" "{{.DiskBlockName2}}" {
  project_id  = data.oxide_project.{{.SupportBlockName}}.id
  description = "a test disk"
  name        = "{{.DiskName2}}"
  size        = 1073741824
  block_size  = 512
}

resource "oxide_instance" "{{.BlockName}}" {
  project_id      = data.oxide_project.{{.SupportBlockName}}.id
  boot_disk_id    = oxide_disk.{{.DiskBlockName}}.id
  description     = "a test instance"
  name            = "{{.InstanceName}}"
  host_name       = "terraform-acc-myhost"
  memory          = 1073741824
  ncpus           = 1
  start_on_create = false
  disk_attachments = [oxide_disk.{{.DiskBlockName}}.id]
}
`
	instanceDiskName := newResourceName()
	diskName := newResourceName()
	diskName2 := newResourceName()
	supportBlockName := newBlockName("support")
	supportBlockName2 := newBlockName("support-update")
	blockNameInstance := newBlockName("instance")
	blockNameInstanceDisk := newBlockName("instance-disk")
	blockNameInstanceDisk2 := newBlockName("instance-disk-2")
	resourceNameInstanceDisk := fmt.Sprintf("oxide_instance.%s", blockNameInstance)
	configDisk, err := parsedAccConfig(
		resourceInstanceDiskConfig{
			BlockName:        blockNameInstance,
			DiskBlockName:    blockNameInstanceDisk,
			DiskBlockName2:   blockNameInstanceDisk2,
			DiskName:         diskName,
			DiskName2:        diskName2,
			InstanceName:     instanceDiskName,
			SupportBlockName: supportBlockName,
		},
		resourceInstanceDiskConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}
	configDiskUpdate, err := parsedAccConfig(
		resourceInstanceDiskConfig{
			BlockName:        blockNameInstance,
			DiskBlockName:    blockNameInstanceDisk,
			DiskBlockName2:   blockNameInstanceDisk2,
			DiskName:         diskName,
			DiskName2:        diskName2,
			InstanceName:     instanceDiskName,
			SupportBlockName: supportBlockName2,
		},
		resourceInstanceDiskConfigUpdateTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: configDisk,
				Check:  checkResourceInstanceDisk(resourceNameInstanceDisk, instanceDiskName),
			},
			{
				// Detach a disk
				Config: configDiskUpdate,
				Check:  checkResourceInstanceDiskUpdate(resourceNameInstanceDisk, instanceDiskName),
			},
			{
				// Reattach disk
				Config: configDisk,
				Check:  checkResourceInstanceDisk(resourceNameInstanceDisk, instanceDiskName),
			},
			{
				ResourceName:      resourceNameInstanceDisk,
				ImportState:       true,
				ImportStateVerify: true,
				// This option is only relevant for create, this means that it will
				// never be imported
				ImportStateVerifyIgnore: []string{"start_on_create"},
			},
		},
	})
}

func checkResourceInstance(resourceName, instanceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test instance"),
		resource.TestCheckResourceAttr(resourceName, "name", instanceName),
		resource.TestCheckResourceAttr(resourceName, "host_name", "terraform-acc-myhost"),
		resource.TestCheckResourceAttr(resourceName, "memory", "1073741824"),
		resource.TestCheckResourceAttr(resourceName, "ncpus", "1"),
		resource.TestCheckResourceAttr(resourceName, "start_on_create", "false"),
		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		resource.TestCheckNoResourceAttr(resourceName, "ssh_public_keys"),
	}...)
}

func checkResourceInstanceFull(resourceName, instanceName, nicName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttrSet(resourceName, "boot_disk_id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test instance"),
		resource.TestCheckResourceAttr(resourceName, "name", instanceName),
		resource.TestCheckResourceAttr(resourceName, "host_name", "terraform-acc-myhost"),
		resource.TestCheckResourceAttr(resourceName, "memory", "1073741824"),
		resource.TestCheckResourceAttr(resourceName, "ncpus", "1"),
		resource.TestCheckResourceAttr(resourceName, "start_on_create", "true"),
		resource.TestCheckResourceAttr(resourceName, "external_ips.0.type", "ephemeral"),
		resource.TestCheckResourceAttrSet(resourceName, "disk_attachments.0"),
		resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.description", "a sample nic"),
		resource.TestCheckResourceAttrSet(resourceName, "network_interfaces.0.id"),
		resource.TestCheckResourceAttrSet(resourceName, "network_interfaces.0.ip_address"),
		resource.TestCheckResourceAttrSet(resourceName, "network_interfaces.0.mac_address"),
		resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.name", nicName),
		resource.TestCheckResourceAttrSet(resourceName, "network_interfaces.0.primary"),
		resource.TestCheckResourceAttrSet(resourceName, "network_interfaces.0.subnet_id"),
		resource.TestCheckResourceAttrSet(resourceName, "network_interfaces.0.vpc_id"),
		resource.TestCheckResourceAttrSet(resourceName, "network_interfaces.0.time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "network_interfaces.0.time_modified"),
		resource.TestCheckResourceAttrSet(resourceName, "ssh_public_keys.0"),
		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.read", "1m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.delete", "2m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.create", "3m"),
	}...)
}

func checkResourceInstanceIP(resourceName, instanceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test instance"),
		resource.TestCheckResourceAttr(resourceName, "name", instanceName),
		resource.TestCheckResourceAttr(resourceName, "host_name", "terraform-acc-myhost"),
		resource.TestCheckResourceAttr(resourceName, "memory", "1073741824"),
		resource.TestCheckResourceAttr(resourceName, "ncpus", "1"),
		resource.TestCheckResourceAttr(resourceName, "external_ips.0.type", "ephemeral"),
		resource.TestCheckResourceAttr(resourceName, "start_on_create", "false"),
		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
	}...)
}

func checkResourceInstanceDisk(resourceName, instanceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttrSet(resourceName, "boot_disk_id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test instance"),
		resource.TestCheckResourceAttr(resourceName, "name", instanceName),
		resource.TestCheckResourceAttr(resourceName, "host_name", "terraform-acc-myhost"),
		resource.TestCheckResourceAttr(resourceName, "memory", "1073741824"),
		resource.TestCheckResourceAttr(resourceName, "ncpus", "1"),
		resource.TestCheckResourceAttr(resourceName, "start_on_create", "false"),
		resource.TestCheckResourceAttrSet(resourceName, "disk_attachments.0"),
		resource.TestCheckResourceAttrSet(resourceName, "disk_attachments.1"),
		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
	}...)
}

func checkResourceInstanceDiskUpdate(resourceName, instanceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttrSet(resourceName, "boot_disk_id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test instance"),
		resource.TestCheckResourceAttr(resourceName, "name", instanceName),
		resource.TestCheckResourceAttr(resourceName, "host_name", "terraform-acc-myhost"),
		resource.TestCheckResourceAttr(resourceName, "memory", "1073741824"),
		resource.TestCheckResourceAttr(resourceName, "ncpus", "1"),
		resource.TestCheckResourceAttr(resourceName, "start_on_create", "false"),
		resource.TestCheckResourceAttrSet(resourceName, "disk_attachments.0"),
		resource.TestCheckNoResourceAttr(resourceName, "disk_attachments.1"),
		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
	}...)
}

func checkResourceInstanceNic(resourceName, instanceName, nicName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test instance"),
		resource.TestCheckResourceAttr(resourceName, "name", instanceName),
		resource.TestCheckResourceAttr(resourceName, "host_name", "terraform-acc-myhost"),
		resource.TestCheckResourceAttr(resourceName, "memory", "1073741824"),
		resource.TestCheckResourceAttr(resourceName, "ncpus", "1"),
		resource.TestCheckResourceAttr(resourceName, "start_on_create", "false"),
		resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.description", "a sample nic"),
		resource.TestCheckResourceAttrSet(resourceName, "network_interfaces.0.id"),
		resource.TestCheckResourceAttrSet(resourceName, "network_interfaces.0.ip_address"),
		resource.TestCheckResourceAttrSet(resourceName, "network_interfaces.0.mac_address"),
		resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.name", nicName),
		resource.TestCheckResourceAttrSet(resourceName, "network_interfaces.0.primary"),
		resource.TestCheckResourceAttrSet(resourceName, "network_interfaces.0.subnet_id"),
		resource.TestCheckResourceAttrSet(resourceName, "network_interfaces.0.vpc_id"),
		resource.TestCheckResourceAttrSet(resourceName, "network_interfaces.0.time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "network_interfaces.0.time_modified"),
		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
	}...)
}

func checkResourceInstanceNicUpdate(resourceName, instanceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test instance"),
		resource.TestCheckResourceAttr(resourceName, "name", instanceName),
		resource.TestCheckResourceAttr(resourceName, "host_name", "terraform-acc-myhost"),
		resource.TestCheckResourceAttr(resourceName, "memory", "1073741824"),
		resource.TestCheckResourceAttr(resourceName, "ncpus", "1"),
		resource.TestCheckResourceAttr(resourceName, "start_on_create", "false"),
		resource.TestCheckNoResourceAttr(resourceName, "network_interfaces.0"),
		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
	}...)
}

func checkResourceInstanceSSHKeys(resourceName, instanceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test instance"),
		resource.TestCheckResourceAttr(resourceName, "name", instanceName),
		resource.TestCheckResourceAttr(resourceName, "host_name", "terraform-acc-myhost"),
		resource.TestCheckResourceAttr(resourceName, "memory", "1073741824"),
		resource.TestCheckResourceAttr(resourceName, "ncpus", "1"),
		resource.TestCheckResourceAttr(resourceName, "start_on_create", "false"),
		resource.TestCheckResourceAttrSet(resourceName, "ssh_public_keys.0"),
		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
	}...)
}

func testAccInstanceDestroy(s *terraform.State) error {
	client, err := newTestClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "oxide_instance" {
			continue
		}

		// TODO: check for block name

		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()

		params := oxide.InstanceViewParams{
			Instance: oxide.NameOrId(rs.Primary.Attributes["id"]),
		}
		res, err := client.InstanceView(ctx, params)
		if err != nil && is404(err) {
			continue
		}

		return fmt.Errorf("instance (%v) still exists", &res.Name)
	}

	return nil
}

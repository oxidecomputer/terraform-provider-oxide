// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
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
		BlockName                  string
		InstanceName               string
		DiskBlockName              string
		DiskName                   string
		SSHKeyName                 string
		AntiAffinityGroupName      string
		SupportBlockName           string
		SupportBlockName2          string
		SSHBlockName               string
		AntiAffinityGroupBlockName string
		AutoRestartPolicy          string
		NicName                    string
	}

	resourceInstanceConfigTpl := `
data "oxide_project" "{{.SupportBlockName}}" {
	name = "tf-acc-test"
}

resource "oxide_instance" "{{.BlockName}}" {
  project_id      = data.oxide_project.{{.SupportBlockName}}.id
  description     = "a test instance"
  name            = "{{.InstanceName}}"
  hostname        = "terraform-acc-myhost"
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

resource "oxide_anti_affinity_group" "{{.AntiAffinityGroupBlockName}}" {
  project_id  = data.oxide_project.{{.SupportBlockName}}.id
  description = "a test anti-affinity group"
  name        = "{{.AntiAffinityGroupName}}"
  policy      = "allow"
}

resource "oxide_instance" "{{.BlockName}}" {
  anti_affinity_groups = [oxide_anti_affinity_group.{{.AntiAffinityGroupBlockName}}.id]
  project_id       	   = data.oxide_project.{{.SupportBlockName}}.id
  auto_restart_policy  = "{{.AutoRestartPolicy}}"
  boot_disk_id     	   = oxide_disk.{{.DiskBlockName}}.id
  description      	   = "a test instance"
  name             	   = "{{.InstanceName}}"
  hostname         	   = "terraform-acc-myhost"
  memory           	   = 1073741824
  ncpus            	   = 1
  start_on_create  	   = true
  ssh_public_keys  	   = [oxide_ssh_key.{{.SSHBlockName}}.id]
  disk_attachments 	   = [oxide_disk.{{.DiskBlockName}}.id]
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
	instanceAaGroupName := newResourceName()
	blockName2 := newBlockName("instance")
	diskBlockName := newBlockName("disk")
	supportBlockName3 := newBlockName("support")
	supportBlockName2 := newBlockName("support")
	supportBlockNameSSHKeys := newBlockName("support-instance-ssh-keys")
	supportBlockNameAaGroup := newBlockName("support-instance-anti-affinity-group")
	resourceName2 := fmt.Sprintf("oxide_instance.%s", blockName2)
	autoRestartPolicy := "best_effort"
	config2, err := parsedAccConfig(
		resourceInstanceFullConfig{
			BlockName:                  blockName2,
			InstanceName:               instanceName2,
			DiskBlockName:              diskBlockName,
			DiskName:                   instanceDiskName,
			SSHKeyName:                 instanceSshKeyName,
			SupportBlockName:           supportBlockName3,
			SupportBlockName2:          supportBlockName2,
			NicName:                    instanceNicName,
			SSHBlockName:               supportBlockNameSSHKeys,
			AntiAffinityGroupBlockName: supportBlockNameAaGroup,
			AntiAffinityGroupName:      instanceAaGroupName,
			AutoRestartPolicy:          autoRestartPolicy,
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

// TestAccCloudResourceInstance_extIPs tests whether Terraform can create
// `oxide_instance` resources with the `external_ips` attribute populated. It
// assumes an IP pool named `default` already exists in the silo the test is
// running against. The `OXIDE_TEST_IP_POOL_NAME` environment variable can be
// used to override the IP pool if necessary.
//
// This test is also meant to catch regressions of the following issue.
// https://github.com/oxidecomputer/terraform-provider-oxide/issues/459.
func TestAccCloudResourceInstance_extIPs(t *testing.T) {
	type resourceInstanceConfig struct {
		BlockName        string
		InstanceName     string
		SupportBlockName string
		IPPoolBlockName  string
		IPPoolName       string
	}

	resourceInstanceExternalIPConfigTpl := `
data "oxide_project" "{{.SupportBlockName}}" {
	name = "tf-acc-test"
}

data "oxide_vpc_subnet" "default" {
  project_name = data.oxide_project.{{.SupportBlockName}}.name
  vpc_name     = "default"
  name         = "default"
}

data "oxide_ip_pool" "{{.IPPoolBlockName}}" {
	name = "{{.IPPoolName}}"
}

resource "oxide_instance" "{{.BlockName}}" {
  project_id      = data.oxide_project.{{.SupportBlockName}}.id
  description     = "a test instance"
  name            = "{{.InstanceName}}"
  hostname        = "terraform-acc-myhost"
  memory          = 1073741824
  ncpus           = 1
  start_on_create = false
  external_ips = [
	{
	  type = "ephemeral"
	  id   = data.oxide_ip_pool.{{.IPPoolBlockName}}.id
	}
  ]
  network_interfaces = [
    {
      name        = "net0"
      description = "net0"
      subnet_id   = data.oxide_vpc_subnet.default.id
      vpc_id      = data.oxide_vpc_subnet.default.vpc_id
    }
  ]
}
`

	resourceInstanceExternalIPConfigUpdate1Tpl := `
data "oxide_project" "{{.SupportBlockName}}" {
	name = "tf-acc-test"
}

data "oxide_vpc_subnet" "default" {
  project_name = data.oxide_project.{{.SupportBlockName}}.name
  vpc_name     = "default"
  name         = "default"
}

resource "oxide_instance" "{{.BlockName}}" {
  project_id      = data.oxide_project.{{.SupportBlockName}}.id
  description     = "a test instance"
  name            = "{{.InstanceName}}"
  hostname        = "terraform-acc-myhost"
  memory          = 1073741824
  ncpus           = 1
  start_on_create = false
  external_ips = [
	{
	  type = "ephemeral"
	}
  ]
  network_interfaces = [
    {
      name        = "net0"
      description = "net0"
      subnet_id   = data.oxide_vpc_subnet.default.id
      vpc_id      = data.oxide_vpc_subnet.default.vpc_id
    }
  ]
}
`

	resourceInstanceExternalIPConfigUpdate2Tpl := `
data "oxide_project" "{{.SupportBlockName}}" {
	name = "tf-acc-test"
}

resource "oxide_instance" "{{.BlockName}}" {
  project_id      = data.oxide_project.{{.SupportBlockName}}.id
  description     = "a test instance"
  name            = "{{.InstanceName}}"
  hostname        = "terraform-acc-myhost"
  memory          = 1073741824
  ncpus           = 1
  start_on_create = false
}
`

	// Some test environments may not have an IP pool named `default`. This allows a
	// user to override the IP pool name used for this test.
	ipPoolName, ok := os.LookupEnv("OXIDE_TEST_IP_POOL_NAME")
	if !ok || ipPoolName == "" {
		ipPoolName = "default"
	}

	instanceName := newResourceName()
	blockName := newBlockName("instance")
	supportBlockName := newBlockName("support")
	ipPoolBlockName := newBlockName("ip-pool")
	resourceName := fmt.Sprintf("oxide_instance.%s", blockName)
	initialConfig, err := parsedAccConfig(
		resourceInstanceConfig{
			BlockName:        blockName,
			InstanceName:     instanceName,
			SupportBlockName: supportBlockName,
			IPPoolBlockName:  ipPoolBlockName,
			IPPoolName:       ipPoolName,
		},
		resourceInstanceExternalIPConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing initial config template data: %e", err)
	}

	updateConfig1, err := parsedAccConfig(
		resourceInstanceConfig{
			BlockName:        blockName,
			InstanceName:     instanceName,
			SupportBlockName: supportBlockName,
		},
		resourceInstanceExternalIPConfigUpdate1Tpl,
	)
	if err != nil {
		t.Errorf("error parsing first update config template data: %e", err)
	}

	updateConfig2, err := parsedAccConfig(
		resourceInstanceConfig{
			BlockName:        blockName,
			InstanceName:     instanceName,
			SupportBlockName: supportBlockName,
		},
		resourceInstanceExternalIPConfigUpdate2Tpl,
	)
	if err != nil {
		t.Errorf("error parsing second update config template data: %e", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccInstanceDestroy,
		Steps: []resource.TestStep{
			// Ephemeral external IP with specified IP pool ID.
			{
				Config: initialConfig,
				Check:  checkResourceInstanceIP(resourceName, instanceName),
			},
			// Ephemeral external IP with default silo IP pool ID.
			{
				Config: updateConfig1,
				Check:  checkResourceInstanceIPUpdate1(resourceName, instanceName),
			},
			// Ephemeral external IP with specified IP pool ID.
			{
				Config: initialConfig,
				Check:  checkResourceInstanceIP(resourceName, instanceName),
			},
			// Detach all external IPs.
			{
				Config: updateConfig2,
				Check:  checkResourceInstanceIPUpdate2(resourceName, instanceName),
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
  hostname        = "terraform-acc-myhost"
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
  hostname         = "terraform-acc-myhost"
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
  hostname         = "terraform-acc-myhost"
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
  hostname        = "terraform-acc-myhost"
  memory          = 1073741824
  ncpus           = 1
  start_on_create = true
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
  hostname        = "terraform-acc-myhost"
  memory          = 1073741824
  ncpus           = 1
  start_on_create = true
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

func TestAccCloudResourceInstance_update(t *testing.T) {
	type resourceInstanceUpdateConfig struct {
		BlockName        string
		InstanceName     string
		SupportBlockName string
	}

	resourceInstanceConfigTpl := `
data "oxide_project" "{{.SupportBlockName}}" {
	name = "tf-acc-test"
}

resource "oxide_instance" "{{.BlockName}}" {
  project_id      = data.oxide_project.{{.SupportBlockName}}.id
  description     = "a test instance"
  name            = "{{.InstanceName}}"
  hostname        = "terraform-acc-myhost"
  memory          = 1073741824
  ncpus           = 1
  start_on_create = true
}
`

	resourceInstanceConfigUpdateTpl := `
data "oxide_project" "{{.SupportBlockName}}" {
	name = "tf-acc-test"
}

resource "oxide_instance" "{{.BlockName}}" {
  project_id      = data.oxide_project.{{.SupportBlockName}}.id
  description     = "a test instance"
  name            = "{{.InstanceName}}"
  hostname        = "terraform-acc-myhost"
  memory          = 1073741824
  ncpus           = 2
  start_on_create = true
}
`

	resourceInstanceConfigUpdate2Tpl := `
data "oxide_project" "{{.SupportBlockName}}" {
	name = "tf-acc-test"
}

resource "oxide_instance" "{{.BlockName}}" {
  project_id      = data.oxide_project.{{.SupportBlockName}}.id
  description     = "a test instance"
  name            = "{{.InstanceName}}"
  hostname        = "terraform-acc-myhost"
  memory          = 2147483648
  ncpus           = 2
  start_on_create = true
}
`
	instanceName := newResourceName()
	supportBlockName := newBlockName("support")
	supportBlockName2 := newBlockName("support-update")
	blockNameInstance := newBlockName("instance")
	resourceNameInstance := fmt.Sprintf("oxide_instance.%s", blockNameInstance)
	config, err := parsedAccConfig(
		resourceInstanceUpdateConfig{
			BlockName:        blockNameInstance,
			InstanceName:     instanceName,
			SupportBlockName: supportBlockName,
		},
		resourceInstanceConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	configUpdate, err := parsedAccConfig(
		resourceInstanceUpdateConfig{
			BlockName:        blockNameInstance,
			InstanceName:     instanceName,
			SupportBlockName: supportBlockName2,
		},
		resourceInstanceConfigUpdateTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	configUpdate2, err := parsedAccConfig(
		resourceInstanceUpdateConfig{
			BlockName:        blockNameInstance,
			InstanceName:     instanceName,
			SupportBlockName: supportBlockName2,
		},
		resourceInstanceConfigUpdate2Tpl,
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
				Check:  checkResourceInstanceUpdate(resourceNameInstance, instanceName),
			},
			{
				// Update NCPUs
				Config: configUpdate,
				Check:  checkResourceInstanceUpdate2(resourceNameInstance, instanceName),
			},
			{
				// Update memory
				Config: configUpdate2,
				Check:  checkResourceInstanceUpdate3(resourceNameInstance, instanceName),
			},
			{
				// Update all
				Config: config,
				Check:  checkResourceInstanceUpdate(resourceNameInstance, instanceName),
			},
			{
				ResourceName:      resourceNameInstance,
				ImportState:       true,
				ImportStateVerify: true,
				// This option is only relevant for create, this means that it will
				// never be imported
				ImportStateVerifyIgnore: []string{"start_on_create"},
			},
		},
	})
}

func TestAccCloudResourceInstance_no_boot_disk(t *testing.T) {
	type resourceInstanceNoBootDiskConfig struct {
		BlockName        string
		InstanceName     string
		DiskBlockName    string
		DiskName         string
		SupportBlockName string
	}

	resourceInstanceNoBootDiskConfigTpl := `
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

resource "oxide_instance" "{{.BlockName}}" {
  project_id      = data.oxide_project.{{.SupportBlockName}}.id
  description     = "a test instance"
  name            = "{{.InstanceName}}"
  hostname        = "terraform-acc-myhost"
  memory          = 1073741824
  ncpus           = 1
  start_on_create = false
  disk_attachments = [oxide_disk.{{.DiskBlockName}}.id]
}
`

	instanceName := newResourceName()
	diskName := newResourceName()
	blockName := newBlockName("instance-no-boot-disk")
	diskBlockName := newBlockName("disk")
	supportBlockName := newBlockName("support")
	resourceName := fmt.Sprintf("oxide_instance.%s", blockName)
	config, err := parsedAccConfig(
		resourceInstanceNoBootDiskConfig{
			BlockName:        blockName,
			InstanceName:     instanceName,
			DiskBlockName:    diskBlockName,
			DiskName:         diskName,
			SupportBlockName: supportBlockName,
		},
		resourceInstanceNoBootDiskConfigTpl,
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
				Check:  checkResourceInstanceNoBootDisk(resourceName, instanceName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// This option is only relevant for create, this means that it will
				// never be imported
				ImportStateVerifyIgnore: []string{"start_on_create"},
			},
		},
	})
}

func TestAccCloudResourceInstance_antiAffinityGroups(t *testing.T) {
	type resourceInstanceAntiAffinityGroupsConfig struct {
		BlockName                   string
		InstanceName                string
		AntiAffinityGroupName       string
		AntiAffinityGroupName2      string
		SupportBlockName            string
		AntiAffinityGroupBlockName  string
		AntiAffinityGroupBlockName2 string
	}

	resourceInstanceAntiAffinityGroupsConfigTpl := `
data "oxide_project" "{{.SupportBlockName}}" {
	name = "tf-acc-test"
}

resource "oxide_anti_affinity_group" "{{.AntiAffinityGroupBlockName}}" {
  project_id  = data.oxide_project.{{.SupportBlockName}}.id
  description = "a test anti-affinity group"
  name        = "{{.AntiAffinityGroupName}}"
  policy      = "allow"
}

resource "oxide_instance" "{{.BlockName}}" {
  anti_affinity_groups = [oxide_anti_affinity_group.{{.AntiAffinityGroupBlockName}}.id]
  project_id      	   = data.oxide_project.{{.SupportBlockName}}.id
  description     	   = "a test instance"
  name            	   = "{{.InstanceName}}"
  hostname        	   = "terraform-acc-myhost"
  memory          	   = 1073741824
  ncpus           	   = 1
  start_on_create 	   = false
}
`

	resourceInstanceAntiAffinityGroupsConfigTplUpdate := `
data "oxide_project" "{{.SupportBlockName}}" {
	name = "tf-acc-test"
}

resource "oxide_anti_affinity_group" "{{.AntiAffinityGroupBlockName}}" {
  project_id  = data.oxide_project.{{.SupportBlockName}}.id
  description = "a test anti-affinity group"
  name        = "{{.AntiAffinityGroupName}}"
  policy      = "allow"
}

resource "oxide_anti_affinity_group" "{{.AntiAffinityGroupBlockName2}}" {
  project_id  = data.oxide_project.{{.SupportBlockName}}.id
  description = "a test anti-affinity group"
  name        = "{{.AntiAffinityGroupName2}}"
  policy      = "allow"
}

resource "oxide_instance" "{{.BlockName}}" {
  anti_affinity_groups = [oxide_anti_affinity_group.{{.AntiAffinityGroupBlockName}}.id, oxide_anti_affinity_group.{{.AntiAffinityGroupBlockName2}}.id]
  project_id      	   = data.oxide_project.{{.SupportBlockName}}.id
  description     	   = "a test instance"
  name            	   = "{{.InstanceName}}"
  hostname        	   = "terraform-acc-myhost"
  memory          	   = 1073741824
  ncpus           	   = 1
  start_on_create 	   = false
}
`

	resourceInstanceAntiAffinityGroupsConfigTplUpdate2 := `
data "oxide_project" "{{.SupportBlockName}}" {
	name = "tf-acc-test"
}

resource "oxide_anti_affinity_group" "{{.AntiAffinityGroupBlockName}}" {
  project_id  = data.oxide_project.{{.SupportBlockName}}.id
  description = "a test anti-affinity group"
  name        = "{{.AntiAffinityGroupName}}"
  policy      = "allow"
}

resource "oxide_anti_affinity_group" "{{.AntiAffinityGroupBlockName2}}" {
  project_id  = data.oxide_project.{{.SupportBlockName}}.id
  description = "a test anti-affinity group"
  name        = "{{.AntiAffinityGroupName2}}"
  policy      = "allow"
}

resource "oxide_instance" "{{.BlockName}}" {
  anti_affinity_groups = [oxide_anti_affinity_group.{{.AntiAffinityGroupBlockName}}.id]
  project_id      	   = data.oxide_project.{{.SupportBlockName}}.id
  description     	   = "a test instance"
  name            	   = "{{.InstanceName}}"
  hostname        	   = "terraform-acc-myhost"
  memory          	   = 1073741824
  ncpus           	   = 1
  start_on_create 	   = false
}
`

	instanceAntiAffinityGroupsName := newResourceName()
	antiAffinityGroupName1 := newResourceName()
	antiAffinityGroupName2 := newResourceName()
	blockNameAntiAffinityGroups := newBlockName("instance-anti-affinity-groups")
	supportBlockNameAntiAffinityGroups := newBlockName("support-instance-anti-affinity-groups")
	supportBlockNameAntiAffinityGroup1 := newBlockName("support-instance-anti-affinity-group")
	supportBlockNameAntiAffinityGroup2 := newBlockName("support-instance-anti-affinity-group")
	resourceName := fmt.Sprintf("oxide_instance.%s", blockNameAntiAffinityGroups)
	configAntiAffinityGroups, err := parsedAccConfig(
		resourceInstanceAntiAffinityGroupsConfig{
			BlockName:                  blockNameAntiAffinityGroups,
			InstanceName:               instanceAntiAffinityGroupsName,
			SupportBlockName:           supportBlockNameAntiAffinityGroups,
			AntiAffinityGroupName:      antiAffinityGroupName1,
			AntiAffinityGroupBlockName: supportBlockNameAntiAffinityGroup1,
		},
		resourceInstanceAntiAffinityGroupsConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	configAntiAffinityGroupsUpdate, err := parsedAccConfig(
		resourceInstanceAntiAffinityGroupsConfig{
			BlockName:                   blockNameAntiAffinityGroups,
			InstanceName:                instanceAntiAffinityGroupsName,
			SupportBlockName:            supportBlockNameAntiAffinityGroups,
			AntiAffinityGroupName:       antiAffinityGroupName1,
			AntiAffinityGroupBlockName:  supportBlockNameAntiAffinityGroup1,
			AntiAffinityGroupName2:      antiAffinityGroupName2,
			AntiAffinityGroupBlockName2: supportBlockNameAntiAffinityGroup2,
		},
		resourceInstanceAntiAffinityGroupsConfigTplUpdate,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	configAntiAffinityGroupsUpdate2, err := parsedAccConfig(
		resourceInstanceAntiAffinityGroupsConfig{
			BlockName:                   blockNameAntiAffinityGroups,
			InstanceName:                instanceAntiAffinityGroupsName,
			SupportBlockName:            supportBlockNameAntiAffinityGroups,
			AntiAffinityGroupName:       antiAffinityGroupName1,
			AntiAffinityGroupBlockName:  supportBlockNameAntiAffinityGroup1,
			AntiAffinityGroupName2:      antiAffinityGroupName2,
			AntiAffinityGroupBlockName2: supportBlockNameAntiAffinityGroup2,
		},
		resourceInstanceAntiAffinityGroupsConfigTplUpdate2,
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
				Config: configAntiAffinityGroups,
				Check:  checkResourceInstanceAntiAffinityGroups(resourceName, instanceAntiAffinityGroupsName),
			},
			// Add another anti-affinity group
			{
				Config: configAntiAffinityGroupsUpdate,
				Check:  checkResourceInstanceAntiAffinityGroupsUpdate(resourceName, instanceAntiAffinityGroupsName),
			},
			// Remove an anti-affinity group
			{
				Config: configAntiAffinityGroupsUpdate2,
				Check:  checkResourceInstanceAntiAffinityGroups(resourceName, instanceAntiAffinityGroupsName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// start-on-create cannot be imported as it is only present at create time
				ImportStateVerifyIgnore: []string{"start_on_create"},
			},
		},
	})
}

func TestAccCloudResourceInstance_host_nameDeprecation(t *testing.T) {
	resourceName := "oxide_instance.test_instance"

	generateConfig := func(t *testing.T, name string, hostnames map[string]string) string {
		tmplData := struct {
			InstanceName      string
			InstanceHostnames map[string]string
		}{
			InstanceName:      name,
			InstanceHostnames: hostnames,
		}

		config, err := parsedAccConfig(tmplData, `
data "oxide_project" "tf_acc_test" {
	name = "tf-acc-test"
}

resource "oxide_instance" "test_instance" {
  project_id      = data.oxide_project.tf_acc_test.id
  description     = "a test instance"
  name            = "{{.InstanceName}}"
{{range $attr, $val := .InstanceHostnames}}
  {{$attr}}        = "{{$val}}"
{{end}}
  memory          = 1073741824
  ncpus           = 1
  start_on_create = false
}
`)
		if err != nil {
			t.Errorf("error parsing config template data: %e", err)
			return ""
		}

		return config
	}

	// Test no changes to remote state when updating the provider and changing
	// attribute name from host_name to hostname.
	t.Run("host_name migration", func(t *testing.T) {
		instanceName := newResourceName()

		resource.ParallelTest(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			CheckDestroy: testAccInstanceDestroy,
			Steps: []resource.TestStep{
				// Initial state using host_name.
				{
					ExternalProviders: map[string]resource.ExternalProvider{
						"oxide": {
							Source:            "oxidecomputer/oxide",
							VersionConstraint: "0.17.0",
						},
					},
					Config: generateConfig(t, instanceName, map[string]string{"host_name": "terraform-acc-myhost"}),
					Check:  resource.TestCheckResourceAttr(resourceName, "host_name", "terraform-acc-myhost"),
				},
				// Update provider without modifying config.
				// Expect no-op.
				{
					ExternalProviders:        map[string]resource.ExternalProvider{},
					ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
					Config:                   generateConfig(t, instanceName, map[string]string{"host_name": "terraform-acc-myhost"}),
					ConfigPlanChecks: resource.ConfigPlanChecks{
						PreApply: []plancheck.PlanCheck{
							plancheck.ExpectEmptyPlan(),
						},
					},
					Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
						resource.TestCheckResourceAttr(resourceName, "host_name", "terraform-acc-myhost"),
						resource.TestCheckResourceAttr(resourceName, "hostname", "terraform-acc-myhost"),
					}...),
				},
				// Update host_name to hostname.
				// Expect no-op.
				{
					ExternalProviders:        map[string]resource.ExternalProvider{},
					ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
					Config:                   generateConfig(t, instanceName, map[string]string{"hostname": "terraform-acc-myhost"}),
					ConfigPlanChecks: resource.ConfigPlanChecks{
						PreApply: []plancheck.PlanCheck{
							plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
						},
					},
					Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
						resource.TestCheckResourceAttr(resourceName, "host_name", "terraform-acc-myhost"),
						resource.TestCheckResourceAttr(resourceName, "hostname", "terraform-acc-myhost"),
					}...),
				},
			},
		})
	})

	// Test no changes to remote state when changing attribute name from
	// hostname to host_name.
	t.Run("hostname rename", func(t *testing.T) {
		instanceName := newResourceName()

		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
			CheckDestroy:             testAccInstanceDestroy,
			Steps: []resource.TestStep{
				// Initial state using hostname.
				{
					Config: generateConfig(t, instanceName, map[string]string{"hostname": "terraform-acc-myhost"}),
					Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
						resource.TestCheckResourceAttr(resourceName, "hostname", "terraform-acc-myhost"),
						resource.TestCheckResourceAttr(resourceName, "host_name", "terraform-acc-myhost"),
					}...),
				},
				// Update hostname to host_name.
				// Expect no-op.
				{
					Config: generateConfig(t, instanceName, map[string]string{"host_name": "terraform-acc-myhost"}),
					ConfigPlanChecks: resource.ConfigPlanChecks{
						PreApply: []plancheck.PlanCheck{
							plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
						},
					},
					Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
						resource.TestCheckResourceAttr(resourceName, "hostname", "terraform-acc-myhost"),
						resource.TestCheckResourceAttr(resourceName, "host_name", "terraform-acc-myhost"),
					}...),
				},
			},
		})
	})

	// Test instance hostname value modification and value modification with
	// attribute rename.
	testCases := []struct {
		from string
		to   string
	}{
		{from: "host_name", to: "hostname"},
		{from: "hostname", to: "host_name"},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("from %s to %s", tc.from, tc.to), func(t *testing.T) {
			instanceName := newResourceName()

			resource.ParallelTest(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
				CheckDestroy:             testAccInstanceDestroy,
				Steps: []resource.TestStep{
					// Initial state.
					{
						Config: generateConfig(t, instanceName, map[string]string{tc.from: "terraform-acc-myhost"}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttr(resourceName, tc.from, "terraform-acc-myhost"),
							resource.TestCheckResourceAttr(resourceName, tc.to, "terraform-acc-myhost"),
						}...),
					},
					// Update hostname value.
					// Expect a resource replacement.
					{
						Config: generateConfig(t, instanceName, map[string]string{tc.from: "terraform-acc-myhost-updated"}),
						ConfigPlanChecks: resource.ConfigPlanChecks{
							PreApply: []plancheck.PlanCheck{
								plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
							},
						},
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttr(resourceName, tc.from, "terraform-acc-myhost-updated"),
							resource.TestCheckResourceAttr(resourceName, tc.to, "terraform-acc-myhost-updated"),
						}...),
					},
					// Update hostname attribute name and value.
					// Expect a resource replacement.
					{
						Config: generateConfig(t, instanceName, map[string]string{tc.to: "terraform-acc-myhost"}),
						ConfigPlanChecks: resource.ConfigPlanChecks{
							PreApply: []plancheck.PlanCheck{
								plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
							},
						},
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttr(resourceName, tc.to, "terraform-acc-myhost"),
							resource.TestCheckResourceAttr(resourceName, tc.from, "terraform-acc-myhost"),
						}...),
					},
				},
			})
		})
	}

	// Test resource import.
	t.Run("import with hostname", func(t *testing.T) {
		instanceName := newResourceName()

		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
			CheckDestroy:             testAccInstanceDestroy,
			Steps: []resource.TestStep{
				{
					Config: generateConfig(t, instanceName, map[string]string{
						"hostname": "terraform-acc-myhost",
					}),
					Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
						resource.TestCheckResourceAttr(resourceName, "hostname", "terraform-acc-myhost"),
						resource.TestCheckResourceAttr(resourceName, "host_name", "terraform-acc-myhost"),
					}...),
				},
				{
					ImportState:             true,
					ResourceName:            resourceName,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"start_on_create"},
				},
			},
		})
	})

	t.Run("import with host_name", func(t *testing.T) {
		instanceName := newResourceName()

		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
			CheckDestroy:             testAccInstanceDestroy,
			Steps: []resource.TestStep{
				{
					Config: generateConfig(t, instanceName, map[string]string{
						"host_name": "terraform-acc-myhost",
					}),
					Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
						resource.TestCheckResourceAttr(resourceName, "host_name", "terraform-acc-myhost"),
						resource.TestCheckResourceAttr(resourceName, "hostname", "terraform-acc-myhost"),
					}...),
				},
				{
					ImportState:             true,
					ResourceName:            resourceName,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"start_on_create"},
				},
			},
		})
	})

	// Test that either hostname or host_name should be provided.
	t.Run("missing hostname", func(t *testing.T) {
		instanceName := newResourceName()

		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
			CheckDestroy:             testAccInstanceDestroy,
			Steps: []resource.TestStep{
				{
					Config:      generateConfig(t, instanceName, map[string]string{}),
					ExpectError: regexp.MustCompile(`one \(and only one\) of \[hostname\] is required`),
				},
			},
		})
	})

	// Test that only host_name or hostname can be set.
	t.Run("host_name and hostname not allowed together", func(t *testing.T) {
		instanceName := newResourceName()

		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
			CheckDestroy:             testAccInstanceDestroy,
			Steps: []resource.TestStep{
				{
					Config: generateConfig(t, instanceName, map[string]string{
						"hostname":  "terraform-acc-myhost",
						"host_name": "terraform-acc-myhost",
					}),
					ExpectError: regexp.MustCompile(`one \(and only one\) of \[hostname\] is required`),
				},
			},
		})
	})
}

func checkResourceInstance(resourceName, instanceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test instance"),
		resource.TestCheckResourceAttr(resourceName, "name", instanceName),
		resource.TestCheckResourceAttr(resourceName, "hostname", "terraform-acc-myhost"),
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
		resource.TestCheckResourceAttr(resourceName, "hostname", "terraform-acc-myhost"),
		resource.TestCheckResourceAttr(resourceName, "memory", "1073741824"),
		resource.TestCheckResourceAttr(resourceName, "ncpus", "1"),
		resource.TestCheckResourceAttr(resourceName, "start_on_create", "true"),
		resource.TestCheckResourceAttr(resourceName, "external_ips.0.type", "ephemeral"),
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
		resource.TestCheckResourceAttrSet(resourceName, "auto_restart_policy"),
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
		resource.TestCheckResourceAttr(resourceName, "hostname", "terraform-acc-myhost"),
		resource.TestCheckResourceAttr(resourceName, "memory", "1073741824"),
		resource.TestCheckResourceAttr(resourceName, "ncpus", "1"),
		resource.TestCheckResourceAttr(resourceName, "external_ips.0.type", "ephemeral"),
		resource.TestCheckResourceAttrSet(resourceName, "external_ips.0.id"),
		resource.TestCheckResourceAttr(resourceName, "start_on_create", "false"),
		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
	}...)
}

func checkResourceInstanceIPUpdate1(resourceName, instanceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test instance"),
		resource.TestCheckResourceAttr(resourceName, "name", instanceName),
		resource.TestCheckResourceAttr(resourceName, "hostname", "terraform-acc-myhost"),
		resource.TestCheckResourceAttr(resourceName, "memory", "1073741824"),
		resource.TestCheckResourceAttr(resourceName, "ncpus", "1"),
		resource.TestCheckResourceAttr(resourceName, "start_on_create", "false"),
		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		resource.TestCheckResourceAttr(resourceName, "external_ips.0.type", "ephemeral"),
		resource.TestCheckResourceAttr(resourceName, "external_ips.0.id", ""),
	}...)
}

func checkResourceInstanceIPUpdate2(resourceName, instanceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test instance"),
		resource.TestCheckResourceAttr(resourceName, "name", instanceName),
		resource.TestCheckResourceAttr(resourceName, "hostname", "terraform-acc-myhost"),
		resource.TestCheckResourceAttr(resourceName, "memory", "1073741824"),
		resource.TestCheckResourceAttr(resourceName, "ncpus", "1"),
		resource.TestCheckResourceAttr(resourceName, "start_on_create", "false"),
		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		resource.TestCheckNoResourceAttr(resourceName, "external_ips"),
	}...)
}

func checkResourceInstanceDisk(resourceName, instanceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttrSet(resourceName, "boot_disk_id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test instance"),
		resource.TestCheckResourceAttr(resourceName, "name", instanceName),
		resource.TestCheckResourceAttr(resourceName, "hostname", "terraform-acc-myhost"),
		resource.TestCheckResourceAttr(resourceName, "memory", "1073741824"),
		resource.TestCheckResourceAttr(resourceName, "ncpus", "1"),
		resource.TestCheckResourceAttr(resourceName, "start_on_create", "true"),
		resource.TestCheckResourceAttrSet(resourceName, "disk_attachments.0"),
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
		resource.TestCheckResourceAttr(resourceName, "hostname", "terraform-acc-myhost"),
		resource.TestCheckResourceAttr(resourceName, "memory", "1073741824"),
		resource.TestCheckResourceAttr(resourceName, "ncpus", "1"),
		resource.TestCheckResourceAttr(resourceName, "start_on_create", "true"),
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
		resource.TestCheckResourceAttr(resourceName, "hostname", "terraform-acc-myhost"),
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
		resource.TestCheckResourceAttr(resourceName, "hostname", "terraform-acc-myhost"),
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
		resource.TestCheckResourceAttr(resourceName, "hostname", "terraform-acc-myhost"),
		resource.TestCheckResourceAttr(resourceName, "memory", "1073741824"),
		resource.TestCheckResourceAttr(resourceName, "ncpus", "1"),
		resource.TestCheckResourceAttr(resourceName, "start_on_create", "false"),
		resource.TestCheckResourceAttrSet(resourceName, "ssh_public_keys.0"),
		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
	}...)
}

func checkResourceInstanceUpdate(resourceName, instanceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test instance"),
		resource.TestCheckResourceAttr(resourceName, "name", instanceName),
		resource.TestCheckResourceAttr(resourceName, "hostname", "terraform-acc-myhost"),
		resource.TestCheckResourceAttr(resourceName, "memory", "1073741824"),
		resource.TestCheckResourceAttr(resourceName, "ncpus", "1"),
		resource.TestCheckResourceAttr(resourceName, "start_on_create", "true"),
		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
	}...)
}

func checkResourceInstanceUpdate2(resourceName, instanceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test instance"),
		resource.TestCheckResourceAttr(resourceName, "name", instanceName),
		resource.TestCheckResourceAttr(resourceName, "hostname", "terraform-acc-myhost"),
		resource.TestCheckResourceAttr(resourceName, "memory", "1073741824"),
		resource.TestCheckResourceAttr(resourceName, "ncpus", "2"),
		resource.TestCheckResourceAttr(resourceName, "start_on_create", "true"),
		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
	}...)
}

func checkResourceInstanceUpdate3(resourceName, instanceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test instance"),
		resource.TestCheckResourceAttr(resourceName, "name", instanceName),
		resource.TestCheckResourceAttr(resourceName, "hostname", "terraform-acc-myhost"),
		resource.TestCheckResourceAttr(resourceName, "memory", "2147483648"),
		resource.TestCheckResourceAttr(resourceName, "ncpus", "2"),
		resource.TestCheckResourceAttr(resourceName, "start_on_create", "true"),
		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
	}...)
}

func checkResourceInstanceAntiAffinityGroups(resourceName, instanceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test instance"),
		resource.TestCheckResourceAttr(resourceName, "name", instanceName),
		resource.TestCheckResourceAttr(resourceName, "hostname", "terraform-acc-myhost"),
		resource.TestCheckResourceAttr(resourceName, "memory", "1073741824"),
		resource.TestCheckResourceAttr(resourceName, "ncpus", "1"),
		resource.TestCheckResourceAttr(resourceName, "start_on_create", "false"),
		resource.TestCheckResourceAttrSet(resourceName, "anti_affinity_groups.0"),
		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
	}...)
}

func checkResourceInstanceAntiAffinityGroupsUpdate(resourceName, instanceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test instance"),
		resource.TestCheckResourceAttr(resourceName, "name", instanceName),
		resource.TestCheckResourceAttr(resourceName, "hostname", "terraform-acc-myhost"),
		resource.TestCheckResourceAttr(resourceName, "memory", "1073741824"),
		resource.TestCheckResourceAttr(resourceName, "ncpus", "1"),
		resource.TestCheckResourceAttr(resourceName, "start_on_create", "false"),
		resource.TestCheckResourceAttrSet(resourceName, "anti_affinity_groups.0"),
		resource.TestCheckResourceAttrSet(resourceName, "anti_affinity_groups.1"),
		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
	}...)
}

func checkResourceInstanceNoBootDisk(resourceName, instanceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test instance"),
		resource.TestCheckResourceAttr(resourceName, "name", instanceName),
		resource.TestCheckResourceAttr(resourceName, "hostname", "terraform-acc-myhost"),
		resource.TestCheckResourceAttr(resourceName, "memory", "1073741824"),
		resource.TestCheckResourceAttr(resourceName, "ncpus", "1"),
		resource.TestCheckResourceAttr(resourceName, "start_on_create", "false"),
		resource.TestCheckNoResourceAttr(resourceName, "boot_disk_id"),
		resource.TestCheckResourceAttrSet(resourceName, "disk_attachments.0"),
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

func TestFilterBootDiskFromDisks(t *testing.T) {
	boot_disk := oxide.InstanceDiskAttachment{
		Name: "testboot01",
		Type: oxide.InstanceDiskAttachmentTypeAttach,
	}

	tests := []struct {
		boot_disk oxide.InstanceDiskAttachment
		disks     []oxide.InstanceDiskAttachment
		want      []oxide.InstanceDiskAttachment
	}{
		{
			boot_disk: boot_disk,
			disks: []oxide.InstanceDiskAttachment{
				boot_disk,
				{
					Name: "testdisk01",
					Type: oxide.InstanceDiskAttachmentTypeAttach,
				},
				{
					Name: "testdisk01",
					Type: oxide.InstanceDiskAttachmentTypeAttach,
				},
			},
			want: []oxide.InstanceDiskAttachment{
				{
					Name: "testdisk01",
					Type: oxide.InstanceDiskAttachmentTypeAttach,
				},
				{
					Name: "testdisk01",
					Type: oxide.InstanceDiskAttachmentTypeAttach,
				},
			},
		},
		{
			boot_disk: oxide.InstanceDiskAttachment{},
			disks: []oxide.InstanceDiskAttachment{
				{
					Name: "testdisk01",
					Type: oxide.InstanceDiskAttachmentTypeAttach,
				},
				{
					Name: "testdisk01",
					Type: oxide.InstanceDiskAttachmentTypeAttach,
				},
				{
					Name: "testboot01",
					Type: oxide.InstanceDiskAttachmentTypeAttach,
				},
			},
			want: []oxide.InstanceDiskAttachment{
				{
					Name: "testdisk01",
					Type: oxide.InstanceDiskAttachmentTypeAttach,
				},
				{
					Name: "testdisk01",
					Type: oxide.InstanceDiskAttachmentTypeAttach,
				},
				{
					Name: "testboot01",
					Type: oxide.InstanceDiskAttachmentTypeAttach,
				},
			},
		},
		{
			boot_disk: oxide.InstanceDiskAttachment{},
			disks:     []oxide.InstanceDiskAttachment{},
			want:      []oxide.InstanceDiskAttachment{},
		},
	}
	for _, tt := range tests {
		disks := filterBootDiskFromDisks(tt.disks, &tt.boot_disk)
		if !reflect.DeepEqual(disks, tt.want) {
			t.Errorf("want: %+v, got: %+v", tt.want, disks)
		}
	}
}

---
page_title: "oxide_instance Resource - terraform-provider-oxide"
---

# oxide_instance (Resource)

This resource manages instances.

!> This resource currently only provides create, read and delete actions. Once update endpoints have been added to the API, that functionality will be added here as well.

## Example Usage

### Basic instance

```hcl
resource "oxide_instance" "example" {
  organization_name = "staff"
  project_name      = "test"
  description       = "a test instance"
  name              = "myinstance"
  host_name         = "<host value>"
  memory            = 1073741824
  ncpus             = 1
}
```

### Assign an IP pool for the instance

```hcl
resource "oxide_instance" "example" {
  organization_name = "staff"
  project_name      = "test"
  description       = "a test instance"
  name              = "myinstance"
  host_name         = "<host value>"
  memory            = 1073741824
  ncpus             = 1
  external_ips      = ["myippool"]
}
```

### Attach two disks to the instance

```hcl
resource "oxide_instance" "example" {
  organization_name = "staff"
  project_name      = "test"
  description       = "a test instance"
  name              = "myinstance"
  host_name         = "<host value>"
  memory            = 1073741824
  ncpus             = 1
  attach_to_disks   = ["disk1", "disk2"]
}
```

### Attach a network interface to the instance

```hcl
resource "oxide_instance" "example" {
  organization_name = "staff"
  project_name      = "test"
  description       = "a test instance"
  name              = "myinstance"
  host_name         = "<host value>"
  memory            = 1073741824
  ncpus             = 1
  network_interface {
    description = "a network interface"
    name        = "mynetworkinterface"
    subnet_name = "default"
    vpc_name    = "default"
  }
}
```

## Schema

### Required

- `description` (String) Description for the instance.
- `host_name` (String) Host name of the instance.
- `memory` (Number) Instance memory in bytes.
- `name` (String) Name of the instance.
- `ncpus` (Number) Number of CPUs allocated for this instance.
- `organization_name` (String) Name of the organization.
- `project_name` (String) Name of the project.

### Optional

- `attach_to_disks` (List of String, Optional) Disks to be attached to this instance.
- `external_ips` (List of String, Optional) External IP addresses provided to this instance. List of IP pools from which to draw addresses.
- `network_interface` (List of Object, Optional) Attaches network interfaces to an instance at the time the instance is created. (see [below for nested schema](#nestedblock--network_interface))
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String) Unique, immutable, system-controlled identifier of the instance.
- `project_id` (String) Unique, immutable, system-controlled identifier of the project.
- `run_state` (String) Running state of an Instance (primarily: booted or stopped). This typically reflects whether it's starting, running, stopping, or stopped, but also includes states related to the instance's lifecycle.
- `time_created` (String) Timestamp of when this instance was created.
- `time_modified` (String) Timestamp of when this instance last modified.
- `time_run_state_updated` (String) Timestamp of when the run state of this instance was last modified.

<a id="nestedblock--network_interface"></a>

### Nested Schema for `network_interface`

Required:

- `description` (String) Description for the network interface.
- `name` (String) Name of the network interface.
- `subnet_name` (String) Name of the VPC Subnet in which to create the network interface.
- `vpc_name` (String) Name of the VPC in which to create the network interface.

Read-Only:

- `ip` (String) IP address for the network interface.
- `subnet_id` (String) ID of the VPC Subnet to which the interface belongs.
- `vpc_id` (String) ID of the VPC in which to which the interface belongs.

<a id="nestedblock--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `default` (String)

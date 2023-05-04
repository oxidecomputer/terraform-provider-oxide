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
  project_id  = "c1dee930-a8e4-11ed-afa1-0242ac120002"
  description = "a test instance"
  name        = "myinstance"
  host_name   = "<host value>"
  memory      = 1073741824
  ncpus       = 1
  timeouts = {
    read   = "1m"
    create = "3m"
    delete = "2m"
  }
}
```

### Assign an IP pool for the instance and do not start instance on creation

```hcl
resource "oxide_instance" "example" {
  project_id      = "c1dee930-a8e4-11ed-afa1-0242ac120002"
  description     = "a test instance"
  name            = "myinstance"
  host_name       = "<host value>"
  memory          = 1073741824
  ncpus           = 1
  start_on_create = false
  external_ips    = ["myippool"]
}
```

### Attach two disks to the instance and define user data

```hcl
resource "oxide_instance" "example" {
  project_id      = "c1dee930-a8e4-11ed-afa1-0242ac120002"
  description     = "a test instance"
  name            = "myinstance"
  host_name       = "<host value>"
  memory          = 1073741824
  ncpus           = 1
  user_data       = filebase64("path/to/init.sh")
  attach_to_disks = ["disk1", "disk2"]
}
```

## Schema

### Required

- `description` (String) Description for the instance.
- `host_name` (String) Host name of the instance.
- `memory` (Number) Instance memory in bytes.
- `name` (String) Name of the instance.
- `ncpus` (Number) Number of CPUs allocated for this instance.
- `project_id` (String) ID of the project that will contain the instance.
- `start_on_create` (Boolean, Default `true`) Starts the instance on creation when set to true.

### Optional

- `attach_to_disks` (List of String, Optional) Disks to be attached to this instance.
- `external_ips` (List of String, Optional) External IP addresses provided to this instance. List of IP pools from which to draw addresses.
- `timeouts` (Attribute, Optional) (see [below for nested schema](#nestedatt--timeouts))
- `user_data` (String) User data for instance initialization systems (such as cloud-init). Must be a Base64-encoded string, as specified in [RFC 4648 § 4](https://datatracker.ietf.org/doc/html/rfc4648#section-4) (+ and / characters with padding). Maximum 32 KiB unencoded data.

### Read-Only

- `id` (String) Unique, immutable, system-controlled identifier of the instance.
- `run_state` (String) Running state of an Instance (primarily: booted or stopped). This typically reflects whether it's starting, running, stopping, or stopped, but also includes states related to the instance's lifecycle.
- `time_created` (String) Timestamp of when this instance was created.
- `time_modified` (String) Timestamp of when this instance last modified.
- `time_run_state_updated` (String) Timestamp of when the run state of this instance was last modified.

<a id="nestedatt--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `create` (String, Default `10m`)
- `delete` (String, Default `10m`)
- `read` (String, Default `10m`)

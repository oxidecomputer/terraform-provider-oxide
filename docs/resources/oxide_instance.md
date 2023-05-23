---
page_title: "oxide_instance Resource - terraform-provider-oxide"
---

# oxide_instance (Resource)

This resource manages instances.

!> Instances must be stopped before updating disk attachments and deleting

## Example Usage

### Basic instance with attached disks

```hcl
resource "oxide_instance" "example" {
  project_id       = "c1dee930-a8e4-11ed-afa1-0242ac120002"
  description      = "a test instance"
  name             = "myinstance"
  host_name        = "<host value>"
  memory           = 1073741824
  ncpus            = 1
  disk_attachments = ["611bb17d-6883-45be-b3aa-8a186fdeafe8", "1aa748cb-26f0-4bf5-8faf-b202dc74d698"]
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

### Define user data

```hcl
resource "oxide_instance" "example" {
  project_id  = "c1dee930-a8e4-11ed-afa1-0242ac120002"
  description = "a test instance"
  name        = "myinstance"
  host_name   = "<host value>"
  memory      = 1073741824
  ncpus       = 1
  user_data   = filebase64("path/to/init.sh")
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

- `disk_attachments` (Set of String, Optional) IDs of the disks to be attached to the instance.
- `external_ips` (List of String, Optional) External IP addresses provided to this instance. List of IP pools from which to draw addresses.
- `timeouts` (Attribute, Optional) (see [below for nested schema](#nestedatt--timeouts))
- `user_data` (String) User data for instance initialization systems (such as cloud-init). Must be a Base64-encoded string, as specified in [RFC 4648 § 4](https://datatracker.ietf.org/doc/html/rfc4648#section-4) (+ and / characters with padding). Maximum 32 KiB unencoded data.

### Read-Only

- `id` (String) Unique, immutable, system-controlled identifier of the instance.
- `time_created` (String) Timestamp of when this instance was created.
- `time_modified` (String) Timestamp of when this instance last modified.

<a id="nestedatt--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `create` (String, Default `10m`)
- `delete` (String, Default `10m`)
- `read` (String, Default `10m`)

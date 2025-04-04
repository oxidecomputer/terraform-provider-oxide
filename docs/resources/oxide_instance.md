---
page_title: "oxide_instance Resource - terraform-provider-oxide"
---

# oxide_instance (Resource)

This resource manages instances.

!> Updates will stop and start the instance.

-> When setting a boot disk using `boot_disk_id`, the boot disk ID must also be
present in `disk_attachments`.

## Example Usage

### Instance minimal example

```hcl
resource "oxide_instance" "example" {
  project_id       = "c1dee930-a8e4-11ed-afa1-0242ac120002"
  description      = "Example instance."
  name             = "myinstance"
  host_name        = "myhostname"
  memory           = 10737418240
  ncpus            = 1
  disk_attachments = ["611bb17d-6883-45be-b3aa-8a186fdeafe8"]
}
```

### Instance with user data and an SSH public key and anti-affinity group

```hcl
resource "oxide_instance" "example" {
  project_id           = "c1dee930-a8e4-11ed-afa1-0242ac120002"
  description          = "Example instance."
  name                 = "myinstance"
  host_name            = "myhostname"
  memory               = 10737418240
  ncpus                = 1
  anti_affinity_groups = ["9b9f9be1-96bf-44ad-864a-0dedae3b3999"]
  disk_attachments     = ["611bb17d-6883-45be-b3aa-8a186fdeafe8"]
  ssh_public_keys      = ["066cab1b-c550-4aea-8a80-8422fd3bfc40"]
  user_data            = filebase64("path/to/init.sh")
}
```

### Instance with a custom network interface with resource timeouts

```hcl
resource "oxide_instance" "example" {
  project_id       = "c1dee930-a8e4-11ed-afa1-0242ac120002"
  description      = "Example instance."
  name             = "myinstance"
  host_name        = "myhostname"
  memory           = 10737418240
  ncpus            = 1
  disk_attachments = ["611bb17d-6883-45be-b3aa-8a186fdeafe8"]
  network_interfaces = [
    {
      subnet_id   = "066cab1b-c550-4aea-8a80-8422fd3bfc40"
      vpc_id      = "9b9f9be1-96bf-44ad-864a-0dedae3b3999"
      description = "Example network interface."
      name        = "mynic"
    },
  ]
  timeouts = {
    read   = "1m"
    create = "3m"
    delete = "2m"
  }
}
```

### Instance with external IPs that does not start when created

```hcl
resource "oxide_instance" "example" {
  project_id       = "c1dee930-a8e4-11ed-afa1-0242ac120002"
  description      = "Example instance."
  name             = "myinstance"
  host_name        = "myhostname"
  memory           = 10737418240
  ncpus            = 1
  disk_attachments = ["611bb17d-6883-45be-b3aa-8a186fdeafe8"]
  start_on_create  = false
  external_ips = [
    {
      type = "ephemeral"
    },
    {
      id   = "eb65d5cb-d8c5-4eae-bcf3-a0e89a633042"
      type = "floating"
    }
  ]
}
```

## Schema

### Required

- `description` (String) Description for the instance.
- `host_name` (String) Hostname of the instance.
- `memory` (Number) Instance memory in bytes.
- `name` (String) Name of the instance.
- `ncpus` (Number) Number of CPUs allocated for this instance.
- `project_id` (String) ID of the project that will contain the instance.

### Optional

- `anti_affinity_groups` (Set of String, Optional) The IDs of the anti-affinity groups this instance should belong to.
- `boot_disk_id` (String, Optional) ID of the disk to boot the instance from. When provided, this ID must also be present in `disk_attachments`.
- `disk_attachments` (Set of String, Optional) IDs of the disks to be attached to the instance. When multiple disk IDs are provided, set `book_disk_id` to specify the boot disk for the instance. Otherwise, a boot disk will be chosen randomly.
- `external_ips` (Set of Object, Optional) External IP addresses associated with the instance. See [below for nested schema](#nestedatt--ips).
- `network_interfaces` (Set of Object, Optional) Network interface devices attached to the instance. See [below for nested schema](#nestedatt--nics).
- `ssh_public_keys` (Set of String, Optional) The IDs of the SSH public keys to be transferred to the instance via cloud-init.
- `start_on_create` (Boolean, Default `true`) Whether to start the instance on creation.
- `timeouts` (Attribute, Optional) Timeouts for performing API operations. See [below for nested schema](#nestedatt--timeouts).
- `user_data` (String) User data for instance initialization systems (e.g., cloud-init). Must be a Base64-encoded string as specified in [RFC 4648 § 4](https://datatracker.ietf.org/doc/html/rfc4648#section-4). Must be no larger than 32 KiB unencoded.

### Read-Only

- `id` (String) Unique, immutable, system-controlled identifier of the instance.
- `time_created` (String) Timestamp of when this instance was created.
- `time_modified` (String) Timestamp of when this instance last modified.

<a id="nestedatt--ips"></a>

### Nested Schema for `external_ips`

### Required

- `type` (String) Type of external IP. Must be one of `ephemeral` or `floating`.

### Optional

- `id` (String) If `type` is `ephemeral`, the ID of the IP pool to retrieve addresses from, or the silo's default pool if not specified. If `type` is `floating`, the ID of the floating IP.

<a id="nestedatt--nics"></a>

### Nested Schema for `network_interfaces`

### Required

- `description` (String) Description for the network interface.
- `name` (String) Name of the network interface.
- `subnet_id` (String) ID of the VPC subnet in which to create the network interface.
- `vpc_id` (String) ID of the VPC in which to create the network interface.

### Optional

- `ip_address` (String) IP address for the network interface. One will be auto-assigned if not provided.

### Read-Only

- `id` (String) Unique, immutable, system-controlled identifier of the network interface.
- `mac_address` (String) MAC address assigned to the network interface.
- `primary` (Boolean) True if this is the primary network interface for the instance to which it's attached to.
- `time_created` (String) Timestamp of when this network interface was created.
- `time_modified` (String) Timestamp of when this network interface was last modified.

<a id="nestedatt--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `create` (String, Default `10m`)
- `delete` (String, Default `10m`)
- `read` (String, Default `10m`)
- `update` (String, Default `10m`)

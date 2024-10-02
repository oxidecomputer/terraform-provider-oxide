---
page_title: "oxide_instance Resource - terraform-provider-oxide"
---

# oxide_instance (Resource)

This resource manages instances.

<!-- TODO: TBD on this behaviour or require replace -->
-> Boot disk updates will stop and reboot the instance.

-> When setting a boot disk, the boot disk ID should also be included as part of `disk_attachments`.

## Example Usage

### Basic instance with attached disks and a network interface

```hcl
resource "oxide_instance" "example" {
  project_id       = "c1dee930-a8e4-11ed-afa1-0242ac120002"
  boot_disk_id     = "611bb17d-6883-45be-b3aa-8a186fdeafe8"
  description      = "a test instance"
  name             = "myinstance"
  host_name        = "<host value>"
  memory           = 1073741824
  ncpus            = 1
  ssh_public_keys  = ["066cab1b-c550-4aea-8a80-8422fd3bfc40", "1aa748cb-26f0-4bf5-8faf-b202dc74d698"]
  disk_attachments = ["611bb17d-6883-45be-b3aa-8a186fdeafe8", "eb65d5cb-d8c5-4eae-bcf3-a0e89a633042"]
  network_interfaces = [
    {
      subnet_id   = "066cab1b-c550-4aea-8a80-8422fd3bfc40"
      vpc_id      = "9b9f9be1-96bf-44ad-864a-0dedae3b3999"
      description = "a sample nic"
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

### Optional

- `boot_disk_id` (String, Optional) ID of the disk the instance should be booted from. This ID must also be present in `disk_attachments`.
- `disk_attachments` (Set of String, Optional) IDs of the disks to be attached to the instance.
- `external_ips` (Set of Object, Optional) External IP addresses provided to this instance. (see [below for nested schema](#nestedatt--ips))
- `network_interfaces` (Set of Object, Optional) Virtual network interface devices attached to an instance. (see [below for nested schema](#nestedatt--nics))
- `ssh_public_keys` (Set of String, Optional) An allowlist of IDs of the saved SSH public keys to be transferred to the instance via cloud-init during instance creation.
- `start_on_create` (Boolean, Default `true`) Starts the instance on creation when set to true.
- `timeouts` (Attribute, Optional) (see [below for nested schema](#nestedatt--timeouts))
- `user_data` (String) User data for instance initialization systems (such as cloud-init). Must be a Base64-encoded string, as specified in [RFC 4648 § 4](https://datatracker.ietf.org/doc/html/rfc4648#section-4) (+ and / characters with padding). Maximum 32 KiB unencoded data.

### Read-Only

- `id` (String) Unique, immutable, system-controlled identifier of the instance.
- `time_created` (String) Timestamp of when this instance was created.
- `time_modified` (String) Timestamp of when this instance last modified.

<a id="nestedatt--ips"></a>

### Nested Schema for `external_ips`

### Required

- `type` (String) Type of external IP. Possible values are: ephemeral or floating.

### Optional

- `id` (String) If type is ephemeral, ID of the IP pool to retrieve addresses from, or the current silo's default pool if not specified. If type is floating, id of the floating IP.

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

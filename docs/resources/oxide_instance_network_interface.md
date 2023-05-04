---
page_title: "oxide_instance_network_interface Resource - terraform-provider-oxide"
---

# oxide_instance_network_interface (Resource)

This resource manages virtual network interface devices attached to an instance.

!> The associated instance must be stopped when the network interface is attached.

## Example Usage

```hcl
resource "oxide_instance_network_interface" "example" {
  instance_id = "c1dee930-a8e4-11ed-afa1-0242ac120002"
  subnet_id   = "611bb17d-6883-45be-b3aa-8a186fdeafe8"
  vpc_id      = "1aa748cb-26f0-4bf5-8faf-b202dc74d698"
  description = "a sample nic"
  name        = "mynic"
  ip_address  = "172.20.15.249"
}
```

## Schema

### Required

- `description` (String) Description for the instance network interface.
- `instance_id` (String) ID of the instance to which the network interface will belong to.
- `name` (String) Name of the VPC.
- `subnet_id` (String) ID of the VPC subnet in which to create the instance network interface.
- `vpc_id` (String) ID of the VPC in which to create the instance network interface.

### Optional

- `ip_address` (String) IP address for the instance network interface. One will be auto-assigned if not provided.
- `timeouts` (Attribute) (see [below for nested schema](#nestedatt--timeouts))

### Read-Only

- `id` (String) Unique, immutable, system-controlled identifier of the instance network interface.
- `mac_address` (String) MAC address assigned to the instance network interface.
- `primary` (Boolean) True if this is the primary network interface for the instance to which it's attached to.
- `time_created` (String) Timestamp of when this VPC was created.
- `time_modified` (String) Timestamp of when this VPC was last modified.

<a id="nestedatt--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `create` (String, Default `10m`)
- `delete` (String, Default `10m`)
- `read` (String, Default `10m`)
- `update` (String, Default `10m`)

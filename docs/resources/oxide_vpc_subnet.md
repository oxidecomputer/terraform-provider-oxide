---
page_title: "oxide_vpc_subnet Resource - terraform-provider-oxide"
---

# oxide_vpc_subnet (Resource)

This resource manages VPC subnets.

## Example Usage

```hcl
resource "oxide_vpc_subnet" "example" {
  vpc_id      = "c1dee930-a8e4-11ed-afa1-0242ac120002"
  description = "a sample vpc subnet"
  name        = "mysubnet"
  ipv4_block  = "192.168.0.0/16"
  ipv6_block  = "fdfe:f6a5:5f06:a643::/64"
  timeouts = {
    read   = "1m"
    create = "3m"
    delete = "2m"
    update = "2m"
  }
}
```

## Schema

### Required

- `description` (String) Description for the VPC subnet.
- `ipv4_block` (String) IPv4 address range for this VPC subnet. It must be allocated from an RFC 1918 private address range, and must not overlap with any other existing subnet in the VPC.
- `name` (String) Name of the VPC subnet.
- `vpc_id` (String) ID of the VPC that will contain the subnet.

### Optional

- `ipv6_block` (String, Optional) IPv6 address range for this VPC subnet. It must be allocated from the RFC 4193 Unique Local Address range, with the prefix equal to the parent VPC's prefix. A random `/64` block will be assigned if one is not provided. It must not overlap with any existing subnet in the VPC.
- `timeouts` (Attribute, Optional) (see [below for nested schema](#nestedatt--timeouts))

### Read-Only

- `id` (String) Unique, immutable, system-controlled identifier of the VPC subnet.
- `time_created` (String) Timestamp of when this VPC subnet was created.
- `time_modified` (String) Timestamp of when this VPC subnet was last modified.

<a id="nestedatt--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `create` (String, Default `10m`)
- `delete` (String, Default `10m`)
- `read` (String, Default `10m`)
- `update` (String, Default `10m`)

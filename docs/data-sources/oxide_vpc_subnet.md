---
page_title: "oxide_vpc_subnet Data Source - terraform-provider-oxide"
---

# oxide_vpc_subnet (Data Source)

Retrieve information about a specified VPC subnet.

## Example Usage

```hcl
data "oxide_vpc_subnet" "example" {
  project_name = "my-project"
  name         = "default"
  vpc_name     = "default"
  timeouts = {
    read = "1m"
  }
}
```

## Schema

### Required

- `name` (String) Name of the subnet.
- `project_name` (String) Name of the project that contains the subnet.
- `vpc_name` (String) Name of the VPC that contains the subnet.

### Optional

- `timeouts` (Attribute, Optional) (see [below for nested schema](#nestedatt--timeouts))

### Read-Only

- `description` (String) Description for the VPC subnet.
- `id` (String) Unique, immutable, system-controlled identifier of the VPC subnet.
- `ipv4_block` (String) IPv4 address range for this VPC subnet. It must be allocated from an RFC 1918 private address range, and must not overlap with any other existing subnet in the VPC.
- `name` (String) Name of the VPC subnet.
- `ipv6_block` (String) IPv6 address range for this VPC subnet. It must be allocated from the RFC 4193 Unique Local Address range, with the prefix equal to the parent VPC's prefix.
- `vpc_id` (String) ID of the VPC that contains the subnet.
- `time_created` (String) Timestamp of when this VPC subnet was created.
- `time_modified` (String) Timestamp of when this VPC subnet was last modified.

<a id="nestedatt--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `read` (String, Default `10m`)

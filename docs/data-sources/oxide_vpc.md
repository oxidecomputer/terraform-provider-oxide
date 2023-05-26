---
page_title: "oxide_vpc Data Source - terraform-provider-oxide"
---

# oxide_vpc (Data Source)

Retrieve information about a specified VPC.

## Example Usage

```hcl
data "oxide_vpc" "example" {
  project_name = "my-project"
  name         = "default"
  timeouts = {
    read = "1m"
  }
}
```

## Schema

### Required

- `name` (String) Name of the VPC.
- `project_name` (String) Name of the project that contains the VPC.

### Optional

- `timeouts` (Attribute, Optional) (see [below for nested schema](#nestedatt--timeouts))

### Read-Only

- `description` (String) Description for the VPC.
- `dns_name` (String) DNS name of the VPC.
- `id` (String) Unique, immutable, system-controlled identifier of the VPC.
- `ipv6_prefix` (String) All IPv6 subnets created from this VPC must be taken from this range, which should be a unique local address in the range `fd00::/48`. The default VPC Subnet will have the first `/64` range from this prefix.
- `project_id` (String) ID of the project that will contain the VPC.
- `system_router_id` (String) ID for the system router where subnet default routes are registered.
- `time_created` (String) Timestamp of when this VPC was created.
- `time_modified` (String) Timestamp of when this VPC was last modified.

<a id="nestedatt--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `read` (String, Default `10m`)

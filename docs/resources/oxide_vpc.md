---
page_title: "oxide_vpc Resource - terraform-provider-oxide"
---

# oxide_vpc (Resource)

This resource manages VPCs.

## Example Usage

```hcl
resource "oxide_vpc" "example" {
  organization_name = "staff"
  project_name      = "test"
  description       = "a test vpc"
  name              = "myvpc"
  dns_name          = "my-vpc-dns"
  ipv6_prefix       = "fd1e:4947:d4a1::/48"
}
```

## Schema

### Required

- `description` (String) Description for the VPC.
- `dns_name` (String) DNS name of the VPC.
- `name` (String) Name of the VPC.
- `organization_name` (String) Name of the organization.
- `project_name` (String) Name of the project.

### Optional

- `ipv6_prefix` (String, Optional) All IPv6 subnets created from this VPC must be taken from this range, which should be a unique local address in the range `fd00::/48`. The default VPC Subnet will have the first `/64` range from this prefix. If no `ipv6_prefix` is defined, a default one will be set.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String) Unique, immutable, system-controlled identifier of the VPC.
- `project_id` (String) Unique, immutable, system-controlled identifier of the project.
- `system_router_id` (String) ID for the system router where subnet default routes are registered.
- `time_created` (String) Timestamp of when this VPC was created.
- `time_modified` (String) Timestamp of when this VPC was last modified.

<a id="nestedblock--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `default` (String)

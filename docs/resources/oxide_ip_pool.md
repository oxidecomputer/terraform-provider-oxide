---
page_title: "oxide_ip_pool Resource - terraform-provider-oxide"
---

# oxide_ip_pool (Resource)

This resource manages IP pools.

!> This resource currently only provides create, read and delete actions.

## Example Usage

```hcl
resource "oxide_ip_pool" "example" {
  description       = "a test ippool"
  name              = "myippool"
  ranges {
    first_address = "172.20.15.227"
    last_address  = "172.20.15.239"
  }
}
```

## Schema

### Required

- `description` (String) Description for the IP pool.
- `name` (String) Name of the IP pool.

### Optional

- `organization_name` (String, Optional) Name of the organization.
- `project_name` (String, Optional) Name of the project.
- `ranges` (List of Object, Optional) Adds IP ranges to the created IP pool. Can be IPv4 or IPv6. (see [below for nested schema](#nestedblock--ranges))
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String) Unique, immutable, system-controlled identifier of the IP pool.
- `project_id` (String) Unique, immutable, system-controlled identifier of the project.
- `time_created` (String) Timestamp of when this IP pool was created.
- `time_modified` (String) Timestamp of when this IP pool was last modified.

<a id="nestedblock--ranges"></a>

### Nested Schema for `ranges`

Required:

- `first_address` (String) First address in the range.
- `last_address` (String) Last address in the range.

Read-Only:

- `id` (String) Unique, immutable, system-controlled identifier.
- `time_created` (String) Timestamp of when this range was added to the IP pool.

<a id="nestedblock--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `default` (String)

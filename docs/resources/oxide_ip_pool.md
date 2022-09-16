---
page_title: "oxide_ip_pool Resource - terraform-provider-oxide"
---

# oxide_ip_pool (Resource)

This resource manages IP pools.

## Example Usage

```hcl
resource "oxide_ip_pool" "example" {
  description       = "a test ippool"
  name              = "myippool"
}
```

## Schema

### Required

- `description` (String) Description for the IP pool.
- `name` (String) Name of the IP pool.

### Optional

- `organization_name` (String, Optional) Name of the organization.
- `project_name` (String, Optional) Name of the project.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String) Unique, immutable, system-controlled identifier of the IP pool.
- `project_id` (String) Unique, immutable, system-controlled identifier of the project.
- `time_created` (String) Timestamp of when this IP pool was created.
- `time_modified` (String) Timestamp of when this IP pool was last modified.

<a id="nestedblock--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `default` (String)

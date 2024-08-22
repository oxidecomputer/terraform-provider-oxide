---
page_title: "oxide_ip_pool Resource - terraform-provider-oxide"
---

# oxide_ip_pool (Resource)

This resource manages IP pools.

## Example Usage

```hcl
resource "oxide_ip_pool" "example" {
  description = "a test IP pool"
  name        = "myippool"
  ranges = [
    {
      first_address = "172.20.18.227"
      last_address  = "172.20.18.239"
    }
  ]
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

- `description` (String) Description for the IP pool.
- `name` (String) Name of the IP pool.

### Optional

- `ranges` (List of Object, Optional) Adds IP ranges to the created IP pool. (see [below for nested schema](#nestedblock--ranges))
- `timeouts` (Attribute, Optional) (see [below for nested schema](#nestedatt--timeouts))

### Read-Only

- `id` (String) Unique, immutable, system-controlled identifier of the IP pool.
- `time_created` (String) Timestamp of when this IP pool was created.
- `time_modified` (String) Timestamp of when this IP pool was last modified.

<a id="nestedblock--ranges"></a>

### Nested Schema for `ranges`

Required:

- `first_address` (String) First address in the range.
- `last_address` (String) Last address in the range.

<a id="nestedatt--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `create` (String, Default `10m`)
- `delete` (String, Default `10m`)
- `read` (String, Default `10m`)
- `update` (String, Default `10m`)
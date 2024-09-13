---
page_title: "oxide_ip_pool_silo_link Resource - terraform-provider-oxide"
---

# oxide_ip_pool_silo_link (Resource)

This resource manages IP pool to silo links.

## Example Usage

```hcl
resource "oxide_ip_pool_silo_link" "example" {
  silo_id = "1fec2c21-cf22-40d8-9ebd-e5b57ebec80f"
  ip_pool_id = "081a331d-5ee4-4a23-ac8b-328af5e15cdc"
  is_default = true
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

- `silo_id` (String) Description for the IP pool.
- `ip_pool_id` (String) Name of the IP pool.
- `is_default` (Boolean) Whether a pool is the default for a silo. All floating IPs and instance ephemeral IPs will come from that pool when no other pool is specified. 

-> There can only be one default pool for a given silo.

### Optional

- `timeouts` (Attribute, Optional) (see [below for nested schema](#nestedatt--timeouts))

### Read-Only

- `id` (String) Unique, immutable, system-controlled, terraform-specific identifier of the resource.

<a id="nestedatt--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `create` (String, Default `10m`)
- `delete` (String, Default `10m`)
- `read` (String, Default `10m`)
- `update` (String, Default `10m`)


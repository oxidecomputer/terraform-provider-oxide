---
page_title: "oxide_ip_pool Data Source - terraform-provider-oxide"
---

# oxide_ip_pool (Data Source)

Retrieve information about a specified IP pool.

## Example Usage

```hcl
data "oxide_ip_pool" "example" {
  name = "default"
  timeouts = {
    read = "1m"
  }
}
```

## Schema

### Required

- `name` (String) Name of the IP pool.

### Optional

- `timeouts` (Attribute, Optional) (see [below for nested schema](#nestedatt--timeouts))

### Read-Only

- `description` (String) Description for the IP pool.
- `id` (String) Unique, immutable, system-controlled identifier of the IP pool.
- `is_default` (Bool) If a pool is the default for a silo, floating IPs and instance ephemeral IPs
  will come from that pool when no other pool is specified. There can be at most one default for a
  given silo.
- `time_created` (String) Timestamp of when this IP pool was created.
- `time_modified` (String) Timestamp of when this IP pool was last modified.

<a id="nestedatt--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `read` (String, Default `10m`)

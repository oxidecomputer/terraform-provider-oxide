---
page_title: "oxide_system_ip_pools Data Source - terraform-provider-oxide"
---

# oxide_system_ip_pools (Data Source)

Retrieve all configured ip pools for the Oxide system

## Example Usage

```hcl
data "oxide_system_ip_pools" "example" {
  timeouts = {
    read = "1m"
  }
}
```

## Schema

### Required

### Optional

- `timeouts` (Attribute, Optional) (see [below for nested schema](#nestedatt--timeouts))

### Read-Only

- `ip_pools` (List of Objects) A list of ip pool objects configured for the system (see [below for nested schema](#nestedatt--ip_pools))

<a id="nestedatt--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `read` (String, Default `10m`)

<a id="nestedatt--ip_pools"></a>

### Nested Schema for `ip_pools`

- `description` (String) Description for the IP pool.
- `id` (String) Unique, immutable, system-controlled identifier of the IP pool.
- `name` (String) Name of the IP pool.
- `time_created` (String) Timestamp of when this IP pool was created.
- `time_modified` (String) Timestamp of when this IP pool was last modified.
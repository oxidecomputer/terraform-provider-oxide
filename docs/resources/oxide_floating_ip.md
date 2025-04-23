---
page_title: "oxide_floating_ip Resource - terraform-provider-oxide"
---

# oxide_floating_ip (Resource)

This resource manages Oxide floating IPs.

## Example Usage

### Allocate a floating IP from the silo's default IP pool

```hcl
resource "oxide_floating_ip" "example" {
  project_id  = "5476ccc9-464d-4dc4-bfc0-5154de1c986f"
  name        = "app-ingress"
  description = "Ingress for application."
}
```

### Allocate a floating IP from the specified IP pool 

```hcl
resource "oxide_floating_ip" "example" {
  project_id  = "5476ccc9-464d-4dc4-bfc0-5154de1c986f"
  name        = "app-ingress"
  description = "Ingress for application."
  ip_pool_id  = "a4720b36-006b-49fc-a029-583528f18a4d"
}
```

### Allocate a specific floating IP from the specified IP pool

```hcl
resource "oxide_floating_ip" "example" {
  project_id  = "5476ccc9-464d-4dc4-bfc0-5154de1c986f"
  name        = "app-ingress"
  description = "Ingress for application."
  ip_pool_id  = "a4720b36-006b-49fc-a029-583528f18a4d"
  ip          = "172.21.252.128"
}
```

## Schema

### Required

- `name` (String) Name of the floating IP.
- `description` (String) Description for the floating IP.
- `project_id` (String) ID of the project that will contain the floating IP.

### Optional

- `ip` (String) IP address for this floating IP. If unset, an IP address will be chosen from the given `ip_pool_id`, or the silo's default IP pool if the `ip_pool_id` attribute is unset.
- `ip_pool_id` (String) ID of the IP pool where the floating IP will be allocated from. If unset, the silo's default IP pool is used.
- `timeouts` (Attribute, Optional) Timeouts for performing API operations. See [below for nested schema](#nestedatt--timeouts).

### Read-Only

- `id` (String) Unique, immutable, system-controlled identifier for the floating IP.
- `instance_id` (String) Instance ID that this floating IP is attached to, if presently attached.
- `time_created` (String) Timestamp of when the floating IP was created.
- `time_modified` (String) Timestamp of when the floating IP was modified.

<a id="nestedatt--timeouts"></a>

### Nested Schema for `timeouts`

#### Optional

- `create` (String, Default `10m`)
- `delete` (String, Default `10m`)
- `read` (String, Default `10m`)
- `update` (String, Default `10m`)

---
page_title: "oxide_floating_ip Data Source - terraform-provider-oxide"
---

# oxide_floating_ip (Data Source)

Retrieve information about a specified floating IP.

## Example Usage

```hcl
data "oxide_floating_ip" "example" {
  project_name = "my-project"
  name         = "app-ingress"
}
```

## Schema

### Required

- `name` (String) Name of the floating IP.
- `project_name` (String) Name of the project that contains the floating IP.

### Optional

- `timeouts` (Attribute, Optional) (see [below for nested schema](#nestedatt--timeouts))

### Read-Only

- `id` (String) Unique, immutable, system-controlled identifier for the floating IP.
- `description` (String) Description for the floating IP.
- `ip` (String) IP address for this floating IP.
- `ip_pool_id` (String) ID of the IP pool where the floating IP was allocated from.
- `instance_id` (String) Instance ID that this floating IP is attached to, if presently attached.
- `time_created` (String) Timestamp of when the floating IP was created.
- `time_modified` (String) Timestamp of when the floating IP was modified.
- `timeouts` (Attribute, Optional) Timeouts for performing API operations. See [below for nested schema](#nestedatt--timeouts).

<a id="nestedatt--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `read` (String, Default `10m`)

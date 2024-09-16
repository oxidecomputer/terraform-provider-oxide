---
page_title: "oxide_silo Data Source - terraform-provider-oxide"
---

# oxide_silo (Data Source)

Retrieve information about a specified Silo.

## Example Usage

```hcl
data "oxide_silo" "example" {
  name = "default"
  timeouts = {
    read = "1m"
  }
}
```

## Schema

### Required

- `name` (String) Name of the Silo.

### Optional

- `timeouts` (Attribute, Optional) (see [below for nested schema](#nestedatt--timeouts))

### Read-Only

- `description` (String) Description for the Silo.
- `id` (String) Unique, immutable, system-controlled identifier of the Silo.
- `identity_mode` (String) How users and groups are managed in this Silo.
- `discoverable` (Bool) A silo where discoverable is false can be retrieved only by its ID - it will not be part of the 'list all silos' output.
- `time_created` (String) Timestamp of when this Silo was created.
- `time_modified` (String) Timestamp of when this Silo was last modified.

<a id="nestedatt--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `read` (String, Default `10m`)

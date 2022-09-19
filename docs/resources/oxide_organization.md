---
page_title: "oxide_organization Resource - terraform-provider-oxide"
---

# oxide_organization (Resource)

This resource manages organizations.

## Example Usage

```hcl
resource "oxide_organization" "example" {
  description       = "a test org"
  name              = "myorg"
}
```

## Schema

### Required

- `description` (String) Description for the organization.
- `name` (String) Name of the organization.

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String) Unique, immutable, system-controlled identifier of the organization.
- `time_created` (String) Timestamp of when this organization was created.
- `time_modified` (String) Timestamp of when this organization was last modified.

<a id="nestedblock--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `default` (String)

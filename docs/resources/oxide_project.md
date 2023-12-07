---
page_title: "oxide_project Resource - terraform-provider-oxide"
---

# oxide_project (Resource)

This resource manages projects.

## Example Usage

```hcl
resource "oxide_project" "example" {
  description = "a test project"
  name        = "myproject"
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

- `description` (String) Description for the project.
- `name` (String) Name of the project.

### Optional

- `timeouts` (Attribute, Optional) (see [below for nested schema](#nestedatt--timeouts))

### Read-Only

- `id` (String) Unique, immutable, system-controlled identifier of the project.
- `time_created` (String) Timestamp of when this project was created.
- `time_modified` (String) Timestamp of when this project was last modified.

<a id="nestedatt--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `create` (String, Default `10m`)
- `delete` (String, Default `10m`)
- `read` (String, Default `10m`)
- `update` (String, Default `10m`)

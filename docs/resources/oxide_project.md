---
page_title: "oxide_project Resource - terraform-provider-oxide"
---

# oxide_project (Resource)

This resource manages projects.

## Example Usage

```hcl
resource "oxide_project" "example" {
  description       = "a test org"
  name              = "myorg"
  organization_name = "staff"
}
```

## Schema

### Required

- `description` (String) Description for the project.
- `name` (String) Name of the project.
- `organization_name` (String) Name of the organization.

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String) Unique, immutable, system-controlled identifier of the project.
- `time_created` (String) Timestamp of when this project was created.
- `time_modified` (String) Timestamp of when this project was last modified.

<a id="nestedblock--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `default` (String)

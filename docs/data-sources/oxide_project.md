---
page_title: "oxide_project Data Source - terraform-provider-oxide"
---

# oxide_project (Data Source)

Retrieve information about a specified project.

## Example Usage

```hcl
data "oxide_project" "example" {
  name = "test"
  timeouts = {
    read = "1m"
  }
}
```

## Schema

### Required

- `name` (String) Name of the project.

### Optional

- `timeouts` (Attribute, Optional) (see [below for nested schema](#nestedatt--timeouts))

### Read-Only

- `description` (String) Description for the project.
- `id` (String) Unique, immutable, system-controlled identifier of the project.
- `time_created` (String) Timestamp of when this project was created.
- `time_modified` (String) Timestamp of when this project was last modified.

<a id="nestedatt--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `read` (String, Default `10m`)

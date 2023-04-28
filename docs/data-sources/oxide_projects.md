---
page_title: "oxide_projects Data Source - terraform-provider-oxide"
---

# oxide_projects (Data Source)

Retrieve a list of projects.

## Example Usage

```hcl
data "oxide_projects" "example" {}
```

## Schema

### Optional

- `timeouts` (Attribute, Optional) (see [below for nested schema](#nestedatt--timeouts))

### Read-Only

- `id` (String) The ID of this resource.
- `projects` (List of Object) A list of all projects (see [below for nested schema](#nestedatt--projects))

<a id="nestedatt--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `read` (String, Default `10m`)

<a id="nestedatt--projects"></a>

### Nested Schema for `projects`

Read-Only:

- `description` (String) Description for the project.
- `id` (String) Unique, immutable, system-controlled identifier of the project.
- `name` (String) Name of the project.
- `time_created` (String) Timestamp of when this project was created.
- `time_modified` (String) Timestamp of when this project was last modified.

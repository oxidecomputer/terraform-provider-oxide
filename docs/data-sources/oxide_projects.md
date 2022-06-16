---
page_title: "oxide_projects Data Source - terraform-provider-oxide"
---

# oxide_projects (Data Source)

Retrieve a list of projects within an organization.

## Example Usage

```hcl
data "oxide_projects" "example" {
  organization_name = "staff"
}
```

## Schema

### Required

- `organization_name` (String) Name of the organization.

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String) The ID of this resource.
- `projects` (List of Object) A list of all projects (see [below for nested schema](#nestedatt--projects))

<a id="nestedblock--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `default` (String)

<a id="nestedatt--projects"></a>

### Nested Schema for `projects`

Read-Only:

- `description` (String) Description for the project.
- `id` (String) Unique, immutable, system-controlled identifier of the project.
- `name` (String) Name of the project.
- `organization_id` (String) Unique, immutable, system-controlled identifier of the organization.
- `time_created` (String) Timestamp of when this project was created.
- `time_modified` (String) Timestamp of when this project was last modified.

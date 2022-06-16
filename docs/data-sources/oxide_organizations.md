---
page_title: "oxide_organizations Data Source - terraform-provider-oxide"
---

# oxide_organizations (Data Source)

Retrieve a list of all organizations.

## Example Usage

```hcl
data "oxide_organizations" "example" {}
```

## Schema

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String) The ID of this resource.
- `organizations` (List of Object) A list of all organizations (see [below for nested schema](#nestedatt--organizations))

<a id="nestedblock--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `default` (String)

<a id="nestedatt--organizations"></a>

### Nested Schema for `organizations`

Read-Only:

- `description` (String) Description of the organization.
- `id` (String) Unique, immutable, system-controlled identifier of the organization.
- `name` (String) Name of the organization.
- `time_created` (String) Timestamp of when this organization was created.
- `time_modified` (String) Timestamp of when this organization was last modified.

---
page_title: "oxide_anti_affinity_group Resource - terraform-provider-oxide"
---

# oxide_anti_affinity_group (Resource)

This resource manages anti-affinity groups.

## Example Usage

```hcl
resource "oxide_anti_affinity_group" "example" {
  project_id  = "c1dee930-a8e4-11ed-afa1-0242ac120002"
  description = "a test anti-affinity group"
  name        = "my-anti-affinty-group"
  policy      = "allow"
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

- `description` (String) Description for the anti-affinity group.
- `name` (String) Name of the anti-affinity group.
- `policy` (String) Affinity policy used to describe what to do when a request cannot be satisfied. Possible values are: `allow` or `fail`.
- `project_id` (String) ID of the project that will contain the anti-affinity group.

### Optional

- `timeouts` (Attribute, Optional) (see [below for nested schema](#nestedatt--timeouts))

### Read-Only

- `id` (String) Unique, immutable, system-controlled identifier of the anti-affinity group.
- `failure_domain` (String) Describes the scope of affinity for the purposes of co-location.
- `time_created` (String) Timestamp of when this anti-affinity group was created.
- `time_modified` (String) Timestamp of when this anti-affinity group was last modified.

<a id="nestedatt--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `create` (String, Default `10m`)
- `delete` (String, Default `10m`)
- `read` (String, Default `10m`)
- `update` (String, Default `10m`)

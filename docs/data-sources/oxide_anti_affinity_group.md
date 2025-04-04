---
page_title: "oxide_anti_affinity_group Data Source - terraform-provider-oxide"
---

# oxide_anti_affinity_group (Data Source)

Retrieve information about a specified anti-affinity group.

## Example Usage

```hcl
data "oxide_anti_affinity_group" "example" {
  project_name = "my-project"
  name         = "my-group"
  timeouts = {
    read = "1m"
  }
}
```

## Schema

### Required

- `name` (String) Name of the anti-affinity group.
- `project_name` (String) Name of the project that contains the anti-affinity group.

### Optional

- `timeouts` (Attribute, Optional) (see [below for nested schema](#nestedatt--timeouts))

### Read-Only

- `description` (String) Description of the anti-affinity group.
- `failure_domain` (String) Describes the scope of affinity for the purposes of co-location.
- `id` (String) Unique, immutable, system-controlled identifier for the anti-affinity group.
- `policy` (String) Affinity policy used to describe what to do when a request cannot be satisfied.
- `project_id` (String) ID of the project that contains the anti-affinity group.
- `time_created` (String) Timestamp of when this anti-affinity group was created.
- `time_modified` (String) Timestamp of when this anti-affinity group was last modified.

<a id="nestedatt--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `read` (String, Default `10m`)

<a id="nestedobject--digest"></a>

### Nested Schema for `digest`

Read-Only:

- `type` (String) Digest type.
- `value` (String) Digest type value.

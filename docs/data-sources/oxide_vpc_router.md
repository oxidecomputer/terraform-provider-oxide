---
page_title: "oxide_vpc_router Data Source - terraform-provider-oxide"
---

# oxide_vpc_router (Data Source)

Retrieve information about a specified VPC router.

## Example Usage

```hcl
data "oxide_vpc_router" "example" {
  project_name = "my-project"
  name         = "system"
  vpc_name     = "default"
  timeouts = {
    read = "1m"
  }
}
```

## Schema

### Required

- `name` (String) Name of the router.
- `project_name` (String) Name of the project that contains the router.
- `vpc_name` (String) Name of the VPC that contains the router.

### Optional

- `timeouts` (Attribute, Optional) (see [below for nested schema](#nestedatt--timeouts))

### Read-Only

- `description` (String) Description for the VPC router.
- `id` (String) Unique, immutable, system-controlled identifier of the VPC router.
- `name` (String) Name of the VPC router.
- `vpc_id` (String) ID of the VPC that contains the router.
- `time_created` (String) Timestamp of when this VPC router was created.
- `time_modified` (String) Timestamp of when this VPC router was last modified.

<a id="nestedatt--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `read` (String, Default `10m`)

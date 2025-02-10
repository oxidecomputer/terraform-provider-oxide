---
page_title: "oxide_vpc_router Resource - terraform-provider-oxide"
---

# oxide_vpc_router (Resource)

This resource manages VPC routers.

## Example Usage

```hcl
resource "oxide_vpc_router" "example" {
  vpc_id      = "c1dee930-a8e4-11ed-afa1-0242ac120002"
  description = "a sample vpc router"
  name        = "myrouter"
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

- `description` (String) Description for the VPC router.
- `name` (String) Name of the VPC router.
- `vpc_id` (String) ID of the VPC that will contain the router.

### Optional

- `timeouts` (Attribute, Optional) (see [below for nested schema](#nestedatt--timeouts))

### Read-Only

- `id` (String) Unique, immutable, system-controlled identifier of the VPC router.
- `kind` (String) Whether the VPC router is custom or system created.
- `time_created` (String) Timestamp of when this VPC router was created.
- `time_modified` (String) Timestamp of when this VPC router was last modified.

<a id="nestedatt--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `create` (String, Default `10m`)
- `delete` (String, Default `10m`)
- `read` (String, Default `10m`)
- `update` (String, Default `10m`)

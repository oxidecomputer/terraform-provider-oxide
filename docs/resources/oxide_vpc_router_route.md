---
page_title: "oxide_vpc_router_route Resource - terraform-provider-oxide"
---

# oxide_vpc_router_route (Resource)

This resource manages VPC router routes.

## Example Usage

```hcl
resource "oxide_vpc_router_route" "example" {
  vpc_router_id = "c1dee930-a8e4-11ed-afa1-0242ac120002"
  description   = "a sample VPC router route"
  name          = "myroute"
  destination = {
    type  = "ip_net"
    value = "::/0"
  }
  target = {
    type  = "ip"
    value = "::1"
  }
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

- `description` (String) Description for the VPC router route.
- `name` (String) Name of the VPC router route.
- `destination` (Object) Selects which traffic this routing rule will apply to. (see [below for nested schema](#nestedatt--destination))
- `target` (Object) Location that matched packets should be forwarded to. (see [below for nested schema](#nestedatt--target))
- `vpc_router_id` (String) ID of the VPC router that will contain the router route.

### Optional

- `timeouts` (Attribute, Optional) (see [below for nested schema](#nestedatt--timeouts))

### Read-Only

- `id` (String) Unique, immutable, system-controlled identifier of the VPC router.
- `kind` (String) Whether the VPC router is custom or system created.
- `time_created` (String) Timestamp of when this VPC router was created.
- `time_modified` (String) Timestamp of when this VPC router was last modified.

<a id="nestedatt--destination"></a>

### Nested Schema for `destination`

Required:

- `type` (String) Route destination type. Possible values: `vpc`, `subnet`, `ip`, and `ip_net`.
- `value` (String) Depending on the type, it will be one of the following:
  - `vpc`: Name of the VPC
  - `subnet`: Name of the VPC subnet
  - `ip`: IP address
  - `ip_net`: IPv4 or IPv6 subnet

<a id="nestedatt--target"></a>

### Nested Schema for `target`

Required:

- `type` (String) Route target type. Possible values: `vpc`, `subnet`, `instance`, `ip`, `internet_gateway`, and `drop`.

Optional:

- `value` (String) Depending on the type, it will be one of the following:
  - `vpc`: Name of the VPC
  - `subnet`: Name of the VPC subnet
  - `instance`: Name of the instance
  - `ip`: IP address
  - `internet_gateway`: Name of the internet gateway

<a id="nestedatt--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `create` (String, Default `10m`)
- `delete` (String, Default `10m`)
- `read` (String, Default `10m`)
- `update` (String, Default `10m`)

---
page_title: "oxide_vpc_router_route Data Source - terraform-provider-oxide"
---

# oxide_vpc_router_route (Data Source)

Retrieve information about a specified VPC router route.

## Example Usage

```hcl
data "oxide_vpc_router_route" "example" {
  project_name    = "my-project"
  name            = "default-v4"
  vpc_name        = "default"
  vpc_router_name = "system"
  timeouts = {
    read = "1m"
  }
}
```

## Schema

### Required

- `name` (String) Name of the router.
- `project_name` (String) Name of the project that contains the VPC router route.
- `vpc_router_name` (String) Name of the VPC router that contains the VPC router route.
- `vpc_name` (String) Name of the VPC that contains the VPC router route.

### Optional

- `timeouts` (Attribute, Optional) (see [below for nested schema](#nestedatt--timeouts))

### Read-Only

- `description` (String) Description for the VPC router.
- `id` (String) Unique, immutable, system-controlled identifier of the VPC router.
- `destination` (Object) Selects which traffic this routing rule will apply to. (see [below for nested schema](#nestedatt--destination))
- `kind` (String) Whether the VPC router is custom or system created.
- `name` (String) Name of the VPC router.
- `target` (Object) Location that matched packets should be forwarded to. (see [below for nested schema](#nestedatt--target))
- `time_created` (String) Timestamp of when this VPC router was created.
- `time_modified` (String) Timestamp of when this VPC router was last modified.
- `vpc_router_id` (String) ID of the VPC router that contains the VPC router route.

<a id="nestedatt--destination"></a>

### Nested Schema for `destination`

Read-Only:

- `type` (String) Route destination type. Possible values: `vpc`, `subnet`, `ip`, and `ip_net`.
- `value` (String) Depending on the type, it will be one of the following:
  - `vpc`: Name of the VPC
  - `subnet`: Name of the VPC subnet
  - `ip`: IP address
  - `ip_net`: IPv4 or IPv6 subnet

<a id="nestedatt--target"></a>

### Nested Schema for `target`

Read-Only:

- `type` (String) Route destination type. Possible values: `vpc`, `subnet`, `instance`, `ip`, `internet_gateway`, and `drop`.
- `value` (String) Depending on the type, it will be one of the following:
  - `vpc`: Name of the VPC
  - `subnet`: Name of the VPC subnet
  - `instance`: Name of the instance
  - `ip`: IP address
  - `internet_gateway`: Name of the internet gateway

<a id="nestedatt--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `read` (String, Default `10m`)

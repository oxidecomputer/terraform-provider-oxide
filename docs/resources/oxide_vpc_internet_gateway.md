---
page_title: "oxide_vpc_internet_gateway Resource - terraform-provider-oxide"
---

# oxide_vpc_internet_gateway (Resource)

This resource manages VPC internet gateways.

## Example Usage

```hcl
resource "oxide_vpc_internet_gateway" "example" {
  vpc_id      = "c1dee930-a8e4-11ed-afa1-0242ac120002"
  description = "a sample VPC internet gateway"
  name        = "myinternetgateway"
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

- `description` (String) Description for the VPC internet gateway.
- `name` (String) Name of the VPC internet gateway.
- `vpc_id` (String) ID of the VPC that will contain the internet gateway.

### Optional

- `cascade_delete` (Bool, Default `false`) Whether to also delete routes targeting the
VPC internet gateway when deleting the VPC internet gateway.
- `timeouts` (Attribute, Optional) (see [below for nested schema](#nestedatt--timeouts))

### Read-Only

- `id` (String) Unique, immutable, system-controlled identifier of the VPC internet gateway.
- `time_created` (String) Timestamp of when this VPC internet gateway was created.
- `time_modified` (String) Timestamp of when this VPC internet gateway was last modified.

<a id="nestedatt--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `create` (String, Default `10m`)
- `delete` (String, Default `10m`)
- `read` (String, Default `10m`)
- `update` (String, Default `10m`)

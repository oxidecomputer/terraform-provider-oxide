---
page_title: "oxide_address_lot Resource - terraform-provider-oxide"
---

# oxide_address_lot (Resource)

This resource manages address lots.

## Example Usage

```hcl
resource "oxide_address_lot" "example" {
  description = "a test address lot"
  name        = "test-address-lot"
  kind        = "pool"
  blocks = [
    {
      first_address = "172.20.18.227"
      last_address  = "172.20.18.239"
    }
  ]
}
```

## Schema

### Required

- `blocks` (Attributes Set) (see [below for nested schema](#nestedatt--blocks))
- `description` (String) Description for the address lot.
- `kind` (String) Kind for the address lot. Must be one of "infra" or "pool".
- `name` (String) Name of the address lot.

### Optional

- `timeouts` (Attributes) (see [below for nested schema](#nestedatt--timeouts))

### Read-Only

- `id` (String) Unique, immutable, system-controlled identifier of the address lot.
- `time_created` (String) Timestamp of when this address lot was created.
- `time_modified` (String) Timestamp of when this address lot was last modified.

<a id="nestedatt--blocks"></a>

### Nested Schema for `blocks`

Required:

- `first_address` (String) First address in the lot.
- `last_address` (String) Last address in the lot.

Read-Only:

- `id` (String) ID of the address lot block.

<a id="nestedatt--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `create` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
- `delete` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Setting a timeout for a Delete operation is only applicable if changes are saved into state before the destroy operation occurs.
- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Read operations occur during any refresh or planning operation when refresh is enabled.
- `update` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).

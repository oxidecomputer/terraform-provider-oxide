---
page_title: "oxide_address_lot Data Source - terraform-provider-oxide"
---

# oxide_address_lot (Data Source)

Retrieve information about a specified address lot.

## Example Usage

```hcl
data "oxide_address_lot" "example" {
  name = "test-address-lot"
}
```

## Schema

### Required

- `name` (String) Name of the address lot.

### Optional

- `timeouts` (Attributes) (see [below for nested schema](#nestedatt--timeouts))

### Read-Only

- `blocks` (Attributes Set) (see [below for nested schema](#nestedatt--blocks))
- `description` (String) Description for the address lot.
- `id` (String) Unique, immutable, system-controlled identifier of the address lot.
- `kind` (String) Kind for the address lot.
- `time_created` (String) Timestamp of when this address lot was created.
- `time_modified` (String) Timestamp of when this address lot was last modified.

<a id="nestedatt--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).

<a id="nestedatt--blocks"></a>

### Nested Schema for `blocks`

Read-Only:

- `first_address` (String) First address in the lot.
- `id` (String) ID of the address lot block.
- `last_address` (String) Last address in the lot.

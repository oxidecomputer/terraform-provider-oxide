---
page_title: "oxide_instance_external_ips Data Source - terraform-provider-oxide"
---

# oxide_instance_external_ips (Data Source)

Retrieve information of all external IPs associated to an instance.

## Example Usage

```hcl
data "oxide_instance_external_ips" "example" {
  instance_id = "c1dee930-a8e4-11ed-afa1-0242ac120002"
}
```

## Schema

### Required

- `instance_id` (String) ID of the instance to which the external IPs belong to.

### Optional

- `timeouts` (Attribute, Optional) (see [below for nested schema](#nestedatt--timeouts))

### Read-Only

- `external_ips` (List of Object) A list of all external IPs (see [below for nested schema](#nestedatt--images))
- `id` (String) The ID of this resource.

<a id="nestedatt--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `read` (String, Default `10m`)

<a id="nestedatt--images"></a>

### Nested Schema for `images`

Read-Only:

- `ip` (String) External IP address.
- `kind` (String) Kind of external IP address.

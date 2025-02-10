---
page_title: "oxide_vpc_internet_gateway Data Source - terraform-provider-oxide"
---

# oxide_vpc_internet_gateway (Data Source)

Retrieve information about a specified VPC internet gateway.

## Example Usage

```hcl
data "oxide_vpc_internet_gateway" "example" {
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

- `name` (String) Name of the internet gateway.
- `project_name` (String) Name of the project that contains the internet gateway.
- `vpc_name` (String) Name of the VPC that contains the internet gateway.

### Optional

- `timeouts` (Attribute, Optional) (see [below for nested schema](#nestedatt--timeouts))

### Read-Only

- `description` (String) Description for the VPC internet gateway.
- `id` (String) Unique, immutable, system-controlled identifier of the VPC internet gateway.
- `name` (String) Name of the VPC internet gateway.
- `vpc_id` (String) ID of the VPC that contains the internet gateway.
- `time_created` (String) Timestamp of when this VPC internet gateway was created.
- `time_modified` (String) Timestamp of when this VPC internet gateway was last modified.

<a id="nestedatt--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `read` (String, Default `10m`)

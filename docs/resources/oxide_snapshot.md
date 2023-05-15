---
page_title: "oxide_snapshot Resource - terraform-provider-oxide"
---

# oxide_snapshot (Resource)

This resource manages snapshots.

-> This resource currently only provides create, read and delete actions. An update requires a resource replacement

## Example Usage

```hcl
resource "oxide_snapshot" "example2" {
  project_id  = "c1dee930-a8e4-11ed-afa1-0242ac120002"
  description = "a test snapshot"
  name        = "mysnapshot"
  disk_id     = "49118786-ca55-49b1-ae9a-e03f7ce41d8c"
  timeouts = {
    read   = "1m"
    create = "3m"
    delete = "2m"
  }
}
```

## Schema

### Required

- `description` (String) Description for the snapshot.
- `name` (String) Name of the snapshot.
- `project_id` (String) ID of the project that will contain the snapshot.
- `size` (Number) Size of the snapshot in bytes.
- `disk_id` (String) ID of the disk to create the snapshot from.

### Optional

- `timeouts` (Attribute) (see [below for nested schema](#nestedatt--timeouts))

### Read-Only

- `id` (String) Unique, immutable, system-controlled identifier of the snapshot.
- `size` (Number) Size of the snapshot in bytes.
- `time_created` (String) Timestamp of when this snapshot was created.
- `time_modified` (String) Timestamp of when this snapshot was last modified.

<a id="nestedatt--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `create` (String, Default `10m`)
- `delete` (String, Default `10m`)
- `read` (String, Default `10m`)

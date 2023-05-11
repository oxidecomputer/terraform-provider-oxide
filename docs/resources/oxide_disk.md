---
page_title: "oxide_disk Resource - terraform-provider-oxide"
---

# oxide_disk (Resource)

This resource manages disks.

To create a blank disk it's necessary to set `block_size`. Otherwise, one of `source_image_id` or `source_snapshot_id` must be set; `block_size` will be automatically calculated.

!> Disks cannot be deleted while attached to instances. Please detach or delete associated instances before attempting to delete.

-> This resource currently only provides create, read and delete actions. An update requires a resource replacement

## Example Usage

```hcl
resource "oxide_disk" "example" {
  project_id  = "c1dee930-a8e4-11ed-afa1-0242ac120002"
  description = "a test disk"
  name        = "mydisk"
  size        = 1073741824
  block_size  = 512
}

resource "oxide_disk" "example2" {
  project_id      = "c1dee930-a8e4-11ed-afa1-0242ac120002"
  description     = "a test disk"
  name            = "mydisk2"
  size            = 1073741824
  source_image_id = "49118786-ca55-49b1-ae9a-e03f7ce41d8c"
  timeouts = {
    read   = "1m"
    create = "3m"
    delete = "2m"
  }
}
```

## Schema

### Required

- `description` (String) Description for the disk.
- `name` (String) Name of the disk.
- `project_id` (String) ID of the project that will contain the disk.
- `size` (Number) Size of the disk in bytes.

### Optional

- `block_size` (Number) Size of blocks in bytes. To be set only when creating a blank disk, will be computed otherwise.
- `source_image_id` (String) ID of the disk source image. To be set only when creating a disk from an image.
- `source_snapshot_id` (String) ID of the disk source snapshot. To be set only when creating a disk from a snapshot.
- `timeouts` (Attribute) (see [below for nested schema](#nestedatt--timeouts))

### Read-Only

- `device_path` (String) Path of the disk.
- `id` (String) Unique, immutable, system-controlled identifier of the disk.
- `time_created` (String) Timestamp of when this disk was created.
- `time_modified` (String) Timestamp of when this disk was last modified.

<a id="nestedatt--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `create` (String, Default `10m`)
- `delete` (String, Default `10m`)
- `read` (String, Default `10m`)

---
page_title: "oxide_disk Resource - terraform-provider-oxide"
---

# oxide_disk (Resource)

This resource manages disks.

## Example Usage

```hcl
resource "oxide_disk" "example" {
  organization_name = "staff"
  project_name = "test"
  description = "a test disk"
  name = "mydisk"
  size = 1024
  disk_source = {
    blank = 512
  }
}

resource "oxide_disk" "example2" {
  organization_name = "staff"
  project_name = "test"
  description = "a test disk"
  name = "mydisk2"
  size = 104857600
  disk_source = {
    global_image = "611bb17d-6883-45be-b3aa-8a186fdeafe8"
  }
}
```

## Schema

### Required

- `description` (String) Description for the disk.
- `disk_source` (Map of String) Source of a disk. Can be one of `blank = block_size`, `image = "image_id"`, `global_image = "image_id"`, or `snapshot = "snapshot_id"`.
- `name` (String) Name of the disk.
- `organization_name` (String) Name of the organization.
- `project_name` (String) Name of the project.
- `size` (Number) Size of the disk in bytes.

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `block_size` (Number) Size of blocks in bytes.
- `device_path` (String) Path of the disk.
- `id` (String) Unique, immutable, system-controlled identifier of the disk.
- `image_id` (String) Unique, immutable, system-controlled identifier of the disk source image.
- `project_id` (String) Unique, immutable, system-controlled identifier of the project.
- `snapshot_id` (String) Unique, immutable, system-controlled identifier of the disk source snapshot.
- `state` (List of Object) State of a Disk (primarily: attached or not). (see [below for nested schema](#nestedatt--state))
- `time_created` (String) Timestamp of when this disk was created.
- `time_modified` (String) Timestamp of when this disk was last modified.

<a id="nestedblock--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `default` (String)

<a id="nestedatt--state"></a>

### Nested Schema for `state`

Read-Only:

- `instance` (String)
- `state` (String)

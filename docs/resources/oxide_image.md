---
page_title: "oxide_image Resource - terraform-provider-oxide"
---

# oxide_image (Resource)

This resource manages images.

To create an image it's necessary to set one of `source_url` or `source_snapshot_id`.

!> This resource does not support updates or deletes.

## Example Usage

```hcl
resource "oxide_image" "example" {
  project_id  = "c1dee930-a8e4-11ed-afa1-0242ac120002"
  description = "a test image"
  name        = "myimage"
  source_url  = "myimage.example.com"
  block_size  = 512
  os          = "alpine"
  version     = "3.15"
}

resource "oxide_image" "example2" {
  project_id         = "c1dee930-a8e4-11ed-afa1-0242ac120002"
  description        = "a test image"
  name               = "myimage2"
  source_snapshot_id = "eb65d5cb-d8c5-4eae-bcf3-a0e89a633042"
  block_size         = 512
  os                 = "ubuntu"
  version            = "20.04"
  timeouts = {
    read   = "1m"
    create = "3m"
  }
}
```

## Schema

### Required

- `project_id` (String) ID of the project that will contain the instance.
- `description` (String) Description for the image.
- `os` (String) OS image distribution. Example: "alpine".
- `version` (String) OS image version. Example: "3.16".
- `name` (String) Name of the image.
- `block_size` (Number) Size of blocks in bytes.

### Optional

- `source_snapshot_id` (String) Snapshot ID of the image source if applicable. To be set only when creating an image from a snapshot.
- `source_url` (String) "URL of the image source if applicable. To be set only when creating an image from a URL.
- `timeouts` (Attribute, Optional) (see [below for nested schema](#nestedatt--timeouts))

### Read-Only

- `digest` (Object) Hash of the image contents, if applicable (see [below for nested schema](#nestedobject--digest)).
- `id` (String) Unique, immutable, system-controlled identifier of the image.
- `size` (Number) Total size in bytes.
- `time_created` (String) Timestamp of when this image was created.
- `time_modified` (String) Timestamp of when this image was last modified.

<a id="nestedatt--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `create` (String, Default `10m`)
- `read` (String, Default `10m`)

### Nested Schema for `digest`

Read-Only:

- `type` (String) Digest type.
- `value` (String) Digest type value.

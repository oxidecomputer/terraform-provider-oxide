---
page_title: "oxide_global_image Resource - terraform-provider-oxide"
---

# oxide_global_image (Resource)

This resource manages global images.

!> This resource currently only provides create and read actions. Once update and delete endpoints have been added to the API, that functionality will be added here as well.

## Example Usage

```hcl
resource "oxide_global_image" "example" {
  description          = "a test global_image"
  name                 = "myglobalimage"
  image_source         = { url = "myimage.example.com" }
  block_size           = 512
  distribution         = "alpine"
  distribution_version = "3.15"
}

resource "oxide_global_image" "example2" {
  description          = "a test global_image"
  name                 = "myglobalimage2"
  image_source         = { snapshot = "eb65d5cb-d8c5-4eae-bcf3-a0e89a633042" }
  block_size           = 512
  distribution         = "ubuntu"
  distribution_version = "20.04"
}
```

## Schema

### Required

- `description` (String) Description for the global image.
- `distribution` (String) OS image distribution. Example: "alpine".
- `distribution_version` (String) OS image version. Example: "3.16".
- `image_source` (Map of String) Source of an image. Can be one of `url = <URL>` or `snapshot = <snapshot_id>`.
- `name` (String) Name of the global image.
- `block_size` (Number) Size of blocks in bytes.

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `digest` (String) Digest is hash of the image contents, if applicable.
- `id` (String) Unique, immutable, system-controlled identifier of the global image.
- `size` (Number) Total size in bytes.
- `url` (String) URL is URL source of this image, if any.
- `time_created` (String) Timestamp of when this global image was created.
- `time_modified` (String) Timestamp of when this global image was last modified.

<a id="nestedblock--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `default` (String)

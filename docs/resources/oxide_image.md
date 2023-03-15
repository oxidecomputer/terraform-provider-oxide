---
page_title: "oxide_image Resource - terraform-provider-oxide"
---

# oxide_image (Resource)

This resource manages images.

!> This resource currently only provides create and read actions. Once update and delete endpoints have been added to the API, that functionality will be added here as well.

## Example Usage

```hcl
resource "oxide_image" "example" {
  project_id   = "c1dee930-a8e4-11ed-afa1-0242ac120002"
  description  = "a test image"
  name         = "myimage"
  image_source = { url = "myimage.example.com" }
  block_size   = 512
  os           = "alpine"
  version      = "3.15"
}

resource "oxide_image" "example2" {
  project_id   = "c1dee930-a8e4-11ed-afa1-0242ac120002"
  description  = "a test image"
  name         = "myimage2"
  image_source = { snapshot = "eb65d5cb-d8c5-4eae-bcf3-a0e89a633042" }
  block_size   = 512
  os           = "ubuntu"
  version      = "20.04"
}
```

## Schema

### Required

- `project_id` (String) ID of the project that will contain the instance.
- `description` (String) Description for the image.
- `os` (String) OS image distribution. Example: "alpine".
- `version` (String) OS image version. Example: "3.16".
- `image_source` (Map of String) Source of an image. Can be one of `url = <URL>` or `snapshot = <snapshot_id>`.
- `name` (String) Name of the image.
- `block_size` (Number) Size of blocks in bytes.

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `digest` (String) Digest is hash of the image contents, if applicable.
- `id` (String) Unique, immutable, system-controlled identifier of the image.
- `size` (Number) Total size in bytes.
- `url` (String) URL is URL source of this image, if any.
- `time_created` (String) Timestamp of when this image was created.
- `time_modified` (String) Timestamp of when this image was last modified.

<a id="nestedblock--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `default` (String)

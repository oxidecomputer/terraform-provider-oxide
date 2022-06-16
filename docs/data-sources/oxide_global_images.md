---
page_title: "oxide_global_images Data Source - terraform-provider-oxide"
---

# oxide_global_images (Data Source)

Retrieve a list of all global images.

## Example Usage

```hcl
data "oxide_global_images" "example" {}
```

## Schema

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `global_images` (List of Object) A list of all global images (see [below for nested schema](#nestedatt--global_images))
- `id` (String) The ID of this resource.

<a id="nestedblock--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `default` (String)

<a id="nestedatt--global_images"></a>

### Nested Schema for `global_images`

Read-Only:

- `block_size` (Number) Block size in bytes.
- `description` (String) Description of the image.
- `digest` (Map of String) Hash of the image contents, if applicable.
- `distribution` (String) Image distribution.
- `id` (String) Unique, immutable, system-controlled identifier for the image.
- `name` (String) Name of the image.
- `size` (Number) Size of the image in bytes.
- `time_created` (String) Timestamp of when this image was created.
- `time_modified` (String) Timestamp of when this image was last modified.
- `url` (String) URL source of this image, if any.
- `version` (String) Image version.

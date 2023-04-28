---
page_title: "oxide_images Data Source - terraform-provider-oxide"
---

# oxide_images (Data Source)

Retrieve a list of all images belonging to a project.

## Example Usage

```hcl
data "oxide_images" "example" {}
```

## Schema

### Required

- `project_id` (String) ID of the project that contains the images.

### Optional

- `timeouts` (Attribute, Optional) (see [below for nested schema](#nestedatt--timeouts))

### Read-Only

- `images` (List of Object) A list of all global images (see [below for nested schema](#nestedatt--images))
- `id` (String) The ID of this resource.

<a id="nestedatt--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `read` (String, Default `10m`)

<a id="nestedatt--images"></a>

### Nested Schema for `images`

Read-Only:

- `block_size` (Number) Block size in bytes.
- `description` (String) Description of the image.
- `digest` (Map of String) Hash of the image contents, if applicable.
- `os` (String) OS image distribution.
- `id` (String) Unique, immutable, system-controlled identifier for the image.
- `name` (String) Name of the image.
- `size` (Number) Size of the image in bytes.
- `time_created` (String) Timestamp of when this image was created.
- `time_modified` (String) Timestamp of when this image was last modified.
- `url` (String) URL source of this image, if any.
- `version` (String) Version of the OS.

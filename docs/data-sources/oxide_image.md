---
page_title: "oxide_image Data Source - terraform-provider-oxide"
---

# oxide_image (Data Source)

Retrieve information about a specified image.

## Example Usage

```hcl
data "oxide_image" "example" {
  project_name = "my-project"
  name         = "my-image"
  timeouts = {
    read = "1m"
  }
}
```

## Schema

### Required

- `project_name` (String) Name of the project that contains the image.
- `name` (String) Name of the image.

### Optional

- `timeouts` (Attribute, Optional) (see [below for nested schema](#nestedatt--timeouts))

### Read-Only

- `block_size` (Number) Block size in bytes.
- `description` (String) Description of the image.
- `digest` (Object) Hash of the image contents, if applicable (see [below for nested schema](#nestedobject--digest)).
- `id` (String) Unique, immutable, system-controlled identifier for the image.
- `os` (String) OS image distribution.
- `project_id` (String) ID of the project that contains the image.
- `size` (Number) Size of the image in bytes.
- `time_created` (String) Timestamp of when this image was created.
- `time_modified` (String) Timestamp of when this image was last modified.
- `url` (String) URL source of this image, if any.
- `version` (String) Version of the OS.

<a id="nestedatt--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `read` (String, Default `10m`)

<a id="nestedobject--digest"></a>

### Nested Schema for `digest`

Read-Only:

- `type` (String) Digest type.
- `value` (String) Digest type value.

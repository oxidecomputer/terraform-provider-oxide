---
page_title: "oxide_instance_disk_attachment Resource - terraform-provider-oxide"
---

# oxide_instance_disk_attachment (Resource)

This resource manages instance disk attachments.

!> This resource does not delete disks, only detaches them.

-> Associated instance must be in a stopped state before attempting to attach disks to it.

## Example Usage

```hcl
resource "oxide_instance_disk_attachment" "sample_attachment" {
  disk_id     = "c1dee930-a8e4-11ed-afa1-0242ac120002"
  instance_id = "611bb17d-6883-45be-b3aa-8a186fdeafe8"
  timeouts = {
    read   = "1m"
    create = "3m"
    delete = "2m"
  }
}
```

## Schema

### Required

- `instance_id` (String) ID of the instance the disk will be attached to.
- `disk_id` (String) ID of the disk to be attached.

### Optional

- `timeouts` (Attribute, Optional) (see [below for nested schema](#nestedatt--timeouts))

### Read-Only

- `id` (String) Unique, immutable, system-controlled identifier of the terraform resource.
- `disk_name` (String) Name of the disk that is attached to the designated instance.

<a id="nestedatt--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `create` (String, Default `10m`)
- `delete` (String, Default `10m`)
- `read` (String, Default `10m`)

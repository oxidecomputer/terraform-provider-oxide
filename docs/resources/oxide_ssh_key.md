---
page_title: "oxide_ssh_key Resource - terraform-provider-oxide"
---

# oxide_ssh_key (Resource)

This resource manages SSH keys.

## Example Usage

```hcl
resource "oxide_ssh_key" "example" {
  name        = "example"
  description = "Example SSH key."
  public_key  = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIE1clIQrzlQNqxgvpCCUFFOcTTFDOaqV+aocfsDZvxqB"
}
```

## Schema

### Required

- `name` (String) Name of the SSH key. Names must begin with a lower case ASCII
  letter, be composed exclusively of lowercase ASCII, uppercase ASCII, numbers,
  and `-`, and may not end with a `-`. Names cannot be a UUID though they may
  contain a UUID.
- `description` (String) Description for the SSH key.
- `public_key` (String) The SSH public key (e.g., `ssh-ed25519 AAAAC3NzaC...`).

### Optional

- `timeouts` (Attribute, Optional) (see [below for nested schema](#nestedatt--timeouts))

### Read-Only

- `id` (String) Unique, immutable, system-controlled identifier of the SSH key.
- `silo_user_id` (String) The ID of the user to whom this SSH key belongs.
- `time_created` (String) Timestamp of when this SSH key was created.
- `time_modified` (String) Timestamp of when this SSH key was last modified.

<a id="nestedatt--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `create` (String, Default `10m`)
- `delete` (String, Default `10m`)
- `read` (String, Default `10m`)
- `update` (String, Default `10m`)

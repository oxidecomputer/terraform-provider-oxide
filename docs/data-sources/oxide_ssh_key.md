---
page_title: "oxide_ssh_key Data Source - terraform-provider-oxide"
---

# oxide_ssh_key (Data Source)

Retrieve information about a specified SSH key.

## Example Usage

```hcl
data "oxide_ssh_key" "example" {
  name = "example"
  timeouts = {
    read = "1m"
  }
}
```

## Schema

### Required

- `name` (String) Name of the SSH key.

### Optional

- `timeouts` (Attribute, Optional) (see [below for nested schema](#nestedatt--timeouts))

### Read-Only

- `id` (String) Unique, immutable, system-controlled identifier of the SSH key.
- `description` (String) Description for the SSH key.
- `public_key` (String) The SSH public key (e.g., `ssh-ed25519 AAAAC3NzaC...`).
- `silo_user_id` (String) The ID of the user to whom this SSH key belongs.
- `time_created` (String) Timestamp of when this SSH key was created.
- `time_modified` (String) Timestamp of when this SSH key was last modified.

<a id="nestedatt--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `read` (String, Default `10m`)

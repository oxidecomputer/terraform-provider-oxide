---
page_title: "oxide_silo Resource - terraform-provider-oxide"
---

# oxide_silo (Resource)

This resource manages the creation of an Oxide Silo.

-> Only the `quotas` attribute supports in-place modification. Changes to other attributes will result in the silo being destroyed and created anew.

## Example Usage

```hcl
resource "oxide_silo" "example" {
  description      = "a test silo"
  name             = "{{.SiloName}}"
  admin_group_name = "test_admin"
  identity_mode    = "saml_jit"
  discoverable     = true
  mapped_fleet_roles = {
    admin  = ["admin", "collaborator"]
    viewer = ["viewer"]
  }
  quotas = {
    cpus    = 8
    memory  = 34359738368
    storage = 34359738368
  }
  tls_certificates = [
    {
      name        = "silo_cert_1"
      description = "test cert 1"
      cert        = file("cert.pem")
      key         = file("key.pem")
      service     = "external_api"
    },
  ]
}
```

## Schema

### Required

- `name` (String) Name of the Oxide Silo.
- `discoverable` (Boolean) Whether this silo is present in the silo_list output. Defaults to `true`.
- `identity_mode` (String) Describes how identities are managed and users are authenticated in this Silo. Only valid values are `saml_jit` and `local_only`.
- `quotas` (Set of Object) Limits the amount of provisionable CPU, memory, and storage in the Silo. (see [below for nested schema](#nestedatt--quotas))
- `description` (String) Description for the Oxide Silo.
- `tls_certificates` (String) Initial TLS certificates to be used for the new Silo's console and API endpoints. (see [below for nested schema](#nestedatt--tls))

### Optional

- `admin_group_name` (String) Admin group name for the Oxide Silo. Identity providers can assert that users belong to this group and those users can log in and further initialize the Silo.
- `mapped_fleet_roles` (Set of Object) Mapping of which Fleet roles are conferred by each Silo role. The default is that no Fleet roles are conferred by any Silo roles unless there's a corresponding entry in this map.

### Read-Only

- `id` (String) Unique, immutable, system-controlled identifier of the Silo.
- `time_created` (String) Timestamp of when this Silo was created.
- `time_modified` (String) Timestamp of when this Silo was last modified.

<a id=""></a>

### Nested Schema for `quotas`

### Required

- `cpus` (Number) The amount of virtual CPUs available for running instances in the Silo
- `memory` (Number) The amount of RAM (in bytes) available for running instances in the Silo
- `storage` (Number) The amount of storage (in bytes) available for disks or snapshots

<a id="nestedatt--quotas"></a>

### Nested Schema for `tls_certificates`

### Required

- `cert` (String) PEM-formatted string containing public certificate chain.
- `description` (String) Description of the certificate.
- `key` (String) PEM-formatted string containing private key.
- `name` (String) The name associated with the certificate.
- `service` (String) The service associated with the certificate. Only valid value is `external_api`.

<a id="nestedatt--tls"></a>

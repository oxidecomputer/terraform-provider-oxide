---
page_title: "oxide_silo Resource - terraform-provider-oxide"
---

# oxide_silo (Resource)

This resource manages the creation of an Oxide silo.

-> Only the `quotas` attribute supports in-place modification. Changes to other
attributes will result in the silo being destroyed and created anew.

## Example Usage

```hcl
resource "oxide_silo" "example" {
  name             = "showcase"
  description      = "Demo and event silo."
  admin_group_name = "showcase_admin"
  identity_mode    = "saml_jit"
  discoverable     = true
  mapped_fleet_roles = {
    admin  = ["admin", "collaborator"]
    viewer = ["viewer"]
  }
  quotas = {
    cpus    = 64
    memory  = 137438953472 # 128 GiB
    storage = 549755813888 # 512 GiB
  }
  tls_certificates = [
    {
      name        = "wildcard_cert"
      description = "Wildcard cert for *.sys.oxide.example.com."
      cert        = file("cert.pem")
      key         = file("key.pem")
      service     = "external_api"
    },
  ]
}
```

## Schema

### Required

- `name` (String) Name of the Oxide silo.
- `description` (String) Description for the Oxide silo.
- `quotas` (Set of Object) Limits the amount of provisionable CPU, memory, and storage in the silo. (See [below for nested schema](#nestedatt--quotas).)
- `tls_certificates` (String, Write-only) Initial TLS certificates to be used for the new silo's console and API endpoints. This attribute is a [write-only attribute](https://developer.hashicorp.com/terraform/plugin/framework/resources/write-only-arguments) and can only be modified by updating its configuration. (https://developer.hashicorp.com/terraform/cli/state/taint). (See [below for nested schema](#nestedatt--tls).)
- `discoverable` (Boolean) Whether this silo is present in the silo_list output. Defaults to `true`.

### Optional

- `identity_mode` (String) How identities are managed and users are authenticated in this silo. Valid values are `saml_jit` and `local_only`. Defaults to `local_only`.
- `admin_group_name` (String) This group will be created during silo creation and granted the "Silo Admin" role. Identity providers can assert that users belong to this group and those users can log in and further initialize the Silo.
- `mapped_fleet_roles` (Map) Setting that defines the association between silo roles and fleet roles. By default, silo roles do not grant any fleet roles. To establish a connection, you create entries in this map. The key for each entry must be a silo role: `admin`, `collaborator`, or `viewer`. The value is a list of fleet roles (`admin`, `collaborator`, or `viewer`) that the key silo role will grant.
- `timeouts` (Attribute, Optional) Timeouts for performing API operations. See [below for nested schema](#nestedatt--timeouts).
- `service` (String) The service associated with the certificate. The only valid value is `external_api`. Defaults to `external_api`.

### Read-Only

- `id` (String) Unique, immutable, system-controlled identifier of the silo.
- `time_created` (String) Timestamp of when this Silo was created.
- `time_modified` (String) Timestamp of when this Silo was last modified.

<a id="nestedatt--quotas"></a>

### Nested Schema for `quotas`

### Required

- `cpus` (Number) The amount of virtual CPUs available for running instances in the silo.
- `memory` (Number) The amount of RAM, in bytes, available for running instances in the silo.
- `storage` (Number) The amount of storage, in bytes, available for disks or snapshots.

<a id="nestedatt--tls"></a>

### Nested Schema for `tls_certificates`

### Required

- `name` (String) The name associated with the certificate.
- `description` (String) Description of the certificate.
- `cert` (String) PEM-formatted string containing public certificate chain.
- `key` (String) PEM-formatted string containing private key.

<a id="nestedatt--timeouts"></a>

### Nested Schema for `timeouts`

#### Optional

- `create` (String, Default `10m`)
- `delete` (String, Default `10m`)
- `read` (String, Default `10m`)
- `update` (String, Default `10m`)

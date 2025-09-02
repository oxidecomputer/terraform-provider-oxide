---
page_title: "oxide_silo_saml_identity_provider Resource - terraform-provider-oxide"
---

# oxide_silo_saml_identity_provider (Resource)

Manages a SAML identity provider (IdP) for an Oxide silo.

-> This resource does not support updates. All attributes are immutable once
created.

-> This resource does not support deletion from the Oxide API. When destroyed in
Terraform, it will be removed from state but will continue to exist in Oxide.

## Example Usage

### With URL Metadata Source

```hcl
resource "oxide_silo_saml_identity_provider" "example" {
  silo                    = oxide_silo.example.id
  name                    = "keycloak"
  description             = "Managed by Terraform."
  group_attribute_name    = "groups"
  idp_entity_id           = "https://keycloak.example.com/realms/oxide"
  acs_url                 = "https://example.com/saml/acs"
  slo_url                 = "https://example.com/saml/logout"
  sp_client_id            = "oxide-sp"
  technical_contact_email = "admin@example.com"

  idp_metadata_source = {
    type = "url"
    url  = "https://keycloak.example.com/realms/oxide/protocol/saml/descriptor"
  }
}
```

### With Base64-Encoded XML Metadata

```hcl
resource "oxide_silo_saml_identity_provider" "example" {
  silo                    = oxide_silo.example.id
  name                    = "custom-idp"
  description             = "Custom SAML identity provider"
  idp_entity_id           = "https://idp.example.com"
  acs_url                 = "https://example.com/saml/acs"
  slo_url                 = "https://example.com/saml/logout"
  sp_client_id            = "oxide-sp"
  technical_contact_email = "admin@example.com"

  idp_metadata_source = {
    type = "base64_encoded_xml"
    data = base64encode(file("${path.module}/idp-metadata.xml"))
  }

  signing_keypair = {
    private_key = base64encode(file("${path.module}/saml-key.pem"))
    public_cert = base64encode(file("${path.module}/saml-cert.pem"))
  }
}
```

## Schema

### Required

- `acs_url` (String) URL where the identity provider should send the SAML response.
- `description` (String) Free-form text describing the SAML identity provider.
- `idp_entity_id` (String) Identity provider's entity ID.
- `idp_metadata_source` (Attributes) Source of identity provider metadata (URL or base64-encoded XML). (see [below for nested schema](#nestedatt--idp_metadata_source))
- `name` (String) Unique, immutable, user-controlled identifier of the SAML identity provider. Maximum length is 63 characters.
- `silo` (String) Name or ID of the silo.
- `slo_url` (String) URL where the identity provider should send logout requests.
- `sp_client_id` (String) Service provider's client ID.
- `technical_contact_email` (String) Technical contact email for SAML configuration.

### Optional

- `group_attribute_name` (String) SAML attribute that holds a user's group membership.
- `signing_keypair` (Attributes) RSA private key and public certificate for signing SAML requests. (see [below for nested schema](#nestedatt--signing_keypair))
- `timeouts` (Attributes) (see [below for nested schema](#nestedatt--timeouts))

### Read-Only

- `id` (String) Unique, immutable, system-controlled identifier of the SAML identity provider.
- `time_created` (String) Timestamp of when this SAML identity provider was created.
- `time_modified` (String) Timestamp of when this SAML identity provider was last modified.

<a id="nestedatt--idp_metadata_source"></a>
### Nested Schema for `idp_metadata_source`

Required:

- `type` (String) The type of metadata source. Must be one of: `url`, `base64_encoded_xml`.

Optional:

- `data` (String) Base64-encoded XML metadata (required when type is `base64_encoded_xml`). Conflicts with `url`.
- `url` (String) URL to fetch metadata from (required when type is `url`). Conflicts with `data`.

<a id="nestedatt--signing_keypair"></a>
### Nested Schema for `signing_keypair`

Required:

- `private_key` (String, Sensitive) RSA private key (base64 encoded).
- `public_cert` (String) Public certificate (base64 encoded).

<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).

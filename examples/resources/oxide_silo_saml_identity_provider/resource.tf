# With URL metadata source.
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

# With base64-encoded XML metadata.
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
    data = filebase64("${path.module}/idp-metadata.xml")
  }

  signing_keypair = {
    private_key = filebase64("${path.module}/saml-key.der")
    public_cert = filebase64("${path.module}/saml-cert.der")
  }
}

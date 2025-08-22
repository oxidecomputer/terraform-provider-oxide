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

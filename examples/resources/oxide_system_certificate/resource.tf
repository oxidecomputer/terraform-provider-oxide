resource "oxide_system_certificate" "example" {
  silo_id     = "1fec2c21-cf22-40d8-9ebd-e5b57ebec80f"
  name        = "wildcard-cert-2026"
  description = "Wildcard cert for *.sys.oxide.example.com."
  cert        = file("cert.pem")
  key         = file("key.pem")
  service     = "external_api"
}

resource "tls_private_key" "example" {
  algorithm = "RSA"
  rsa_bits  = 2048
}

resource "tls_self_signed_cert" "example" {
  private_key_pem       = tls_private_key.example.private_key_pem
  validity_period_hours = 8760
  dns_names             = ["*.sys.oxide.example.com"]
  allowed_uses          = ["key_encipherment", "digital_signature", "server_auth"]

  subject {
    common_name = "*.sys.oxide.example.com"
  }
}

resource "oxide_system_certificate" "example" {
  silo_id     = "1fec2c21-cf22-40d8-9ebd-e5b57ebec80f"
  name        = "wildcard-cert-2026"
  description = "Wildcard cert for *.sys.oxide.example.com."
  cert        = tls_self_signed_cert.example.cert_pem
  key         = tls_private_key.example.private_key_pem
  service     = "external_api"

  lifecycle {
    create_before_destroy = true
  }
}

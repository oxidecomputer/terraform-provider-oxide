# Allocate a floating IP from the silo's default IP pool
resource "oxide_floating_ip" "example" {
  project_id  = "5476ccc9-464d-4dc4-bfc0-5154de1c986f"
  name        = "app-ingress"
  description = "Ingress for application."
}

# Allocate a floating IP from the specified IP pool 
resource "oxide_floating_ip" "example_with_pool" {
  project_id  = "5476ccc9-464d-4dc4-bfc0-5154de1c986f"
  name        = "app-ingress"
  description = "Ingress for application."
  ip_pool_id  = "a4720b36-006b-49fc-a029-583528f18a4d"
}

# Allocate a specific floating IP from the specified IP pool
resource "oxide_floating_ip" "example_with_address" {
  project_id  = "5476ccc9-464d-4dc4-bfc0-5154de1c986f"
  name        = "app-ingress"
  description = "Ingress for application."
  ip_pool_id  = "a4720b36-006b-49fc-a029-583528f18a4d"
  ip          = "172.21.252.128"
}

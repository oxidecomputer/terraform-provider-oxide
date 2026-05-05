# Allocate a floating IP from the silo's default IP pool
resource "oxide_floating_ip" "example" {
  project_id  = data.oxide_project.my_project.id
  name        = "app-ingress"
  description = "Ingress for application."
  ip_version  = "v4"
}

# Allocate a floating IP from the specified IP pool
resource "oxide_floating_ip" "example_with_pool" {
  project_id  = data.oxide_project.my_project.id
  name        = "app-ingress-from-pool"
  description = "Ingress for application."
  ip_pool_id  = data.oxide_ip_pool.default.id
}

# Allocate a specific floating IP from the specified IP pool
resource "oxide_floating_ip" "example_with_address" {
  project_id  = data.oxide_project.my_project.id
  name        = "app-ingress-static"
  description = "Ingress for application."
  ip          = "10.0.1.42"
}

# Prerequisites for the example.
data "oxide_project" "my_project" {
  name = "my-project"
}

data "oxide_ip_pool" "default" {
  name = "default"
}

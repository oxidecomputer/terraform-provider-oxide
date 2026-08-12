resource "oxide_system_ip_pool_silo_link" "example" {
  pool       = "my-pool"
  silo       = "my-silo"
  is_default = false
}

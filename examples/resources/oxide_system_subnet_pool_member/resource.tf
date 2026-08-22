resource "oxide_system_subnet_pool_member" "example" {
  pool              = "my-pool"
  subnet            = "192.0.2.0/24"
  min_prefix_length = 24
  max_prefix_length = 28
}

resource "oxide_system_subnet_pool_member" "example" {
  subnet_pool_id    = "3e2c6e84-bed8-4c94-afc3-1032082d6a90"
  subnet            = "192.0.2.0/24"
  min_prefix_length = 24
  max_prefix_length = 28
}

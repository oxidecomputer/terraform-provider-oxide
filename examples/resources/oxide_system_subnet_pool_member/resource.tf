resource "oxide_system_subnet_pool" "example" {
  name        = "my-subnet-pool"
  description = "Example system subnet pool."
  ip_version  = "v4"
}

resource "oxide_system_subnet_pool_member" "example" {
  subnet_pool_id    = oxide_system_subnet_pool.example.id
  subnet            = "192.0.2.0/24"
  min_prefix_length = 24
  max_prefix_length = 28
}

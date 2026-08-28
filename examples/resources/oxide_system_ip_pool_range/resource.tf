data "oxide_system_ip_pool" "example" {
  pool = "my-pool"
}

resource "oxide_system_ip_pool_range" "example" {
  ip_pool_id    = data.oxide_system_ip_pool.example.id
  first_address = "172.20.18.227"
  last_address  = "172.20.18.239"
}

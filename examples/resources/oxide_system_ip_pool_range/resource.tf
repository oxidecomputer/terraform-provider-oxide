resource "oxide_system_ip_pool_range" "example" {
  pool  = "my-pool"
  first = "172.20.18.227"
  last  = "172.20.18.239"
}

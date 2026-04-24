resource "oxide_ip_pool_silo_link" "example" {
  silo_id    = "1fec2c21-cf22-40d8-9ebd-e5b57ebec80f"
  ip_pool_id = data.oxide_ip_pool.default.id
  is_default = true
  timeouts = {
    read   = "1m"
    create = "3m"
    delete = "2m"
    update = "2m"
  }
}

# Prerequisites for the example.
data "oxide_ip_pool" "default" {
  name = "default"
}

resource "oxide_ip_pool_silo_link" "example" {
  silo_id    = "1fec2c21-cf22-40d8-9ebd-e5b57ebec80f"
  ip_pool_id = "081a331d-5ee4-4a23-ac8b-328af5e15cdc"
  is_default = true
  timeouts = {
    read   = "1m"
    create = "3m"
    delete = "2m"
    update = "2m"
  }
}

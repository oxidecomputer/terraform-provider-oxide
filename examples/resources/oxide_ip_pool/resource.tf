resource "oxide_ip_pool" "example" {
  description = "a test IP pool"
  name        = "myippool"
  ranges = [
    {
      first_address = "172.20.18.227"
      last_address  = "172.20.18.239"
    }
  ]
  timeouts = {
    read   = "1m"
    create = "3m"
    delete = "2m"
    update = "2m"
  }
}

data "oxide_ip_pool" "example" {
  name = "default"
  timeouts = {
    read = "1m"
  }
}

data "oxide_silo" "example" {
  name = "default"
  timeouts = {
    read = "1m"
  }
}

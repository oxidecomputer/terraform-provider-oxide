data "oxide_ssh_key" "example" {
  name = "example"
  timeouts = {
    read = "1m"
  }
}

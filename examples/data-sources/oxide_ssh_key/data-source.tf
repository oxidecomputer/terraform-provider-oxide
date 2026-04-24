data "oxide_ssh_key" "example" {
  name = "my-key"
  timeouts = {
    read = "1m"
  }
}

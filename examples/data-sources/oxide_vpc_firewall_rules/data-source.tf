data "oxide_vpc_firewall_rules" "example" {
  project = "my-project"
  vpc     = "my-vpc"

  timeouts = {
    read = "1m"
  }
}

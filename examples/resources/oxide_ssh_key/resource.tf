resource "oxide_ssh_key" "example" {
  name        = "example"
  description = "Example SSH key."
  public_key  = "ssh-ed25519 AAAC3NzaC1lZI1NTE5AAAAIE1clIQrzlQNqxgvpCCUFOcTTFDOaqV+aocfsDZxqB"
}

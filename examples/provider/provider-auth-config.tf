variable "oxide_token" {
  type        = string
  description = "Oxide API token."
  sensitive   = true
}
provider "oxide" {
  host  = "https://oxide.sys.example.com"
  token = "oxide-token-XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
}

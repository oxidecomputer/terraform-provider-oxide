locals {
  credentials = provider::oxide::credentials(pathexpand("~/.config/oxide/credentials.toml"))

  # Alternatively, pass null or an empty string to use the default
  # configuration file path.
  # credentials = provider::oxide::credentials(null)

  prod_host      = local.credentials["prod"].host
  prod_api_token = local.credentials["prod"].token
}

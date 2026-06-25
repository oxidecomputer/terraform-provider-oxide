locals {
  credentials = provider::oxide::credentials(null)

  prod_host      = local.credentials["prod"].host
  prod_api_token = local.credentials["prod"].token
}

# IP Pools

This example Terraform configuration file performs the following:

1. Creates an IP pool.

To try it out make sure you have the following:

- Previously set `OXIDE_HOST` and `OXIDE_TOKEN` environment variables

Although not recommended, if you do not wish to set the environment variables, you can use the `host` and `token` fields in the `oxide` provider block:

```hcl
provider "oxide" {
  host = "<your-host>"
  token = "<your-token>"
}
```

# Instance with attached disks

This example Terraform configuration file performs the following:

1. Creates two blank disks.
2. Creates an instance and attach the previous two disks to it.

To try it out make sure you have the following:

- An organization named `corp`.
- A project within the `corp` organization called `test`.
- Previously set `OXIDE_HOST` and `OXIDE_TOKEN` environment variables

Alternatively, you can modify the configuration file with your own `organization_name` and `project_name`. Although not recommended, if you do not wish to set the environment variables, you can use the `host` and `token` fields in the `oxide` provider block:

```hcl
provider "oxide" {
  host = "<your-host>"
  token = "<your-token>"
}
```

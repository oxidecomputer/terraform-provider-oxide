# Create basic instance

This example Terraform configuration file performs the following:

1. Creates a basic instance without attached disks or network interfaces.

To try it out make sure you have the following:

- A project.
- Previously set `OXIDE_HOST` and `OXIDE_TOKEN` environment variables

Alternatively, you can modify the configuration file with your own `project_id`. Although not recommended, if you do not wish to set the environment variables, you can use the `host` and `token` fields in the `oxide` provider block:

```hcl
provider "oxide" {
  host = "<your-host>"
  token = "<your-token>"
}
```

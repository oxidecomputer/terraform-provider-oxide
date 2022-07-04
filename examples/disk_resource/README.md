# Disks

This example Terraform configuration file performs the following:

1. Create a blank disk.
2. Create a disk from a global image.

To try it out make sure you have the following:

- Verify that the `global_image` ID within `disk_source` matches the global image you want to use.
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

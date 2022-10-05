# Basic Rack Setup Example

This Terraform configuration file sets up the basic elements on a rack to be able to run the [demo configuration file](../demo/):

1. Creates an organization called "myorg".
2. Creates a project called "myproj" in the "myorg" organization.
3. Creates an IP pool named "mypool" and adds an IP range (172.20.15.227 - 172.20.15.239).
4. Creates several global images.

_IMPORTANT: Currently there is no way to delete a global image. This means that you cannot run `terraform destroy` on this configuration file._

To try out this configuration file follow the [instructions](https://github.com/oxidecomputer/terraform-provider-oxide/#using-the-provider) from the README.

# Terraform Provider Oxide Demo

To try out the provider you'll need to follow these steps:

- Have a Nexus server running.
- Make sure you have set the `$OXIDE_HOST` and `$OXIDE_TOKEN` environment variables.
- Have Terraform [installed](https://www.terraform.io/downloads) locally.
- Run `make install` from the root of this repository. This will install the plugin in your local Terraform plugins directory.
- From the `examples/` directory run `terraform init` and `terraform apply` to use the example Terraform configuration file.

## Running acceptance tests

To run the acceptance testing suite, you need to make sure to have either the `$OXIDE_HOST` and `$OXIDE_TOKEN` environment variables, or `$OXIDE_TEST_HOST` and `$OXIDE_TEST_TOKEN`. If you wish to use the later for a testing environment, make sure you have unset the previous first.

Until all resources have been added you'll need to make sure your testing environment has the following:

- An organization named "corp".
- A project within the "corp" organization named "test".
- At least one global image.

Run `make testacc`.

Eventually we'll have a GitHub action to create a Nexus server and run these tests, but for now testing will have to be run manually.

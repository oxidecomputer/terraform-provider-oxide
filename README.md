# Terraform Provider Oxide Demo

The [scope](https://docs.google.com/document/d/1TNvy5-aqZPcv1PQllzySSIV7KJsGl5llvfc2L1IiD-Y/edit?usp=sharing) of this provider is to only have the necessary [resources](https://www.terraform.io/language/resources) and [data sources](https://www.terraform.io/language/data-sources) for the on-site demo.

**IMPORTANT:** The resources for this provider currently only provide create, read and delete actions. Once update endpoints have been added to the API, that functionality will be added here as well. That said, this is not a blocker for the demo. The requirements for the demo only specify resource creation.

## Requirements

- [Terraform](https://www.terraform.io/downloads) 0.1.x and above, we recommend using the latest stable release whenever possible. When installing on an Illumos machine use the Solaris binary.
- [Go 1.18](https://go.dev/dl/) (to build the provider plugin)

## Building the provider

There are two make targets to build the provider.

`make install` will build the binary in the `/bin` directory in the root of this repository, and install the provider in your local Terraform plugins directory. Using this target will enable you to use the provider you've just built with your own Terraform configuration files.

There is a caveat when installing with an Apple M1 computer. When building the same version with changes you will have to manually delete the previous binary, or change the version in the Makefile.

`make build` will only build the binary in the `/bin` directory. Terraform will not know to look for the provider there, and will not work with Terraform configuration files.

## Using the provider

The documentation for each of the resources and data sources can be found in the `docs/` directory. The formatting may look a bit odd, but it needs to be compliant with the Terraform documentation format for when we publish the provider.

To try out the provider you'll need to follow these steps:

- Make sure you've installed the provider using `make install`.
- Have a Nexus server running.
- Set the `$OXIDE_HOST` and `$OXIDE_TOKEN` environment variables. If you do not wish to use these variables, you have the option to set the host and token directly on the provider block. For security reasons this approach is not recommended.
  ```hcl
  provider "oxide" {
    host = "<host>"
    token = "<token>"
  }
  ```
- Pick an example From the `examples/` directory and cd into it, or create your own Terraform configuration file. You can change the values of the fields of the example files to work with your environment.
- Run `terraform init` and `terraform apply` from within the chosen example directory. This will create resources or read data sources based on a Terraform configuration file.
- To remove all created resources run `terraform destroy`.

**IMPORTANT: Given the update functionality is not enabled; when changing a .tf configuration file on a resource where you have already run `terraform apply`, you MUST run a `terraform destroy` first.**

To try out the demo configuration file, use the [examples/demo/](./examples/demo/) directory.

When trying out the same example with a provider you've recently built with changes, make sure to remove all the files from the example Terraform generated first.

## Running the linters

Before getting started make sure you've set the `$GOBIN` environment variable to where your go binaries are stored and add it to `$PATH`.

Example:

```console
export GOBIN=/Users/username/go/bin
export PATH=${PATH}:/Users/username/go/bin
```

There is a make target to run the linters. All that's needed is `make lint`.

## Running acceptance tests

To run the acceptance testing suite, you need to make sure to have either the `$OXIDE_HOST` and `$OXIDE_TOKEN` environment variables, or `$OXIDE_TEST_HOST` and `$OXIDE_TEST_TOKEN`. If you wish to use the later for a testing environment, make sure you have unset the previous first.

Until all resources have been added you'll need to make sure your testing environment has the following:

- An organization named "corp".
- A project within the "corp" organization named "test".
- At least one global image.

Run `make testacc`.

Eventually we'll have a GitHub action to create a Nexus server and run these tests, but for now testing will have to be run manually.

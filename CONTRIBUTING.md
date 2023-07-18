# Contributing

## Requirements

- [Terraform](https://www.terraform.io/downloads) 1.x and above, we recommend using the latest stable release whenever possible. When installing on an Illumos machine use the Solaris binary.
- [Go](https://go.dev/dl/) 1.20.x and above (to build the provider plugin)

## Building the provider

There are two make targets to build the provider.

`make install` will build the binary in the `./bin` directory in the root of this repository, and install the provider in your local Terraform plugins directory. Using this target will enable you to use the provider you've just built with your own Terraform configuration files.

There is a caveat when installing with an Apple M1 computer. When building the same version with changes you will have to manually delete the previous binary, or change the version in the Makefile.

`make build` will only build the binary in the `./bin` directory. Terraform will not know to look for the provider there, and will not work with Terraform configuration files.

## Using the provider

To try out the provider you'll need to follow these steps:

- Make sure you've installed the provider using `make install`.
- Set the `$OXIDE_HOST` and `$OXIDE_TOKEN` environment variables. If you do not wish to use these variables, you have the option to set the host and token directly on the provider block. For security reasons this approach is not recommended.
  ```hcl
  provider "oxide" {
    host = "<host>"
    token = "<token>"
  }
  ```
- Pick an example From the `examples/` directory and cd into it, or create your own Terraform configuration file. You can change the values of the fields of the example files to work with your environment.
- If you want to try out your local changes, make sure you set the version to the one you just built. This will generally be the current version with "-dev" appended (e.g. `version = "0.1.0-beta-dev"`).
- Run `terraform init` and `terraform apply` from within the chosen example directory. This will create resources or read data sources based on a Terraform configuration file.
- To remove all created resources run `terraform destroy`.

To try out the demo configuration file, use the [examples/demo/](./examples/demo/) directory.

When trying out the same example with a provider you've recently built with changes, make sure to remove all the files from the example Terraform generated first.

## Running the linters

There is a make target to run the linters. All that's needed is `make lint`.

## Running acceptance tests

To run the acceptance testing suite, you need to make sure to have either the `$OXIDE_HOST` and `$OXIDE_TOKEN` environment variables, or `$OXIDE_TEST_HOST` and `$OXIDE_TEST_TOKEN`. If you wish to use the later for a testing environment, make sure you have unset the previous first.

Until all resources have been added you'll need to make sure your testing environment has the following:

- A project named "tf-acc-test".
- At least one image.

Run `make testacc`.

Eventually we'll have a GitHub action to create a Nexus server and run these tests, but for now testing will have to be run manually.

## Releasing a new version

Remember to update the changelog before releasing.

There is a Github action that takes care of creating the artifacts and publishing to the terraform registry.

To trigger the process create a tag from the local main branch and push to Github.

```console
$ git tag v0.1.0
$ git push origin v0.1.0
```

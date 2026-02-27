# Contributing

## Requirements

- [Terraform](https://www.terraform.io/downloads) 1.x and above, we recommend using the latest stable release whenever possible. When installing on an Illumos machine use the Solaris binary.
- [Go](https://go.dev/dl/) 1.20.x and above (to build the provider plugin)

## Building the provider

There are two make targets to build the provider.

`make install` will build the binary using the `go install` command and place
it in the path defined by the environment variables  `$GOBIN`, `$GOPATH/bin`,
or `$HOME/go/bin`, depending on which ones are defined. Refer to `go help
install` for more information.

Once the binary is installed, you will need to create or modify the file
`$HOME/.terraformrc` to point the Oxide provider to its installation path. For
example:

```hcl
provider_installation {
  dev_overrides {
    "registry.terraform.io/oxidecomputer/oxide" = "$HOME/go/bin"
  }

  direct {}
}
```

Refer to the [Terraform
documentation](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-provider#prepare-terraform-for-local-provider-install)
for more information.

`make build` will only build the binary in the `./bin` directory. Terraform
will not know to look for the provider there, and will not work with Terraform
configuration files.

### Building with local SDK changes or other versions

To use the Terraform provider with a local `oxide.go` Go SDK run `make local-api`.
This target assumes both the `oxide.go` and `terraform-provider-oxide` repositories
are checked out to adjacent directories (e.g., share the same parent directory).

```
.
├── oxide.go
└── terraform-provider-oxide
```

To undo those changes run `make unset-local-api`.

To use a specific version of the Go SDK run `SDK_V={GIT_HASH|VERSION} make sdk-version`.

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

To run the acceptance testing suite, you need to make sure to have either the
`OXIDE_HOST` and `OXIDE_TOKEN`, or the `OXIDE_PROFILE` environment variables
exported.

Until all resources have been added you'll need to make sure your testing environment has the following:

- A project named "tf-acc-test".
- At least one image.

Tests that exercise the `oxide_silo` resource need a tls cert that's
valid for the domain of the Oxide server used for acceptance tests. The
tests will generate a self-signed cert, but need to know which DNS name
to use for it. We default to the \*.sys.oxide-dev.test wildcard used by
the [simulated omicron
environment](https://github.com/oxidecomputer/omicron/blob/main/docs/how-to-run-simulated.adoc).
To override when testing against a different environment, set the
`$OXIDE_TEST_SILO_DNS_NAME` environment variable to the relevant DNS name.

Run `make testacc`.

To run tests against an empty simulated omicron environment, first provision
the Docker containers with `make testacc-sim` and run the test suite with `make
testacc-local`.

To run specific test cases, set the test name pattern as the variable
`TEST_ACC_NAME` .

```
TEST_ACC_NAME=TestAccCloudDataSourceInstanceExternalIPs_full make testacc
TEST_ACC_NAME=TestAccCloudDataSourceInstanceExternalIPs_full make testacc-local
```

## Releasing a new version

Steps prefixed with `.0 ->` only need to be done for major and minor releases
where the patch version is (`vX.Y.0`).

- [ ] Create release branch `release-vX.Y.Z`.
  ```
  git checkout main
  git pull origin main
  git checkout -b release-vX.Y.Z
  ```
- [ ] Update `CHANGELOG.md`.
  ```
  make changelog
  ```
- [ ] `.0 ->` Update the `Build status` table in `README.md` to point to the new release line branch.
- [ ] Commit changes and open a PR.
  ```
  git add CHANGELOG.md README.md
  git commit -m 'release vX.Y.Z'
  git push origin release-vX.Y.Z
  ```
- [ ] From the release branch, create and push release tag.
  ```
  git checkout release-vX.Y.Z
  git tag vX.Y.Z
  git push origin vX.Y.Z
  ```
- [ ] `.0 ->` From the release branch, create and push the new release line branch.
  ```
  git checkout release-vX.Y.Z
  git checkout -b rel/vX.Y
  git push origin rel/vX.Y
  ```
- [ ] Approve and monitor the [Release](https://github.com/oxidecomputer/terraform-provider-oxide/actions/workflows/release.yml) workflow.
- [ ] Update the [Release](https://github.com/oxidecomputer/terraform-provider-oxide/releases) description with the changelog content for the version and publish the release.
- [ ] `.0 ->` Create new backport [label](https://github.com/oxidecomputer/terraform-provider-oxide/labels) with color `#b1d0c7` and named `backport/vX.Y`.

## Backporting changes

The repository is organized with multiple release branches, each targeting a
specific release line. The release branches are named `rel/vX.Y` where `X.Y`
represents the release line version.

Pull requests should target the `main` branch and be backported to release
lines as necessary.

To backport a PR to the branch `rel/vX.Y` add the label
`backport/vX.Y` to the PR. Once merged, the backport automation will create a
new PR backporting the changes to the release branch. The backport label can
also be added after the PR is merged.

If a backport has merge conflicts, the conflicts are committed to the PR and
you can checkout the branch to fix them. Once the changes are clean, you can
merge the backport PR.

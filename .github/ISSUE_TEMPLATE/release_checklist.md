---
name: Release checklist
about: Steps to take when releasing a new version (only for Oxide release team).
labels: release

---

## Release checklist
<!--
 Please follow all of these steps in the order below.
 After completing each task put an `x` in the corresponding box,
 and paste the link to the relevant PR.
-->
- [ ] Update the Terraform configuration version constraints in the following files to use the version you want to release.
    - [ ] [`examples`](https://github.com/oxidecomputer/terraform-provider-oxide/tree/main/examples)
    - [ ] [`docs`](https://github.com/oxidecomputer/terraform-provider-oxide/tree/main/docs)
- [ ] Update the following files with the version you want to release.
    - [ ] [`VERSION`](https://github.com/oxidecomputer/terraform-provider-oxide/blob/main/VERSION)
    - [ ] [`internal/provider/version.go`](https://github.com/oxidecomputer/terraform-provider-oxide/blob/main/oxide/version.go)
- [ ] Update the `github.com/oxidecomputer/oxide.go` dependency to the latest release.
- [ ] Run the acceptance tests against an environment with the same Omicron version as the Go SDK version the provider is on.
- [ ] Generate the `CHANGELOG.md` file.
    - [ ] Run `make changelog`
    - [ ] Add the date of the release to the title
- [ ] Release the new version draft by running `make tag`. Approve and monitor the `Release` GitHub Actions workflow.
- [ ] Verify the release is correct, it's being created from the correct tag on GitHub, add the changelog entry to the release notes, and make the release live. Note: Terraform registry will silently fail to publish if the tag is incorrect, and GitHub has a habit of messing up the tag a release is created from on occasion. 
- [ ] Verify the release is available on the Terraform provider registry.
- [ ] If this is a major or minor release, create a new `MAJOR.MINOR` branch from the newly created tag.
- [ ] Update the following files with the upcoming version.
    - [ ] [`VERSION`](https://github.com/oxidecomputer/terraform-provider-oxide/blob/main/VERSION)
    - [ ] [`internal/provider/version.go`](https://github.com/oxidecomputer/terraform-provider-oxide/blob/main/oxide/version.go)
- [ ] Create new changelog file in [.changelog/](https://github.com/oxidecomputer/terraform-provider-oxide/blob/main/changelog/) on the relevant branches.

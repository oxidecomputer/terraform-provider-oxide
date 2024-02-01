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
- [ ] Make sure all examples and docs reference the new provider version.
- [ ] Make sure the [VERSION](https://github.com/oxidecomputer/terraform-provider-oxide/blob/main/VERSION) and [internal/provider/version.go](https://github.com/oxidecomputer/terraform-provider-oxide/blob/main/oxide/version.go) files have the new version you want to release.
- [ ] Make sure the `oxide` SDK dependency is up to date with the latest release.
- [ ] Generate changelog by running `make changelog` and add date of the release to the title.
- [ ] Release the new version draft by running `make tag`.
- [ ] Verify the release is correct on GitHub and make the release live.
- [ ] Verify the release is available on the Terraform provider registry.
- [ ] If this is not a minor patch, create a new branch with the current version
- [ ] Update to upcoming version in [VERSION](https://github.com/oxidecomputer/terraform-provider-oxide/blob/main/VERSION), [internal/provider/version.go](https://github.com/oxidecomputer/terraform-provider-oxide/blob/main/oxide/version.go), and
    create new changelog tracker file in [.changelog/](https://github.com/oxidecomputer/terraform-provider-oxide/blob/main/changelog/) for the relevant branches.
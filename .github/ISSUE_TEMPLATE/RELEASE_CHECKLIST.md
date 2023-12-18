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
- [ ] Make sure version is set to the new version in VERSION and version.go files.
- [ ] Make sure the `oxide` SDK dependency is up to date with the latest release.
- [ ] Generate changelog by running `make changelog` and add date of the release to the title.
- [ ] Release the new version by running `make tag`.
- [ ] If this is not a minor patch, create a new branch with the current version
- [ ] Update to upcoming version in `VERSION`, `internal/provider/version.go`, and
    create new changelog tracker file in .changelog/ for the relevant branches.
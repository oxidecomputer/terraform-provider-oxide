---
page_title: "oxide_instance Resource - terraform-provider-oxide"
---

# oxide_instance (Resource)

This resource manages instances.

## Example Usage

```hcl
resource "oxide_instance" "example" {
  organization_name = "staff"
  project_name      = "test"
  description       = "a test instance"
  name              = "myinstance"
  host_name         = "<host value>"
  memory            = 512
  ncpus             = 1
}
```

## Schema

### Required

- `description` (String) Description for the instance.
- `host_name` (String) Host name of the instance.
- `memory` (Number) Instance memory in bytes.
- `name` (String) Name of the instance.
- `ncpus` (Number) Number of CPUs allocated for this instance.
- `organization_name` (String) Name of the organization.
- `project_name` (String) Name of the project.

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String) Unique, immutable, system-controlled identifier of the instance.
- `project_id` (String) Unique, immutable, system-controlled identifier of the project.
- `run_state` (String) Running state of an Instance (primarily: booted or stopped). This typically reflects whether it's starting, running, stopping, or stopped, but also includes states related to the instance's lifecycle.
- `time_created` (String) Timestamp of when this instance was created.
- `time_modified` (String) Timestamp of when this instance last modified.
- `time_run_state_updated` (String) Timestamp of when the run state of this instance was last modified.

<a id="nestedblock--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `default` (String)

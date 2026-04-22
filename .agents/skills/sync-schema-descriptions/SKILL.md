---
name: sync-schema-descriptions
description: |
  Sources Terraform schema descriptions from the pinned oxide.go SDK's Go doc
  comments, which already mirror the OpenAPI spec. Adapts the text to the TF
  surface (translating reshaped attribute names, clarifying NameOrId inputs,
  trimming API-internal paragraphs, and reconciling any divergence between
  provider and API default behavior) so the descriptions stay accurate and
  aligned with the version of the SDK the provider actually calls into.
---

# Sync Schema Descriptions

## Overview

Terraform schema `Description` / `MarkdownDescription` strings are hand-written
and drift from the canonical descriptions in the OpenAPI spec. The oxide.go
SDK, pinned via `go.mod`, has the same descriptions baked into Go doc comments
on each field. This skill walks through sourcing the right description,
adapting it, and applying the change.

Not every TF attribute maps 1:1 to a single SDK field (e.g. TF's `boot_disk_id`
corresponds to the SDK's richer `boot_disk`), so blind copying is wrong. The
skill codifies the adaptation process rather than automating it blindly.

## Workflow

### 1. Resolve the SDK source

    go list -m -f '{{.Dir}}' github.com/oxidecomputer/oxide.go

Descriptions live in `<dir>/oxide/types.go`. Do not use a local checkout of
the SDK, which might not match the pinned version.

### 2. Identify scope

The user will usually ask about one attribute (e.g. `instance.ssh_public_keys`),
a whole resource, or all resources. For each in-scope attribute, work through
steps 3–8.

### 3. Locate the TF attribute

In `internal/provider/<resource>/resource.go` (or `data_source.go`), find the
schema block and note the current `Description` / `MarkdownDescription`. Find
the TF model struct field via its `tfsdk:"<attr_name>"` tag.

Do not touch schema-version-compat files (`resource_v0.go`, `resource_v1.go`,
…). They represent frozen prior schemas kept for state upgrades; descriptions
there are not rendered to current docs, so the files should stay as they
were when that version shipped.

### 4. Find the matching SDK field

**Input attributes (set in Create/Update):** grep the Create method for the
model field. You'll see a line like `params.Body.SshPublicKeys = ...`. The SDK
struct is the type of `params.Body` (e.g. `oxide.InstanceCreate`); the SDK
field is the LHS accessor.

**Computed/output attributes (populated in Read):** grep the Read method for
how `state.<Field>` is set from the API response — e.g.
`state.TimeCreated = types.StringValue(resp.TimeCreated.String())`. The SDK
struct is the response type (e.g. `oxide.Instance`), the field is the `resp.<X>`
accessor.

If the TF attribute has no clean 1:1 SDK counterpart (flattened, split, or
composed from multiple API fields), stop and surface this to the user — it
needs a hand-written description, not an adaptation.

### 5. Read the SDK doc comment

Open `<sdk_dir>/oxide/types.go`, find `type <StructName> struct`, and read the
Go doc comment on the field. Comments are hard-wrapped; reflow them into a
single logical description.

Also note the field's Go type — it drives adaptation rule (b) below.

### 6. Adapt

**The default is verbatim.** Copy the SDK doc comment's prose into the TF
description word-for-word. Do not insert em dashes, merge sentences,
paraphrase verbs, or restructure for flow. The value of sourcing from the SDK
is fidelity to a single maintained source of truth; editorial drift
reintroduces exactly the hand-written divergence the skill is meant to
eliminate.

Only deviate when one of the rules below applies, and only to the extent the
rule requires. If you aren't sure whether a change is required, keep the
source wording and ask the operator.

**Do not invent text from sibling attributes.** If the SDK has no usable
description for a given TF attribute (e.g. the SDK doc comment is a
tautology like "X is a variant of Y", or the corresponding SDK type/field is
internal wrapping with no prose), the skill's job is done — the attribute
needs a hand-written description, which is outside this skill's scope.
Surface the gap to the operator. Never copy text from a TF sibling
(`v4` → `v6`, `ephemeral` → `floating`, etc.) and pass it off as adapted; it
is still hand-written, just plagiarized.

**a. Translate field-name references.** If the comment references an API field
name that differs in TF (e.g. "set the `boot_disk` attribute"), rewrite it to
the TF name (`boot_disk_id`). Same for any other cross-referenced field.

**b. `NameOrId` phrasing.** The API accepts both names and IDs for
`NameOrId` / `[]NameOrId` inputs. The right phrasing depends on what the TF
attribute represents:

- If the attribute is the resource's own `name` (typically a required
  top-level attribute on a resource or data source), it stores the name.
  Describe as "Name of …".
- If the attribute is a reference to another resource (e.g. `vpc_id`,
  `disk_attachments`, `boot_disk_id`), the provider writes the resolved ID
  back into state on Read, so a user who supplies a name will see a
  perpetual diff (or recreate) on the next plan. Until the provider
  preserves the user-supplied form, keep descriptions honest: say "ID of
  …" / "IDs of …" regardless of whether the SDK field type is `string` or
  `NameOrId`. If the source SDK comment says "name or ID", rewrite to "ID".

**c. Trim non-user-facing paragraphs.** Drop:
- internal API behavior the TF user can't observe
- roadmap/hedging language ("currently, …", "in the future, …")
- information already encoded by the TF schema (e.g. defaults — TF renders
  those separately)

Keep paragraphs describing user-visible behavior: accepted value forms,
interactions with other attributes. Paragraphs about omission/defaults need
verification first — see rule (e).

**d. Preserve TF-specific additions.** If the current TF description contains
information that isn't in the SDK (e.g. cross-refs to other TF attrs, validator
behavior), keep it — it usually reflects provider logic that doesn't exist at
the API level.

**e. Provider defaults may diverge from API defaults.** Any SDK phrase like
"if not provided, …", "if null, …", or "defaults to …" describes what the
_API_ does when the field is absent from the request — but the provider may
not actually omit the field. Helpers like `shared.NewNameOrIdList` return
`[]T{}` (never nil), and whether that serializes as `[]` or gets dropped
depends on the SDK struct's JSON tag (`omitempty` / `omitzero` vs. bare). An
empty list reaching the API is semantically different from an absent field.

If the SDK description includes a default-behavior clause, decide how to
handle it:

- **Inspect the code.** Trace the provider's Create/Update path for the field
  and confirm whether a null TF value produces an omitted request field or a
  zero-value one. If the provider's behavior matches the SDK clause, keep it;
  if it doesn't, rewrite to describe what the provider actually does.
- **Ask the operator.** If the behavior isn't clear from the code, or if the
  mismatch looks like a provider bug worth fixing separately, surface it.
- **Omit the clause.** When in doubt, drop the default-behavior sentence
  entirely rather than risk a misleading description.

**f. Add undocumented TF-specific constraints.** Rule (d) is about keeping
what's already in the current description; this rule is about adding what's
missing. If the schema has validators or plan modifiers whose effect isn't
mentioned in the current description, fold a short clause in — the
framework's auto-generated docs don't surface these. Examples:

- `stringvalidator.LengthBetween(1, 63)` → "Must be 1–63 characters."
- `stringvalidator.RegexMatches(pattern, "…")` → reuse the validator's
  message, or describe the pattern.
- `stringvalidator.OneOf("a", "b", "c")` → "Must be one of `a`, `b`, `c`."

Skip constraints the framework already renders (`Required` / `Optional` /
`Computed` flags, `Default` values — same reasoning as rule (c)).

### 7. Apply, then hand off for review

Apply all proposed changes with `Edit` to the current `resource.go` /
`data_source.go` only (never `resource_v*.go`). Run `make fmt` to regenerate
`docs/resources/*.md` and `docs/data-sources/*.md`, then `make lint` to verify.

Report what was changed in a short summary and tell the operator to review via
`git diff`. The diff is a better review surface for multi-attribute text
changes than an in-chat bulleted list.

### 8. Show the provenance table

At the end of the pass, emit a table with one row per touched attribute:

| Attribute | SDK source (field, verbatim) | Provider description (as applied) |

For each row:
- **SDK source**: the type + field (e.g. `InstanceCreate.Memory`) and the
  doc-comment prose as it appears in `oxide/types.go`, unedited. For rule (f)
  additions, use "(none — rule f)" and name the validator.
- **Provider description**: the final string now in the schema, so the
  operator can see the adaptation deltas inline without chasing through
  `git diff` and the SDK side-by-side.

One row per attribute updated — skip attributes left untouched.

Exceptions — pause and surface to the operator **before** editing rather than
after:
- Step 4 found no clean 1:1 SDK counterpart (reshaped, split, or composed
  attribute) — the description needs to be hand-written, not adapted.
- Rule (e) flagged a behavior-claim mismatch that implies a provider fix is
  needed rather than a description change. Describe the mismatch and ask
  whether to narrow the description to match current provider behavior or open
  a separate issue to fix the provider first.

## Notes

- The SDK pins to a specific Omicron release; the descriptions on disk match
  the API the provider actually calls at runtime. There is no need to fetch
  the OpenAPI spec directly.
- When `go.mod` bumps `oxide.go`, upstream descriptions may change. Re-running
  this skill after an SDK bump is a reasonable drift-detection pass.

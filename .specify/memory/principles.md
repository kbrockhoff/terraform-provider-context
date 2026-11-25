<!--
Sync Impact Report
==================
Version change: (unversioned template) → 1.0.0
Bump rationale: first ratified version; every placeholder replaced with project-specific content.

Modified principles (template placeholder → new title):
- ARCHITECTURE_PRINCIPLE_1 → Compute Deterministically, Without Side Effects
- ARCHITECTURE_PRINCIPLE_2 → Abstract Cloud Differences Behind One Contract
- ARCHITECTURE_PRINCIPLE_3 → Make Every Resource Identifiable and Accountable
- CODE_PRINCIPLE_1 → Keep Business Logic in the Reusable Library
- CODE_PRINCIPLE_2 → Validate Early and Fail Clearly
- CODE_PRINCIPLE_3 → Test the Contract, Not the Implementation
- APPROACH_1 → Library-First Change Flow
- APPROACH_2 → Compatibility-Driven Evolution

Added sections:
- Architecture: Protect Information Exposed Through Tags; Preserve Consumer Compatibility
- Code: Enforce Quality Gates Automatically; Document from the Schema;
  Release Reproducibly and Verifiably
- Approach: Progressive Cloud Coverage
- Governance: full rule set (authority, compliance, complexity, exceptions, amendments,
  operational guidance)

Removed sections: none (template comments replaced by concrete content)

Sources used: README.md, docs/design.md, .agents/agents.md (Terraform module, style and
test guides), CLAUDE.md, Makefile, .github/workflows/*.yml, .golangci.yml, .goreleaser.yml,
pkg/context/*.go, internal/**.go, examples/.

Templates and commands:
- ✅ .specify/templates/plan-template.md — "Principles Check" reads gates from this file
  generically; no change required.
- ✅ .specify/templates/spec-template.md — no principle-specific mandatory sections; no change.
- ✅ .specify/templates/tasks-template.md — task categories remain generic; no change.
- ✅ .claude/commands/iac.*.md — no outdated principle references found.
- ✅ .agents/agents.md — replaced the generic module template with the project-specific
  AGENTS.md content (now identical to AGENTS.md).
- ⚠ docs/design.md — component tree lists internal/models and tests/ which do not exist,
  and omits pkg/context; update to reflect the library-first layout.

Follow-up TODOs:
- pkg/context/git.go convertSSHToHTTPS does not strip userinfo from HTTPS remotes
  (https://user:token@host/...), which conflicts with "Protect Information Exposed Through
  Tags"; file a bug bead.
- No acceptance tests (internal/provider, internal/datasource) exist yet; required by
  "Test the Contract, Not the Implementation" at the Enhanced level.
-->

# Context Terraform Provider Principles

## Cloud Architecture Principles

### Compute Deterministically, Without Side Effects

The `brockhoff_context` data source is evaluated on every plan in every consuming
configuration. If its outputs drift between runs, every downstream resource sees a perpetual
tag diff, and teams stop trusting their plans. The provider therefore MUST behave as a pure
function: the same provider configuration and data source inputs MUST yield the same
`name_prefix`, `tags`, and `data_tags`. It MUST NOT create, read, or modify cloud resources,
MUST NOT require cloud credentials, and MUST NOT persist anything beyond what Terraform
stores for a data source.

- **Baseline**: Outputs depend only on declared inputs. Every input that is not declared
  (wall-clock time for an `Ephemeral` deletion date, local git repository metadata) is
  isolated in one function, documented, and can be overridden by an explicit input or turned
  off with a feature toggle.
- **Enhanced**: Changes that add a new implicit input require a documented justification and a
  way for consumers to pin the value, so that `Production` and `MissionCritical` configurations
  can make their plans fully reproducible.

### Abstract Cloud Differences Behind One Contract

Consumers write one data source block and receive tags that are valid on the target cloud.
Each cloud has different limits on tag key and value characters, maximum length, delimiter, and
the value used for "not applicable". These differences MUST stay behind a single cloud provider
contract, so that naming and tagging logic never branches on the cloud identifier itself.

- **Baseline**: Every supported `cloud_provider` identifier resolves to an implementation
  of the contract. Identifiers that have no dedicated implementation resolve to a conservative
  default that produces values every target accepts.
- **Enhanced**: A cloud is promoted to a dedicated implementation only when its rules differ
  from the default in a way consumers need, and only together with tests that cover its
  sanitization, key validation and length limits.

### Make Every Resource Identifiable and Accountable

The provider's purpose is resource identity. Its name prefixes and tags drive cost attribution,
ownership lookup, data classification, and automated clean-up across every infrastructure
repository that uses it. A naming or tagging defect is therefore a governance defect in every
consumer. Generated identity MUST follow the Brockhoff naming standard and MUST be predictable
from the inputs.

- **Baseline**: Name prefixes stay within the documented length and character limits.
  Required identity tags (environment, management platform, source repository where it is
  available) are always emitted. Unknown values follow one documented rule: they are
  substituted with the cloud's N/A value, or they are omitted.
- **Enhanced**: For `Production` and `MissionCritical` environment types, ownership, cost
  center, ITSM and data classification tags are first-class outputs. Organization-level tag
  policies are expressible through the provider rather than through ad hoc
  `additional_tags`.

### Protect Information Exposed Through Tags

Tags are visible to anyone with read access to the cloud account, and they are copied into
billing exports, inventories and logs. Anything the provider emits MUST be safe to disclose that
widely.

- **Baseline**: Every tag value is sanitized for the target cloud. Git-derived values (remote
  URL, branch, commit) MUST be normalized to public, credential-free forms: SSH remotes are
  converted to HTTPS and any embedded userinfo or token is stripped. The provider holds no
  credentials and writes no secrets to state.
- **Enhanced**: Inputs that look like secrets (tokens, keys, passwords) are rejected rather
  than emitted, and the sensitivity and data regulation tags accurately reflect the declared
  classification.

### Preserve Consumer Compatibility

This provider replaces the `kbrockhoff/terraform-external-context` module and is also
consumed as a Go library (`pkg/context`) by non-Terraform clients. Any change to an output
propagates to the tags of every resource in every consuming repository.

- **Baseline**: For equivalent inputs, outputs match the module's outputs unless a documented,
  versioned change says otherwise. Schema attributes and exported Go identifiers are not
  removed or renamed outside a MAJOR release.
- **Enhanced**: Behavior changes that alter existing tag keys or values ship behind an
  opt-in flag or a deprecation period first, with migration notes in the changelog.

## IaC Code Principles

### Keep Business Logic in the Reusable Library

Naming, tagging, validation, cloud rules, and git integration live in `pkg/context`, a
framework-independent Go package that is shared by the provider and by standalone clients
(see `examples/client-app`). The provider packages (`internal/provider`,
`internal/datasource`) are thin adapters that map `terraform-plugin-framework` schema and types
to library calls. `internal/core` exists only as a backward-compatible set of re-exports and
MUST NOT gain new logic. When logic is duplicated in an adapter, it diverges, and Terraform users
and library users end up receiving different tags for the same inputs.

### Validate Early and Fail Clearly

Invalid identity inputs MUST stop the plan with a diagnostic that names the attribute, the value
it received, and the rule it broke. A valid-looking prefix that is silently truncated or a tag
that is silently dropped is not acceptable. Validation of enumerations (cloud provider,
environment type, availability, sensitivity), formats (namespace, environment, emails, deletion
date), and length is done once, in the library, and the schema validators reuse it. The
exception is environmental context the consumer cannot control, such as git metadata that is
unavailable: the provider logs a warning and continues without those tags.

### Test the Contract, Not the Implementation

Tests assert what consumers depend on: the generated prefixes, tag maps, alternate output formats,
and validation errors for a given set of inputs. They do not assert internal helper structure.
Unit tests in `pkg/context` are table-driven, they run with `-race`, and the target is more
than 90% coverage. Each supported cloud and each special rule (ephemeral deletion date, system
prefixes, N/A substitution, owner and source repository toggles) has explicit cases. Acceptance
tests exercise the data source through Terraform, using the configurations in `examples/`. These
are plan-only tests and never need cloud credentials.

### Enforce Quality Gates Automatically

Quality is checked by machines before review. Every change MUST pass `gofmt -s` and `goimports`,
`go vet`, `golangci-lint` (with the version pinned in CI), `go mod verify`, and unit tests with
race detection. CI runs these checks on every pull request. `make check` and `make pre-commit`
reproduce them locally. Fix lint findings rather than suppress them. A suppression needs an
inline justification.

### Document from the Schema

Attribute descriptions in the provider schema are the single source of reference documentation.
Registry docs are generated with `tfplugindocs` (`make docs-generate`), and every data source has
a runnable example under `examples/`. A schema change is incomplete until the generated docs,
the README attribute tables, and at least one example reflect it.

### Release Reproducibly and Verifiably

Consumers install this provider from the Terraform Registry and trust its checksums. Releases
are cut only through release-please and GoReleaser from `main`. They use semantic versioning
derived from Conventional Commits, and every artifact is signed with GPG. The Go toolchain and CI
actions are pinned to explicit versions. Dependabot updates them, and they merge only when CI
passes.

## Implementation Approaches

### Library-First Change Flow

Every functional change follows the same order. First change `pkg/context` and its tests.
Then wire the change into the data source schema and model. Last, update generated docs and
examples. This keeps the library usable without Terraform and makes the adapter diff small and
reviewable. Use this flow for all new tags, inputs, output formats, and validation rules. Pure
plumbing changes, such as framework upgrades, CI, and release tooling, may skip the library step.

### Compatibility-Driven Evolution

Classify each change before you build it:

- A new optional input, a new output, or a new cloud implementation is additive: a MINOR
  release.
- A bug fix that makes outputs match the documented standard or the legacy module is a PATCH
  release, and the changelog calls out every tag value it changes.
- Removing or renaming an attribute or an exported identifier, or changing default tag keys or
  values, is breaking: a MAJOR release, preceded by a deprecation warning in at least one MINOR
  release.

Pick the least disruptive classification that still delivers the change.

### Progressive Cloud Coverage

Start every cloud identifier on the default implementation. Promote a cloud to a dedicated
implementation (as `aws`, `az` and `gcp` are today) only when a consumer needs its specific
limits, and include the tests required by the cloud contract principle. This keeps the
supported-cloud list broad without carrying untested, speculative rule sets.

## Governance

**Authority and Precedence**: These principles govern every specification, plan, task list and
code change in this repository. When they conflict with agent guidance in `.agents/agents.md`,
`CLAUDE.md`, or generic Terraform module guidance, these principles take precedence for this
provider.

**Compliance and Accountability**: The plans produced by `/iac.plan` MUST pass a Principles
Check against each principle above. Code review and the CI quality gates verify ongoing
compliance. Reviewers reject changes that add logic outside `pkg/context`, add undocumented
implicit inputs, or change outputs without a matching version classification.

**Justification for Complexity**: A new dependency, a new implicit input, a new cloud
implementation, or a new output format needs a written rationale in the plan or the pull request
that explains the consumer need it serves.

**Deviation and Exception Process**: Any deviation is recorded in the plan's complexity tracking
table, or in a beads issue, together with its rationale and the conditions for removing it.
Exceptions are reviewed when the principles are next amended.

**Amendment and Evolution**: Amendments are made through `/iac.principles` and follow semantic
versioning: MAJOR for removing or redefining a principle, MINOR for adding a principle or
materially expanding one, and PATCH for clarifications. Each amendment updates the Sync Impact
Report and the dependent templates.

**Relationship to Operational Guidance**: These principles set out WHAT and WHY.
`CLAUDE.md`, `.agents/agents.md`, the `Makefile` and `docs/design.md` cover the day-to-day
HOW, and change more often to follow tooling.

**Version**: 1.0.0 | **Ratified**: 2026-11-24 | **Last Amended**: 2026-11-24

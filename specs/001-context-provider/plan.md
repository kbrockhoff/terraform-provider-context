# Architecture Plan: Context Provider for Standardized Naming and Tagging

**Branch**: `001-context-provider` | **Date**: 2026-11-24 | **Spec**: [spec.md](./spec.md)
**Input**: Infrastructure specification from `/specs/001-context-provider/spec.md`

**Note**: This template is filled in by the `/iac.plan` command. For deep research, module
specs, and quickstart guide, run `/iac.enrichplan` after planning.

## Summary

The Context provider is a credential-free Terraform provider with a single data source,
`brockhoff_context`. It returns three outputs:

- a standardized name prefix;
- a resource tag set;
- a data tag set.

Tag values are formatted for the target cloud. Contexts support parent/child inheritance. All
of the logic is also available as the standalone Go library `pkg/context`.

The code already implements most of the spec. This plan is a **conformance plan**: it closes
the gaps between the spec and the existing code. The gaps fall into these groups:

1. **Legacy parity (gaps 2–6)**:
   - omit absent PM and ITSM tags;
   - environment tag fallback;
   - `never` as the default deletion date;
   - fix `dataowners` when owner tags are off;
   - `propagate_at_launch` in list-of-maps outputs.
2. **Security and robustness**:
   - remove credentials from repository URLs;
   - warn when git information is missing;
   - truncate values without splitting characters;
   - enforce tag key rules;
   - report errors against the attribute that caused them.
3. **New capability (FR-028)**: an optional data sovereignty input, emitted as the
   `sovereignty` data tag only when it is supplied.
4. **Test coverage**:
   - acceptance tests and data source tests, which do not exist today;
   - a benchmark;
   - a test that the git cache works under concurrent use.
5. **Documentation and CI hygiene**:
   - environment type documentation;
   - migration notes for gap 1;
   - pinned tool versions in CI.

Every change follows the library-first flow (principles: *Library-First Change Flow*). The
library changes first, then the data source adapter, then the docs and examples.

## Technical Context

**Cloud Provider**: None. The provider is cloud-agnostic and calls no cloud APIs. It encodes
tag rules for four targets: `aws`, `az`, `gcp`, and a default rule set used by `dc`, `oci`,
`ibm`, `do`, `vul`, `ali` and `cv`.
**IaC Tool**: Terraform CLI, plugin protocol 6.0. Local and CI testing use Terraform 1.16.5,
the current stable release. Any Terraform version ≥ 1.0 that speaks protocol 6 is supported.
**Provider Versions** (Go dependencies, latest stable versions checked through the Go module
proxy):

- Go toolchain: 1.27.4, pinned in `go.mod` and in the CI `GO_VERSION`
- `github.com/hashicorp/terraform-plugin-framework`: v1.19.0, already the latest
- `github.com/hashicorp/terraform-plugin-log`: v0.10.0, already in use
- `github.com/hashicorp/terraform-plugin-testing`: v1.16.0, **new** test-only dependency for
  acceptance tests
- `terraform-plugin-framework-validators` is **not** added. Validation stays in `pkg/context`
  (principles: *Validate Early and Fail Clearly*).

**Module Versions**: Not applicable. This project ships a provider, not modules.
**Curated Modules**: Not applicable. `kbrockhoff/terraform-external-context` (the legacy module)
is the reference implementation for parity.
**State Backend**: Not applicable. The provider has no state of its own. Data source outputs
live in each consumer's state.
**Environment Strategy**: The pipeline has three stages: a local developer install
(`make install` writes `.terraformrc` with a dev override), then CI on pull requests, then a
release to the public registry from `main` through release-please and GoReleaser.
**Testing**:

- `go test -race` for `pkg/context` unit tests, which are table-driven
- terraform-plugin-testing acceptance tests in `internal/provider`, which need no cloud
  credentials and are gated by `TF_ACC=1`
- `make test-examples` to validate the examples
- Go benchmarks for the performance SLO

**Security Scanning**:

- `golangci-lint` v2.15
- `gosec`, pinned to a release tag instead of `@master`, with SARIF upload
- `go mod verify`

**Cost Estimation**: Not applicable. Runtime cost is $0 and CI runs on hosted runners for a
public repository.
**Target Environments**: The provider has no environments of its own. It serves consumers in
all seven environment types (spec FR-027). Releases target the platforms in `.goreleaser.yml`:
freebsd, linux, darwin and windows on amd64, 386, arm and arm64, excluding darwin/386 and
windows/arm. This is a superset of the five platforms the spec requires.
**Compliance**:

- Terraform Registry publishing: a GPG-signed `SHA256SUMS` file and a
  `terraform-registry-manifest.json` declaring protocol 6.0
- Apache License 2.0
- Owner email addresses are personal data and are controlled by the consumer through
  `owner_tags_enabled`

## Principles Check

Gates come from `.specify/memory/principles.md` v1.0.0.

| Principle | Status | How this plan satisfies it |
|-----------|--------|----------------------------|
| Compute Deterministically, Without Side Effects | ✅ with justified exceptions | No network calls and no resources. The ephemeral date (wall clock) and git metadata (local repository) are the only implicit inputs. Each lives in one function and can be overridden by `deletion_date` or `source_repo_tags_enabled = false`. Git lookup is made injectable for tests. See Complexity Tracking. |
| Abstract Cloud Differences Behind One Contract | ✅ | All rules stay behind the `CloudProvider` interface. Key validation (`ValidateTagKey`) is now actually called. Rune-safe truncation lives in shared code, not in each cloud implementation. |
| Make Every Resource Identifiable and Accountable | ✅ | `environment`, `availability`, `managedby` and `deletiondate` are always emitted. N/A rules are limited to the seven tags in FR-020. Name prefix rules are unchanged. |
| Protect Information Exposed Through Tags | ✅ after fix | `sourcerepo` is rebuilt from a parsed URL with userinfo removed (bead `tfpctx-6ce`). Additional-tag keys are validated and all values sanitized. |
| Preserve Consumer Compatibility | ✅ | Gaps 2–6 move outputs to match the legacy module. Gap 1 and name non-inheritance are documented divergences with migration notes. The sovereignty addition (C17) is opt-in, so contexts that do not set it are unchanged. The output changes ship together as one MINOR release with per-tag changelog entries. Library API changes are additive only. |
| Keep Business Logic in the Reusable Library | ✅ | All behavior changes land in `pkg/context`. The data source only maps types, reports diagnostics and logs warnings. The provider's duplicate cloud list is replaced with `ValidateCloudProvider`. `internal/core` only gets new aliases. |
| Validate Early and Fail Clearly | ✅ | Diagnostics move to `AddAttributeError` with attribute paths. Bad tag keys and bad prefixes fail the read. Missing git information only produces a `tflog.Warn`. |
| Test the Contract, Not the Implementation | ✅ after tasks | Acceptance tests are added (bead `tfpctx-24o`). Every FR-009 to FR-023 rule gets a test case, with a traceability table in tasks. The `pkg/context` coverage gate is ≥ 90%. |
| Enforce Quality Gates Automatically | ✅ | The existing CI gates stay. A coverage threshold check is added to the test job. |
| Document from the Schema | ✅ | Schema descriptions are expanded (environment type with the RPO/RTO table, `deletion_date` with `never`). Docs are regenerated with `tfplugindocs` v0.25.0, and the README tables and examples are updated. |
| Release Reproducibly and Verifiably | ⚠ → ✅ | CI currently uses `terraform_version: latest`, `gosec@master` and `tfplugindocs@latest`. This plan pins all three. |

## Infrastructure Architecture

### Component Design

The runtime is the Terraform CLI talking over gRPC (protocol 6) to the provider binary
(`main.go`). Inside the binary, the layers are:

- `internal/provider`: provider schema and Configure. Produces a `ProviderConfig` with
  `cloud_provider` and `tag_prefix`.
- `internal/datasource`: schema, the parent/child merge, the model ↔ config mapping,
  diagnostics and `tflog` calls.
- `internal/core`: type aliases and re-exports only.
- `pkg/context`: the shared library, also used by `examples/client-app`. It contains:
  - `naming.go`: `NameGenerator`
  - `validation.go`: the `Validate*` functions and allowed values
  - `tags.go`: `TagProcessor` and format converters
  - `cloud.go`: the `CloudProvider` interface and its implementations
  - `git.go`: `GetGitInfo`, with a cache and URL normalization

Read pipeline, in order (unchanged except where marked ★):

1. Merge the inputs: child value, then parent value, then default.
2. Validate the inputs. ★ Errors now use attribute paths.
3. Resolve the deletion date. ★ Explicit date, then Ephemeral + 90 days, then `never`.
4. Generate the name prefix.
5. Generate tags. ★ Key validation, rune-safe truncation, and git warnings collected.
6. Generate data tags. ★ `dataowners` fixed.
7. Convert formats. ★ Adds `propagate_at_launch`.
8. Build the resolved context. ★ Environment-name fallback.

#### Change inventory (by spec requirement)

| # | Requirement | Location | Change |
|---|-------------|----------|--------|
| C1 | FR-016 | `pkg/context/tags.go` | Add `ResolveDeletionDate(config)`, which returns the explicit date, else Ephemeral + 90 days, else `never`. `ProcessEphemeralEnvironment` stays as a thin wrapper for API compatibility. |
| C2 | FR-009, FR-016 | `pkg/context/validation.go` | `ValidateDeletionDate` accepts the literal `never`, so that a resolved context passed back in as a parent context still validates. |
| C3 | FR-013 | `pkg/context/tags.go` | The `environment` tag uses EnvironmentName, then Environment, then `dev`. `environment`, `availability`, `managedby` and `deletiondate` are always emitted, whatever the N/A setting. |
| C4 | FR-014 | `pkg/context/tags.go` | `projectmgmtid` is emitted only when a PM project code exists. Each ITSM tag is emitted only when its identifier exists. The platform prefix is applied per identifier. |
| C5 | FR-015, FR-020 | `pkg/context/tags.go` | `dataowners` is gated on `OwnerTagsEnabled` first, then on the N/A rule, the same as the product and code owner tags. N/A substitution is limited to the seven FR-020 keys. |
| C6 | FR-017 | `pkg/context/git.go` | Normalize URLs with `net/url`. Convert SSH forms (`git@host:path` and `ssh://`) to `https://host/path`. Strip `User` from HTTPS URLs and drop the trailing `.git`. Unparseable URLs are dropped and produce a warning. Add a package variable `gitCommand` (an exec seam) for tests. |
| C7 | FR-017 | `pkg/context/tags.go` | `TagProcessor` gets an optional `GitInfoFunc` field (default `GetGitInfo`) and a `Warnings []string` field. If git information is missing, a warning is appended and the git tags are skipped. The `Process()` signature is unchanged. |
| C8 | FR-021 | `pkg/context/tags.go` | A shared `truncateRunes(value, max)` truncates on rune boundaries. Truncation is applied after sanitization, as it is today. |
| C9 | FR-022 | `pkg/context/tags.go` | After the prefix is applied, each key is checked with `CloudProvider.ValidateTagKey`. The first invalid key returns an error that names it. |
| C10 | FR-023 | `pkg/context/tags.go`, `internal/datasource` | `ConvertTagsToListOfMaps` adds `propagate_at_launch: "true"` to each entry. The schema element type (map of strings) is unchanged. |
| C11 | FR-006, FR-013 | `internal/datasource/context.go` | `context_output.environment_name` gets the fallback value. `context_output.environment` stays as supplied. `context_output.deletion_date` holds the resolved value. |
| C12 | FR-009 | `internal/datasource/context.go` | Replace `AddError` with `AddAttributeError(path.Root(attr), …)`. Turn collected `Warnings` into `tflog.Warn` calls. |
| C13 | FR-001 | `internal/provider/provider.go` | Replace the inline allowed-cloud map with `ctx.ValidateCloudProvider`. |
| C14 | FR-027 | `internal/datasource/context.go` | Expand the `environment_type` description with its purpose and the RPO/RTO table. Keep it out of naming in the README and agents guide (bead `tfpctx-am0`). |
| C15 | FR-024, FR-025 | `internal/core/*.go` | Add aliases for `ResolveDeletionDate`. No logic. |
| C17 | FR-004, FR-009, FR-019, FR-028 | `pkg/context/validation.go`, `pkg/context/tags.go`, `internal/core`, `internal/datasource/context.go` | Add the data sovereignty input. The library gets `DataSourceConfig.DataSovereignty` and `ValidateDataSovereignty`, which rejects empty or whitespace-only values and nothing else; ISO 3166-1/3166-2 is a documented convention only. `ProcessDataTags()` emits `sovereignty` only when the value is non-empty, with no N/A path, and never in resource tags. The data source gets the optional `data_sovereignty` attribute at the top level and in `parent_context` and `context_output`, merged with `mergeStringValue`. This addition is new and has no counterpart in the legacy module. |
| C16 | FR-026 | `docs/`, `README.md`, `examples/`, `docs/design.md` | Regenerate docs. Update the tag tables and the deletion date description. Add migration notes covering gap 1 and the non-inheritance of `name`. Fix the `design.md` layout (bead `tfpctx-6he`). Fix the AGENTS claim that Ephemeral overrides an explicit date. |

### Compute Resources

There is no cloud compute.

- **Runtime**: the provider binary runs as a short-lived plugin process that the Terraform CLI
  starts for each command. It is CGO-free and built with `-trimpath`.
- **Concurrency**: Terraform may read many `brockhoff_context` instances in parallel. The git
  cache uses a double-checked `sync.RWMutex`, which stays. It is verified by a race-enabled test
  that runs 100 concurrent `Process()` calls and asserts exactly one invocation of
  `gitCommand` for each git query.
- **CI compute**: GitHub-hosted `ubuntu-latest` runners. The jobs are test, lint, the 5-way
  build matrix, acceptance, examples, security and docs. Each keeps its existing timeout, and
  the test job stays within 15 minutes.
- **Performance budget**: `BenchmarkTagProcessorProcess` uses git disabled and the full input
  set, and the target is under 50 ms per operation. The current in-memory pipeline is expected
  to take well under 1 ms, so the budget has a wide margin.

### Data Storage

No data is persisted.

- **In-process cache**: one `GitInfo` value with a 5-minute TTL. It is process-local and
  discarded when the plugin exits. `ClearGitCache()` stays available for tests.
- **Terraform state**: data source outputs are stored in each consumer's state. No secrets or
  credentials are written: C6 removes the only potential source, URL userinfo.
- **Release artifacts**: zip archives, `SHA256SUMS` and the `.sig` file are stored as GitHub
  release assets and ingested by the Terraform Registry.

### Networking

- **Runtime**: none. The only external interactions are two local subprocess calls:
  `git config --get remote.origin.url` and `git rev-parse HEAD`.
- **Provider protocol**: gRPC over a local socket between the Terraform CLI and the plugin,
  managed by `terraform-plugin-go`.
- **Release path**: a GitHub Actions release job publishes to GitHub Releases, and the
  Terraform Registry picks up the release through its webhook. Module dependencies are
  fetched through the Go module proxy and checked against `go.sum` (`go mod verify`).

### Security

- **Credentials**: the provider has no credential attributes. The Configure step reads only
  `cloud_provider` and `tag_prefix`.
- **Tag content**:
  - C6 strips URL userinfo.
  - Every value, including additional tags, is sanitized per cloud.
  - C9 rejects invalid keys.
  - The provider never emits commit author or email.
- **Subprocess safety**: git runs with fixed arguments and no shell, and its output is trimmed
  and parsed rather than interpolated.
- **Supply chain**:
  - Workflow permissions stay at `contents: read`; only release and docs are elevated.
  - Release checksums are GPG-signed through `crazy-max/ghaction-import-gpg` and GoReleaser
    `signs`.
  - Pin `securego/gosec` to a release tag, `tfplugindocs` to `v0.25.0`, and
    `hashicorp/setup-terraform` to `terraform_version: 1.16.5`.
  - Dependabot keeps the action and Go module versions current. Auto-merge happens only on
    green CI.
- **Static analysis**: golangci-lint (errcheck, govet, ineffassign, staticcheck, unused) and
  gosec SARIF in GitHub code scanning.

### Environment Configuration

The provider has no deployable environments. "Environment" appears in two places.

1. **The consumer-facing environment type (FR-027)**. The provider validates it and passes it
   through. Its only effect on outputs is the Ephemeral deletion date. Consuming modules map
   each type to resiliency and cost defaults using the RPO/RTO table in the spec. Examples
   cover `Ephemeral` (`examples/ephemeral`) and `Production` (`examples/full`).
2. **The development pipeline**:

   | Stage | How | Provider source |
   |-------|-----|-----------------|
   | Local | `make install`, then `make test-examples` / `make acceptance-test` | Dev override in a generated `.terraformrc` |
   | CI | Pull request workflow `ci.yml` | Built from the PR, with `make install` in CI |
   | Release | `release.yml` on release-please tag | Terraform Registry `kbrockhoff/context` |

   Pre-release validation uses `make release-snapshot`, which runs GoReleaser in snapshot mode.

### Complexity Level

**Baseline**, with Enhanced controls where they protect consumers.

- **Baseline** applies to most of the work. The provider is a small, single-purpose binary.
  The changes are local edits to an existing layered design. No new packages, services or
  runtime dependencies are added; `terraform-plugin-testing` is test-only.
- **Enhanced** applies to compatibility and release integrity, because a defect reaches the
  tags of every consuming resource in every repository:
  - legacy parity tests;
  - per-tag changelog entries;
  - signed releases;
  - pinned toolchain versions;
  - the ≥ 90% coverage gate.

Rejected as over-engineering for now:

- a plugin architecture for clouds (the `CloudProvider` interface is enough);
- organization tag policies, multi-parent inheritance and region rules (out of scope in the
  spec);
- a separate validators package.

### State Management

Not applicable to the provider itself.

- The data source is recomputed on every plan, and Terraform stores its outputs in the
  consumer's state, where they are overwritten on refresh. There is no provider-side
  migration and no state upgrade logic.
- **Plan stability**: outputs are deterministic for fixed inputs, commit and calendar day.
  A consumer pinning `deletion_date` (or not using `Ephemeral`) gets fully stable plans.
  Acceptance tests verify this with a second plan step that expects an empty plan.
- **Upgrade impact**: the gap 2–6 changes will show tag diffs on consumers' resources once,
  on the first plan after upgrading. The release notes must say this explicitly.

## Project Structure

### Documentation (this infrastructure)

```text
specs/001-context-provider/
├── spec.md              # Infrastructure specification - /iac.specify
├── plan.md              # This file - architecture plan - /iac.plan
├── checklists/
│   └── requirements.md  # Spec quality checklist
└── tasks.md             # Implementation tasks - /iac.tasks (next)
```

### Source Code (repository root)

```text
.
├── main.go                         # Plugin entry point (protocol 6, registry address)
├── internal/
│   ├── provider/
│   │   ├── provider.go             # C13: reuse ValidateCloudProvider
│   │   ├── provider_test.go        # NEW: schema/config unit tests, ProtoV6 factories
│   │   └── datasource_acc_test.go  # NEW: acceptance tests (TF_ACC), plan-stability steps
│   ├── datasource/
│   │   ├── context.go              # C11, C12, C14, C17
│   │   └── context_test.go         # NEW: merge helpers, context_output mapping
│   └── core/                       # Aliases only (C15)
├── pkg/context/
│   ├── naming.go                   # Unchanged (gap 1 kept)
│   ├── validation.go               # C2, C17
│   ├── tags.go                     # C1, C3–C5, C7–C10, C17
│   ├── cloud.go                    # Unchanged interface; ValidateTagKey now used
│   ├── git.go                      # C6, gitCommand seam
│   ├── *_test.go                   # Extended: parity, security, rune, concurrency
│   └── bench_test.go               # NEW: performance SLO benchmark
├── examples/                       # minimal, full, cloud-providers, ephemeral,
│                                   # parent-child, data-sources, client-app (C16)
├── docs/                           # tfplugindocs output + design.md (C16)
├── templates/                      # tfplugindocs templates (if introduced for RPO/RTO table)
├── .github/workflows/              # ci.yml (pins), release.yml, docs.yml
├── .goreleaser.yml
├── terraform-registry-manifest.json
└── Makefile
```

**Structure Decision**: Keep the existing Go provider layout. None of the template's
Terraform, Pulumi or CloudFormation layouts apply, because this repository ships a provider
binary and a Go library, not infrastructure configuration. New files are test files only, plus
an optional `templates/` directory if the RPO/RTO table needs a tfplugindocs template instead
of a schema description.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| Implicit input: wall-clock date for the Ephemeral `deletiondate` | FR-016, legacy parity. Ephemeral environments must self-expire without manual dates. | Requiring an explicit date would break legacy behavior and push the work onto every consumer. Mitigation: an explicit `deletion_date` always wins, and the value only changes once per day. |
| Implicit input: local git metadata (`sourcerepo`, `sourcecommit`) | FR-017. Traceability from resource to source. | An explicit input would need CI plumbing in every consumer. Mitigation: it can be turned off with `source_repo_tags_enabled`, it is cached, it is injectable for tests, and it warns when missing. |
| Package-level test seam (`gitCommand` variable) | Lets the concurrency/cache SLO be verified deterministically. | Spawning real git in tests is slow and environment-dependent, and cannot count calls. |

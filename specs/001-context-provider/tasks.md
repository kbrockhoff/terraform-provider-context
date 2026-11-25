---
description: "Conformance tasks for the Context provider (spec 001)"
---

# Tasks: Context Provider for Standardized Naming and Tagging

**Input**: Design documents from `/specs/001-context-provider/`
**Prerequisites**: plan.md (change inventory C1–C16), spec.md (FR-001 to FR-027)

**Organization**: This repository ships a Go provider binary, not deployed infrastructure, so
the template's infrastructure tiers become dependency layers of the code:

| Template tier | Layer used here |
|---------------|-----------------|
| Setup | dependencies, CI pins |
| Network (foundation that blocks everything else) | library security and validation primitives |
| Compute & Data | core tagging library |
| Application | provider and data source adapter |
| Application validation | acceptance tests and examples |
| Polish | docs |

Every phase ends with a validation checkpoint. The checkpoints use `go build`, `go vet` and
`go test` instead of `terraform validate`. Acceptance tests and `make test-examples` provide
the Terraform-level validation.

## Format: `[ID] [P?] Description`

- **[ID]**: Sequential task number in execution order
- **[P]**: Can run in parallel with the other [P] tasks in the same phase. These tasks touch
  different files and have no data dependency on each other.
- Paths are relative to the repository root
  (`/Users/kbrockhoff/git/kbrockhoff/terraform-provider-context/reverse-spec`).

## Ground Rules for Every Task

- **Library first**: behavior lives in `pkg/context`. `internal/core` holds aliases only, and
  `internal/datasource` holds type mapping, diagnostics and logging only.
- **Tests**: table-driven `t.Run` subtests, matching the style of the existing `*_test.go`
  files. Write or adjust the test before or alongside the code change.
- **Formatting**: run `gofmt -s -w` and `goimports` on every edited Go file.
- **No commits** unless the user asks. Output-changing commits use the `feat:` type so that
  release-please cuts a MINOR release (principles: *Compatibility-Driven Evolution*).

---

## Phase 1: Setup

**Purpose**: Add the test dependency, run the library tests in CI, and pin unpinned tools
(principles: *Release Reproducibly and Verifiably*).

- [ ] T001 Add the acceptance-test dependency to `go.mod`/`go.sum`: run
  `go get github.com/hashicorp/terraform-plugin-testing@v1.16.0 && go mod tidy`. Confirm that
  `terraform-plugin-framework` stays at v1.19.0.
- [ ] T002 [P] In `.github/workflows/ci.yml`, make these changes:
  - Change the unit test step to
    `go test -v -race -coverprofile=coverage.out ./internal/... ./pkg/...`. Today
    `pkg/context` tests are not run in CI.
  - Add a following step that fails when `pkg/context` statement coverage is below 90%:
    `go test -coverprofile=pkg.out ./pkg/context/... && go tool cover -func=pkg.out | awk '/^total:/ {sub("%","",$3); if ($3+0 < 90) {print "pkg/context coverage " $3 "% < 90%"; exit 1}}'`
  - Replace `terraform_version: latest` with `terraform_version: 1.16.5` in both
    `setup-terraform` steps.
  - Replace `securego/gosec@master` with `securego/gosec@v2.30.0`.
  - Replace `tfplugindocs@latest` with `tfplugindocs@v0.25.0`.
- [ ] T003 [P] In `.github/workflows/docs.yml`, replace `tfplugindocs@latest` with
  `tfplugindocs@v0.25.0`.
- [ ] T004 [P] In `Makefile`, replace every `golangci-lint/cmd/golangci-lint@latest` with
  `@v2.15.0` and every `tfplugindocs@latest` with `@v0.25.0`. These appear in the `lint`,
  `docs-generate`, `docs-validate`, `dev-setup` and `install-tools` targets.
- [ ] T005 Setup checkpoint: run `go mod verify && go build ./... && go vet ./... && go test -race ./...`.
  All existing tests must pass before any behavior changes.

**Checkpoint**: The toolchain is pinned and CI exercises the library.

---

## Phase 2: Foundation (library security and validation primitives)

**Purpose**: Fix the git URL credential leak, make git lookups testable, and accept `never` as
a deletion date. Phase 3 depends on all three.

- [ ] T006 Add table-driven tests in `pkg/context/git_test.go` for a new
  `normalizeRepoURL(raw string) (string, error)`:

  | Input | Expected |
  |-------|----------|
  | `git@github.com:org/repo.git` | `https://github.com/org/repo` |
  | `ssh://git@bitbucket.org/org/repo.git` | `https://bitbucket.org/org/repo` |
  | `ssh://git@host:2222/org/repo.git` | `https://host/org/repo` (port dropped) |
  | `https://user:token@github.com/org/repo.git` | `https://github.com/org/repo` |
  | `https://user@github.com/org/repo` | `https://github.com/org/repo` |
  | `https://github.com/org/repo.git` | `https://github.com/org/repo` |
  | `http://host/org/repo` | `http://host/org/repo` |
  | `""` | error |
  | `"::not a url"` | error |

  Remove the tests that are tied to `convertSSHToHTTPS`.
- [ ] T007 Implement `normalizeRepoURL` in `pkg/context/git.go` using `net/url`:
  1. Rewrite the scp-like `git@host:path` form to `ssh://git@host/path` before parsing.
  2. For the `ssh` scheme, set the scheme to `https` and drop the port.
  3. Always set `u.User = nil`.
  4. Trim the `.git` suffix from `u.Path`.
  5. Return `u.String()`.

  Delete `convertSSHToHTTPS`. Callers use `normalizeRepoURL`; a URL that fails to normalize
  leaves `RepoURL` empty. This covers FR-017 and bead `tfpctx-6ce`.
- [ ] T008 Add a test seam to `pkg/context/git.go`:
  `var gitCommand = func(args ...string) (string, error)`, which runs
  `exec.Command("git", args...).Output()` and trims the output. Then change `GetGitInfo`:
  - Use `gitCommand("config", "--get", "remote.origin.url")` and
    `gitCommand("rev-parse", "HEAD")`.
  - Return a non-nil error, `errGitUnavailable`, when both URL and commit are empty.
  - Cache the result and its error for the TTL, so a missing repository is not re-queried
    within 5 minutes.
- [ ] T009 Add `TestGetGitInfoConcurrentSingleLookup` to `pkg/context/git_test.go`:
  1. Replace `gitCommand` with a counting stub guarded by `sync/atomic`.
  2. Call `ClearGitCache()`.
  3. Start 100 goroutines that call `GetGitInfo()`.
  4. Assert exactly one call per git subcommand and identical results.
  5. Restore the original with `t.Cleanup`.

  The test must pass under `go test -race`. This covers the scalability NFR and the
  performance validation criterion.
- [ ] T010 [P] In `pkg/context/validation.go`, make `ValidateDeletionDate` accept the literal
  `never`. Add the cases `"never"` (valid), `"Never"` (invalid) and `"2025-02-30"` (invalid)
  to `pkg/context/validation_test.go`. This covers C2.
- [ ] T011 Foundation checkpoint: run `go test -race -run 'Git|Validate' ./pkg/context/...`
  and `go vet ./pkg/...`.

**Checkpoint**: Repository URLs are credential-free and git lookups can be faked in tests.

---

## Phase 3: Core Tagging Library (`pkg/context/tags.go`)

**Purpose**: Legacy parity for gaps 2–6, plus FR-016, FR-020, FR-021 and FR-022. Tasks
T013–T021 all edit `tags.go`, so they run sequentially.

- [ ] T012 Create `pkg/context/tags_parity_test.go` with failing tests that encode the legacy
  module's behavior. All of them use `GitInfoFunc` stubbed to return an error, or
  `SourceRepoTagsEnabled=false`. Cases:
  - (a) No PM code or ITSM ids, with N/A on: none of `projectmgmtid`, `systemid`,
    `componentid`, `instanceid` is present.
  - (b) A PM platform without a code: `projectmgmtid` is absent.
  - (c) Environment-name fallback:
    - `EnvironmentName="Production"` gives `environment=Production`.
    - Only `Environment="prd"` gives `prd`.
    - Neither set gives `dev`.
    - All three hold with N/A off.
  - (d) `deletiondate`:
    - is `never` for `Development` with no date;
    - is the explicit date when set;
    - is a fixed clock + 90 days for `Ephemeral`.
  - (e) `OwnerTagsEnabled=false` with `DataOwners` set and N/A on: `dataowners` is absent.
  - (f) N/A on and every optional value empty: exactly the 7 FR-020 keys hold the cloud's N/A
    value. Check this for `aws` (`N/A`), `az` (`NotApplicable`), `gcp` (`not_applicable`)
    and `dc` (`N/A`).
  - (g) N/A off: none of the 7 FR-020 keys is present, but `environment`, `availability`,
    `managedby` and `deletiondate` are present.
- [ ] T013 In `pkg/context/tags.go`, add `var now = time.Now`, which tests can override, and
  `func ResolveDeletionDate(config *DataSourceConfig)`:
  - an explicit `DeletionDate` is kept;
  - otherwise, `EnvironmentType == "Ephemeral"` gives `now().UTC().Add(90*24*time.Hour)`
    formatted `2006-01-02`;
  - otherwise the value is `"never"`.

  Re-implement `ProcessEphemeralEnvironment` as a deprecated wrapper that calls
  `ResolveDeletionDate`, and add a `// Deprecated:` doc comment. This covers C1 and FR-016.
- [ ] T014 In `pkg/context/tags.go`, add `func EffectiveEnvironmentName(config *DataSourceConfig) string`,
  which returns EnvironmentName, else Environment, else `"dev"`. In `Process()`:
  - always set `environment` to `EffectiveEnvironmentName(...)`;
  - always set `availability`, `managedby` and `deletiondate` from the config, with no N/A
    path.

  This covers C3 and FR-013.
- [ ] T015 In `Process()` in `pkg/context/tags.go`, change the PM and ITSM handling:
  - emit `projectmgmtid` only when `PMProjectCode != ""`, prefixed with
    `PMPlatform + delimiter` when `SystemPrefixesEnabled && PMPlatform != ""`;
  - emit each of `systemid`, `componentid` and `instanceid` only when its id is non-empty,
    prefixed with `ITSMPlatform + delimiter` under the same condition;
  - remove the N/A fallbacks for these four keys.

  This covers C4 and FR-014.
- [ ] T016 In `pkg/context/tags.go`, rename `addTag` to `addOptionalTag` and use it only for
  `costcenter`, `securityreview`, `privacyreview`, `dataregulations` and the owner keys.
  Change `ProcessDataTags()` so that `dataowners` is evaluated only inside
  `if tp.Config.OwnerTagsEnabled { … }`, matching `productowners` and `codeowners`. This
  covers C5, FR-015 and FR-020.
- [ ] T017 In `pkg/context/tags.go`, add the exported fields
  `GitInfoFunc func() (*GitInfo, error)` (nil means `GetGitInfo`) and `Warnings []string` to
  `TagProcessor`. In `Process()`:
  1. Reset `tp.Warnings`.
  2. When `SourceRepoTagsEnabled` is set, call the git function.
  3. On error, append
     `"source repository tags skipped: " + err.Error()` and emit neither git tag.
  4. When git succeeds, emit `sourcerepo` and `sourcecommit` only when they are non-empty.

  This covers C7 and FR-017.
- [ ] T018 In `pkg/context/tags.go`, extract the duplicated prefix, sanitize and truncate loop
  from `Process()` and `ProcessDataTags()` into
  `func (tp *TagProcessor) finalize(tags map[string]string) (map[string]string, error)`. Add
  `func truncateRunes(s string, max int) string`, which counts runes and never splits a
  multi-byte character, and use it in place of the byte slicing. Add tests to
  `pkg/context/tags_test.go`:
  - a 300-rune value containing `é` on `az`, truncated to 256 runes, with `utf8.ValidString`
    true;
  - a 70-character value on `gcp`, truncated to 63.

  This covers C8 and FR-021.
- [ ] T019 In `finalize()` in `pkg/context/tags.go`, validate every prefixed key with
  `tp.CloudProvider.ValidateTagKey(key)`. Return the first invalid key, in sorted order, as
  `fmt.Errorf("invalid tag key %q for target cloud rules", key)`. Add tests to
  `pkg/context/tags_test.go`:
  - `gcp` with `AdditionalTags{"Team":"x"}` gives an error mentioning `bc-Team`;
  - `dc` with `TagPrefix="bc:"` gives an error;
  - `aws` with `AdditionalTags{"team":"x"}` gives no error.

  This covers C9 and FR-022.
- [ ] T020 In `pkg/context/tags.go`, make `ConvertTagsToListOfMaps` add
  `"propagate_at_launch": "true"` to every entry, keeping the sort by key. Update the existing
  expectations in `pkg/context/tags_test.go`. This covers C10 and FR-023.
- [ ] T021 Update the existing cases in `pkg/context/tags_test.go` that assumed N/A for absent
  PM/ITSM ids, N/A environment or N/A deletion date. Set `GitInfoFunc` stubs in every
  `TagProcessor` test so that no test runs real git. The tests from T012 must now pass.
- [ ] T022 [P] In `internal/core/tags.go`, add the re-exports
  `func ResolveDeletionDate(config *DataSourceConfig)` and
  `func EffectiveEnvironmentName(config *DataSourceConfig) string`, which delegate to
  `pkg/context`. Mark `ProcessEphemeralEnvironment` deprecated in its doc comment. Add no
  logic. This covers C15 and FR-025.
- [ ] T023 [P] Create `pkg/context/bench_test.go` with two benchmarks:
  - `BenchmarkTagProcessorProcess`: every `DataSourceConfig` field populated, three owners,
    five additional tags, the `aws` provider, and `GitInfoFunc` returning a fixed
    `GitInfo`.
  - `BenchmarkNameGenerator`.

  Also add `TestTagProcessorPerformanceBudget`, which runs `testing.Benchmark` on the first
  benchmark and fails if `NsPerOp() > 50_000_000`. This is the performance SLO.
- [ ] T024 [P] Update `examples/client-app/main.go` to call `context.ResolveDeletionDate` in
  place of `ProcessEphemeralEnvironment`, and to print `tp.Warnings`. Verify with
  `cd examples/client-app && go build ./... && go run .`.
- [ ] T025 Library checkpoint: run `go test -race -cover ./pkg/context/...`, with coverage at
  least 90% and the T012 parity tests passing. Also run `go build ./... && go vet ./...`.

**Checkpoint**: Library behavior conforms to FR-013 to FR-023. The adapter can now be wired up.

---

## Phase 4: Provider and Data Source Adapter

**Purpose**: Wire the library changes into the Terraform layer, using attribute-scoped
diagnostics and warnings.

- [ ] T026 [P] In `internal/provider/provider.go`, change `Configure`:
  - Replace the inline `validProviders` map with `core.ValidateCloudProvider(cloudProvider)`.
    On failure, call
    `resp.Diagnostics.AddAttributeError(path.Root("cloud_provider"), "Invalid cloud provider", err.Error())`.
  - Validate `tag_prefix`: build `prefix + "environment"` and check it with
    `core.GetCloudProvider(cloudProvider).ValidateTagKey(...)`. On failure, add an attribute
    error on `path.Root("tag_prefix")` whose message names the cloud.

  This covers C13, FR-001 and FR-002.
- [ ] T027 In `Read` in `internal/datasource/context.go`:
  - Replace `core.ProcessEphemeralEnvironment(config)` with
    `core.ResolveDeletionDate(config)`.
  - Convert each validation `AddError` to `AddAttributeError` with `path.Root(...)` on
    `namespace`, `environment`, `environment_type`, `availability`, `sensitivity`,
    `deletion_date`, `product_owners`, `code_owners` and `data_owners`.
  - Report a name prefix failure on `path.Root("name")`.
  - Report a `Process()` error on `path.Root("additional_tags")` and a `ProcessDataTags()`
    error on `path.Root("additional_data_tags")`.

  This covers C12 and FR-009.
- [ ] T028 In `internal/datasource/context.go`, after `tagProcessor.Process()`, call
  `tflog.Warn(ctx, w)` for each entry in `tagProcessor.Warnings`. Use `tflog`, not diagnostics,
  so that plans outside git stay quiet. This covers C12 and FR-017.
- [ ] T029 In the `context_output` population in `internal/datasource/context.go`:
  - set `environment_name` to `core.EffectiveEnvironmentName(config)`;
  - leave `environment` as the merged input, null-to-empty as today, with no `dev`;
  - set `deletion_date` to the resolved value from T027.

  This covers C11, FR-006 and FR-013.
- [ ] T030 In `internal/datasource/context.go`, update the schema descriptions in both the
  top-level `Schema()` and `getContextAttributes()`:
  - `environment_type`: deployment-stage classification used by consuming modules to pick
    resiliency and cost defaults. List the 7 values with their RPO/RTO. State that it does not
    affect the name prefix, is not emitted as a tag, and that `Ephemeral` sets a +90-day
    deletion date.
  - `deletion_date`: `YYYY-MM-DD` or `never`. When unset, it defaults to `never`, or to
    today + 90 days for Ephemeral.
  - `environment_name`: tag value that falls back to environment, then `dev`.
  - `tags_as_list_of_maps` and `data_tags_as_list_of_maps`: entries contain `key`, `value`
    and `propagate_at_launch`.

  This covers C14 and FR-027.
- [ ] T031 [P] Create `internal/datasource/context_test.go` with unit tests:
  - `mergeStringValue`, `mergeBoolValue`, `mergeListValue` and `mergeMapValue`: child set,
    parent only, neither, and a map merge where the child key wins;
  - `Schema()` returns no diagnostics, using `datasource.SchemaRequest`.
- [ ] T032 Adapter checkpoint: run `go build ./... && go vet ./... && golangci-lint run ./...`,
  then `go test -race ./internal/...`.

**Checkpoint**: The provider binary carries all conformance changes.

---

## Phase 5: Acceptance Tests and Examples

**Purpose**: Terraform-level validation without cloud credentials, plus plan stability
(bead `tfpctx-24o`). Each acceptance test uses `resource.UnitTest` or `resource.Test` with
`TF_ACC`, and sets `source_repo_tags_enabled = false` unless it is testing git.

- [ ] T033 Create `internal/provider/provider_test.go` with:
  - `testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){"brockhoff": providerserver.NewProtocol6WithError(New("test")())}`;
  - `testAccPreCheck(t)`, a no-op that exists for future checks;
  - `TestProviderSchema`, which validates the provider schema with no diagnostics.
- [ ] T034 [P] Create `internal/provider/datasource_context_acc_test.go`. Each test ends with a
  `PlanOnly: true` step that expects an empty plan, for the plan stability SLO.
  - `TestAccContext_Minimal`: `name = "webapp"`. Check that `name_prefix = webapp`,
    `tags.bc-environment = dev`, `tags.bc-deletiondate = never`, `tags.bc-availability =
    preemptable`, `data_tags.bc-sensitivity = confidential`, and that the
    `bc-projectmgmtid` key is absent.
  - `TestAccContext_Full`: `cloud_provider = "aws"` and every input set. Check:
    - `name_prefix = myorg-payment-api-prod`;
    - `bc-projectmgmtid = "JIRA PAY"`, using the space delimiter;
    - owners joined with a space;
    - `tags_as_list_of_maps.0.propagate_at_launch = "true"`.
  - `TestAccContext_Ephemeral`: Ephemeral with no date gives `bc-deletiondate` equal to the
    UTC date + 90 days, computed in the test. Ephemeral with `deletion_date = "2030-01-01"`
    keeps the explicit date.
  - `TestAccContext_ParentChild`:
    - the child inherits namespace, environment and cost center;
    - a child override wins;
    - the child's `name` is not inherited;
    - `additional_tags` are merged;
    - with a parent that has no environment, `context_output.environment` is null or empty
      and `context_output.environment_name = dev`.
  - `TestAccContext_NamePrefixPartial`: `namespace = "ns"`, `name = "app"` and no
    environment gives `name_prefix = ns-app` (gap 1, kept).
- [ ] T035 [P] Create `internal/provider/datasource_context_validation_acc_test.go`, using
  `ExpectError` regexps. One step per invalid input:
  - namespace `TooLongNS1`;
  - environment `Prod`;
  - `environment_type = "Staging"`;
  - `availability = "high"`;
  - `sensitivity = "secret"`;
  - `deletion_date = "2025-13-01"`;
  - `product_owners = ["not-an-email"]`;
  - provider `cloud_provider = "azure"`;
  - provider `cloud_provider = "gcp"` with `additional_tags = { Team = "x" }`;
  - provider `tag_prefix = "bc:"` with `cloud_provider = "dc"`.

  Each regexp must match the attribute name or the rejected value.
- [ ] T036 [P] Create `internal/provider/datasource_context_cloud_acc_test.go`. For each of
  `aws`, `az`, `gcp` and `dc`, assert:
  - the N/A value of `bc-costcenter`;
  - the delimiter in `bc-codeowners` with two owners;
  - sanitization of `additional_tags = { note = "a<b>&c?d/e#f:g" }`;
  - truncation of a 300-character value to 256 for aws/az and 63 for gcp/dc;
  - lowercase values on `gcp`.
- [ ] T037 [P] Update `examples/parent-child/README.md` and `examples/parent-child/main.tf`
  comments to state that `name` is never inherited, and that `environment_name` falls back to
  `dev` in `context_output` while `environment` does not.
- [ ] T038 [P] Update the comments in `examples/ephemeral/ephemeral.tf`: an explicit
  `deletion_date` wins, and non-Ephemeral contexts default to `never`.
- [ ] T039 Acceptance checkpoint: run
  `make install && TF_ACC=1 go test -v -timeout 30m ./internal/provider/...` and
  `make test-examples`. Every example must init, validate and plan successfully, and a second
  `terraform plan` in each example directory must report `No changes`.

**Checkpoint**: Terraform-level behavior is verified against the spec.

---

## Phase 6: Polish and Documentation

**Purpose**: Make the docs match the spec (principles: *Document from the Schema*), close the
beads, and run the final gates.

- [ ] T040 Update `templates/data-sources/context.md.tmpl`:
  - add an "Environment Type" section with the FR-027 table;
  - add a "Migrating from terraform-external-context" section that lists the divergences
    (gap 1 prefix composition, `name` not inherited, provider-level `cloud_provider` and
    `tag_prefix`, default prefix `bc-`).

  Then run `make docs-generate` and confirm that `docs/data-sources/context.md` and
  `docs/index.md` changed as expected.
- [ ] T041 [P] Update `README.md`:
  - move `environment_type` out of "Naming Configuration" into a new "Environment
    Classification" section with the RPO/RTO table;
  - document `deletion_date` defaults (`never`, Ephemeral + 90 days);
  - list the generated tag keys and when each is emitted, per FR-013 to FR-020;
  - show `propagate_at_launch` in the list-of-maps description;
  - add a short "Upgrade notes" section that lists the tag values changed by gaps 2–6.

  This covers bead `tfpctx-am0`.
- [ ] T042 [P] Update `.agents/agents.md`, which `AGENTS.md` symlinks to:
  - move `environment_type` out of "Naming Configuration";
  - fix "Ephemeral Environment Handling" so that an explicit date wins and the default is
    `never`;
  - list the actual tag keys (`projectmgmtid`, `systemid`, `componentid`, `instanceid`);
  - document the N/A scope, the `dataowners` gating and `propagate_at_launch`;
  - add a pointer to `specs/001-context-provider/spec.md` and
    `.specify/memory/principles.md`.

  This covers bead `tfpctx-am0`.
- [ ] T043 [P] Update `docs/design.md`:
  - replace the component tree with the real layout (`pkg/context` as the library,
    `internal/core` as aliases, no `internal/models` or `tests/`);
  - fix the tag key names in "System Prefix Logic";
  - fix the Ephemeral rule;
  - change Git Errors to "log warning via tflog and skip git tags";
  - reference spec 001.

  This covers bead `tfpctx-6he`.
- [ ] T044 [P] Update `pkg/context/README.md` to document `ResolveDeletionDate`,
  `EffectiveEnvironmentName`, `TagProcessor.GitInfoFunc`, `TagProcessor.Warnings`, the
  deprecated `ProcessEphemeralEnvironment`, key validation errors and the
  `propagate_at_launch` entry.
- [ ] T045 Final validation:
  - `gofmt -s -l .` produces no output;
  - `go vet ./...`, `golangci-lint run ./...`, `go mod verify`;
  - `go test -race -cover ./...`, with `pkg/context` at 90% or more;
  - `make install && TF_ACC=1 go test ./internal/provider/...`;
  - `make test-examples`, `make docs-validate`;
  - `make release-snapshot`, if goreleaser is installed: confirm the archives and
    `SHA256SUMS` are produced.

  Record any failures with their output.
- [ ] T046 Traceability review: confirm every row of the table below has a passing test.
  Then:
  - set the `spec.md` Status to `Implemented`;
  - tick the matching items in the spec's Success Criteria checklist;
  - remove the fixed items from spec Notes → "Code defects found".
- [ ] T047 Close the beads `tfpctx-6ce`, `tfpctx-24o`, `tfpctx-am0` and `tfpctx-6he` with
  `bd close`. Report the changed files and propose `feat:`/`fix:`/`ci:`/`docs:` commit
  groupings, but do not commit until asked.

---

## Requirement → Task Traceability

| Requirement | Implementation | Verification |
|-------------|----------------|--------------|
| FR-001 | T026 | T035 |
| FR-002 | T026 | T035 |
| FR-003 | existing | T033, T034 (no credentials needed) |
| FR-004/005 | existing | T031, T034 ParentChild |
| FR-006 | T029 | T034 ParentChild |
| FR-007 | existing | T034 Minimal |
| FR-008 | existing | T034 ParentChild (`context_output.enabled`) |
| FR-009 | T010, T027 | T035 |
| FR-010/011 | existing | existing `naming_test.go`, T034 Full |
| FR-012 | existing (gap 1 kept) | T034 NamePrefixPartial |
| FR-013 | T014, T029 | T012(c,g), T034 Minimal |
| FR-014 | T015 | T012(a,b), T034 Full |
| FR-015 | T016 | T012(e) |
| FR-016 | T013 | T012(d), T034 Ephemeral |
| FR-017 | T006–T009, T017, T028 | T006, T009, T017 tests |
| FR-018 | existing | T034 ParentChild |
| FR-019 | T016 | T012(e,f) |
| FR-020 | T016 | T012(f,g), T036 |
| FR-021 | T018 | T018 tests, T036 |
| FR-022 | T019, T026 | T019 tests, T035 |
| FR-023 | T020 | T020 tests, T034 Full |
| FR-024/025 | T022, T024 | T024 (client-app builds) |
| FR-026 | T037, T038, T040–T044 | T045 docs-validate, test-examples |
| FR-027 | T030, T040–T042 | T045 docs-validate |
| Perf / scalability NFR | T008, T023 | T009, T023 |
| Release / supply chain | T002–T004 | T045 release-snapshot |

## Dependencies and Execution Order

### Phase Dependencies

```
Phase 1 Setup ──► Phase 2 Foundation ──► Phase 3 Library ──► Phase 4 Adapter ──► Phase 5 Acceptance ──► Phase 6 Polish
   (T001–T005)       (T006–T011)            (T012–T025)          (T026–T032)          (T033–T039)            (T040–T047)
```

- Phase 2 blocks Phase 3, because T017 depends on the git seam and error from T008 and on the
  URL fix from T007, and T013 depends on the `never` validation from T010 through round-trip
  inheritance.
- Phase 3 blocks Phase 4, because the adapter calls `ResolveDeletionDate`,
  `EffectiveEnvironmentName` and `Warnings`.
- Phase 4 blocks Phase 5, because the acceptance tests exercise the wired adapter.
- In Phase 6, T040 must run before T045 (generated docs), and T046 and T047 run last.

### Sequential Chains Within Phases

- T006 → T007 → T008 → T009 (all edit `git.go` and `git_test.go`)
- T012 → T013 → … → T021 (all edit `tags.go`, plus `tags_test.go` or `tags_parity_test.go`)
- T027 → T028 → T029 → T030 (all edit `internal/datasource/context.go`)
- T033 before T034–T036 (they share the provider factories)

### Parallel Opportunities

- **Phase 1**: T002, T003 and T004 (three different files), after T001.
- **Phase 2**: T010 (validation files) alongside the T006–T009 chain.
- **Phase 3**: T022, T023 and T024 after T021 (core aliases, benchmark file, client app).
- **Phase 4**: T026 (provider) and T031 (new test file) alongside the T027–T030 chain.
- **Phase 5**: T034, T035, T036, T037 and T038 after T033.
- **Phase 6**: T041, T042, T043 and T044 alongside T040.

Example, Phase 5 after T033:

```
Task: "T034 Create internal/provider/datasource_context_acc_test.go …"
Task: "T035 Create internal/provider/datasource_context_validation_acc_test.go …"
Task: "T036 Create internal/provider/datasource_context_cloud_acc_test.go …"
Task: "T037 Update examples/parent-child/README.md and main.tf …"
Task: "T038 Update examples/ephemeral/ephemeral.tf …"
```

### Validation Checkpoints

| After | Checkpoint | Command |
|-------|------------|---------|
| Phase 1 | T005 | `go mod verify && go build ./... && go vet ./... && go test -race ./...` |
| Phase 2 | T011 | `go test -race -run 'Git\|Validate' ./pkg/context/...` |
| Phase 3 | T025 | `go test -race -cover ./pkg/context/...` (≥ 90%) |
| Phase 4 | T032 | `golangci-lint run ./... && go test -race ./internal/...` |
| Phase 5 | T039 | `TF_ACC=1 go test ./internal/provider/...` + `make test-examples` |
| Phase 6 | T045 | full gate, including `make docs-validate` and `make release-snapshot` |

## Implementation Strategy

1. Land Phases 1–2 first. They are independently valuable: CI runs the library tests, tools
   are pinned, and the credential leak is fixed. They are safe to ship as `fix:` and `ci:`
   commits in a PATCH release.
2. Land Phases 3–5 together. They change consumer-visible tag outputs (gaps 2–6), so they ship
   as a single MINOR release with per-tag upgrade notes (T041).
3. Phase 6 docs can land with step 2 or immediately after it.

**Environment order**: the provider has no dev, staging or prod deployments. Its promotion path
is local install (`make install`), then PR CI, then `make release-snapshot`, then a
release-please tag, then the Terraform Registry.

## Notes

- This is a conformance task list. Requirements marked "existing" in the traceability table
  already work and only gain test coverage.
- `internal/core` must never gain logic. Any new library function needs only an alias there.
- Use `t.Setenv`/`t.Cleanup` to restore any package-level seam (`now`, `gitCommand`) changed
  in a test.
- `[P]` tasks touch different files. Never run two tasks that edit `tags.go` or
  `internal/datasource/context.go` at the same time.

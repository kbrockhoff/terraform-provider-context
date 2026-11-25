# Infrastructure Specification: Context Provider for Standardized Naming and Tagging

**Spec ID**: `001-context-provider`
**Created**: 2026-11-24
**Status**: Draft
**Input**: User description: "reverse engineer the existing code and docs to generate specifications"

## Executive Summary *(mandatory)*

This specification describes the capabilities of the Context provider as it exists today. It
was reverse-engineered from the provider code (`internal/`), the reusable library
(`pkg/context`), the project documentation (README, `docs/design.md`, AGENTS.md), and the
legacy `terraform-external-context` module that the provider replaces.

The provider is a credential-free, side-effect-free building block. Infrastructure
configurations call it to get three things:

- a standardized **name prefix**;
- a **resource tag set** covering environment, ownership, billing, project management, ITSM,
  review and source-repository metadata;
- a **data tag set** covering sensitivity, data owners and regulations.

Tags are formatted for the target cloud's constraints. Contexts can be layered: a parent
context supplies shared values, and child contexts override individual fields. The same logic
is also published as a standalone library for non-infrastructure clients such as serverless
functions and command-line tools.

## Problem Statement *(mandatory)*

### Current State

Consumers produce names and tags with the legacy `terraform-external-context` module. That
module has several drawbacks:

- It runs an external script to read git metadata.
- It recomputes the ephemeral deletion date from the current timestamp on every plan, so the
  value is never known until apply.
- Cloud and tag-prefix settings must be repeated in every context instead of once per
  configuration.
- Its logic is available only to Terraform code.

The provider reimplements the module natively, but its behavior has only been described
informally. The README, `docs/design.md` and AGENTS.md disagree with each other and with the
code on several points, including tag key names, the deletion date override rule and the
package layout.

### Desired State

There is one authoritative, testable specification of what the provider and library MUST do.
Planning, tests, documentation and future changes all trace to it. Every point where the
current behavior differs from the legacy module is either adopted deliberately as a versioned
change or scheduled for correction.

### Business Impact

- **Benefits**:
  - Consistent names and tags across every consuming repository, which enables reliable cost
    attribution, ownership lookup, data classification and automated clean-up of ephemeral
    environments.
  - Faster planning, because no external script runs on each evaluation.
  - One code path shared by Terraform and non-Terraform clients.
- **Risks if not specified**:
  - Silent tag drift when consumers migrate from the module, which corrupts cost reports and
    ownership data.
  - Documentation that contradicts the provider's actual behavior.
  - Regressions that no test catches, because the provider and data source layers have no
    tests.

## Infrastructure Requirements *(mandatory)*

### Functional Requirements

#### Provider configuration

- **FR-001**: The provider MUST accept an optional target-cloud identifier with one of these
  values: `dc`, `aws`, `az`, `gcp`, `oci`, `ibm`, `do`, `vul`, `ali`, `cv`. The default is
  `dc`. Any other value MUST be rejected at configuration time with an error that lists the
  allowed values.
- **FR-002**: The provider MUST accept an optional tag key prefix, defaulting to `bc-`, and
  MUST apply it to every generated tag key. The prefix MUST be applied to both resource tags and
  data tags.
- **FR-003**: The provider MUST require no credentials, MUST make no network calls, and MUST
  NOT create or modify any resource. It exposes exactly one data source (`brockhoff_context`)
  and no managed resources.

#### Context inputs and inheritance

- **FR-004**: The data source MUST accept the following optional inputs:
  - **Naming**: namespace, name, environment, environment name.
  - **Environment classification**: environment type (see FR-027). It is not a naming input.
  - **Resource management**: an enabled flag, availability, managed-by, deletion date.
  - **Project management**: project management platform and project code.
  - **ITSM**: ITSM platform, and system, component and instance identifiers.
  - **Ownership**: cost center, and product, code and data owner lists.
  - **Classification**: sensitivity, data regulations list, data sovereignty jurisdiction
    (FR-028), security review and privacy review.
  - **Feature toggles**: source-repository tags, system prefixes, not-applicable
    substitution, owner tags.
  - **Extra tags**: additional tag maps for resource tags and for data tags.
- **FR-005**: The data source MUST accept an optional parent context. For every scalar,
  list and flag input, the explicitly set child value wins, then the parent's value, then the
  default. For additional-tag maps, the parent map and child map MUST be merged, with child
  keys winning. `name` MUST NOT be inherited from the parent.
- **FR-006**: The data source MUST emit a resolved context output. It has the same shape as
  the parent context input, holds the effective values after inheritance, defaults and
  ephemeral processing, and can be passed as the parent context of another data source.
- **FR-007**: When no value is supplied, defaults MUST be:
  - availability: `preemptable`
  - managed-by: `terraform`
  - sensitivity: `confidential`
  - enabled: true
  - every feature toggle: true
- **FR-008**: The enabled flag MUST be passed through to the resolved context so that
  consumers can make their own resource creation conditional. It MUST NOT suppress or alter
  any generated output. This matches the legacy module.

#### Environment type

- **FR-027**: The environment type MUST classify the deployment stage so that consuming
  configurations can choose resource defaults suited to it. Resiliency defaults follow the
  RPO and RTO targets below, and cost-optimization defaults focus on shutting resources down
  outside working hours. The provider MUST carry the effective environment type, after
  inheritance, in the resolved context (FR-006). Applying the per-type defaults is the
  consuming configuration's job; the provider MUST NOT apply them itself.

  | Type | Use case | Configuration focus | RPO | RTO |
  |------|----------|---------------------|-----|-----|
  | `None` | Custom configuration | No defaults applied; use individual config values | N/A | N/A |
  | `Ephemeral` | Temporary environments | Cost-optimized, minimal durability requirements | N/A | 48h |
  | `Development` | Developer workspaces | Balanced cost and functionality for active development | 24h | 48h |
  | `Testing` | Automated testing | Consistent, repeatable configurations | 24h | 48h |
  | `UAT` | User acceptance testing | Production-like settings with some cost optimization | 12h | 24h |
  | `Production` | Live systems | High availability, durability and performance | 1h | 4h |
  | `MissionCritical` | Critical production | Maximum reliability, redundancy and monitoring | 5m | 1h |

  The environment type MUST NOT contribute to the name prefix (FR-010 to FR-012) and MUST NOT
  be emitted as a tag (FR-013). Its only effect on outputs is the deletion date for
  `Ephemeral` (FR-016). Values outside this table MUST be rejected (FR-009).

#### Validation (fail fast)

- **FR-009**: Invalid inputs MUST stop evaluation with a diagnostic that names the attribute
  and the rejected value. The rules are:
  - **namespace and environment**: 1–8 characters, lowercase, starting with a letter,
    containing only letters, digits and hyphens, and not ending in a hyphen.
  - **environment type**: one of `None`, `Ephemeral`, `Development`, `Testing`, `UAT`,
    `Production`, `MissionCritical`.
  - **availability**: one of `preemptable`, `spot`, `standard`, `dedicated`, `isolated`.
  - **sensitivity**: one of `public`, `internal`, `confidential`, `restricted`, `critical`.
  - **deletion date**: a real calendar date in `YYYY-MM-DD` form.
  - **owner lists**: every entry is a syntactically valid email address.
  - **data sovereignty**: when supplied, not empty and not only whitespace.

#### Name prefix

- **FR-010**: The name prefix MUST be lowercase, 2–24 characters long, start with a letter,
  end with a letter or digit, and contain only letters, digits and hyphens. A prefix that cannot
  meet these rules after truncation MUST cause an error. It MUST NOT be emitted silently.
- **FR-011**: If namespace, name and environment are all present, the prefix is
  `namespace-name-environment`. If it exceeds 24 characters, only the name component is
  truncated, any trailing hyphen is removed, and namespace and environment are kept intact.
- **FR-012**: If namespace or environment is absent, the prefix MUST join the components that
  are present, in namespace, name, environment order, with hyphens. For example, `ns-name` or
  `name-env`. If none are present, evaluation MUST fail. This deliberately diverges from the
  legacy module, which uses `name` alone in this case (parity gap 1). The divergence MUST be
  documented in the migration guidance, because it changes resource names for consumers who
  migrate.

#### Resource tags

- **FR-013**: Resource tags MUST use these base keys, each prefixed per FR-002:
  - `environment` (from environment name)
  - `availability`
  - `managedby`
  - `deletiondate`
  - `costcenter`
  - `projectmgmtid`
  - `systemid`
  - `componentid`
  - `instanceid`
  - `productowners`
  - `codeowners`
  - `securityreview`
  - `privacyreview`
  - `sourcerepo`
  - `sourcecommit`

  The environment type MUST NOT be emitted as a tag.

  The `environment`, `availability`, `managedby` and `deletiondate` tags MUST always be
  emitted:

  - `environment` is the environment name. If that is unset, it is the environment. If both
    are unset, it is `dev`. The fallback applies to the tag and to the resolved context's
    environment name only. The resolved context's environment MUST stay as supplied, so that
    child contexts never inherit `dev` into their name prefix. The fallback MUST NOT affect
    any name prefix.
  - `deletiondate` follows FR-016.
- **FR-014**: When system prefixes are enabled and the relevant platform is set, the PM
  project code and each ITSM identifier MUST be emitted as `platform<delimiter>id`, using the
  target cloud's delimiter. Otherwise the bare identifier is emitted. When the PM project code
  is absent, the `projectmgmtid` tag MUST be omitted, whatever the not-applicable setting.
  Likewise, each of `systemid`, `componentid` and `instanceid` MUST be omitted when its
  identifier is absent.
- **FR-015**: Owner lists MUST be joined with the target cloud's delimiter. When owner tags are
  disabled, the product, code and data owner tags MUST be omitted, including the data owner
  tag in the data tag set.
- **FR-016**: When the environment type is `Ephemeral` and no deletion date is supplied
  directly or through inheritance, the deletion date MUST be set to the evaluation date plus
  90 days. An explicitly supplied deletion date MUST always take precedence. For any other
  environment type with no deletion date, the value MUST be `never`.
- **FR-017**: When source-repository tags are enabled and the working directory is inside a
  git repository, the tags MUST contain the origin remote URL and the current commit
  identifier. The URL MUST be:
  - converted from SSH to HTTPS form;
  - stripped of any embedded user name, password or token;
  - stripped of a trailing `.git`.

  If repository information is unavailable, these tags MUST be skipped and a warning MUST be
  logged. Evaluation MUST NOT fail.
- **FR-018**: Additional tags MUST be merged after generated tags. An additional tag with the
  same base key overrides the generated value.

#### Data tags

- **FR-019**: Data tags MUST contain `sensitivity`, `dataregulations` (joined with the target
  delimiter), `dataowners` (per FR-015) and, when supplied, `sovereignty` (per FR-028),
  followed by the merged additional data tags. All of these are prefixed per FR-002.
- **FR-028**: The data source MUST accept an optional data sovereignty input. Its value is the
  jurisdiction whose laws govern where the data may be stored and processed. By convention it
  is an ISO 3166-1 country code such as `DE` or an ISO 3166-2 subdivision code such as
  `US-CA`. The value is free text and the convention is not validated. The input is
  inherited from the parent context like every other scalar input (FR-005) and is carried in the
  resolved context (FR-006).
  - When a value is supplied, after inheritance, it MUST be emitted as the data tag
    `sovereignty`. It MUST NOT be emitted as a resource tag.
  - When no value is supplied, the `sovereignty` tag MUST be omitted. It is never substituted
    with the N/A value, whatever the not-applicable setting (FR-020).
  - A supplied value that is empty or only whitespace MUST be rejected (FR-009).
  - The value is normalized and truncated for the target cloud like any other tag value
    (FR-021).

#### Missing values and cloud formatting

- **FR-020**: Not-applicable substitution applies only to these tags: `costcenter`,
  `productowners`, `codeowners`, `securityreview`, `privacyreview`, `dataregulations` and
  `dataowners`. Owner tags additionally require owner tags to be enabled. When substitution is
  enabled, an absent value in one of these tags MUST be emitted as the target cloud's N/A value.
  When substitution is disabled, the tag MUST be omitted. All other tags follow FR-013, FR-014,
  FR-017 and FR-028.
- **FR-021**: Every resource and data tag value, including additional tags, MUST be
  normalized for the target cloud and then truncated to that cloud's maximum length, as shown
  in the table below. Truncation MUST NOT split a multi-byte character.

  | Cloud id | Max value length | Lowercase | Disallowed characters → replacement | List delimiter | N/A value |
  |----------|------------------|-----------|---------------------------------------|----------------|-----------|
  | `aws` | 256 | no | anything outside letters, digits, space and `.:=+@_/-` → `_` | space | `N/A` |
  | `az` | 256 | no | space and `<>%&\?/#:` → removed | `;` | `NotApplicable` |
  | `gcp` | 63 | yes | anything outside lowercase letters, digits, `_` and `-` → `-` | `_` | `not_applicable` |
  | all others | 63 | no | `<>%&\?` → `_` | `;` | `N/A` |

- **FR-022**: Generated tag keys, after the prefix is applied, MUST satisfy the target cloud's
  key rules. Additional-tag keys that violate those rules MUST be rejected with a diagnostic.

#### Alternate output formats

- **FR-023**: For both the resource tag set and the data tag set, the data source MUST emit:
  - a list of `{key, value, propagate_at_launch = "true"}` maps, sorted by key, matching the
    legacy module's shape;
  - a sorted list of `key=value` strings;
  - a single comma-separated string of those pairs.

#### Reusable library

- **FR-024**: The library MUST expose all naming, validation, tagging, cloud-rule, git and
  format-conversion capabilities independently of the Terraform plugin framework. For the same
  inputs, it MUST produce identical results to the data source.
- **FR-025**: The legacy internal re-export package (`internal/core`) MUST remain a
  behavior-free alias layer over the library.

#### Documentation and examples

- **FR-026**: Reference documentation MUST be generated from the schema descriptions. Runnable
  examples MUST exist for minimal, full, cloud-specific, ephemeral and parent/child usage, plus
  a standalone library client.

### Non-Functional Requirements

#### Performance

- A data source evaluation without git lookup completes in under 50 ms.
- Git metadata is looked up at most once per 5-minute window per provider process, however many
  data source instances are evaluated.
- Evaluation calls no network services. The only external processes are the two local git
  queries.

#### Availability

- The provider works fully offline and outside a git repository. If repository information is
  missing, only the source-repository tags are affected.
- Releases are published to the public provider registry for linux (amd64, arm64), darwin
  (amd64, arm64) and windows (amd64).

#### Security

- No credentials are accepted, stored or written to state.
- Credentials embedded in repository URLs are never emitted (FR-017).
- Every emitted tag value is sanitized for the target cloud (FR-021).
- Release artifacts are checksummed and signed with the publisher's GPG key.
- The build toolchain and CI action versions are pinned, and dependency integrity is verified
  in CI.

#### Scalability

- Hundreds of data source instances in one configuration evaluate concurrently without data
  races. The git cache is safe for concurrent access.
- Adding a new target cloud requires only a new cloud-rule implementation and its tests. Naming,
  tagging and data source code are unchanged.

## Service Level Objectives (SLOs) *(mandatory)*

- **Determinism**: 100% of repeated evaluations with identical inputs, at the same commit and
  on the same calendar day, produce byte-identical outputs. This is verified by an acceptance
  test that plans twice and expects no diff.
- **Plan stability**: 0 spurious diffs on downstream resources when nothing relevant changed.
  This is measured by a second plan against the examples reporting "No changes".
- **Correctness**: 100% of the rules in FR-009 to FR-023 are covered by at least one passing
  automated test case. This is measured by a traceability table in the test plan.
- **Library coverage**: at least 90% statement coverage for the reusable library, measured by the
  CI coverage report on every pull request.
- **Pipeline health**: every pull request runs formatting, static analysis, lint and race-enabled
  unit tests, and finishes within 15 minutes on the CI runner.
- **Release integrity**: 100% of published release artifacts carry a valid signature and a
  checksum entry, verified by the registry at ingestion.

## Cost Constraints *(mandatory)*

### Budget

- **Runtime**: $0. The provider provisions no billable resources and calls no paid APIs.
- **CI/CD**: runs within the free tier of the source host's hosted runners for a public
  repository. No paid build infrastructure.

### Cost Optimization

- Git lookups are cached rather than repeated per data source instance.
- The CI matrix is limited to the 5 supported OS/architecture pairs.
- Tests are plan-only and never create cloud resources.

## Compliance Requirements *(include if applicable)*

### Regulatory Frameworks

The provider is not subject to a regulatory framework itself. It is a classification enabler:
the sensitivity, data regulation and data sovereignty tags let consumers meet their own
obligations, such as GDPR or CCPA, or data residency rules, which consumers declare as input
values. Provider registry publishing
requirements are mandatory:

- signed releases;
- a registry manifest declaring protocol version 6.0;
- semantic version tags.

### Data Requirements

- Owner tags contain email addresses, which are personal data. Consumers control them through
  the owner tags toggle (FR-015). The provider MUST NOT add personal data that the consumer did
  not supply. Git commit author information MUST NOT be emitted.
- The project is distributed under the Apache License 2.0.

## Success Criteria *(mandatory)*

### Code Validation

- [ ] All quality gates pass on the release branch: formatting, static analysis, lint and
  dependency verification.
- [ ] Schema descriptions, generated reference docs and README attribute tables list the same
  inputs, defaults and outputs.
- [ ] Every `[NEEDS CLARIFICATION]` item and every legacy parity gap has a recorded decision,
  and divergent behavior is called out in the changelog.

### Security Validation

- [ ] A test proves that HTTPS remotes with embedded credentials produce credential-free
  `sourcerepo` values.
- [ ] A test proves that every tag value, including additional tags, is sanitized and truncated
  for each of the 4 rule sets.
- [ ] Release pipeline output shows a signed checksum file for every release.

### Performance Validation

- [ ] A benchmark shows a single evaluation without git lookup under 50 ms.
- [ ] A test proves that 100 concurrent evaluations perform at most one git lookup inside the
  cache window, with no race detector findings.

### Operational Validation

- [ ] Acceptance tests plan every example configuration successfully, without cloud
  credentials.
- [ ] A second plan of each example reports no changes.
- [ ] The standalone library client example builds and produces output identical to the data
  source for the same inputs.

## Assumptions *(include if making assumptions)*

- The code under `pkg/context` and `internal/` is the primary source of truth for current
  behavior. Where documentation contradicts it, the documentation is stale. Examples: AGENTS.md
  says an Ephemeral environment overrides an explicit deletion date, and `docs/design.md` uses
  `pmproject`/`itsmsystem` keys.
- The difference from the legacy module in where the target cloud and tag prefix are set
  (provider level rather than per context) is intentional. So is the different default prefix
  (`bc-` rather than empty). Neither is treated as a parity gap.
- The 90-day ephemeral retention period, the 24-character prefix limit and the 1–8 character
  namespace and environment limits come from the Brockhoff naming standard and are fixed.
- The 50 ms evaluation target is a new budget. The current code has no measured baseline; the
  only reference is the pure in-memory computation it performs. The 15-minute pipeline
  target matches the existing CI job timeout.
- The data sovereignty value is free text naming a jurisdiction. The recommended convention is
  an ISO 3166-1 country code (`DE`, `CA`) or an ISO 3166-2 subdivision code (`US-CA`,
  `CA-QC`). This convention is not enforced: validation only rejects empty or whitespace-only
  values (FR-009).
- One specification covers the whole provider. It can be split later, for example into naming,
  tagging and inheritance specs, if planning shows that is useful.

## Out of Scope *(include if explicitly excluding items)*

- Managed resources, and enforcing tags on cloud resources. Consumers apply the outputs.
- Organization tag policies, region-specific tag rules and hierarchical tag inheritance beyond
  one parent. These are listed as future enhancements in `docs/design.md`.
- Dedicated rule sets for `oci`, `ibm`, `do`, `vul`, `ali` and `cv`. These use the default rule
  set until a consumer needs more.
- Migration tooling that rewrites module calls into data source blocks.

## Dependencies *(include if external dependencies exist)*

- The infrastructure-as-code CLI and its plugin protocol 6.0, through the official plugin
  framework.
- A local `git` executable for the source-repository tags. It is optional.
- The public provider registry and a GPG signing key held as a CI secret, used for releases.
- The legacy `kbrockhoff/terraform-external-context` module, used as the reference for
  compatibility.

## Notes

### Legacy parity gaps

These gaps were found by comparing the provider with the legacy module's `main.tf`.
Decision (2026-11-24): keep the provider's behavior for gap 1, and match the legacy module for
gaps 2–6.

| # | Behavior | Legacy module | Provider today | Decision |
|---|----------|---------------|----------------|----------|
| 1 | Prefix when only one of namespace or environment is set | uses `name` only | joins the parts that are present (for example `ns-name`) | Keep the provider's behavior (FR-012) |
| 2 | Absent PM or ITSM identifiers | tag omitted | `N/A` emitted when substitution is enabled | Match legacy (FR-014) |
| 3 | Environment tag when environment name is unset | falls back to environment, which defaults to `dev` | `N/A` | Match legacy (FR-013) |
| 4 | Deletion date tag for non-Ephemeral contexts without a date | `never` | `N/A` | Match legacy (FR-016) |
| 5 | `dataowners` when owner tags are disabled | omitted | `N/A` emitted | Match legacy (FR-015) |
| 6 | List-of-maps entries | include `propagate_at_launch = "true"` | `key` and `value` only | Match legacy (FR-023) |
| 7 | `name` inheritance from the parent context | inherited | not inherited | Keep the provider's behavior (FR-005) |

Gaps 2–6 change the provider's current outputs. Ship them as one MINOR release and list each
changed tag in the changelog.

### Additions beyond the legacy module

- FR-028 (data sovereignty): a new optional input and data tag added on 2026-11-25. It has no
  counterpart in the legacy module or the current code. It is additive, so it ships as a MINOR
  release, and contexts that do not set it produce unchanged outputs.

### Code defects found (spec-conformance work)

- FR-013, FR-014, FR-015, FR-016, FR-020, FR-023: these need code changes to close legacy
  parity gaps 2–6.
- FR-017: HTTPS remotes keep embedded credentials (bead `tfpctx-6ce`), and missing git
  information is skipped silently with no warning logged.
- FR-021: values are truncated at byte length, which can split multi-byte characters.
- FR-022: tag key rules exist in the library but are never applied, so additional-tag keys are
  not checked.
- FR-027: the README and AGENTS.md list environment type under "Naming Configuration". The
  schema description lists only the allowed values and does not explain the purpose, RPO or
  RTO. The generated reference docs inherit that description.
- FR-001: the provider repeats the allowed-cloud list instead of reusing the library's
  validation.
- Tests: there are none for `internal/provider` or `internal/datasource` (bead `tfpctx-24o`).

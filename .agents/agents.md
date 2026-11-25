# Context Terraform Provider

> **Authoritative sources**: behavior is specified in `specs/001-context-provider/spec.md`
> (FR-001 to FR-028) and governed by `.specify/memory/principles.md`. When this guide and the
> spec disagree, the spec wins.

## Overview

This Terraform provider replicates the functionality of the `kbrockhoff/terraform-external-context` module. The provider should generate the same three primary outputs: `name_prefix`, `tags`, and `data_tags` while providing a more native Terraform experience.

## Provider Schema Requirements

### Provider Configuration

The provider should accept only the following configuration attributes for global cloud-specific settings:

- `cloud_provider` (string, optional) - Cloud provider identifier: `dc`, `aws`, `az`, `gcp`, `oci`, `ibm`, `do`, `vul`, `ali`, `cv`
- `tag_prefix` (string, optional, default: "bc-") - Prefix for all generated tags

## Data Source Requirements

### Context Data Source

The `context` data source contains all configuration fields and provides the computed outputs:

```hcl
provider "brockhoff" {
  cloud_provider = "aws"
  tag_prefix     = "bc-"
}

data "brockhoff_context" "main" {
  # Naming Configuration
  namespace       = "myorg"
  name            = "myapp"
  environment     = "prod"
  environment_name = "Production"
  environment_type = "Production"
  
  # Resource Management
  enabled        = true
  availability   = "preemptable"
  managedby      = "terraform"
  deletion_date  = "2024-12-31"
  
  # ... other attributes
}
```

#### Data Source Configuration Attributes

##### Naming Configuration
- `namespace` (string, optional) - Organization or business unit identifier (1-8 chars, lowercase alphanumeric with hyphens)
- `name` (string, optional) - Unique resource name (combined name_prefix must be 2-24 chars)
- `environment` (string, optional) - Environment abbreviation (1-8 chars, lowercase alphanumeric with hyphens)
- `environment_name` (string, optional) - Full environment name; the `environment` tag falls back to `environment`, then `dev`

##### Environment Classification
- `environment_type` (string, optional) - One of: None, Ephemeral, Development, Testing, UAT, Production, MissionCritical.
  Classifies the deployment stage so consuming configurations can choose resiliency (RPO/RTO) and cost defaults.
  Not a naming input: it never affects `name_prefix` and is not emitted as a tag. Only `Ephemeral` changes
  outputs (deletion date). See spec FR-027 for the RPO/RTO table.

##### Resource Management
- `enabled` (bool, optional, default: true) - Enable/disable resource creation
- `availability` (string, optional, default: "preemptable") - Availability requirement from predefined list
- `managedby` (string, optional, default: "terraform") - Management platform identifier
- `deletion_date` (string, optional) - Resource deletion date (YYYY-MM-DD format) or `never`

##### Project Management Integration
- `pm_platform` (string, optional) - Project management platform (e.g., JIRA, SNOW)
- `pm_project_code` (string, optional) - Project code/prefix

##### ITSM Integration
- `itsm_platform` (string, optional) - IT Service Management platform
- `itsm_system_id` (string, optional) - ITSM system identifier
- `itsm_component_id` (string, optional) - ITSM component identifier
- `itsm_instance_id` (string, optional) - ITSM instance identifier

##### Ownership and Billing
- `cost_center` (string, optional) - Cost center for billing
- `product_owners` (list(string), optional) - Product owner email addresses
- `code_owners` (list(string), optional) - Code owner email addresses
- `data_owners` (list(string), optional) - Data owner email addresses

##### Data Classification
- `sensitivity` (string, optional, default: "confidential") - Data sensitivity level from predefined list
- `data_regs` (list(string), optional) - Data compliance regulations
- `security_review` (string, optional) - Security review identifier/date
- `privacy_review` (string, optional) - Privacy review identifier/date

##### Feature Toggles
- `source_repo_tags_enabled` (bool, optional, default: true) - Include git repository tags
- `system_prefixes_enabled` (bool, optional, default: true) - Add platform prefixes to system IDs
- `not_applicable_enabled` (bool, optional, default: true) - Include N/A tags for null values
- `owner_tags_enabled` (bool, optional, default: true) - Include owner tags

##### Additional Tags
- `additional_tags` (map(string), optional) - Custom tags to merge
- `additional_data_tags` (map(string), optional) - Custom data-specific tags to merge

## Computed Attributes

### Primary Outputs

1. **name_prefix** (string) - Computed name prefix following Brockhoff standards:
   - Format: `namespace-name-environment`, joining only the parts that are set (e.g. `ns-app`)
   - Must match regex: `/^[a-z][a-z0-9-]{0,22}[a-z0-9]$/`
   - Maximum 24 characters with intelligent truncation

2. **tags** (map(string)) - Normalized tag map including:
   - Environment and availability tags
   - Management and ownership tags
   - Project management and ITSM tags
   - Source repository information (if enabled)
   - Custom additional tags

3. **data_tags** (map(string)) - Data-specific tags including:
   - Data sensitivity classification
   - Data owner information
   - Data regulation compliance tags
   - Custom additional data tags

### Additional Outputs

4. **tags_as_list_of_maps** (list(object)) - Tags as `{key, value, propagate_at_launch = "true"}` maps, sorted by key
5. **tags_as_kvp_list** (list(string)) - Tags as key=value pairs
6. **tags_as_comma_separated_string** (string) - Tags as comma-separated string
7. **data_tags_as_list_of_maps** (list(object)) - Data tags as `{key, value, propagate_at_launch = "true"}` maps, sorted by key
8. **data_tags_as_kvp_list** (list(string)) - Data tags as key=value pairs
9. **data_tags_as_comma_separated_string** (string) - Data tags as comma-separated string

## Business Logic Requirements

### Name Prefix Generation

1. **Standard Format**: If `namespace`, `name`, and `environment` are all provided:
   - Combine as: `namespace-name-environment`
   - Convert to lowercase
   
2. **Partial Format**: If `namespace` or `environment` are null:
   - Join the parts that are present, in namespace, name, environment order (e.g. `ns-app`)
   - This deliberately differs from the legacy module, which used `name` alone
   
3. **Length Validation**: Ensure final prefix matches `/^[a-z][a-z0-9-]{0,22}[a-z0-9]$/`
   - Maximum 24 characters total
   - If too long, truncate the `name` component to fit

### Cloud Provider Constraints

Apply different tag formatting rules based on `cloud_provider`:

- **AWS**: 256 char limit, replace `/[^a-zA-Z0-9 \\.:=+@_/-]/` with `_`, delimiter: ` `
- **Azure**: 256 char limit, replace `/[ <>%&\\?/#:]/` with `""`, delimiter: `;`, N/A: "NotApplicable"  
- **GCP**: 63 char limit, lowercase tags, replace `/[^a-z0-9_-]/` with `-`, delimiter: `_`, N/A: "not_applicable"
- **Default/DC**: 63 char limit, replace `/[<>%&\\?]/` with `_`, delimiter: `;`

### Tag Processing

1. **Tag Keys**: resource tags are `environment`, `availability`, `managedby`, `deletiondate`, `costcenter`,
   `projectmgmtid`, `systemid`, `componentid`, `instanceid`, `productowners`, `codeowners`, `securityreview`,
   `privacyreview`, `sourcerepo`, `sourcecommit`; data tags are `sensitivity`, `dataregulations`, `dataowners`,
   `sovereignty`. All are prefixed with `tag_prefix`.

2. **Conditional Tags**:
   - `environment`, `availability`, `managedby` and `deletiondate` are always emitted
   - N/A substitution applies only to `costcenter`, `productowners`, `codeowners`, `securityreview`,
     `privacyreview`, `dataregulations` and `dataowners`
   - Owner tags, including `dataowners`, are omitted entirely when `owner_tags_enabled` is false
   - `projectmgmtid` and the ITSM id tags are emitted only when their identifier is set (never N/A)
   - `sovereignty` (from `data_sovereignty`) is emitted only when set (never N/A); it is free text with
     ISO 3166-1/3166-2 codes as the recommended, unvalidated convention

3. **System Prefixes**: When `system_prefixes_enabled` is true:
   - Prepend platform names to ITSM and PM codes
   - Format: `platform<delimiter>code`

4. **Git Repository Integration**: When `source_repo_tags_enabled` is true:
   - Execute git commands to get repository URL and commit hash
   - Convert SSH URLs to HTTPS format and strip any embedded credentials
   - Include `sourcerepo` and `sourcecommit` tags
   - When git information is unavailable, skip these tags and log a warning (never fail)

5. **Key Validation**: prefixed tag keys must satisfy the target cloud's key rules; invalid keys are rejected

### Ephemeral Environment Handling

Deletion date resolution:
- An explicit `deletion_date` (directly or inherited) always wins
- Otherwise, for `environment_type = "Ephemeral"`, today (UTC) plus 90 days
- Otherwise `never`

### Validation Requirements

1. **Input Validation**: 
   - Validate cloud provider against allowed list
   - Validate environment type against allowed list
   - Validate name patterns and lengths
   - Validate sensitivity and availability against allowed values

2. **Output Validation**:
   - Ensure name_prefix matches required regex pattern
   - Apply cloud provider-specific tag constraints
   - Truncate tag values that exceed length limits, never splitting a multi-byte character

## Implementation Considerations

### State Management
- The provider should be stateless and re-compute outputs on each plan/apply
- Git repository information should be fetched during plan phase
- Consider caching git information to avoid repeated calls

### Error Handling
- Gracefully handle missing git repository information
- Provide clear error messages for validation failures
- Support partial functionality when optional features fail

### Performance
- Minimize external command execution
- Cache computed values during single terraform run
- Avoid blocking operations where possible

### Backwards Compatibility
- Maintain exact output compatibility with existing module
- Support all existing input parameters and validation rules
- Preserve tag formatting and naming conventions

## Testing Requirements

The provider should include comprehensive tests covering:

1. **Unit Tests**: All business logic functions
2. **Integration Tests**: End-to-end data source functionality  
3. **Cloud Provider Tests**: Validation of cloud-specific constraints
4. **Edge Cases**: Truncation, missing inputs, invalid configurations
5. **Git Integration Tests**: Repository information extraction

## Documentation Requirements

1. **Provider Documentation**: Complete attribute reference
2. **Data Source Documentation**: Usage examples and output descriptions
3. **Migration Guide**: Instructions for migrating from existing module
4. **Examples**: Common usage patterns and configurations

<!-- BEGIN BEADS INTEGRATION v:1 profile:minimal hash:46cd31e7 -->
## Beads Issue Tracker

This project uses **bd (beads)** for issue tracking. Run `bd prime` to see full workflow context and commands.

### Quick Reference

```bash
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --claim  # Claim work
bd close <id>         # Complete work
```

### Rules

- Use `bd` for ALL task tracking — do NOT use TodoWrite, TaskCreate, or markdown TODO lists
- Run `bd prime` for detailed command reference and session close protocol
- Use `bd remember` for persistent knowledge — do NOT use MEMORY.md files

**Architecture in one line:** issues live in a local Dolt DB; sync uses `refs/dolt/data` on your git remote; `.beads/issues.jsonl` is a passive export. See https://github.com/gastownhall/beads/blob/main/docs/core-concepts/sync-concepts.md for details and anti-patterns.

## Agent Context Profiles

The managed Beads block is task-tracking guidance, not permission to override repository, user, or orchestrator instructions.

- **Conservative (default)**: Use `bd` for task tracking. Do not run git commits, git pushes, or Dolt remote sync unless explicitly asked. At handoff, report changed files, validation, and suggested next commands.
- **Minimal**: Keep tool instruction files as pointers to `bd prime`; use the same conservative git policy unless active instructions say otherwise.
- **Team-maintainer**: Only when the repository explicitly opts in, agents may close beads, run quality gates, commit, and push as part of session close. A current "do not commit" or "do not push" instruction still wins.

## Session Completion

This protocol applies when ending a Beads implementation workflow. It is subordinate to explicit user, repository, and orchestrator instructions.

1. **File issues for remaining work** - Create beads for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **Handle git/sync by active profile**:
   ```bash
   # Conservative/minimal/default: report status and proposed commands; wait for approval.
   git status

   # Team-maintainer opt-in only, unless current instructions forbid it:
   git pull --rebase
   bd dolt push
   git push
   git status
   ```
5. **Hand off** - Summarize changes, validation, issue status, and any blocked sync/commit/push step

**Critical rules:**
- Explicit user or orchestrator instructions override this Beads block.
- Do not commit or push without clear authority from the active profile or the current user request.
- If a required sync or push is blocked, stop and report the exact command and error.
<!-- END BEADS INTEGRATION -->

<!-- BEGIN BEADS CODEX SETUP: generated by bd setup codex -->
## Beads Issue Tracker

Use Beads (`bd`) for durable task tracking in repositories that include it. Use the `beads` skill at `.agents/skills/beads/SKILL.md` (project install) or `~/.agents/skills/beads/SKILL.md` (global install) for Beads workflow guidance, then use the `bd` CLI for issue operations.

### Quick Reference

```bash
bd ready                # Find available work
bd show <id>            # View issue details
bd update <id> --claim  # Claim work
bd close <id>           # Complete work
bd prime                # Refresh Beads context
```

### Rules

- Use `bd` for all task tracking; do not create markdown TODO lists.
- Run `bd prime` when Beads context is missing or stale. Codex 0.129.0+ can load Beads context automatically through native hooks; use `/hooks` to inspect or toggle them.
- Keep persistent project memory in Beads via `bd remember`; do not create ad hoc memory files.

**Architecture in one line:** issues live in a local Dolt DB; sync uses `refs/dolt/data` on your git remote; `.beads/issues.jsonl` is a passive export. See https://github.com/gastownhall/beads/blob/main/docs/core-concepts/sync-concepts.md for details and anti-patterns.
<!-- END BEADS CODEX SETUP -->

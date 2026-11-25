# Terraform Provider: Context

A Terraform provider that generates standardized naming conventions and cloud-provider-specific tags for infrastructure resources, replacing the functionality of the `kbrockhoff/terraform-external-context` module with a native Terraform provider implementation.

## Features

- **Standardized Naming**: Generates name prefixes following Brockhoff naming standards
- **Cloud Provider Support**: AWS, Azure, GCP, and more with provider-specific tag formatting
- **Flexible Configuration**: Comprehensive set of configuration options for various use cases
- **Git Integration**: Automatically includes source repository information in tags
- **Multiple Output Formats**: Tags available in various formats (map, list, CSV, etc.)

## Quick Start

```hcl
terraform {
  required_providers {
    brockhoff = {
      source = "kbrockhoff/context"
    }
  }
}

# Provider configuration
provider "brockhoff" {
  cloud_provider = "aws"
  tag_prefix     = "bc-"
}

# Basic usage
data "brockhoff_context" "app" {
  namespace   = "myorg"
  name        = "webapp"
  environment = "prod"
}

# Use the generated outputs
resource "aws_instance" "example" {
  # ... other configuration ...
  
  tags = data.brockhoff_context.app.tags
}
```

## Provider Configuration

The provider accepts minimal configuration focused on cloud-specific settings:

| Argument | Description | Type | Default |
|----------|-------------|------|---------|
| `cloud_provider` | Cloud provider identifier (`dc`, `aws`, `az`, `gcp`, `oci`, `ibm`, `do`, `vul`, `ali`, `cv`) | `string` | `"dc"` |
| `tag_prefix` | Prefix for all generated tags | `string` | `"bc-"` |

## Data Source: `brockhoff_context`

### Configuration Arguments

#### Naming Configuration
- `namespace` (Optional) - Organization or business unit identifier (1-8 chars)
- `name` (Optional) - Unique resource name
- `environment` (Optional) - Environment abbreviation (1-8 chars)
- `environment_name` (Optional) - Full environment name, emitted as the `environment` tag (falls back to `environment`, then `dev`)

#### Environment Classification
- `environment_type` (Optional) - Deployment stage classification. Consuming configurations use it to choose
  resiliency and cost defaults; the provider only validates it and passes it through `context_output`. It does
  not affect `name_prefix` and is not emitted as a tag. `Ephemeral` sets a deletion date of today (UTC) + 90 days.

| Type | Use Case | Configuration Focus | RPO | RTO |
|------|----------|---------------------|-----|-----|
| `None` | Custom configuration | No defaults applied, use individual config values | N/A | N/A |
| `Ephemeral` | Temporary environments | Cost-optimized, minimal durability requirements | N/A | 48h |
| `Development` | Developer workspaces | Balanced cost and functionality for active development | 24h | 48h |
| `Testing` | Automated testing | Consistent, repeatable configurations | 24h | 48h |
| `UAT` | User acceptance testing | Production-like settings with some cost optimization | 12h | 24h |
| `Production` | Live systems | High availability, durability, and performance | 1h | 4h |
| `MissionCritical` | Critical production | Maximum reliability, redundancy, and monitoring | 5m | 1h |

#### Resource Management
- `enabled` (Optional) - Enable/disable resource creation (default: `true`)
- `availability` (Optional) - Availability level (default: `"preemptable"`)
- `managedby` (Optional) - Management platform identifier (default: `"terraform"`)
- `deletion_date` (Optional) - Resource deletion date (`YYYY-MM-DD`) or `never`. Defaults to today (UTC) + 90 days
  for `Ephemeral` environments, otherwise `never`. An explicit value always wins.

#### Integration & Ownership
- `pm_platform` / `pm_project_code` - Project management integration
- `itsm_platform` / `itsm_system_id` / `itsm_component_id` / `itsm_instance_id` - ITSM integration
- `cost_center` - Cost center for billing
- `product_owners` / `code_owners` / `data_owners` - Owner email addresses

#### Data Classification
- `sensitivity` (Optional) - Data sensitivity level (default: `"confidential"`)
- `data_regs` - Data compliance regulations
- `data_sovereignty` (Optional) - Jurisdiction governing where the data may be stored and processed. By convention an
  ISO 3166-1 country code (`DE`) or ISO 3166-2 subdivision code (`US-CA`); free text, not validated against ISO.
  Emitted as the `sovereignty` data tag only when set, never as N/A.
- `security_review` / `privacy_review` - Review identifiers/dates

#### Feature Toggles
- `source_repo_tags_enabled` (Optional) - Include git repository tags (default: `true`)
- `system_prefixes_enabled` (Optional) - Add platform prefixes to system IDs (default: `true`)
- `not_applicable_enabled` (Optional) - Include N/A tags for null values (default: `true`)
- `owner_tags_enabled` (Optional) - Include owner tags (default: `true`)

#### Additional Tags
- `additional_tags` - Custom tags to merge
- `additional_data_tags` - Custom data-specific tags to merge

### Computed Attributes

#### Primary Outputs
- `name_prefix` - Generated name prefix
- `tags` - Main tags map
- `data_tags` - Data-specific tags map

#### Alternative Formats
- `tags_as_list_of_maps` - Tags as `{key, value, propagate_at_launch = "true"}` maps, sorted by key
- `tags_as_kvp_list` - Tags as key=value pairs
- `tags_as_comma_separated_string` - Tags as comma-separated string
- `data_tags_as_list_of_maps` - Data tags as `{key, value, propagate_at_launch = "true"}` maps, sorted by key
- `data_tags_as_kvp_list` - Data tags as key=value pairs
- `data_tags_as_comma_separated_string` - Data tags as comma-separated string

### Generated Tag Keys

Every key is prefixed with the provider's `tag_prefix`.

| Tag set | Key | When emitted |
|---------|-----|--------------|
| `tags` | `environment`, `availability`, `managedby`, `deletiondate` | Always |
| `tags` | `costcenter`, `securityreview`, `privacyreview` | Value, or the cloud's N/A value when `not_applicable_enabled` |
| `tags` | `productowners`, `codeowners` | When `owner_tags_enabled`: value, or N/A as above |
| `tags` | `projectmgmtid`, `systemid`, `componentid`, `instanceid` | Only when the identifier is set (platform-prefixed when `system_prefixes_enabled`) |
| `tags` | `sourcerepo`, `sourcecommit` | When `source_repo_tags_enabled` and git information is available (credential-free URL) |
| `data_tags` | `sensitivity` | Always |
| `data_tags` | `dataregulations` | Value, or N/A when `not_applicable_enabled` |
| `data_tags` | `dataowners` | When `owner_tags_enabled`: value, or N/A as above |
| `data_tags` | `sovereignty` | Only when `data_sovereignty` is set |

Additional tags are merged last and override generated keys. Tag keys that are invalid for the target cloud
(for example uppercase keys on `gcp`) are rejected.

## Examples

### Minimal Configuration

```hcl
data "brockhoff_context" "minimal" {
  name = "myapp"
}

output "name_prefix" {
  value = data.brockhoff_context.minimal.name_prefix
}
```

### Full Configuration

```hcl
provider "brockhoff" {
  cloud_provider = "aws"
  tag_prefix     = "bc-"
}

data "brockhoff_context" "full" {
  namespace        = "myorg"
  name             = "payment-api"
  environment      = "prod"
  environment_type = "Production"

  cost_center     = "engineering"
  product_owners  = ["product@example.com"]
  code_owners     = ["dev@example.com"]

  sensitivity     = "confidential"
  data_regs       = ["GDPR", "CCPA"]

  additional_tags = {
    team = "platform"
    tier = "1"
  }
}
```

### Cloud Provider Specific

```hcl
# AWS Configuration
provider "brockhoff" {
  alias          = "aws"
  cloud_provider = "aws"
  tag_prefix     = "aws-"
}

# GCP Configuration  
provider "brockhoff" {
  alias          = "gcp"
  cloud_provider = "gcp"
  tag_prefix     = "gcp_"
}
```

### Ephemeral Environments

```hcl
data "brockhoff_context" "ephemeral" {
  namespace        = "test"
  name             = "feature-branch"
  environment      = "eph"
  environment_type = "Ephemeral" # deletion_date defaults to today (UTC) + 90 days

  availability = "spot"
}
```

## Cloud Provider Differences

### AWS
- 256 character limit for tag values
- Delimiter: space ` `
- N/A value: `"N/A"`
- Sanitization: Replace non-alphanumeric/space/allowed chars with `_`

### Azure
- 256 character limit for tag values
- Delimiter: semicolon `;`
- N/A value: `"NotApplicable"`
- Sanitization: Remove spaces and special characters

### GCP
- 63 character limit for tag values
- Delimiter: underscore `_`
- N/A value: `"not_applicable"`
- Sanitization: Convert to lowercase, replace non-alphanumeric with hyphens

## Development

### Building

```bash
go build
```

### Testing

```bash
go test ./...
```

### Examples

See the [examples/](examples/) directory for various usage patterns.

## Migration from Module

This provider replaces the `kbrockhoff/terraform-external-context` module. Key differences:

1. **Provider Configuration**: Only `cloud_provider` and `tag_prefix` are at provider level
2. **Data Source**: All other configuration moved to the data source
3. **Native Terraform**: No external script dependencies
4. **Enhanced Performance**: Reduced external command execution
5. **Name prefix composition**: When only one of `namespace` or `environment` is set, `name_prefix` joins the
   parts that are present (for example `ns-app`); the module used `name` alone. This can rename resources.
6. **No `name` inheritance**: `name` is never inherited from `parent_context`.
7. **Tag key validation**: Tag keys that are invalid for the target cloud are rejected.

### Upgrade Notes

Upgrading from earlier provider releases changes these tag values once, on the first plan after upgrading, to
match the legacy module:

- `projectmgmtid`, `systemid`, `componentid` and `instanceid` are omitted when unset (previously N/A).
- `environment` falls back to `environment`, then `dev` (previously N/A).
- `deletiondate` is `never` for non-Ephemeral contexts without a date (previously N/A).
- `dataowners` is omitted when `owner_tags_enabled = false` (previously N/A).
- `tags_as_list_of_maps` and `data_tags_as_list_of_maps` entries gain `propagate_at_launch = "true"`.
- The new optional `data_sovereignty` input adds a `sovereignty` data tag only when set.

### Migration Example

```hcl
# Before (Module)
module "context" {
  source = "kbrockhoff/terraform-external-context"
  
  namespace   = "myorg"
  name        = "myapp"
  environment = "prod"
}

# After (Provider)
provider "brockhoff" {
  cloud_provider = "aws"
  tag_prefix     = "bc-"
}

data "brockhoff_context" "main" {
  namespace   = "myorg"
  name        = "myapp"
  environment = "prod"
}
```

## License

This project is licensed under the ASL2 License.
 See [LICENSE](LICENSE) for details.

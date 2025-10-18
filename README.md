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
- `environment_name` (Optional) - Full environment name
- `environment_type` (Optional) - Environment type: `None`, `Ephemeral`, `Development`, `Testing`, `UAT`, `Production`, `MissionCritical`

#### Resource Management
- `enabled` (Optional) - Enable/disable resource creation (default: `true`)
- `availability` (Optional) - Availability level (default: `"preemptable"`)
- `managedby` (Optional) - Management platform identifier (default: `"terraform"`)
- `deletion_date` (Optional) - Resource deletion date (YYYY-MM-DD format)

#### Integration & Ownership
- `pm_platform` / `pm_project_code` - Project management integration
- `itsm_platform` / `itsm_system_id` / `itsm_component_id` / `itsm_instance_id` - ITSM integration
- `cost_center` - Cost center for billing
- `product_owners` / `code_owners` / `data_owners` - Owner email addresses

#### Data Classification
- `sensitivity` (Optional) - Data sensitivity level (default: `"confidential"`)
- `data_regs` - Data compliance regulations
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
- `tags_as_list_of_maps` - Tags formatted for AWS resources
- `tags_as_kvp_list` - Tags as key=value pairs
- `tags_as_comma_separated_string` - Tags as comma-separated string
- `data_tags_as_list_of_maps` - Data tags formatted for AWS resources
- `data_tags_as_kvp_list` - Data tags as key=value pairs  
- `data_tags_as_comma_separated_string` - Data tags as comma-separated string

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
  environment      = "ephemeral"
  environment_type = "Ephemeral" # Auto-calculates deletion_date

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

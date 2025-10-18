terraform {
  required_providers {
    brockhoff = {
      source = "kbrockhoff/context"
    }
  }
}

provider "brockhoff" {
  cloud_provider = "aws"
  tag_prefix     = "app-"
}

# Parent context for the entire application stack
data "brockhoff_context" "parent" {
  namespace        = "platform"
  environment      = "prod"
  environment_name = "Production"
  environment_type = "Production"

  # Common settings for all components
  availability  = "dedicated"
  managedby     = "terraform"
  cost_center   = "engineering"
  product_owners = ["product@example.com"]
  code_owners    = ["platform-team@example.com"]

  # Project tracking
  pm_platform     = "JIRA"
  pm_project_code = "PLAT-100"

  # ITSM
  itsm_platform  = "ServiceNow"
  itsm_system_id = "SYS-PLATFORM"

  # Common tags
  additional_tags = {
    team        = "platform"
    application = "payment-stack"
    tier        = "production"
  }
}

# Child context for API component
# Inherits all parent settings but overrides name and adds component-specific tags
data "brockhoff_context" "api" {
  # Use parent context as base
  context = data.brockhoff_context.parent.context_output

  # Override only what's different for this component
  name = "payment-api"

  # Add component-specific ITSM ID
  itsm_component_id = "COMP-API"

  # Add component-specific tags
  additional_tags = {
    component = "api"
    language  = "go"
    port      = "8080"
  }
}

# Child context for database component
# Inherits from parent, different name and component tags
data "brockhoff_context" "database" {
  # Use parent context as base
  context = data.brockhoff_context.parent.context_output

  # Override for this component
  name = "payment-db"

  # Database has higher sensitivity
  sensitivity = "critical"

  # Add component-specific ITSM ID
  itsm_component_id = "COMP-DB"

  # Add data regulations
  data_regs = ["PCI-DSS", "SOC2"]

  # Add component-specific tags
  additional_tags = {
    component = "database"
    engine    = "postgresql"
    version   = "15"
  }
}

# Child context for cache component
data "brockhoff_context" "cache" {
  # Use parent context as base
  context = data.brockhoff_context.parent.context_output

  # Override for this component
  name = "payment-cache"

  # Cache can use lower availability
  availability = "standard"

  # Add component-specific ITSM ID
  itsm_component_id = "COMP-CACHE"

  # Add component-specific tags
  additional_tags = {
    component = "cache"
    engine    = "redis"
    version   = "7"
  }
}

# Outputs showing the inheritance
output "parent_name_prefix" {
  description = "Parent context name prefix"
  value       = data.brockhoff_context.parent.name_prefix
}

output "api_name_prefix" {
  description = "API inherits namespace and environment from parent"
  value       = data.brockhoff_context.api.name_prefix
}

output "database_name_prefix" {
  description = "Database inherits namespace and environment from parent"
  value       = data.brockhoff_context.database.name_prefix
}

output "cache_name_prefix" {
  description = "Cache inherits namespace and environment from parent"
  value       = data.brockhoff_context.cache.name_prefix
}

output "parent_tags" {
  description = "Parent context tags"
  value       = data.brockhoff_context.parent.tags
}

output "api_tags" {
  description = "API tags (inherited + component-specific)"
  value       = data.brockhoff_context.api.tags
}

output "database_tags" {
  description = "Database tags (inherited + overridden sensitivity)"
  value       = data.brockhoff_context.database.tags
}

output "cache_tags" {
  description = "Cache tags (inherited + overridden availability)"
  value       = data.brockhoff_context.cache.tags
}

output "api_data_tags" {
  description = "API data tags"
  value       = data.brockhoff_context.api.data_tags
}

output "database_data_tags" {
  description = "Database data tags (includes PCI-DSS)"
  value       = data.brockhoff_context.database.data_tags
}

terraform {
  required_providers {
    brockhoff = {
      source = "kbrockhoff/context"
    }
  }
}

# Provider configuration for AWS
provider "brockhoff" {
  cloud_provider = "aws"
  tag_prefix     = "bc-"
}

# Full configuration example
data "brockhoff_context" "full" {
  # Naming Configuration
  namespace        = "myorg"
  name             = "payment-api"
  environment      = "prod"
  environment_name = "Production"
  environment_type = "Production"

  # Resource Management
  enabled       = true
  availability  = "dedicated"
  managedby     = "terraform"
  deletion_date = "2024-12-31"

  # Project Management Integration
  pm_platform     = "JIRA"
  pm_project_code = "PAY-123"

  # ITSM Integration
  itsm_platform     = "ServiceNow"
  itsm_system_id    = "SYS-001"
  itsm_component_id = "COMP-PAY-001"
  itsm_instance_id  = "INST-001"

  # Ownership and Billing
  cost_center     = "engineering"
  product_owners  = ["product@example.com", "manager@example.com"]
  code_owners     = ["dev@example.com", "lead@example.com"]
  data_owners     = ["data@example.com"]

  # Data Classification
  sensitivity     = "confidential"
  data_regs       = ["GDPR", "CCPA", "PCI-DSS"]
  security_review = "2024-01-15"
  privacy_review  = "2024-01-20"

  # Feature Toggles
  source_repo_tags_enabled = true
  system_prefixes_enabled  = true
  not_applicable_enabled   = true
  owner_tags_enabled       = true

  # Additional Tags
  additional_tags = {
    team        = "platform"
    tier        = "1"
    compliance  = "high"
    environment = "production"
  }

  additional_data_tags = {
    encryption = "AES256"
    retention  = "7years"
    backup     = "daily"
  }
}

# Outputs
output "name_prefix" {
  description = "Generated name prefix"
  value       = data.brockhoff_context.full.name_prefix
}

output "tags" {
  description = "Generated tags map"
  value       = data.brockhoff_context.full.tags
}

output "data_tags" {
  description = "Generated data tags map"
  value       = data.brockhoff_context.full.data_tags
}

output "tags_as_list" {
  description = "Tags formatted as list of maps for AWS"
  value       = data.brockhoff_context.full.tags_as_list_of_maps
}

output "tags_kvp" {
  description = "Tags as key=value pairs"
  value       = data.brockhoff_context.full.tags_as_kvp_list
}

output "tags_csv" {
  description = "Tags as comma-separated string"
  value       = data.brockhoff_context.full.tags_as_comma_separated_string
}
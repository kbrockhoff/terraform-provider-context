terraform {
  required_providers {
    brockhoff = {
      source = "kbrockhoff/context"
    }
  }
}

provider "brockhoff" {
  cloud_provider = "aws"
}

# Ephemeral environment example - auto-calculates deletion date
data "brockhoff_context" "ephemeral" {
  namespace        = "test"
  name             = "feature-branch"
  environment      = "eph"
  environment_type = "Ephemeral" # This triggers auto-calculation of deletion_date

  availability = "spot"
  
  additional_tags = {
    branch = "feature/new-feature"
    ci_run = "12345"
  }
}

output "ephemeral_name" {
  value = data.brockhoff_context.ephemeral.name_prefix
}

output "ephemeral_deletion_date" {
  description = "Auto-calculated deletion date (90 days from now)"
  value       = lookup(data.brockhoff_context.ephemeral.tags, "bc-DeletionDate", "not set")
}

output "ephemeral_tags" {
  value = data.brockhoff_context.ephemeral.tags
}
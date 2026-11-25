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

# Ephemeral environment example - auto-calculates deletion date.
# When deletion_date is unset, Ephemeral contexts get today (UTC) plus 90 days.
# An explicit deletion_date always wins, and non-Ephemeral contexts default to "never".
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
  description = "Auto-calculated deletion date (90 days from now, UTC)"
  value       = lookup(data.brockhoff_context.ephemeral.tags, "bc-deletiondate", "not set")
}

output "ephemeral_tags" {
  value = data.brockhoff_context.ephemeral.tags
}
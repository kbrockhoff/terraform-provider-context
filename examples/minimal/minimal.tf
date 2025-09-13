terraform {
  required_providers {
    brockhoff = {
      source = "kbrockhoff/context"
    }
  }
}

# Minimal configuration with defaults
provider "brockhoff" {}

data "brockhoff_context" "minimal" {
  name = "myapp"
}

output "name_prefix" {
  value = data.brockhoff_context.minimal.name_prefix
}

output "tags" {
  value = data.brockhoff_context.minimal.tags
}
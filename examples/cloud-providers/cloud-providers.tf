terraform {
  required_providers {
    brockhoff = {
      source = "kbrockhoff/context"
    }
  }
}

# AWS Configuration
provider "brockhoff" {
  alias          = "aws"
  cloud_provider = "aws"
  tag_prefix     = "aws-"
}

data "brockhoff_context" "aws_app" {
  provider = brockhoff.aws

  namespace   = "myorg"
  name        = "webapp"
  environment = "staging"

  cost_center = "engineering"
  sensitivity = "internal"

  additional_tags = {
    stack = "aws"
  }
}

# Azure Configuration
provider "brockhoff" {
  alias          = "azure"
  cloud_provider = "az"
  tag_prefix     = "az-"
}

data "brockhoff_context" "azure_app" {
  provider = brockhoff.azure

  namespace   = "myorg"
  name        = "webapp"
  environment = "staging"

  cost_center = "engineering"
  sensitivity = "internal"

  additional_tags = {
    stack = "azure"
  }
}

# GCP Configuration
provider "brockhoff" {
  alias          = "gcp"
  cloud_provider = "gcp"
  tag_prefix     = "gcp_"
}

data "brockhoff_context" "gcp_app" {
  provider = brockhoff.gcp

  namespace   = "myorg"
  name        = "webapp"
  environment = "staging"

  cost_center = "engineering"
  sensitivity = "internal"

  additional_tags = {
    stack = "gcp"
  }
}

# Outputs to show cloud-specific differences
output "aws_tags" {
  description = "AWS-formatted tags"
  value       = data.brockhoff_context.aws_app.tags
}

output "azure_tags" {
  description = "Azure-formatted tags"
  value       = data.brockhoff_context.azure_app.tags
}

output "gcp_tags" {
  description = "GCP-formatted tags (lowercase)"
  value       = data.brockhoff_context.gcp_app.tags
}
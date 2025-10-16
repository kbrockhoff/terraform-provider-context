data "brockhoff_context" "example" {
  namespace        = "myorg"
  name             = "myapp"
  environment      = "prod"
  environment_name = "Production"
  environment_type = "Production"
}

# Use the computed outputs
resource "aws_s3_bucket" "example" {
  bucket = data.brockhoff_context.example.name_prefix

  tags = data.brockhoff_context.example.tags
}

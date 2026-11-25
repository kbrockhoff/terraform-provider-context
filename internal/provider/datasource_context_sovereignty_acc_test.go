package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccContext_DataSovereignty(t *testing.T) {
	const awsConfig = `
provider "brockhoff" {
  cloud_provider = "aws"
}

data "brockhoff_context" "supplied" {
  name                     = "app"
  data_sovereignty         = "US-CA"
  source_repo_tags_enabled = false
}

data "brockhoff_context" "absent" {
  name                     = "app"
  not_applicable_enabled   = true
  source_repo_tags_enabled = false
}

data "brockhoff_context" "parent" {
  name                     = "parent"
  data_sovereignty         = "CA-QC"
  source_repo_tags_enabled = false
}

data "brockhoff_context" "child" {
  parent_context = data.brockhoff_context.parent.context_output
  name           = "child"
}
`
	const gcpConfig = `
provider "brockhoff" {
  cloud_provider = "gcp"
}

data "brockhoff_context" "supplied" {
  name                     = "app"
  data_sovereignty         = "US-CA"
  source_repo_tags_enabled = false
}
`
	const blankConfig = `
data "brockhoff_context" "blank" {
  name                     = "app"
  data_sovereignty         = "   "
  source_repo_tags_enabled = false
}
`
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: awsConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.brockhoff_context.supplied", "data_tags.bc-sovereignty", "US-CA"),
					resource.TestCheckNoResourceAttr("data.brockhoff_context.supplied", "tags.bc-sovereignty"),
					resource.TestCheckNoResourceAttr("data.brockhoff_context.absent", "data_tags.bc-sovereignty"),
					resource.TestCheckResourceAttr("data.brockhoff_context.child", "data_tags.bc-sovereignty", "CA-QC"),
					resource.TestCheckResourceAttr("data.brockhoff_context.child", "context_output.data_sovereignty", "CA-QC"),
				),
			},
			planStable(awsConfig),
			{
				Config: gcpConfig,
				Check:  resource.TestCheckResourceAttr("data.brockhoff_context.supplied", "data_tags.bc-sovereignty", "us-ca"),
			},
			{
				Config:      blankConfig,
				ExpectError: regexp.MustCompile(`data_sovereignty`),
			},
		},
	})
}

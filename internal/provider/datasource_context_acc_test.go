package provider

import (
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// planStable returns a plan-only step that expects no changes for the same configuration
func planStable(config string) resource.TestStep {
	return resource.TestStep{
		Config:   config,
		PlanOnly: true,
	}
}

func TestAccContext_Minimal(t *testing.T) {
	const config = `
data "brockhoff_context" "test" {
  name                     = "webapp"
  source_repo_tags_enabled = false
}
`
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.brockhoff_context.test", "name_prefix", "webapp"),
					resource.TestCheckResourceAttr("data.brockhoff_context.test", "tags.bc-environment", "dev"),
					resource.TestCheckResourceAttr("data.brockhoff_context.test", "tags.bc-deletiondate", "never"),
					resource.TestCheckResourceAttr("data.brockhoff_context.test", "tags.bc-availability", "preemptable"),
					resource.TestCheckResourceAttr("data.brockhoff_context.test", "tags.bc-managedby", "terraform"),
					resource.TestCheckResourceAttr("data.brockhoff_context.test", "tags.bc-costcenter", "N/A"),
					resource.TestCheckResourceAttr("data.brockhoff_context.test", "data_tags.bc-sensitivity", "confidential"),
					resource.TestCheckNoResourceAttr("data.brockhoff_context.test", "tags.bc-projectmgmtid"),
					resource.TestCheckNoResourceAttr("data.brockhoff_context.test", "tags.bc-systemid"),
					resource.TestCheckNoResourceAttr("data.brockhoff_context.test", "tags.bc-sourcerepo"),
					resource.TestCheckNoResourceAttr("data.brockhoff_context.test", "data_tags.bc-sovereignty"),
					resource.TestCheckResourceAttr("data.brockhoff_context.test", "context_output.enabled", "true"),
				),
			},
			planStable(config),
		},
	})
}

func TestAccContext_Full(t *testing.T) {
	const config = `
provider "brockhoff" {
  cloud_provider = "aws"
  tag_prefix     = "bc-"
}

data "brockhoff_context" "test" {
  namespace        = "myorg"
  name             = "payment-api"
  environment      = "prod"
  environment_name = "Production"
  environment_type = "Production"

  enabled       = false
  availability  = "dedicated"
  managedby     = "terraform"
  deletion_date = "2030-12-31"

  pm_platform       = "JIRA"
  pm_project_code   = "PAY"
  itsm_platform     = "SNOW"
  itsm_system_id    = "S1"
  itsm_component_id = "C1"
  itsm_instance_id  = "I1"

  cost_center    = "engineering"
  product_owners = ["product@example.com"]
  code_owners    = ["dev1@example.com", "dev2@example.com"]
  data_owners    = ["data@example.com"]

  sensitivity      = "restricted"
  data_regs        = ["GDPR", "CCPA"]
  data_sovereignty = "DE"
  security_review  = "SR-1"
  privacy_review   = "PR-1"

  source_repo_tags_enabled = false

  additional_tags      = { team = "platform" }
  additional_data_tags = { retention = "7y" }
}
`
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.brockhoff_context.test", "name_prefix", "myorg-payment-api-prod"),
					resource.TestCheckResourceAttr("data.brockhoff_context.test", "tags.bc-environment", "Production"),
					resource.TestCheckResourceAttr("data.brockhoff_context.test", "tags.bc-deletiondate", "2030-12-31"),
					resource.TestCheckResourceAttr("data.brockhoff_context.test", "tags.bc-projectmgmtid", "JIRA PAY"),
					resource.TestCheckResourceAttr("data.brockhoff_context.test", "tags.bc-systemid", "SNOW S1"),
					resource.TestCheckResourceAttr("data.brockhoff_context.test", "tags.bc-componentid", "SNOW C1"),
					resource.TestCheckResourceAttr("data.brockhoff_context.test", "tags.bc-instanceid", "SNOW I1"),
					resource.TestCheckResourceAttr("data.brockhoff_context.test", "tags.bc-codeowners", "dev1@example.com dev2@example.com"),
					resource.TestCheckResourceAttr("data.brockhoff_context.test", "tags.bc-team", "platform"),
					resource.TestCheckResourceAttr("data.brockhoff_context.test", "data_tags.bc-dataregulations", "GDPR CCPA"),
					resource.TestCheckResourceAttr("data.brockhoff_context.test", "data_tags.bc-dataowners", "data@example.com"),
					resource.TestCheckResourceAttr("data.brockhoff_context.test", "data_tags.bc-sovereignty", "DE"),
					resource.TestCheckResourceAttr("data.brockhoff_context.test", "data_tags.bc-retention", "7y"),
					resource.TestCheckResourceAttr("data.brockhoff_context.test", "tags_as_list_of_maps.0.key", "bc-availability"),
					resource.TestCheckResourceAttr("data.brockhoff_context.test", "tags_as_list_of_maps.0.propagate_at_launch", "true"),
					resource.TestCheckResourceAttr("data.brockhoff_context.test", "data_tags_as_list_of_maps.0.propagate_at_launch", "true"),
					resource.TestCheckResourceAttr("data.brockhoff_context.test", "tags_as_kvp_list.0", "bc-availability=dedicated"),
					resource.TestCheckResourceAttr("data.brockhoff_context.test", "context_output.enabled", "false"),
				),
			},
			planStable(config),
		},
	})
}

func TestAccContext_Ephemeral(t *testing.T) {
	const computed = `
data "brockhoff_context" "test" {
  name                     = "feature"
  environment_type         = "Ephemeral"
  source_repo_tags_enabled = false
}
`
	const explicit = `
data "brockhoff_context" "test" {
  name                     = "feature"
  environment_type         = "Ephemeral"
  deletion_date            = "2030-01-01"
  source_repo_tags_enabled = false
}
`
	want := time.Now().UTC().Add(90 * 24 * time.Hour).Format("2006-01-02")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: computed,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.brockhoff_context.test", "tags.bc-deletiondate", want),
					resource.TestCheckResourceAttr("data.brockhoff_context.test", "context_output.deletion_date", want),
				),
			},
			planStable(computed),
			{
				Config: explicit,
				Check:  resource.TestCheckResourceAttr("data.brockhoff_context.test", "tags.bc-deletiondate", "2030-01-01"),
			},
		},
	})
}

func TestAccContext_ParentChild(t *testing.T) {
	const config = `
data "brockhoff_context" "parent" {
  namespace                = "plat"
  name                     = "parent"
  cost_center              = "eng"
  data_sovereignty         = "CA-QC"
  source_repo_tags_enabled = false
  additional_tags          = { team = "platform", owner = "ops" }
}

data "brockhoff_context" "child" {
  parent_context  = data.brockhoff_context.parent.context_output
  name            = "api"
  cost_center     = "api-cc"
  additional_tags = { team = "api", tier = "1" }
}

data "brockhoff_context" "unnamed" {
  parent_context = data.brockhoff_context.parent.context_output
}
`
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Parent: environment stays as supplied, environment name falls back to dev
					resource.TestCheckResourceAttr("data.brockhoff_context.parent", "context_output.environment", ""),
					resource.TestCheckResourceAttr("data.brockhoff_context.parent", "context_output.environment_name", "dev"),
					resource.TestCheckResourceAttr("data.brockhoff_context.parent", "context_output.deletion_date", "never"),

					// Child: inherits namespace, overrides cost center, merges additional tags
					resource.TestCheckResourceAttr("data.brockhoff_context.child", "name_prefix", "plat-api"),
					resource.TestCheckResourceAttr("data.brockhoff_context.child", "tags.bc-costcenter", "api-cc"),
					resource.TestCheckResourceAttr("data.brockhoff_context.child", "tags.bc-team", "api"),
					resource.TestCheckResourceAttr("data.brockhoff_context.child", "tags.bc-owner", "ops"),
					resource.TestCheckResourceAttr("data.brockhoff_context.child", "tags.bc-tier", "1"),
					resource.TestCheckResourceAttr("data.brockhoff_context.child", "tags.bc-environment", "dev"),
					resource.TestCheckNoResourceAttr("data.brockhoff_context.child", "tags.bc-sourcerepo"),

					// Name is never inherited
					resource.TestCheckResourceAttr("data.brockhoff_context.unnamed", "name_prefix", "plat"),
					resource.TestCheckResourceAttr("data.brockhoff_context.unnamed", "tags.bc-costcenter", "eng"),
				),
			},
			planStable(config),
		},
	})
}

func TestAccContext_NamePrefixPartial(t *testing.T) {
	const config = `
data "brockhoff_context" "test" {
  namespace                = "ns"
  name                     = "app"
  source_repo_tags_enabled = false
}
`
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  resource.TestCheckResourceAttr("data.brockhoff_context.test", "name_prefix", "ns-app"),
			},
		},
	})
}

package provider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccContext_CloudRules(t *testing.T) {
	long := strings.Repeat("x", 300)

	tests := []struct {
		cloud      string
		naValue    string
		codeOwners string
		note       string
		maxLen     int
		envTag     string
	}{
		{cloud: "aws", naValue: "N/A", codeOwners: "a@example.com b@example.com", note: "a_b__c_d/e_f:g", maxLen: 256, envTag: "Production"},
		{cloud: "az", naValue: "NotApplicable", codeOwners: "a@example.com;b@example.com", note: "abcdefg", maxLen: 256, envTag: "Production"},
		{cloud: "gcp", naValue: "not_applicable", codeOwners: "a-example-com_b-example-com", note: "a-b--c-d-e-f-g", maxLen: 63, envTag: "production"},
		{cloud: "dc", naValue: "N/A", codeOwners: "a@example.com;b@example.com", note: "a_b__c_d/e#f:g", maxLen: 63, envTag: "Production"},
	}

	for _, tt := range tests {
		t.Run(tt.cloud, func(t *testing.T) {
			config := fmt.Sprintf(`
provider "brockhoff" {
  cloud_provider = %q
}

data "brockhoff_context" "test" {
  name                     = "app"
  environment_name         = "Production"
  code_owners              = ["a@example.com", "b@example.com"]
  source_repo_tags_enabled = false
  additional_tags = {
    note = "a<b>&c?d/e#f:g"
    long = %q
  }
}
`, tt.cloud, long)

			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: config,
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("data.brockhoff_context.test", "tags.bc-costcenter", tt.naValue),
							resource.TestCheckResourceAttr("data.brockhoff_context.test", "tags.bc-codeowners", tt.codeOwners),
							resource.TestCheckResourceAttr("data.brockhoff_context.test", "tags.bc-note", tt.note),
							resource.TestCheckResourceAttr("data.brockhoff_context.test", "tags.bc-environment", tt.envTag),
							resource.TestCheckResourceAttr("data.brockhoff_context.test", "tags.bc-long", strings.Repeat("x", tt.maxLen)),
						),
					},
					planStable(config),
				},
			})
		})
	}
}

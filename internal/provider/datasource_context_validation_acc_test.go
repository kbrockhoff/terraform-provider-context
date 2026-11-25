package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// contextWith returns a minimal data source configuration with one extra attribute
func contextWith(attribute string) string {
	return fmt.Sprintf(`
data "brockhoff_context" "test" {
  name = "app"
  %s
}
`, attribute)
}

func TestAccContext_ValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		wantErr string
	}{
		{
			name:    "namespace too long",
			config:  contextWith(`namespace = "toolongns1"`),
			wantErr: `Invalid namespace`,
		},
		{
			name:    "environment uppercase",
			config:  contextWith(`environment = "Prod"`),
			wantErr: `Invalid environment`,
		},
		{
			name:    "unknown environment type",
			config:  contextWith(`environment_type = "Staging"`),
			wantErr: `Invalid environment_type`,
		},
		{
			name:    "unknown availability",
			config:  contextWith(`availability = "high"`),
			wantErr: `Invalid availability`,
		},
		{
			name:    "unknown sensitivity",
			config:  contextWith(`sensitivity = "secret"`),
			wantErr: `Invalid sensitivity`,
		},
		{
			name:    "invalid deletion date",
			config:  contextWith(`deletion_date = "2025-13-01"`),
			wantErr: `Invalid deletion_date`,
		},
		{
			name:    "invalid owner email",
			config:  contextWith(`product_owners = ["not-an-email"]`),
			wantErr: `Invalid product_owners`,
		},
		{
			name:    "explicit empty data sovereignty",
			config:  contextWith(`data_sovereignty = ""`),
			wantErr: `Invalid data_sovereignty`,
		},
		{
			name: "explicit empty cloud provider",
			config: `
provider "brockhoff" {
  cloud_provider = ""
}
data "brockhoff_context" "test" { name = "app" }`,
			wantErr: `Invalid cloud provider`,
		},
		{
			name: "unknown cloud provider",
			config: `
provider "brockhoff" {
  cloud_provider = "azure"
}
data "brockhoff_context" "test" { name = "app" }`,
			wantErr: `Invalid cloud provider`,
		},
		{
			name: "gcp rejects uppercase additional tag key",
			config: `
provider "brockhoff" {
  cloud_provider = "gcp"
}
data "brockhoff_context" "test" {
  name                     = "app"
  source_repo_tags_enabled = false
  additional_tags          = { Team = "x" }
}`,
			wantErr: `bc-Team`,
		},
		{
			name: "tag prefix invalid for cloud",
			config: `
provider "brockhoff" {
  cloud_provider = "dc"
  tag_prefix     = "bc:"
}
data "brockhoff_context" "test" { name = "app" }`,
			wantErr: `Invalid tag prefix`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config:      tt.config,
						ExpectError: regexp.MustCompile(tt.wantErr),
					},
				},
			})
		})
	}
}

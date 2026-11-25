package provider

import (
	"context"
	"testing"

	fwprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories instantiates the provider for acceptance tests
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"brockhoff": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccPreCheck validates prerequisites for acceptance tests.
// The provider needs no credentials, so there is nothing to check yet.
func testAccPreCheck(t *testing.T) {
	t.Helper()
}

func TestProviderSchema(t *testing.T) {
	ctx := context.Background()
	resp := &fwprovider.SchemaResponse{}
	New("test")().Schema(ctx, fwprovider.SchemaRequest{}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() diagnostics: %v", resp.Diagnostics)
	}
	if diags := resp.Schema.ValidateImplementation(ctx); diags.HasError() {
		t.Fatalf("schema implementation invalid: %v", diags)
	}
}

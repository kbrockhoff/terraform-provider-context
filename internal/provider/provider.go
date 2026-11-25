package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/kbrockhoff/terraform-provider-context/internal/core"
	ctxdatasource "github.com/kbrockhoff/terraform-provider-context/internal/datasource"
)

// Ensure ContextProvider satisfies various provider interfaces.
var _ provider.Provider = &ContextProvider{}

// ContextProvider defines the provider implementation.
type ContextProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// ContextProviderModel describes the provider data model.
type ContextProviderModel struct {
	CloudProvider types.String `tfsdk:"cloud_provider"`
	TagPrefix     types.String `tfsdk:"tag_prefix"`
}

func (p *ContextProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "brockhoff"
	resp.Version = p.version
}

func (p *ContextProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The Context provider generates standardized naming conventions and cloud-provider-specific tags for infrastructure resources.",
		Attributes: map[string]schema.Attribute{
			"cloud_provider": schema.StringAttribute{
				Description: "Cloud provider identifier: dc, aws, az, gcp, oci, ibm, do, vul, ali, cv",
				Optional:    true,
			},
			"tag_prefix": schema.StringAttribute{
				Description: "Prefix for all generated tag keys (default \"bc-\"). The prefixed keys must be valid for the target cloud.",
				Optional:    true,
			},
		},
	}
}

func (p *ContextProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Debug(ctx, "Configuring Context provider")

	var data ContextProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Set defaults
	cloudProvider := "dc"
	if !data.CloudProvider.IsNull() {
		cloudProvider = data.CloudProvider.ValueString()
		// An explicit empty value is invalid; only an omitted attribute gets the dc default
		if cloudProvider == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("cloud_provider"),
				"Invalid cloud provider",
				"cloud provider must not be empty; omit it for the default (dc) or use one of: dc, aws, az, gcp, oci, ibm, do, vul, ali, cv",
			)
			return
		}
	}

	tagPrefix := "bc-"
	if !data.TagPrefix.IsNull() {
		tagPrefix = data.TagPrefix.ValueString()
	}

	// Validate cloud provider
	if err := core.ValidateCloudProvider(cloudProvider); err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("cloud_provider"), "Invalid cloud provider", err.Error())
		return
	}

	// Validate the tag prefix produces keys the target cloud accepts
	if !core.GetCloudProvider(cloudProvider).ValidateTagKey(tagPrefix + "environment") {
		resp.Diagnostics.AddAttributeError(
			path.Root("tag_prefix"),
			"Invalid tag prefix",
			fmt.Sprintf("Tag prefix %q produces tag keys that are not valid for cloud provider %q", tagPrefix, cloudProvider),
		)
		return
	}

	// Create provider configuration
	providerConfig := &ctxdatasource.ProviderConfig{
		CloudProvider: cloudProvider,
		TagPrefix:     tagPrefix,
	}

	tflog.Debug(ctx, "Context provider configured", map[string]interface{}{
		"cloud_provider": cloudProvider,
		"tag_prefix":     tagPrefix,
	})

	// Make provider config available to data sources
	resp.DataSourceData = providerConfig
	resp.ResourceData = providerConfig
}

func (p *ContextProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}

func (p *ContextProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		ctxdatasource.NewContextDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ContextProvider{
			version: version,
		}
	}
}

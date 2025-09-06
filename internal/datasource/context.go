package datasource

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/kbrockhoff/terraform-provider-context/internal/core"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ContextDataSource{}
var _ datasource.DataSourceWithConfigure = &ContextDataSource{}

// ProviderConfig holds provider-level configuration
type ProviderConfig struct {
	CloudProvider string
	TagPrefix     string
}

func NewContextDataSource() datasource.DataSource {
	return &ContextDataSource{}
}

// ContextDataSource defines the data source implementation.
type ContextDataSource struct {
	providerConfig *ProviderConfig
}

// ContextDataSourceModel describes the data source data model.
type ContextDataSourceModel struct {
	// Naming Configuration
	Namespace       types.String `tfsdk:"namespace"`
	Name            types.String `tfsdk:"name"`
	Environment     types.String `tfsdk:"environment"`
	EnvironmentName types.String `tfsdk:"environment_name"`
	EnvironmentType types.String `tfsdk:"environment_type"`

	// Resource Management
	Enabled      types.Bool   `tfsdk:"enabled"`
	Availability types.String `tfsdk:"availability"`
	ManagedBy    types.String `tfsdk:"managedby"`
	DeletionDate types.String `tfsdk:"deletion_date"`

	// Project Management Integration
	PMPlatform    types.String `tfsdk:"pm_platform"`
	PMProjectCode types.String `tfsdk:"pm_project_code"`

	// ITSM Integration
	ITSMPlatform    types.String `tfsdk:"itsm_platform"`
	ITSMSystemID    types.String `tfsdk:"itsm_system_id"`
	ITSMComponentID types.String `tfsdk:"itsm_component_id"`
	ITSMInstanceID  types.String `tfsdk:"itsm_instance_id"`

	// Ownership and Billing
	CostCenter    types.String `tfsdk:"cost_center"`
	ProductOwners types.List   `tfsdk:"product_owners"`
	CodeOwners    types.List   `tfsdk:"code_owners"`
	DataOwners    types.List   `tfsdk:"data_owners"`

	// Data Classification
	Sensitivity    types.String `tfsdk:"sensitivity"`
	DataRegs       types.List   `tfsdk:"data_regs"`
	SecurityReview types.String `tfsdk:"security_review"`
	PrivacyReview  types.String `tfsdk:"privacy_review"`

	// Feature Toggles
	SourceRepoTagsEnabled types.Bool `tfsdk:"source_repo_tags_enabled"`
	SystemPrefixesEnabled types.Bool `tfsdk:"system_prefixes_enabled"`
	NotApplicableEnabled  types.Bool `tfsdk:"not_applicable_enabled"`
	OwnerTagsEnabled      types.Bool `tfsdk:"owner_tags_enabled"`

	// Additional Tags
	AdditionalTags     types.Map `tfsdk:"additional_tags"`
	AdditionalDataTags types.Map `tfsdk:"additional_data_tags"`

	// Computed Outputs
	ID                             types.String `tfsdk:"id"`
	NamePrefix                     types.String `tfsdk:"name_prefix"`
	Tags                           types.Map    `tfsdk:"tags"`
	DataTags                       types.Map    `tfsdk:"data_tags"`
	TagsAsListOfMaps               types.List   `tfsdk:"tags_as_list_of_maps"`
	TagsAsKVPList                  types.List   `tfsdk:"tags_as_kvp_list"`
	TagsAsCommaSeparatedString     types.String `tfsdk:"tags_as_comma_separated_string"`
	DataTagsAsListOfMaps           types.List   `tfsdk:"data_tags_as_list_of_maps"`
	DataTagsAsKVPList              types.List   `tfsdk:"data_tags_as_kvp_list"`
	DataTagsAsCommaSeparatedString types.String `tfsdk:"data_tags_as_comma_separated_string"`
}

func (d *ContextDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_context"
}

func (d *ContextDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Generates standardized naming conventions and cloud-provider-specific tags for infrastructure resources.",

		Attributes: map[string]schema.Attribute{
			// Naming Configuration
			"namespace": schema.StringAttribute{
				Description: "Organization or business unit identifier (1-8 chars, lowercase alphanumeric with hyphens)",
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Description: "Unique resource name (combined name_prefix must be 2-24 chars)",
				Optional:    true,
			},
			"environment": schema.StringAttribute{
				Description: "Environment abbreviation (1-8 chars, lowercase alphanumeric with hyphens)",
				Optional:    true,
			},
			"environment_name": schema.StringAttribute{
				Description: "Full environment name",
				Optional:    true,
			},
			"environment_type": schema.StringAttribute{
				Description: "One of: None, Ephemeral, Development, Testing, UAT, Production, MissionCritical",
				Optional:    true,
			},

			// Resource Management
			"enabled": schema.BoolAttribute{
				Description: "Enable/disable resource creation",
				Optional:    true,
			},
			"availability": schema.StringAttribute{
				Description: "Availability requirement from predefined list",
				Optional:    true,
			},
			"managedby": schema.StringAttribute{
				Description: "Management platform identifier",
				Optional:    true,
			},
			"deletion_date": schema.StringAttribute{
				Description: "Resource deletion date (YYYY-MM-DD format)",
				Optional:    true,
			},

			// Project Management Integration
			"pm_platform": schema.StringAttribute{
				Description: "Project management platform (e.g., JIRA, SNOW)",
				Optional:    true,
			},
			"pm_project_code": schema.StringAttribute{
				Description: "Project code/prefix",
				Optional:    true,
			},

			// ITSM Integration
			"itsm_platform": schema.StringAttribute{
				Description: "IT Service Management platform",
				Optional:    true,
			},
			"itsm_system_id": schema.StringAttribute{
				Description: "ITSM system identifier",
				Optional:    true,
			},
			"itsm_component_id": schema.StringAttribute{
				Description: "ITSM component identifier",
				Optional:    true,
			},
			"itsm_instance_id": schema.StringAttribute{
				Description: "ITSM instance identifier",
				Optional:    true,
			},

			// Ownership and Billing
			"cost_center": schema.StringAttribute{
				Description: "Cost center for billing",
				Optional:    true,
			},
			"product_owners": schema.ListAttribute{
				Description: "Product owner email addresses",
				Optional:    true,
				ElementType: types.StringType,
			},
			"code_owners": schema.ListAttribute{
				Description: "Code owner email addresses",
				Optional:    true,
				ElementType: types.StringType,
			},
			"data_owners": schema.ListAttribute{
				Description: "Data owner email addresses",
				Optional:    true,
				ElementType: types.StringType,
			},

			// Data Classification
			"sensitivity": schema.StringAttribute{
				Description: "Data sensitivity level from predefined list",
				Optional:    true,
			},
			"data_regs": schema.ListAttribute{
				Description: "Data compliance regulations",
				Optional:    true,
				ElementType: types.StringType,
			},
			"security_review": schema.StringAttribute{
				Description: "Security review identifier/date",
				Optional:    true,
			},
			"privacy_review": schema.StringAttribute{
				Description: "Privacy review identifier/date",
				Optional:    true,
			},

			// Feature Toggles
			"source_repo_tags_enabled": schema.BoolAttribute{
				Description: "Include git repository tags",
				Optional:    true,
			},
			"system_prefixes_enabled": schema.BoolAttribute{
				Description: "Add platform prefixes to system IDs",
				Optional:    true,
			},
			"not_applicable_enabled": schema.BoolAttribute{
				Description: "Include N/A tags for null values",
				Optional:    true,
			},
			"owner_tags_enabled": schema.BoolAttribute{
				Description: "Include owner tags",
				Optional:    true,
			},

			// Additional Tags
			"additional_tags": schema.MapAttribute{
				Description: "Custom tags to merge",
				Optional:    true,
				ElementType: types.StringType,
			},
			"additional_data_tags": schema.MapAttribute{
				Description: "Custom data-specific tags to merge",
				Optional:    true,
				ElementType: types.StringType,
			},

			// Computed Outputs
			"id": schema.StringAttribute{
				Description: "Unique identifier for this data source instance",
				Computed:    true,
			},
			"name_prefix": schema.StringAttribute{
				Description: "Computed name prefix following Brockhoff standards",
				Computed:    true,
			},
			"tags": schema.MapAttribute{
				Description: "Normalized tag map",
				Computed:    true,
				ElementType: types.StringType,
			},
			"data_tags": schema.MapAttribute{
				Description: "Data-specific tags",
				Computed:    true,
				ElementType: types.StringType,
			},
			"tags_as_list_of_maps": schema.ListAttribute{
				Description: "Tags formatted for AWS resources",
				Computed:    true,
				ElementType: types.MapType{
					ElemType: types.StringType,
				},
			},
			"tags_as_kvp_list": schema.ListAttribute{
				Description: "Tags as key=value pairs",
				Computed:    true,
				ElementType: types.StringType,
			},
			"tags_as_comma_separated_string": schema.StringAttribute{
				Description: "Tags as comma-separated string",
				Computed:    true,
			},
			"data_tags_as_list_of_maps": schema.ListAttribute{
				Description: "Data tags formatted for AWS resources",
				Computed:    true,
				ElementType: types.MapType{
					ElemType: types.StringType,
				},
			},
			"data_tags_as_kvp_list": schema.ListAttribute{
				Description: "Data tags as key=value pairs",
				Computed:    true,
				ElementType: types.StringType,
			},
			"data_tags_as_comma_separated_string": schema.StringAttribute{
				Description: "Data tags as comma-separated string",
				Computed:    true,
			},
		},
	}
}

func (d *ContextDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider is not configured.
	if req.ProviderData == nil {
		return
	}

	providerConfig, ok := req.ProviderData.(*ProviderConfig)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *ProviderConfig, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.providerConfig = providerConfig
}

func (d *ContextDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ContextDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Convert model to core config
	config := &core.DataSourceConfig{
		Namespace:       data.Namespace.ValueString(),
		Name:            data.Name.ValueString(),
		Environment:     data.Environment.ValueString(),
		EnvironmentName: data.EnvironmentName.ValueString(),
		EnvironmentType: data.EnvironmentType.ValueString(),

		Enabled:      data.Enabled.ValueBoolPointer() != nil && *data.Enabled.ValueBoolPointer(),
		Availability: data.Availability.ValueString(),
		ManagedBy:    data.ManagedBy.ValueString(),
		DeletionDate: data.DeletionDate.ValueString(),

		PMPlatform:    data.PMPlatform.ValueString(),
		PMProjectCode: data.PMProjectCode.ValueString(),

		ITSMPlatform:    data.ITSMPlatform.ValueString(),
		ITSMSystemID:    data.ITSMSystemID.ValueString(),
		ITSMComponentID: data.ITSMComponentID.ValueString(),
		ITSMInstanceID:  data.ITSMInstanceID.ValueString(),

		CostCenter:     data.CostCenter.ValueString(),
		Sensitivity:    data.Sensitivity.ValueString(),
		SecurityReview: data.SecurityReview.ValueString(),
		PrivacyReview:  data.PrivacyReview.ValueString(),

		SourceRepoTagsEnabled: data.SourceRepoTagsEnabled.ValueBoolPointer() != nil && *data.SourceRepoTagsEnabled.ValueBoolPointer(),
		SystemPrefixesEnabled: data.SystemPrefixesEnabled.ValueBoolPointer() != nil && *data.SystemPrefixesEnabled.ValueBoolPointer(),
		NotApplicableEnabled:  data.NotApplicableEnabled.ValueBoolPointer() != nil && *data.NotApplicableEnabled.ValueBoolPointer(),
		OwnerTagsEnabled:      data.OwnerTagsEnabled.ValueBoolPointer() != nil && *data.OwnerTagsEnabled.ValueBoolPointer(),
	}

	// Set defaults
	if config.Availability == "" {
		config.Availability = "preemptable"
	}
	if config.ManagedBy == "" {
		config.ManagedBy = "terraform"
	}
	if config.Sensitivity == "" {
		config.Sensitivity = "confidential"
	}
	if !data.SourceRepoTagsEnabled.IsNull() {
		config.SourceRepoTagsEnabled = data.SourceRepoTagsEnabled.ValueBool()
	} else {
		config.SourceRepoTagsEnabled = true
	}
	if !data.SystemPrefixesEnabled.IsNull() {
		config.SystemPrefixesEnabled = data.SystemPrefixesEnabled.ValueBool()
	} else {
		config.SystemPrefixesEnabled = true
	}
	if !data.NotApplicableEnabled.IsNull() {
		config.NotApplicableEnabled = data.NotApplicableEnabled.ValueBool()
	} else {
		config.NotApplicableEnabled = true
	}
	if !data.OwnerTagsEnabled.IsNull() {
		config.OwnerTagsEnabled = data.OwnerTagsEnabled.ValueBool()
	} else {
		config.OwnerTagsEnabled = true
	}

	// Extract list values
	if !data.ProductOwners.IsNull() {
		productOwners := []string{}
		data.ProductOwners.ElementsAs(ctx, &productOwners, false)
		config.ProductOwners = productOwners
	}
	if !data.CodeOwners.IsNull() {
		codeOwners := []string{}
		data.CodeOwners.ElementsAs(ctx, &codeOwners, false)
		config.CodeOwners = codeOwners
	}
	if !data.DataOwners.IsNull() {
		dataOwners := []string{}
		data.DataOwners.ElementsAs(ctx, &dataOwners, false)
		config.DataOwners = dataOwners
	}
	if !data.DataRegs.IsNull() {
		dataRegs := []string{}
		data.DataRegs.ElementsAs(ctx, &dataRegs, false)
		config.DataRegs = dataRegs
	}

	// Extract map values
	if !data.AdditionalTags.IsNull() {
		additionalTags := map[string]string{}
		data.AdditionalTags.ElementsAs(ctx, &additionalTags, false)
		config.AdditionalTags = additionalTags
	}
	if !data.AdditionalDataTags.IsNull() {
		additionalDataTags := map[string]string{}
		data.AdditionalDataTags.ElementsAs(ctx, &additionalDataTags, false)
		config.AdditionalDataTags = additionalDataTags
	}

	// Validation
	if err := core.ValidateNamespace(config.Namespace); err != nil {
		resp.Diagnostics.AddError("Invalid namespace", err.Error())
		return
	}
	if err := core.ValidateEnvironment(config.Environment); err != nil {
		resp.Diagnostics.AddError("Invalid environment", err.Error())
		return
	}
	if err := core.ValidateEnvironmentType(config.EnvironmentType); err != nil {
		resp.Diagnostics.AddError("Invalid environment_type", err.Error())
		return
	}
	if err := core.ValidateAvailability(config.Availability); err != nil {
		resp.Diagnostics.AddError("Invalid availability", err.Error())
		return
	}
	if err := core.ValidateSensitivity(config.Sensitivity); err != nil {
		resp.Diagnostics.AddError("Invalid sensitivity", err.Error())
		return
	}
	if err := core.ValidateDeletionDate(config.DeletionDate); err != nil {
		resp.Diagnostics.AddError("Invalid deletion_date", err.Error())
		return
	}
	if err := core.ValidateEmails(config.ProductOwners); err != nil {
		resp.Diagnostics.AddError("Invalid product_owners", err.Error())
		return
	}
	if err := core.ValidateEmails(config.CodeOwners); err != nil {
		resp.Diagnostics.AddError("Invalid code_owners", err.Error())
		return
	}
	if err := core.ValidateEmails(config.DataOwners); err != nil {
		resp.Diagnostics.AddError("Invalid data_owners", err.Error())
		return
	}

	// Process ephemeral environment
	core.ProcessEphemeralEnvironment(config)

	// Generate name prefix
	nameGen := &core.NameGenerator{
		Namespace:   config.Namespace,
		Name:        config.Name,
		Environment: config.Environment,
	}
	namePrefix, err := nameGen.Generate()
	if err != nil {
		resp.Diagnostics.AddError("Failed to generate name prefix", err.Error())
		return
	}

	// Get cloud provider
	cloudProvider := d.providerConfig.CloudProvider
	if cloudProvider == "" {
		cloudProvider = "dc"
	}
	cp := core.GetCloudProvider(cloudProvider)

	// Generate tags
	tagProcessor := &core.TagProcessor{
		CloudProvider: cp,
		Config:        config,
		TagPrefix:     d.providerConfig.TagPrefix,
	}

	tags, err := tagProcessor.Process()
	if err != nil {
		resp.Diagnostics.AddError("Failed to generate tags", err.Error())
		return
	}

	dataTags, err := tagProcessor.ProcessDataTags()
	if err != nil {
		resp.Diagnostics.AddError("Failed to generate data tags", err.Error())
		return
	}

	// Convert outputs
	tagsListOfMaps := core.ConvertTagsToListOfMaps(tags)
	tagsKVPList := core.ConvertTagsToKVPList(tags)
	tagsCommaSeparated := core.ConvertTagsToCommaSeparated(tags)

	dataTagsListOfMaps := core.ConvertTagsToListOfMaps(dataTags)
	dataTagsKVPList := core.ConvertTagsToKVPList(dataTags)
	dataTagsCommaSeparated := core.ConvertTagsToCommaSeparated(dataTags)

	// Set computed values
	data.ID = types.StringValue(namePrefix)
	data.NamePrefix = types.StringValue(namePrefix)

	// Convert maps to types.Map
	tagsMap, diags := types.MapValueFrom(ctx, types.StringType, tags)
	resp.Diagnostics.Append(diags...)
	data.Tags = tagsMap

	dataTagsMap, diags := types.MapValueFrom(ctx, types.StringType, dataTags)
	resp.Diagnostics.Append(diags...)
	data.DataTags = dataTagsMap

	// Convert list of maps
	tagsListValue, diags := types.ListValueFrom(ctx, types.MapType{ElemType: types.StringType}, tagsListOfMaps)
	resp.Diagnostics.Append(diags...)
	data.TagsAsListOfMaps = tagsListValue

	dataTagsListValue, diags := types.ListValueFrom(ctx, types.MapType{ElemType: types.StringType}, dataTagsListOfMaps)
	resp.Diagnostics.Append(diags...)
	data.DataTagsAsListOfMaps = dataTagsListValue

	// Convert KVP lists
	tagsKVPListValue, diags := types.ListValueFrom(ctx, types.StringType, tagsKVPList)
	resp.Diagnostics.Append(diags...)
	data.TagsAsKVPList = tagsKVPListValue

	dataTagsKVPListValue, diags := types.ListValueFrom(ctx, types.StringType, dataTagsKVPList)
	resp.Diagnostics.Append(diags...)
	data.DataTagsAsKVPList = dataTagsKVPListValue

	// Set comma-separated strings
	data.TagsAsCommaSeparatedString = types.StringValue(tagsCommaSeparated)
	data.DataTagsAsCommaSeparatedString = types.StringValue(dataTagsCommaSeparated)

	tflog.Debug(ctx, "Context data source read", map[string]interface{}{
		"name_prefix":     namePrefix,
		"tags_count":      len(tags),
		"data_tags_count": len(dataTags),
	})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

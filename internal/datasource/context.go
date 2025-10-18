package datasource

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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

// ContextInputModel describes the context input data model for parent context inheritance.
type ContextInputModel struct {
	// Naming Configuration
	Namespace       types.String `tfsdk:"namespace"`
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
}

// ContextDataSourceModel describes the data source data model.
type ContextDataSourceModel struct {
	// Parent Context Input (optional)
	ParentContext types.Object `tfsdk:"parent_context"`

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
	ContextOutput                  types.Object `tfsdk:"context_output"`
}

func (d *ContextDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_context"
}

// getContextAttributes returns the schema attributes for the context object
func getContextAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"namespace": schema.StringAttribute{
			Description: "Organization or business unit identifier (1-8 chars, lowercase alphanumeric with hyphens)",
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
		"pm_platform": schema.StringAttribute{
			Description: "Project management platform (e.g., JIRA, SNOW)",
			Optional:    true,
		},
		"pm_project_code": schema.StringAttribute{
			Description: "Project code/prefix",
			Optional:    true,
		},
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
	}
}

func (d *ContextDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Generates standardized naming conventions and cloud-provider-specific tags for infrastructure resources. Supports parent/child context inheritance.",

		Attributes: map[string]schema.Attribute{
			// Parent Context Input (optional - for parent context inheritance)
			"parent_context": schema.SingleNestedAttribute{
				Description: "Parent context values to inherit. Child context can override individual fields.",
				Optional:    true,
				Attributes:  getContextAttributes(),
			},

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
			"context_output": schema.SingleNestedAttribute{
				Description: "Resolved context values that can be used as input for child contexts",
				Computed:    true,
				Attributes:  getContextAttributes(),
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

// mergeStringValue returns the individual value if set, otherwise the context value
func mergeStringValue(individualValue, contextValue types.String) string {
	if !individualValue.IsNull() {
		return individualValue.ValueString()
	}
	if !contextValue.IsNull() {
		return contextValue.ValueString()
	}
	return ""
}

// mergeBoolValue returns the individual value if set, otherwise the context value
func mergeBoolValue(individualValue, contextValue types.Bool, defaultValue bool) bool {
	if !individualValue.IsNull() {
		return individualValue.ValueBool()
	}
	if !contextValue.IsNull() {
		return contextValue.ValueBool()
	}
	return defaultValue
}

// mergeListValue returns the individual value if set, otherwise the context value
func mergeListValue(ctx context.Context, individualValue, contextValue types.List) []string {
	if !individualValue.IsNull() {
		values := []string{}
		individualValue.ElementsAs(ctx, &values, false)
		return values
	}
	if !contextValue.IsNull() {
		values := []string{}
		contextValue.ElementsAs(ctx, &values, false)
		return values
	}
	return nil
}

// mergeMapValue returns the individual value if set, otherwise the context value
func mergeMapValue(ctx context.Context, individualValue, contextValue types.Map) map[string]string {
	merged := make(map[string]string)

	if !contextValue.IsNull() {
		parentValues := map[string]string{}
		contextValue.ElementsAs(ctx, &parentValues, false)
		for k, v := range parentValues {
			merged[k] = v
		}
	}

	if !individualValue.IsNull() {
		childValues := map[string]string{}
		individualValue.ElementsAs(ctx, &childValues, false)
		for k, v := range childValues {
			merged[k] = v
		}
	}

	return merged
}

func (d *ContextDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ContextDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Extract parent context if provided
	var parentCtx ContextInputModel
	if !data.ParentContext.IsNull() {
		diag := data.ParentContext.As(ctx, &parentCtx, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diag...)
		if resp.Diagnostics.HasError() {
			return
		}
		tflog.Debug(ctx, "Parent context provided, will merge with individual inputs")
	}

	// Convert model to core config, merging parent context with individual inputs
	// Merge order: defaults -> parent context -> individual inputs
	config := &core.DataSourceConfig{
		// Name is always from individual input (not inherited)
		Name: data.Name.ValueString(),

		// These fields can be inherited from parent context
		Namespace:       mergeStringValue(data.Namespace, parentCtx.Namespace),
		Environment:     mergeStringValue(data.Environment, parentCtx.Environment),
		EnvironmentName: mergeStringValue(data.EnvironmentName, parentCtx.EnvironmentName),
		EnvironmentType: mergeStringValue(data.EnvironmentType, parentCtx.EnvironmentType),

		Availability: mergeStringValue(data.Availability, parentCtx.Availability),
		ManagedBy:    mergeStringValue(data.ManagedBy, parentCtx.ManagedBy),
		DeletionDate: mergeStringValue(data.DeletionDate, parentCtx.DeletionDate),

		PMPlatform:    mergeStringValue(data.PMPlatform, parentCtx.PMPlatform),
		PMProjectCode: mergeStringValue(data.PMProjectCode, parentCtx.PMProjectCode),

		ITSMPlatform:    mergeStringValue(data.ITSMPlatform, parentCtx.ITSMPlatform),
		ITSMSystemID:    mergeStringValue(data.ITSMSystemID, parentCtx.ITSMSystemID),
		ITSMComponentID: mergeStringValue(data.ITSMComponentID, parentCtx.ITSMComponentID),
		ITSMInstanceID:  mergeStringValue(data.ITSMInstanceID, parentCtx.ITSMInstanceID),

		CostCenter:     mergeStringValue(data.CostCenter, parentCtx.CostCenter),
		Sensitivity:    mergeStringValue(data.Sensitivity, parentCtx.Sensitivity),
		SecurityReview: mergeStringValue(data.SecurityReview, parentCtx.SecurityReview),
		PrivacyReview:  mergeStringValue(data.PrivacyReview, parentCtx.PrivacyReview),

		ProductOwners: mergeListValue(ctx, data.ProductOwners, parentCtx.ProductOwners),
		CodeOwners:    mergeListValue(ctx, data.CodeOwners, parentCtx.CodeOwners),
		DataOwners:    mergeListValue(ctx, data.DataOwners, parentCtx.DataOwners),
		DataRegs:      mergeListValue(ctx, data.DataRegs, parentCtx.DataRegs),

		AdditionalTags:     mergeMapValue(ctx, data.AdditionalTags, parentCtx.AdditionalTags),
		AdditionalDataTags: mergeMapValue(ctx, data.AdditionalDataTags, parentCtx.AdditionalDataTags),

		SourceRepoTagsEnabled: mergeBoolValue(data.SourceRepoTagsEnabled, parentCtx.SourceRepoTagsEnabled, true),
		SystemPrefixesEnabled: mergeBoolValue(data.SystemPrefixesEnabled, parentCtx.SystemPrefixesEnabled, true),
		NotApplicableEnabled:  mergeBoolValue(data.NotApplicableEnabled, parentCtx.NotApplicableEnabled, true),
		OwnerTagsEnabled:      mergeBoolValue(data.OwnerTagsEnabled, parentCtx.OwnerTagsEnabled, true),
	}

	// Handle Enabled field specially - default to true
	config.Enabled = mergeBoolValue(data.Enabled, parentCtx.Enabled, true)

	// Apply defaults for fields that are still empty after merging
	if config.Availability == "" {
		config.Availability = "preemptable"
	}
	if config.ManagedBy == "" {
		config.ManagedBy = "terraform"
	}
	if config.Sensitivity == "" {
		config.Sensitivity = "confidential"
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

	// Populate context_output with resolved values for use in child contexts
	contextOutput := ContextInputModel{
		Namespace:       types.StringValue(config.Namespace),
		Environment:     types.StringValue(config.Environment),
		EnvironmentName: types.StringValue(config.EnvironmentName),
		EnvironmentType: types.StringValue(config.EnvironmentType),

		Enabled:      types.BoolValue(config.Enabled),
		Availability: types.StringValue(config.Availability),
		ManagedBy:    types.StringValue(config.ManagedBy),
		DeletionDate: types.StringValue(config.DeletionDate),

		PMPlatform:    types.StringValue(config.PMPlatform),
		PMProjectCode: types.StringValue(config.PMProjectCode),

		ITSMPlatform:    types.StringValue(config.ITSMPlatform),
		ITSMSystemID:    types.StringValue(config.ITSMSystemID),
		ITSMComponentID: types.StringValue(config.ITSMComponentID),
		ITSMInstanceID:  types.StringValue(config.ITSMInstanceID),

		CostCenter:     types.StringValue(config.CostCenter),
		Sensitivity:    types.StringValue(config.Sensitivity),
		SecurityReview: types.StringValue(config.SecurityReview),
		PrivacyReview:  types.StringValue(config.PrivacyReview),

		SourceRepoTagsEnabled: types.BoolValue(config.SourceRepoTagsEnabled),
		SystemPrefixesEnabled: types.BoolValue(config.SystemPrefixesEnabled),
		NotApplicableEnabled:  types.BoolValue(config.NotApplicableEnabled),
		OwnerTagsEnabled:      types.BoolValue(config.OwnerTagsEnabled),
	}

	// Convert list fields - always initialize with proper type even if empty
	listVal, diags := types.ListValueFrom(ctx, types.StringType, config.ProductOwners)
	resp.Diagnostics.Append(diags...)
	contextOutput.ProductOwners = listVal

	listVal, diags = types.ListValueFrom(ctx, types.StringType, config.CodeOwners)
	resp.Diagnostics.Append(diags...)
	contextOutput.CodeOwners = listVal

	listVal, diags = types.ListValueFrom(ctx, types.StringType, config.DataOwners)
	resp.Diagnostics.Append(diags...)
	contextOutput.DataOwners = listVal

	listVal, diags = types.ListValueFrom(ctx, types.StringType, config.DataRegs)
	resp.Diagnostics.Append(diags...)
	contextOutput.DataRegs = listVal

	// Convert map fields - always initialize with proper type even if empty
	mapVal, diags := types.MapValueFrom(ctx, types.StringType, config.AdditionalTags)
	resp.Diagnostics.Append(diags...)
	contextOutput.AdditionalTags = mapVal

	mapVal, diags = types.MapValueFrom(ctx, types.StringType, config.AdditionalDataTags)
	resp.Diagnostics.Append(diags...)
	contextOutput.AdditionalDataTags = mapVal

	// Set context_output
	contextOutputObj, diagsCtx := types.ObjectValueFrom(ctx, map[string]attr.Type{
		"namespace":                types.StringType,
		"environment":              types.StringType,
		"environment_name":         types.StringType,
		"environment_type":         types.StringType,
		"enabled":                  types.BoolType,
		"availability":             types.StringType,
		"managedby":                types.StringType,
		"deletion_date":            types.StringType,
		"pm_platform":              types.StringType,
		"pm_project_code":          types.StringType,
		"itsm_platform":            types.StringType,
		"itsm_system_id":           types.StringType,
		"itsm_component_id":        types.StringType,
		"itsm_instance_id":         types.StringType,
		"cost_center":              types.StringType,
		"product_owners":           types.ListType{ElemType: types.StringType},
		"code_owners":              types.ListType{ElemType: types.StringType},
		"data_owners":              types.ListType{ElemType: types.StringType},
		"sensitivity":              types.StringType,
		"data_regs":                types.ListType{ElemType: types.StringType},
		"security_review":          types.StringType,
		"privacy_review":           types.StringType,
		"source_repo_tags_enabled": types.BoolType,
		"system_prefixes_enabled":  types.BoolType,
		"not_applicable_enabled":   types.BoolType,
		"owner_tags_enabled":       types.BoolType,
		"additional_tags":          types.MapType{ElemType: types.StringType},
		"additional_data_tags":     types.MapType{ElemType: types.StringType},
	}, contextOutput)
	resp.Diagnostics.Append(diagsCtx...)
	data.ContextOutput = contextOutputObj

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

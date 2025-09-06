package core

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// TagProcessor handles tag generation and processing
type TagProcessor struct {
	CloudProvider CloudProvider
	Config        *DataSourceConfig
	TagPrefix     string
}

// DataSourceConfig contains all configuration fields from the data source
type DataSourceConfig struct {
	// Naming
	Namespace       string
	Name            string
	Environment     string
	EnvironmentName string
	EnvironmentType string

	// Resource Management
	Enabled      bool
	Availability string
	ManagedBy    string
	DeletionDate string

	// Integration
	PMPlatform      string
	PMProjectCode   string
	ITSMPlatform    string
	ITSMSystemID    string
	ITSMComponentID string
	ITSMInstanceID  string

	// Ownership
	CostCenter    string
	ProductOwners []string
	CodeOwners    []string
	DataOwners    []string

	// Data Classification
	Sensitivity    string
	DataRegs       []string
	SecurityReview string
	PrivacyReview  string

	// Feature Toggles
	SourceRepoTagsEnabled bool
	SystemPrefixesEnabled bool
	NotApplicableEnabled  bool
	OwnerTagsEnabled      bool

	// Additional Tags
	AdditionalTags     map[string]string
	AdditionalDataTags map[string]string
}

// Process generates the main tags map
func (tp *TagProcessor) Process() (map[string]string, error) {
	tags := make(map[string]string)
	delimiter := tp.CloudProvider.GetDelimiter()
	naValue := tp.CloudProvider.GetNAValue()

	// Environment and resource tags
	tp.addTag(tags, "Environment", tp.Config.Environment, naValue)
	tp.addTag(tags, "EnvironmentName", tp.Config.EnvironmentName, naValue)
	tp.addTag(tags, "EnvironmentType", tp.Config.EnvironmentType, naValue)
	tp.addTag(tags, "Availability", tp.Config.Availability, naValue)
	tp.addTag(tags, "ManagedBy", tp.Config.ManagedBy, naValue)
	tp.addTag(tags, "DeletionDate", tp.Config.DeletionDate, naValue)

	// Billing
	tp.addTag(tags, "CostCenter", tp.Config.CostCenter, naValue)

	// Project Management
	if tp.Config.SystemPrefixesEnabled && tp.Config.PMPlatform != "" && tp.Config.PMProjectCode != "" {
		tags["PMProject"] = fmt.Sprintf("%s%s%s", tp.Config.PMPlatform, delimiter, tp.Config.PMProjectCode)
	} else {
		tp.addTag(tags, "PMProject", tp.Config.PMProjectCode, naValue)
	}

	// ITSM
	if tp.Config.SystemPrefixesEnabled && tp.Config.ITSMPlatform != "" {
		if tp.Config.ITSMSystemID != "" {
			tags["ITSMSystem"] = fmt.Sprintf("%s%s%s", tp.Config.ITSMPlatform, delimiter, tp.Config.ITSMSystemID)
		}
		if tp.Config.ITSMComponentID != "" {
			tags["ITSMComponent"] = fmt.Sprintf("%s%s%s", tp.Config.ITSMPlatform, delimiter, tp.Config.ITSMComponentID)
		}
		if tp.Config.ITSMInstanceID != "" {
			tags["ITSMInstance"] = fmt.Sprintf("%s%s%s", tp.Config.ITSMPlatform, delimiter, tp.Config.ITSMInstanceID)
		}
	} else {
		tp.addTag(tags, "ITSMSystem", tp.Config.ITSMSystemID, naValue)
		tp.addTag(tags, "ITSMComponent", tp.Config.ITSMComponentID, naValue)
		tp.addTag(tags, "ITSMInstance", tp.Config.ITSMInstanceID, naValue)
	}

	// Ownership (if enabled)
	if tp.Config.OwnerTagsEnabled {
		if len(tp.Config.ProductOwners) > 0 {
			tags["ProductOwners"] = strings.Join(tp.Config.ProductOwners, delimiter)
		} else if tp.Config.NotApplicableEnabled {
			tags["ProductOwners"] = naValue
		}

		if len(tp.Config.CodeOwners) > 0 {
			tags["CodeOwners"] = strings.Join(tp.Config.CodeOwners, delimiter)
		} else if tp.Config.NotApplicableEnabled {
			tags["CodeOwners"] = naValue
		}
	}

	// Reviews
	tp.addTag(tags, "SecurityReview", tp.Config.SecurityReview, naValue)
	tp.addTag(tags, "PrivacyReview", tp.Config.PrivacyReview, naValue)

	// Git repository tags (if enabled)
	if tp.Config.SourceRepoTagsEnabled {
		gitInfo, err := GetGitInfo()
		if err == nil && gitInfo != nil {
			tp.addTag(tags, "SourceRepo", gitInfo.RepoURL, naValue)
			tp.addTag(tags, "SourceCommit", gitInfo.CommitHash, naValue)
		}
	}

	// Merge additional tags
	for k, v := range tp.Config.AdditionalTags {
		tags[k] = v
	}

	// Apply tag prefix and sanitization
	prefixedTags := make(map[string]string)
	for k, v := range tags {
		key := tp.TagPrefix + k
		value := tp.CloudProvider.SanitizeTagValue(v)

		// Truncate if necessary
		maxLen := tp.CloudProvider.GetMaxTagLength()
		if len(value) > maxLen {
			value = value[:maxLen]
		}

		prefixedTags[key] = value
	}

	return prefixedTags, nil
}

// ProcessDataTags generates data-specific tags
func (tp *TagProcessor) ProcessDataTags() (map[string]string, error) {
	tags := make(map[string]string)
	delimiter := tp.CloudProvider.GetDelimiter()
	naValue := tp.CloudProvider.GetNAValue()

	// Data classification
	tp.addTag(tags, "DataSensitivity", tp.Config.Sensitivity, naValue)

	if len(tp.Config.DataRegs) > 0 {
		tags["DataRegulations"] = strings.Join(tp.Config.DataRegs, delimiter)
	} else if tp.Config.NotApplicableEnabled {
		tags["DataRegulations"] = naValue
	}

	// Data ownership
	if tp.Config.OwnerTagsEnabled && len(tp.Config.DataOwners) > 0 {
		tags["DataOwners"] = strings.Join(tp.Config.DataOwners, delimiter)
	} else if tp.Config.NotApplicableEnabled {
		tags["DataOwners"] = naValue
	}

	// Merge additional data tags
	for k, v := range tp.Config.AdditionalDataTags {
		tags[k] = v
	}

	// Apply tag prefix and sanitization
	prefixedTags := make(map[string]string)
	for k, v := range tags {
		key := tp.TagPrefix + k
		value := tp.CloudProvider.SanitizeTagValue(v)

		// Truncate if necessary
		maxLen := tp.CloudProvider.GetMaxTagLength()
		if len(value) > maxLen {
			value = value[:maxLen]
		}

		prefixedTags[key] = value
	}

	return prefixedTags, nil
}

// addTag adds a tag if value is not empty or N/A is enabled
func (tp *TagProcessor) addTag(tags map[string]string, key, value, naValue string) {
	if value != "" {
		tags[key] = value
	} else if tp.Config.NotApplicableEnabled {
		tags[key] = naValue
	}
}

// ProcessEphemeralEnvironment handles ephemeral environment special logic
func ProcessEphemeralEnvironment(config *DataSourceConfig) {
	if config.EnvironmentType == "Ephemeral" && config.DeletionDate == "" {
		// Calculate deletion date as 90 days from now
		deletionDate := time.Now().Add(90 * 24 * time.Hour)
		config.DeletionDate = deletionDate.Format("2006-01-02")
	}
}

// ConvertTagsToListOfMaps converts tags map to list of maps for AWS
func ConvertTagsToListOfMaps(tags map[string]string) []map[string]string {
	result := make([]map[string]string, 0, len(tags))

	// Sort keys for consistent output
	keys := make([]string, 0, len(tags))
	for k := range tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		result = append(result, map[string]string{
			"key":   k,
			"value": tags[k],
		})
	}

	return result
}

// ConvertTagsToKVPList converts tags to key=value pairs
func ConvertTagsToKVPList(tags map[string]string) []string {
	result := make([]string, 0, len(tags))

	// Sort keys for consistent output
	keys := make([]string, 0, len(tags))
	for k := range tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		result = append(result, fmt.Sprintf("%s=%s", k, tags[k]))
	}

	return result
}

// ConvertTagsToCommaSeparated converts tags to comma-separated string
func ConvertTagsToCommaSeparated(tags map[string]string) string {
	kvpList := ConvertTagsToKVPList(tags)
	return strings.Join(kvpList, ",")
}

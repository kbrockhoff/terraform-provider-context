package context

import (
	"fmt"
	"maps"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	// DefaultEnvironmentName is the environment tag value when neither environment name
	// nor environment is supplied
	DefaultEnvironmentName = "dev"

	// ephemeralRetention is how long Ephemeral environments are kept before deletion
	ephemeralRetention = 90 * 24 * time.Hour
)

// now returns the current time; it is a variable so tests can fix the clock
var now = time.Now

// TagProcessor handles tag generation and processing
type TagProcessor struct {
	CloudProvider CloudProvider
	Config        *DataSourceConfig
	TagPrefix     string

	// GitInfoFunc supplies source repository information; nil means GetGitInfo
	GitInfoFunc func() (*GitInfo, error)

	// Warnings collects non-fatal problems found by the last Process call
	Warnings []string
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
	Sensitivity     string
	DataRegs        []string
	DataSovereignty string
	SecurityReview  string
	PrivacyReview   string

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
	tp.Warnings = nil
	tags := make(map[string]string)
	delimiter := tp.CloudProvider.GetDelimiter()
	naValue := tp.CloudProvider.GetNAValue()

	// Always-present environment and resource tags
	// Note: environment type is used for calculations but not included in output tags
	tags["environment"] = EffectiveEnvironmentName(tp.Config)
	tags["availability"] = tp.Config.Availability
	tags["managedby"] = tp.Config.ManagedBy
	tags["deletiondate"] = tp.Config.DeletionDate

	// Billing
	tp.addOptionalTag(tags, "costcenter", tp.Config.CostCenter, naValue)

	// Project management and ITSM identifiers are only emitted when supplied
	tp.addSystemTag(tags, "projectmgmtid", tp.Config.PMPlatform, tp.Config.PMProjectCode, delimiter)
	tp.addSystemTag(tags, "systemid", tp.Config.ITSMPlatform, tp.Config.ITSMSystemID, delimiter)
	tp.addSystemTag(tags, "componentid", tp.Config.ITSMPlatform, tp.Config.ITSMComponentID, delimiter)
	tp.addSystemTag(tags, "instanceid", tp.Config.ITSMPlatform, tp.Config.ITSMInstanceID, delimiter)

	// Ownership (if enabled)
	if tp.Config.OwnerTagsEnabled {
		tp.addOptionalTag(tags, "productowners", strings.Join(tp.Config.ProductOwners, delimiter), naValue)
		tp.addOptionalTag(tags, "codeowners", strings.Join(tp.Config.CodeOwners, delimiter), naValue)
	}

	// Reviews
	tp.addOptionalTag(tags, "securityreview", tp.Config.SecurityReview, naValue)
	tp.addOptionalTag(tags, "privacyreview", tp.Config.PrivacyReview, naValue)

	// Git repository tags (if enabled)
	if tp.Config.SourceRepoTagsEnabled {
		tp.addGitTags(tags)
	}

	// Merge additional tags
	maps.Copy(tags, tp.Config.AdditionalTags)

	return tp.finalize(tags)
}

// ProcessDataTags generates data-specific tags
func (tp *TagProcessor) ProcessDataTags() (map[string]string, error) {
	tags := make(map[string]string)
	delimiter := tp.CloudProvider.GetDelimiter()
	naValue := tp.CloudProvider.GetNAValue()

	// Data classification
	tags["sensitivity"] = tp.Config.Sensitivity
	tp.addOptionalTag(tags, "dataregulations", strings.Join(tp.Config.DataRegs, delimiter), naValue)

	// Data sovereignty is only emitted when supplied, never as N/A
	if tp.Config.DataSovereignty != "" {
		tags["sovereignty"] = tp.Config.DataSovereignty
	}

	// Data ownership (if enabled)
	if tp.Config.OwnerTagsEnabled {
		tp.addOptionalTag(tags, "dataowners", strings.Join(tp.Config.DataOwners, delimiter), naValue)
	}

	// Merge additional data tags
	maps.Copy(tags, tp.Config.AdditionalDataTags)

	return tp.finalize(tags)
}

// finalize applies the tag prefix, validates keys, and sanitizes and truncates values
func (tp *TagProcessor) finalize(tags map[string]string) (map[string]string, error) {
	maxLen := tp.CloudProvider.GetMaxTagLength()

	keys := make([]string, 0, len(tags))
	for k := range tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	prefixedTags := make(map[string]string, len(tags))
	for _, k := range keys {
		key := tp.TagPrefix + k
		if !tp.CloudProvider.ValidateTagKey(key) {
			return nil, fmt.Errorf("invalid tag key %q for target cloud rules", key)
		}
		value := tp.CloudProvider.SanitizeTagValue(tags[k])
		prefixedTags[key] = truncateRunes(value, maxLen)
	}

	return prefixedTags, nil
}

// addGitTags adds source repository tags, recording a warning when git is unavailable
func (tp *TagProcessor) addGitTags(tags map[string]string) {
	gitInfoFunc := tp.GitInfoFunc
	if gitInfoFunc == nil {
		gitInfoFunc = GetGitInfo
	}

	gitInfo, err := gitInfoFunc()
	if err != nil || gitInfo == nil {
		reason := "no repository information"
		if err != nil {
			reason = err.Error()
		}
		tp.Warnings = append(tp.Warnings, "source repository tags skipped: "+reason)
		return
	}

	if gitInfo.RepoURL != "" {
		tags["sourcerepo"] = gitInfo.RepoURL
	} else {
		tp.Warnings = append(tp.Warnings, "sourcerepo tag skipped: repository remote URL unavailable")
	}
	if gitInfo.CommitHash != "" {
		tags["sourcecommit"] = gitInfo.CommitHash
	} else {
		tp.Warnings = append(tp.Warnings, "sourcecommit tag skipped: commit unavailable")
	}
}

// addSystemTag adds a project management or ITSM identifier tag when the id is supplied,
// prefixed with its platform when system prefixes are enabled
func (tp *TagProcessor) addSystemTag(tags map[string]string, key, platform, id, delimiter string) {
	if id == "" {
		return
	}
	if tp.Config.SystemPrefixesEnabled && platform != "" {
		tags[key] = platform + delimiter + id
		return
	}
	tags[key] = id
}

// addOptionalTag adds a tag if value is not empty, or the N/A value when N/A is enabled
func (tp *TagProcessor) addOptionalTag(tags map[string]string, key, value, naValue string) {
	if value != "" {
		tags[key] = value
	} else if tp.Config.NotApplicableEnabled {
		tags[key] = naValue
	}
}

// truncateRunes shortens s to at most max characters without splitting a multi-byte character
func truncateRunes(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	runes := []rune(s)
	return string(runes[:max])
}

// EffectiveEnvironmentName returns the environment tag value: the environment name,
// otherwise the environment, otherwise DefaultEnvironmentName
func EffectiveEnvironmentName(config *DataSourceConfig) string {
	if config.EnvironmentName != "" {
		return config.EnvironmentName
	}
	if config.Environment != "" {
		return config.Environment
	}
	return DefaultEnvironmentName
}

// ResolveDeletionDate sets the effective deletion date: an explicit date is kept,
// Ephemeral environments get today (UTC) plus 90 days, and everything else gets "never"
func ResolveDeletionDate(config *DataSourceConfig) {
	if config.DeletionDate != "" {
		return
	}
	if config.EnvironmentType == "Ephemeral" {
		config.DeletionDate = now().UTC().Add(ephemeralRetention).Format("2006-01-02")
		return
	}
	config.DeletionDate = NeverDeletionDate
}

// ProcessEphemeralEnvironment handles ephemeral environment special logic.
//
// Deprecated: use ResolveDeletionDate, which also applies the "never" default.
func ProcessEphemeralEnvironment(config *DataSourceConfig) {
	ResolveDeletionDate(config)
}

// ConvertTagsToListOfMaps converts tags map to a sorted list of maps with key, value
// and propagate_at_launch entries, matching the legacy module's shape
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
			"key":                 k,
			"value":               tags[k],
			"propagate_at_launch": "true",
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

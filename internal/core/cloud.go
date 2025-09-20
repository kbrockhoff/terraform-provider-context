package core

import (
	"regexp"
	"strings"
)

// Precompiled regular expressions
var (
	awsSanitizeRegex        = regexp.MustCompile(`[^a-zA-Z0-9 \\.:=+@_/-]`)
	awsValidateKeyRegex     = regexp.MustCompile(`^[a-zA-Z0-9 +\-=._:/]+$`)
	azureSanitizeRegex      = regexp.MustCompile(`[ <>%&\\?/#:]`)
	azureValidateKeyRegex   = regexp.MustCompile(`[<>%&\\?/]`)
	gcpSanitizeRegex        = regexp.MustCompile(`[^a-z0-9_-]`)
	gcpValidateKeyRegex     = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)
	defaultSanitizeRegex    = regexp.MustCompile(`[<>%&\\?]`)
	defaultValidateKeyRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

// CloudProvider interface defines cloud-specific tag formatting rules
type CloudProvider interface {
	GetMaxTagLength() int
	GetDelimiter() string
	GetNAValue() string
	SanitizeTagValue(value string) string
	ValidateTagKey(key string) bool
}

// AWSProvider implements CloudProvider for AWS
type AWSProvider struct{}

func (p *AWSProvider) GetMaxTagLength() int {
	return 256
}

func (p *AWSProvider) GetDelimiter() string {
	return " "
}

func (p *AWSProvider) GetNAValue() string {
	return "N/A"
}

func (p *AWSProvider) SanitizeTagValue(value string) string {
	// Replace characters not matching /[a-zA-Z0-9 \\.:=+@_/-]/ with _
	return awsSanitizeRegex.ReplaceAllString(value, "_")
}

func (p *AWSProvider) ValidateTagKey(key string) bool {
	// AWS tag keys can contain letters, numbers, spaces, and +-=._:/
	return awsValidateKeyRegex.MatchString(key)
}

// AzureProvider implements CloudProvider for Azure
type AzureProvider struct{}

func (p *AzureProvider) GetMaxTagLength() int {
	return 256
}

func (p *AzureProvider) GetDelimiter() string {
	return ";"
}

func (p *AzureProvider) GetNAValue() string {
	return "NotApplicable"
}

func (p *AzureProvider) SanitizeTagValue(value string) string {
	// Replace /[ <>%&\\?/#:]/ with empty string
	return azureSanitizeRegex.ReplaceAllString(value, "")
}

func (p *AzureProvider) ValidateTagKey(key string) bool {
	// Azure tag keys cannot contain <, >, %, &, \, ?, /
	return !azureValidateKeyRegex.MatchString(key)
}

// GCPProvider implements CloudProvider for GCP
type GCPProvider struct{}

func (p *GCPProvider) GetMaxTagLength() int {
	return 63
}

func (p *GCPProvider) GetDelimiter() string {
	return "_"
}

func (p *GCPProvider) GetNAValue() string {
	return "not_applicable"
}

func (p *GCPProvider) SanitizeTagValue(value string) string {
	// Convert to lowercase and replace non-alphanumeric/underscore/hyphen with hyphen
	value = strings.ToLower(value)
	return gcpSanitizeRegex.ReplaceAllString(value, "-")
}

func (p *GCPProvider) ValidateTagKey(key string) bool {
	// GCP labels must be lowercase letters, numbers, hyphens, underscores
	return gcpValidateKeyRegex.MatchString(key)
}

// DefaultProvider implements CloudProvider for DC and other providers
type DefaultProvider struct{}

func (p *DefaultProvider) GetMaxTagLength() int {
	return 63
}

func (p *DefaultProvider) GetDelimiter() string {
	return ";"
}

func (p *DefaultProvider) GetNAValue() string {
	return "N/A"
}

func (p *DefaultProvider) SanitizeTagValue(value string) string {
	// Replace /[<>%&\\?]/ with _
	return defaultSanitizeRegex.ReplaceAllString(value, "_")
}

func (p *DefaultProvider) ValidateTagKey(key string) bool {
	// Basic validation - no special characters that could cause issues
	return defaultValidateKeyRegex.MatchString(key)
}

// GetCloudProvider returns the appropriate CloudProvider implementation
func GetCloudProvider(provider string) CloudProvider {
	switch provider {
	case "aws":
		return &AWSProvider{}
	case "az":
		return &AzureProvider{}
	case "gcp":
		return &GCPProvider{}
	default:
		return &DefaultProvider{}
	}
}

package core

// This package re-exports from pkg/context for backward compatibility
// New code should import from github.com/kbrockhoff/terraform-provider-context/pkg/context directly

import (
	ctx "github.com/kbrockhoff/terraform-provider-context/pkg/context"
)

// TagProcessor handles tag generation and processing
type TagProcessor = ctx.TagProcessor

// DataSourceConfig contains all configuration fields from the data source
type DataSourceConfig = ctx.DataSourceConfig

// ProcessEphemeralEnvironment handles ephemeral environment special logic
func ProcessEphemeralEnvironment(config *DataSourceConfig) {
	ctx.ProcessEphemeralEnvironment(config)
}

// ConvertTagsToListOfMaps converts tags map to list of maps for AWS
func ConvertTagsToListOfMaps(tags map[string]string) []map[string]string {
	return ctx.ConvertTagsToListOfMaps(tags)
}

// ConvertTagsToKVPList converts tags to key=value pairs
func ConvertTagsToKVPList(tags map[string]string) []string {
	return ctx.ConvertTagsToKVPList(tags)
}

// ConvertTagsToCommaSeparated converts tags to comma-separated string
func ConvertTagsToCommaSeparated(tags map[string]string) string {
	return ctx.ConvertTagsToCommaSeparated(tags)
}

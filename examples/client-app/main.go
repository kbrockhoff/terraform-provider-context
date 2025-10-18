package main

import (
	"fmt"
	"log"

	"github.com/kbrockhoff/terraform-provider-context/pkg/context"
)

// This example demonstrates how to use the terraform-provider-context package
// in a standalone Go application (e.g., a serverless function)
func main() {
	fmt.Println("=== Terraform Provider Context - Client App Example ===\n")

	// Example 1: Generate a name prefix
	fmt.Println("Example 1: Name Generation")
	nameGen := &context.NameGenerator{
		Namespace:   "myorg",
		Name:        "payment-api",
		Environment: "prod",
	}
	namePrefix, err := nameGen.Generate()
	if err != nil {
		log.Fatalf("Failed to generate name: %v", err)
	}
	fmt.Printf("Generated name prefix: %s\n\n", namePrefix)

	// Example 2: Generate tags for a resource
	fmt.Println("Example 2: Tag Generation")

	// Create configuration
	config := &context.DataSourceConfig{
		Namespace:             "myorg",
		Environment:           "prod",
		EnvironmentName:       "Production",
		EnvironmentType:       "Production",
		Availability:          "dedicated",
		ManagedBy:             "terraform",
		CostCenter:            "engineering",
		ProductOwners:         []string{"product@example.com"},
		CodeOwners:            []string{"dev@example.com"},
		SourceRepoTagsEnabled: true,
		NotApplicableEnabled:  true,
		OwnerTagsEnabled:      true,
		AdditionalTags:        map[string]string{"team": "platform"},
		AdditionalDataTags:    make(map[string]string),
	}

	// Get cloud provider (AWS in this example)
	cloudProvider := context.GetCloudProvider("aws")

	// Create tag processor
	tagProcessor := &context.TagProcessor{
		CloudProvider: cloudProvider,
		Config:        config,
		TagPrefix:     "bc-",
	}

	// Generate tags
	tags, err := tagProcessor.Process()
	if err != nil {
		log.Fatalf("Failed to process tags: %v", err)
	}

	fmt.Println("Generated tags:")
	for key, value := range tags {
		fmt.Printf("  %s = %s\n", key, value)
	}

	// Example 3: Convert tags to different formats
	fmt.Println("\nExample 3: Tag Format Conversion")

	// Convert to list of maps (useful for AWS resources)
	tagsAsListOfMaps := context.ConvertTagsToListOfMaps(tags)
	fmt.Printf("Tags as list of maps: %d tags\n", len(tagsAsListOfMaps))

	// Convert to key=value pairs
	tagsAsKVP := context.ConvertTagsToKVPList(tags)
	fmt.Println("Tags as key=value pairs:")
	for _, kvp := range tagsAsKVP[:3] { // Show first 3
		fmt.Printf("  %s\n", kvp)
	}
	fmt.Printf("  ... and %d more\n", len(tagsAsKVP)-3)

	// Convert to comma-separated string
	tagsAsCSV := context.ConvertTagsToCommaSeparated(tags)
	fmt.Printf("\nTags as CSV (first 100 chars): %s...\n", tagsAsCSV[:min(100, len(tagsAsCSV))])

	// Example 4: Get Git information
	fmt.Println("\nExample 4: Git Integration")
	gitInfo, err := context.GetGitInfo()
	if err == nil && gitInfo != nil {
		fmt.Printf("Repository URL: %s\n", gitInfo.RepoURL)
		fmt.Printf("Commit Hash: %s\n", gitInfo.CommitHash)
	} else {
		fmt.Println("Git information not available (not in a git repository)")
	}

	fmt.Println("\n=== Example Complete ===")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

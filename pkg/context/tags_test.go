package context

import (
	"testing"
)

func TestTagProcessor_WithGitTags(t *testing.T) {
	// Setup config with git tags enabled
	config := &DataSourceConfig{
		Namespace:             "test",
		Environment:           "dev",
		EnvironmentName:       "Development",
		EnvironmentType:       "Development",
		Availability:          "standard",
		ManagedBy:             "terraform",
		SourceRepoTagsEnabled: true,
		NotApplicableEnabled:  true,
		AdditionalTags:        make(map[string]string),
		AdditionalDataTags:    make(map[string]string),
	}

	cp := GetCloudProvider("dc")
	processor := &TagProcessor{
		CloudProvider: cp,
		Config:        config,
		TagPrefix:     "test-",
	}

	// Process tags
	tags, err := processor.Process()
	if err != nil {
		t.Fatalf("Failed to process tags: %v", err)
	}

	// Verify git tags are present (if git is available)
	// Note: These tags may not be present in non-git environments
	gitInfo, gitErr := GetGitInfo()
	if gitErr == nil && gitInfo != nil && gitInfo.RepoURL != "" {
		// Git is available, verify tags are included
		if _, ok := tags["test-sourcerepo"]; !ok {
			t.Error("Expected test-sourcerepo tag to be present when git is available")
		}
		if _, ok := tags["test-sourcecommit"]; !ok {
			t.Error("Expected test-sourcecommit tag to be present when git is available")
		}

		// Verify values are not empty
		if tags["test-sourcerepo"] == "" {
			t.Error("Expected test-sourcerepo to have a value")
		}
		if tags["test-sourcecommit"] == "" {
			t.Error("Expected test-sourcecommit to have a value")
		}
	}
}

func TestTagProcessor_WithoutGitTags(t *testing.T) {
	// Setup config with git tags disabled
	config := &DataSourceConfig{
		Namespace:             "test",
		Environment:           "dev",
		EnvironmentName:       "Development",
		EnvironmentType:       "Development",
		Availability:          "standard",
		ManagedBy:             "terraform",
		SourceRepoTagsEnabled: false,
		NotApplicableEnabled:  true,
		AdditionalTags:        make(map[string]string),
		AdditionalDataTags:    make(map[string]string),
	}

	cp := GetCloudProvider("dc")
	processor := &TagProcessor{
		CloudProvider: cp,
		Config:        config,
		TagPrefix:     "test-",
	}

	// Process tags
	tags, err := processor.Process()
	if err != nil {
		t.Fatalf("Failed to process tags: %v", err)
	}

	// Verify git tags are NOT present when disabled
	if _, ok := tags["test-sourcerepo"]; ok {
		t.Error("Expected test-sourcerepo tag to be absent when disabled")
	}
	if _, ok := tags["test-sourcecommit"]; ok {
		t.Error("Expected test-sourcecommit tag to be absent when disabled")
	}
}

func TestTagProcessor_RequiredTags(t *testing.T) {
	// Setup minimal config
	config := &DataSourceConfig{
		Namespace:             "myorg",
		Environment:           "prod",
		EnvironmentName:       "Production",
		EnvironmentType:       "Production",
		Availability:          "dedicated",
		ManagedBy:             "terraform",
		SourceRepoTagsEnabled: true,
		NotApplicableEnabled:  true,
		AdditionalTags:        make(map[string]string),
		AdditionalDataTags:    make(map[string]string),
	}

	cp := GetCloudProvider("aws")
	processor := &TagProcessor{
		CloudProvider: cp,
		Config:        config,
		TagPrefix:     "bc-",
	}

	// Process tags
	tags, err := processor.Process()
	if err != nil {
		t.Fatalf("Failed to process tags: %v", err)
	}

	// Verify required tags are present
	requiredTags := []string{
		"bc-environment",
		"bc-availability",
		"bc-managedby",
	}

	for _, tag := range requiredTags {
		if _, ok := tags[tag]; !ok {
			t.Errorf("Expected required tag %s to be present", tag)
		}
	}
}

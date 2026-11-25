package context

import (
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"
)

// fixedGit is a GitInfoFunc returning deterministic repository information
func fixedGit() (*GitInfo, error) {
	return &GitInfo{RepoURL: "https://github.com/org/repo", CommitHash: "abc123"}, nil
}

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

	processor := &TagProcessor{
		CloudProvider: GetCloudProvider("dc"),
		Config:        config,
		TagPrefix:     "test-",
		GitInfoFunc:   fixedGit,
	}

	// Process tags
	tags, err := processor.Process()
	if err != nil {
		t.Fatalf("Failed to process tags: %v", err)
	}

	if tags["test-sourcerepo"] != "https://github.com/org/repo" {
		t.Errorf("test-sourcerepo = %q, want repository URL", tags["test-sourcerepo"])
	}
	if tags["test-sourcecommit"] != "abc123" {
		t.Errorf("test-sourcecommit = %q, want abc123", tags["test-sourcecommit"])
	}
	if len(processor.Warnings) != 0 {
		t.Errorf("Warnings = %v, want none", processor.Warnings)
	}
}

func TestTagProcessor_GitUnavailableWarns(t *testing.T) {
	config := &DataSourceConfig{
		Availability:          "standard",
		ManagedBy:             "terraform",
		SourceRepoTagsEnabled: true,
		NotApplicableEnabled:  true,
	}

	processor := &TagProcessor{
		CloudProvider: GetCloudProvider("dc"),
		Config:        config,
		TagPrefix:     "test-",
		GitInfoFunc:   noGit,
	}

	tags, err := processor.Process()
	if err != nil {
		t.Fatalf("Process() error = %v, want git problems to be non-fatal", err)
	}
	if _, ok := tags["test-sourcerepo"]; ok {
		t.Error("test-sourcerepo present when git is unavailable")
	}
	if len(processor.Warnings) != 1 || !strings.Contains(processor.Warnings[0], "source repository tags skipped") {
		t.Errorf("Warnings = %v, want one skipped-tags warning", processor.Warnings)
	}

	// Warnings are reset on every call
	processor.GitInfoFunc = fixedGit
	if _, err := processor.Process(); err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if len(processor.Warnings) != 0 {
		t.Errorf("Warnings = %v, want reset to none", processor.Warnings)
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

	processor := &TagProcessor{
		CloudProvider: GetCloudProvider("dc"),
		Config:        config,
		TagPrefix:     "test-",
		GitInfoFunc: func() (*GitInfo, error) {
			t.Error("git must not be queried when source repo tags are disabled")
			return nil, nil
		},
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
		DeletionDate:          NeverDeletionDate,
		SourceRepoTagsEnabled: true,
		NotApplicableEnabled:  true,
		AdditionalTags:        make(map[string]string),
		AdditionalDataTags:    make(map[string]string),
	}

	processor := &TagProcessor{
		CloudProvider: GetCloudProvider("aws"),
		Config:        config,
		TagPrefix:     "bc-",
		GitInfoFunc:   fixedGit,
	}

	// Process tags
	tags, err := processor.Process()
	if err != nil {
		t.Fatalf("Failed to process tags: %v", err)
	}

	// Verify required tags are present
	requiredTags := map[string]string{
		"bc-environment":  "Production",
		"bc-availability": "dedicated",
		"bc-managedby":    "terraform",
		"bc-deletiondate": "never",
	}

	for tag, want := range requiredTags {
		if tags[tag] != want {
			t.Errorf("Expected required tag %s = %q, got %q", tag, want, tags[tag])
		}
	}
}

func TestTagProcessor_AdditionalTagsOverride(t *testing.T) {
	config := baseConfig()
	config.AdditionalTags = map[string]string{"managedby": "pulumi", "team": "platform"}
	config.AdditionalDataTags = map[string]string{"sensitivity": "public"}

	tags, dataTags := mustProcess(t, newParityProcessor("aws", config))
	if tags["bc-managedby"] != "pulumi" {
		t.Errorf("bc-managedby = %q, want additional tag to override", tags["bc-managedby"])
	}
	if tags["bc-team"] != "platform" {
		t.Errorf("bc-team = %q, want platform", tags["bc-team"])
	}
	if dataTags["bc-sensitivity"] != "public" {
		t.Errorf("bc-sensitivity = %q, want additional data tag to override", dataTags["bc-sensitivity"])
	}
}

func TestTagProcessor_OwnersJoinedWithDelimiter(t *testing.T) {
	tests := map[string]string{
		"aws": "a@example.com b@example.com",
		"az":  "a@example.com;b@example.com",
		"dc":  "a@example.com;b@example.com",
	}

	for cloud, want := range tests {
		t.Run(cloud, func(t *testing.T) {
			config := baseConfig()
			config.CodeOwners = []string{"a@example.com", "b@example.com"}
			config.DataRegs = []string{"GDPR", "CCPA"}
			tags, dataTags := mustProcess(t, newParityProcessor(cloud, config))
			if tags["bc-codeowners"] != want {
				t.Errorf("bc-codeowners = %q, want %q", tags["bc-codeowners"], want)
			}
			delimiter := GetCloudProvider(cloud).GetDelimiter()
			if dataTags["bc-dataregulations"] != "GDPR"+delimiter+"CCPA" {
				t.Errorf("bc-dataregulations = %q", dataTags["bc-dataregulations"])
			}
		})
	}
}

func TestTagProcessor_TruncatesOnRuneBoundaries(t *testing.T) {
	t.Run("azure multi-byte value", func(t *testing.T) {
		config := baseConfig()
		config.AdditionalTags = map[string]string{"note": strings.Repeat("é", 300)}
		tags, _ := mustProcess(t, newParityProcessor("az", config))
		got := tags["bc-note"]
		if n := utf8.RuneCountInString(got); n != 256 {
			t.Errorf("rune count = %d, want 256", n)
		}
		if !utf8.ValidString(got) {
			t.Error("truncated value is not valid UTF-8")
		}
	})

	t.Run("gcp ascii value", func(t *testing.T) {
		config := baseConfig()
		config.AdditionalTags = map[string]string{"note": strings.Repeat("a", 70)}
		tags, _ := mustProcess(t, newParityProcessor("gcp", config))
		if n := len(tags["bc-note"]); n != 63 {
			t.Errorf("length = %d, want 63", n)
		}
	})

	t.Run("short value unchanged", func(t *testing.T) {
		if got := truncateRunes("héllo", 10); got != "héllo" {
			t.Errorf("truncateRunes() = %q, want unchanged", got)
		}
	})
}

func TestTagProcessor_ValidatesTagKeys(t *testing.T) {
	tests := []struct {
		name       string
		cloud      string
		prefix     string
		additional map[string]string
		data       map[string]string
		wantErr    string
	}{
		{name: "gcp rejects uppercase additional key", cloud: "gcp", prefix: "bc-", additional: map[string]string{"Team": "x"}, wantErr: "bc-Team"},
		{name: "dc rejects colon prefix", cloud: "dc", prefix: "bc:", wantErr: "bc:"},
		{name: "gcp rejects uppercase data key", cloud: "gcp", prefix: "bc-", data: map[string]string{"Owner": "x"}, wantErr: "bc-Owner"},
		{name: "aws accepts lowercase key", cloud: "aws", prefix: "bc-", additional: map[string]string{"team": "x"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := baseConfig()
			config.DeletionDate = NeverDeletionDate
			config.AdditionalTags = tt.additional
			config.AdditionalDataTags = tt.data
			tp := newParityProcessor(tt.cloud, config)
			tp.TagPrefix = tt.prefix

			_, err := tp.Process()
			if err == nil {
				_, err = tp.ProcessDataTags()
			}
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %v, want mention of %q", err, tt.wantErr)
			}
		})
	}
}

func TestConvertTagsToListOfMaps(t *testing.T) {
	got := ConvertTagsToListOfMaps(map[string]string{"b": "2", "a": "1"})
	want := []map[string]string{
		{"key": "a", "value": "1", "propagate_at_launch": "true"},
		{"key": "b", "value": "2", "propagate_at_launch": "true"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ConvertTagsToListOfMaps() = %v, want %v", got, want)
	}
}

func TestConvertTagsToKVPListAndCommaSeparated(t *testing.T) {
	tags := map[string]string{"b": "2", "a": "1"}
	if got := ConvertTagsToKVPList(tags); !reflect.DeepEqual(got, []string{"a=1", "b=2"}) {
		t.Errorf("ConvertTagsToKVPList() = %v", got)
	}
	if got := ConvertTagsToCommaSeparated(tags); got != "a=1,b=2" {
		t.Errorf("ConvertTagsToCommaSeparated() = %q", got)
	}
	if got := ConvertTagsToCommaSeparated(nil); got != "" {
		t.Errorf("ConvertTagsToCommaSeparated(nil) = %q, want empty", got)
	}
}

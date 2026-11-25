package context

import (
	"errors"
	"testing"
	"time"
)

// Parity tests encode the behavior of the legacy terraform-external-context module
// (spec 001, legacy parity gaps 2-6) plus the data sovereignty tag (FR-028).

// noGit is a GitInfoFunc that reports git as unavailable
func noGit() (*GitInfo, error) {
	return nil, errors.New("git unavailable in test")
}

// newParityProcessor builds a TagProcessor with git disabled via a stub
func newParityProcessor(cloud string, config *DataSourceConfig) *TagProcessor {
	return &TagProcessor{
		CloudProvider: GetCloudProvider(cloud),
		Config:        config,
		TagPrefix:     "bc-",
		GitInfoFunc:   noGit,
	}
}

// baseConfig returns the defaults the data source applies before processing
func baseConfig() *DataSourceConfig {
	return &DataSourceConfig{
		Name:                  "app",
		Availability:          "preemptable",
		ManagedBy:             "terraform",
		Sensitivity:           "confidential",
		SystemPrefixesEnabled: true,
		NotApplicableEnabled:  true,
		OwnerTagsEnabled:      true,
	}
}

func mustProcess(t *testing.T, tp *TagProcessor) (map[string]string, map[string]string) {
	t.Helper()
	if tp.Config.DeletionDate == "" {
		ResolveDeletionDate(tp.Config)
	}
	tags, err := tp.Process()
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	dataTags, err := tp.ProcessDataTags()
	if err != nil {
		t.Fatalf("ProcessDataTags() error = %v", err)
	}
	return tags, dataTags
}

func TestParity_AbsentPMAndITSMOmitted(t *testing.T) {
	keys := []string{"bc-projectmgmtid", "bc-systemid", "bc-componentid", "bc-instanceid"}

	t.Run("no ids with N/A on", func(t *testing.T) {
		tags, _ := mustProcess(t, newParityProcessor("aws", baseConfig()))
		for _, k := range keys {
			if _, ok := tags[k]; ok {
				t.Errorf("tag %s present, want omitted", k)
			}
		}
	})

	t.Run("platform without project code", func(t *testing.T) {
		config := baseConfig()
		config.PMPlatform = "JIRA"
		config.ITSMPlatform = "SNOW"
		tags, _ := mustProcess(t, newParityProcessor("aws", config))
		for _, k := range keys {
			if _, ok := tags[k]; ok {
				t.Errorf("tag %s present, want omitted", k)
			}
		}
	})

	t.Run("ids with platform prefixes", func(t *testing.T) {
		config := baseConfig()
		config.PMPlatform = "JIRA"
		config.PMProjectCode = "PAY"
		config.ITSMPlatform = "SNOW"
		config.ITSMSystemID = "S1"
		config.ITSMInstanceID = "I1"
		tags, _ := mustProcess(t, newParityProcessor("aws", config))
		want := map[string]string{
			"bc-projectmgmtid": "JIRA PAY",
			"bc-systemid":      "SNOW S1",
			"bc-instanceid":    "SNOW I1",
		}
		for k, v := range want {
			if tags[k] != v {
				t.Errorf("tag %s = %q, want %q", k, tags[k], v)
			}
		}
		if _, ok := tags["bc-componentid"]; ok {
			t.Error("bc-componentid present without a component id")
		}
	})

	t.Run("ids without prefixes when disabled", func(t *testing.T) {
		config := baseConfig()
		config.SystemPrefixesEnabled = false
		config.PMPlatform = "JIRA"
		config.PMProjectCode = "PAY"
		config.ITSMPlatform = "SNOW"
		config.ITSMComponentID = "C1"
		tags, _ := mustProcess(t, newParityProcessor("dc", config))
		if tags["bc-projectmgmtid"] != "PAY" || tags["bc-componentid"] != "C1" {
			t.Errorf("got projectmgmtid=%q componentid=%q, want bare ids",
				tags["bc-projectmgmtid"], tags["bc-componentid"])
		}
	})
}

func TestParity_EnvironmentFallback(t *testing.T) {
	tests := []struct {
		name            string
		environment     string
		environmentName string
		want            string
	}{
		{name: "environment name wins", environment: "prd", environmentName: "Production", want: "Production"},
		{name: "falls back to environment", environment: "prd", want: "prd"},
		{name: "falls back to dev", want: "dev"},
	}

	for _, na := range []bool{true, false} {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				config := baseConfig()
				config.NotApplicableEnabled = na
				config.Environment = tt.environment
				config.EnvironmentName = tt.environmentName
				tags, _ := mustProcess(t, newParityProcessor("dc", config))
				if tags["bc-environment"] != tt.want {
					t.Errorf("bc-environment = %q, want %q (N/A=%v)", tags["bc-environment"], tt.want, na)
				}
				if got := EffectiveEnvironmentName(config); got != tt.want {
					t.Errorf("EffectiveEnvironmentName() = %q, want %q", got, tt.want)
				}
			})
		}
	}
}

func TestParity_DeletionDate(t *testing.T) {
	fixed := time.Date(2026, 11, 25, 23, 30, 0, 0, time.UTC)
	original := now
	now = func() time.Time { return fixed }
	t.Cleanup(func() { now = original })

	tests := []struct {
		name     string
		envType  string
		explicit string
		want     string
	}{
		{name: "non-ephemeral defaults to never", envType: "Development", want: "never"},
		{name: "no type defaults to never", want: "never"},
		{name: "explicit date kept", envType: "Production", explicit: "2030-01-01", want: "2030-01-01"},
		{name: "ephemeral explicit date wins", envType: "Ephemeral", explicit: "2030-01-01", want: "2030-01-01"},
		{name: "ephemeral computes plus 90 days", envType: "Ephemeral", want: "2027-02-23"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := baseConfig()
			config.EnvironmentType = tt.envType
			config.DeletionDate = tt.explicit
			ResolveDeletionDate(config)
			if config.DeletionDate != tt.want {
				t.Fatalf("ResolveDeletionDate() = %q, want %q", config.DeletionDate, tt.want)
			}
			tags, _ := mustProcess(t, newParityProcessor("dc", config))
			if tags["bc-deletiondate"] != tt.want {
				t.Errorf("bc-deletiondate = %q, want %q", tags["bc-deletiondate"], tt.want)
			}
		})
	}

	t.Run("deprecated wrapper delegates", func(t *testing.T) {
		config := baseConfig()
		config.EnvironmentType = "Ephemeral"
		ProcessEphemeralEnvironment(config)
		if config.DeletionDate != "2027-02-23" {
			t.Errorf("ProcessEphemeralEnvironment() = %q, want 2027-02-23", config.DeletionDate)
		}
	})
}

func TestParity_DataOwnersGatedByOwnerTags(t *testing.T) {
	config := baseConfig()
	config.OwnerTagsEnabled = false
	config.DataOwners = []string{"data@example.com"}
	config.ProductOwners = []string{"product@example.com"}
	tags, dataTags := mustProcess(t, newParityProcessor("aws", config))
	if _, ok := dataTags["bc-dataowners"]; ok {
		t.Error("bc-dataowners present with owner tags disabled")
	}
	if _, ok := tags["bc-productowners"]; ok {
		t.Error("bc-productowners present with owner tags disabled")
	}
}

func TestParity_NotApplicableScope(t *testing.T) {
	naKeys := []string{
		"bc-costcenter", "bc-productowners", "bc-codeowners",
		"bc-securityreview", "bc-privacyreview",
	}
	naDataKeys := []string{"bc-dataregulations", "bc-dataowners"}
	alwaysKeys := []string{"bc-environment", "bc-availability", "bc-managedby", "bc-deletiondate"}

	clouds := map[string]string{
		"aws": "N/A",
		"az":  "NotApplicable",
		"gcp": "not_applicable",
		"dc":  "N/A",
	}

	for cloud, naValue := range clouds {
		t.Run(cloud+" N/A on", func(t *testing.T) {
			tags, dataTags := mustProcess(t, newParityProcessor(cloud, baseConfig()))
			for _, k := range naKeys {
				if tags[k] != naValue {
					t.Errorf("%s = %q, want %q", k, tags[k], naValue)
				}
			}
			for _, k := range naDataKeys {
				if dataTags[k] != naValue {
					t.Errorf("%s = %q, want %q", k, dataTags[k], naValue)
				}
			}
			// Only the always-present keys and the N/A keys exist
			if len(tags) != len(naKeys)+len(alwaysKeys) {
				t.Errorf("got %d resource tags %v, want %d", len(tags), tags, len(naKeys)+len(alwaysKeys))
			}
		})

		t.Run(cloud+" N/A off", func(t *testing.T) {
			config := baseConfig()
			config.NotApplicableEnabled = false
			tags, dataTags := mustProcess(t, newParityProcessor(cloud, config))
			for _, k := range append(naKeys, naDataKeys...) {
				if _, ok := tags[k]; ok {
					t.Errorf("%s present with N/A off", k)
				}
				if _, ok := dataTags[k]; ok {
					t.Errorf("%s present in data tags with N/A off", k)
				}
			}
			for _, k := range alwaysKeys {
				if _, ok := tags[k]; !ok {
					t.Errorf("%s missing with N/A off", k)
				}
			}
		})
	}
}

func TestDataSovereigntyTag(t *testing.T) {
	t.Run("emitted as data tag when supplied", func(t *testing.T) {
		config := baseConfig()
		config.DataSovereignty = "US-CA"
		tags, dataTags := mustProcess(t, newParityProcessor("aws", config))
		if dataTags["bc-sovereignty"] != "US-CA" {
			t.Errorf("bc-sovereignty = %q, want US-CA", dataTags["bc-sovereignty"])
		}
		if _, ok := tags["bc-sovereignty"]; ok {
			t.Error("bc-sovereignty must not be a resource tag")
		}
	})

	t.Run("normalized for gcp", func(t *testing.T) {
		config := baseConfig()
		config.DataSovereignty = "US-CA"
		_, dataTags := mustProcess(t, newParityProcessor("gcp", config))
		if dataTags["bc-sovereignty"] != "us-ca" {
			t.Errorf("bc-sovereignty = %q, want us-ca", dataTags["bc-sovereignty"])
		}
	})

	for _, cloud := range []string{"aws", "az", "gcp", "dc"} {
		t.Run(cloud+" omitted when not supplied", func(t *testing.T) {
			_, dataTags := mustProcess(t, newParityProcessor(cloud, baseConfig()))
			if v, ok := dataTags["bc-sovereignty"]; ok {
				t.Errorf("bc-sovereignty = %q, want omitted (never N/A)", v)
			}
		})
	}
}

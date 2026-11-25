package context

import "testing"

// benchConfig populates every DataSourceConfig field
func benchConfig() *DataSourceConfig {
	return &DataSourceConfig{
		Namespace:             "myorg",
		Name:                  "payment-api",
		Environment:           "prod",
		EnvironmentName:       "Production",
		EnvironmentType:       "Production",
		Enabled:               true,
		Availability:          "dedicated",
		ManagedBy:             "terraform",
		DeletionDate:          NeverDeletionDate,
		PMPlatform:            "JIRA",
		PMProjectCode:         "PAY",
		ITSMPlatform:          "SNOW",
		ITSMSystemID:          "S1",
		ITSMComponentID:       "C1",
		ITSMInstanceID:        "I1",
		CostCenter:            "engineering",
		ProductOwners:         []string{"a@example.com", "b@example.com", "c@example.com"},
		CodeOwners:            []string{"a@example.com", "b@example.com", "c@example.com"},
		DataOwners:            []string{"a@example.com", "b@example.com", "c@example.com"},
		Sensitivity:           "confidential",
		DataRegs:              []string{"GDPR", "CCPA"},
		DataSovereignty:       "US-CA",
		SecurityReview:        "SR-1",
		PrivacyReview:         "PR-1",
		SourceRepoTagsEnabled: true,
		SystemPrefixesEnabled: true,
		NotApplicableEnabled:  true,
		OwnerTagsEnabled:      true,
		AdditionalTags:        map[string]string{"t1": "v1", "t2": "v2", "t3": "v3", "t4": "v4", "t5": "v5"},
		AdditionalDataTags:    map[string]string{"d1": "v1"},
	}
}

func BenchmarkTagProcessorProcess(b *testing.B) {
	tp := &TagProcessor{
		CloudProvider: GetCloudProvider("aws"),
		Config:        benchConfig(),
		TagPrefix:     "bc-",
		GitInfoFunc:   fixedGit,
	}
	for b.Loop() {
		if _, err := tp.Process(); err != nil {
			b.Fatal(err)
		}
		if _, err := tp.ProcessDataTags(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNameGenerator(b *testing.B) {
	ng := &NameGenerator{Namespace: "myorg", Name: "payment-api", Environment: "prod"}
	for b.Loop() {
		if _, err := ng.Generate(); err != nil {
			b.Fatal(err)
		}
	}
}

// TestTagProcessorPerformanceBudget enforces the 50 ms per-evaluation SLO
func TestTagProcessorPerformanceBudget(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance budget in short mode")
	}
	result := testing.Benchmark(BenchmarkTagProcessorProcess)
	if ns := result.NsPerOp(); ns > 50_000_000 {
		t.Errorf("tag processing took %d ns/op, budget is 50 ms", ns)
	}
}

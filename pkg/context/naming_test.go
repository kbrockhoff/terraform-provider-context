package context

import (
	"testing"
)

func TestNameGenerator_Generate(t *testing.T) {
	tests := []struct {
		name         string
		namespace    string
		resourceName string
		environment  string
		want         string
		wantErr      bool
	}{
		{
			name:         "standard format",
			namespace:    "myorg",
			resourceName: "app",
			environment:  "prod",
			want:         "myorg-app-prod",
			wantErr:      false,
		},
		{
			name:         "name only",
			namespace:    "",
			resourceName: "myapp",
			environment:  "",
			want:         "myapp",
			wantErr:      false,
		},
		{
			name:         "namespace and name",
			namespace:    "org",
			resourceName: "service",
			environment:  "",
			want:         "org-service",
			wantErr:      false,
		},
		{
			name:         "truncation required",
			namespace:    "verylongorg",
			resourceName: "verylongappname",
			environment:  "production",
			want:         "verylongorg-verylongappn", // Should truncate to fit 24 chars
			wantErr:      false,
		},
		{
			name:         "empty inputs",
			namespace:    "",
			resourceName: "",
			environment:  "",
			want:         "",
			wantErr:      true,
		},
		{
			name:         "single char name",
			namespace:    "",
			resourceName: "a",
			environment:  "",
			want:         "",
			wantErr:      true, // Too short (min 2 chars)
		},
		{
			name:         "uppercase converted",
			namespace:    "MyOrg",
			resourceName: "MyApp",
			environment:  "PROD",
			want:         "myorg-myapp-prod",
			wantErr:      false,
		},
		{
			name:         "max length exact",
			namespace:    "ab",
			resourceName: "abcdefghijklmnop",
			environment:  "cd",
			want:         "ab-abcdefghijklmnop-cd",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ng := &NameGenerator{
				Namespace:   tt.namespace,
				Name:        tt.resourceName,
				Environment: tt.environment,
			}
			got, err := ng.Generate()
			if (err != nil) != tt.wantErr {
				t.Errorf("NameGenerator.Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("NameGenerator.Generate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNameGenerator_IntelligentTruncate(t *testing.T) {
	tests := []struct {
		name  string
		ng    *NameGenerator
		input string
		want  string
	}{
		{
			name: "preserve namespace and environment",
			ng: &NameGenerator{
				Namespace:   "myorg",
				Name:        "verylongappname",
				Environment: "prod",
			},
			input: "myorg-verylongappname-prod",
			want:  "myorg-verylongappna-prod",
		},
		{
			name: "remove trailing hyphen",
			ng: &NameGenerator{
				Namespace:   "org",
				Name:        "app-name-test",
				Environment: "dev",
			},
			input: "org-app-name-test-dev",
			want:  "org-app-name-test-dev", // Already within limit
		},
		{
			name: "simple truncation",
			ng: &NameGenerator{
				Name: "verylongapplicationnamethatshouldbetruncated",
			},
			input: "verylongapplicationnamethatshouldbetruncated",
			want:  "verylongapplicationnamet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ng.intelligentTruncate(tt.input)
			if got != tt.want {
				t.Errorf("NameGenerator.intelligentTruncate() = %v, want %v", got, tt.want)
			}
		})
	}
}

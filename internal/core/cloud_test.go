package core

import (
	"fmt"
	"testing"
)

func TestAWSProvider(t *testing.T) {
	p := &AWSProvider{}

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "valid characters",
			input: "test-value_123",
			want:  "test-value_123",
		},
		{
			name:  "invalid characters replaced",
			input: "test#value$123",
			want:  "test_value_123",
		},
		{
			name:  "spaces preserved",
			input: "test value 123",
			want:  "test value 123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.SanitizeTagValue(tt.input)
			if got != tt.want {
				t.Errorf("AWSProvider.SanitizeTagValue() = %v, want %v", got, tt.want)
			}
		})
	}

	// Test other methods
	if p.GetMaxTagLength() != 256 {
		t.Errorf("AWSProvider.GetMaxTagLength() = %v, want 256", p.GetMaxTagLength())
	}
	if p.GetDelimiter() != " " {
		t.Errorf("AWSProvider.GetDelimiter() = %v, want ' '", p.GetDelimiter())
	}
	if p.GetNAValue() != "N/A" {
		t.Errorf("AWSProvider.GetNAValue() = %v, want 'N/A'", p.GetNAValue())
	}
}

func TestAzureProvider(t *testing.T) {
	p := &AzureProvider{}

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "valid characters",
			input: "test-value_123",
			want:  "test-value_123",
		},
		{
			name:  "spaces removed",
			input: "test value 123",
			want:  "testvalue123",
		},
		{
			name:  "special characters removed",
			input: "test<>%&\\?/#:value",
			want:  "testvalue",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.SanitizeTagValue(tt.input)
			if got != tt.want {
				t.Errorf("AzureProvider.SanitizeTagValue() = %v, want %v", got, tt.want)
			}
		})
	}

	// Test other methods
	if p.GetMaxTagLength() != 256 {
		t.Errorf("AzureProvider.GetMaxTagLength() = %v, want 256", p.GetMaxTagLength())
	}
	if p.GetDelimiter() != ";" {
		t.Errorf("AzureProvider.GetDelimiter() = %v, want ';'", p.GetDelimiter())
	}
	if p.GetNAValue() != "NotApplicable" {
		t.Errorf("AzureProvider.GetNAValue() = %v, want 'NotApplicable'", p.GetNAValue())
	}
}

func TestGCPProvider(t *testing.T) {
	p := &GCPProvider{}

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "valid characters lowercase",
			input: "test-value_123",
			want:  "test-value_123",
		},
		{
			name:  "uppercase converted",
			input: "TEST-VALUE",
			want:  "test-value",
		},
		{
			name:  "spaces replaced with hyphen",
			input: "test value 123",
			want:  "test-value-123",
		},
		{
			name:  "special characters replaced",
			input: "test@value#123",
			want:  "test-value-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.SanitizeTagValue(tt.input)
			if got != tt.want {
				t.Errorf("GCPProvider.SanitizeTagValue() = %v, want %v", got, tt.want)
			}
		})
	}

	// Test other methods
	if p.GetMaxTagLength() != 63 {
		t.Errorf("GCPProvider.GetMaxTagLength() = %v, want 63", p.GetMaxTagLength())
	}
	if p.GetDelimiter() != "_" {
		t.Errorf("GCPProvider.GetDelimiter() = %v, want '_'", p.GetDelimiter())
	}
	if p.GetNAValue() != "not_applicable" {
		t.Errorf("GCPProvider.GetNAValue() = %v, want 'not_applicable'", p.GetNAValue())
	}
}

func TestGetCloudProvider(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		wantType string
	}{
		{
			name:     "aws",
			provider: "aws",
			wantType: "*core.AWSProvider",
		},
		{
			name:     "azure",
			provider: "az",
			wantType: "*core.AzureProvider",
		},
		{
			name:     "gcp",
			provider: "gcp",
			wantType: "*core.GCPProvider",
		},
		{
			name:     "default",
			provider: "dc",
			wantType: "*core.DefaultProvider",
		},
		{
			name:     "unknown",
			provider: "unknown",
			wantType: "*core.DefaultProvider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetCloudProvider(tt.provider)
			gotType := fmt.Sprintf("%T", got)
			if gotType != tt.wantType {
				t.Errorf("GetCloudProvider(%s) returned type %v, want %v", tt.provider, gotType, tt.wantType)
			}
		})
	}
}

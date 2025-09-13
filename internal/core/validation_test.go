package core

import (
	"testing"
)

func TestValidateNamespace(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		wantErr   bool
	}{
		{
			name:      "valid namespace",
			namespace: "myorg",
			wantErr:   false,
		},
		{
			name:      "empty namespace",
			namespace: "",
			wantErr:   false, // Optional field
		},
		{
			name:      "single char",
			namespace: "a",
			wantErr:   false,
		},
		{
			name:      "max length",
			namespace: "orgname1",
			wantErr:   false,
		},
		{
			name:      "too long",
			namespace: "verylongorgname",
			wantErr:   true,
		},
		{
			name:      "uppercase",
			namespace: "MyOrg",
			wantErr:   true,
		},
		{
			name:      "with hyphen",
			namespace: "my-org",
			wantErr:   false,
		},
		{
			name:      "invalid characters",
			namespace: "my_org",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNamespace(tt.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNamespace() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateEnvironment(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		wantErr     bool
	}{
		{
			name:        "valid environment",
			environment: "prod",
			wantErr:     false,
		},
		{
			name:        "empty environment",
			environment: "",
			wantErr:     false, // Optional field
		},
		{
			name:        "with hyphen",
			environment: "pre-prod",
			wantErr:     false,
		},
		{
			name:        "too long",
			environment: "production",
			wantErr:     true,
		},
		{
			name:        "uppercase",
			environment: "PROD",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEnvironment(tt.environment)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEnvironment() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateCloudProvider(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		wantErr  bool
	}{
		{
			name:     "valid aws",
			provider: "aws",
			wantErr:  false,
		},
		{
			name:     "valid azure",
			provider: "az",
			wantErr:  false,
		},
		{
			name:     "valid gcp",
			provider: "gcp",
			wantErr:  false,
		},
		{
			name:     "empty",
			provider: "",
			wantErr:  false, // Will use default
		},
		{
			name:     "invalid",
			provider: "invalid",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCloudProvider(tt.provider)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCloudProvider() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateEnvironmentType(t *testing.T) {
	tests := []struct {
		name    string
		envType string
		wantErr bool
	}{
		{
			name:    "valid Production",
			envType: "Production",
			wantErr: false,
		},
		{
			name:    "valid Ephemeral",
			envType: "Ephemeral",
			wantErr: false,
		},
		{
			name:    "empty",
			envType: "",
			wantErr: false,
		},
		{
			name:    "invalid",
			envType: "Invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEnvironmentType(tt.envType)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEnvironmentType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateDeletionDate(t *testing.T) {
	tests := []struct {
		name    string
		date    string
		wantErr bool
	}{
		{
			name:    "valid date",
			date:    "2024-12-31",
			wantErr: false,
		},
		{
			name:    "empty date",
			date:    "",
			wantErr: false, // Optional field
		},
		{
			name:    "invalid format",
			date:    "12/31/2024",
			wantErr: true,
		},
		{
			name:    "invalid date",
			date:    "2024-13-45",
			wantErr: true,
		},
		{
			name:    "not a date",
			date:    "not-a-date",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDeletionDate(tt.date)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDeletionDate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{
			name:    "valid email",
			email:   "user@example.com",
			wantErr: false,
		},
		{
			name:    "empty email",
			email:   "",
			wantErr: false, // Optional field
		},
		{
			name:    "complex valid email",
			email:   "user.name+tag@sub.example.com",
			wantErr: false,
		},
		{
			name:    "invalid no @",
			email:   "userexample.com",
			wantErr: true,
		},
		{
			name:    "invalid no domain",
			email:   "user@",
			wantErr: true,
		},
		{
			name:    "invalid no user",
			email:   "@example.com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateEmails(t *testing.T) {
	tests := []struct {
		name    string
		emails  []string
		wantErr bool
	}{
		{
			name:    "all valid",
			emails:  []string{"user1@example.com", "user2@example.com"},
			wantErr: false,
		},
		{
			name:    "empty list",
			emails:  []string{},
			wantErr: false,
		},
		{
			name:    "one invalid",
			emails:  []string{"user1@example.com", "invalid-email"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmails(tt.emails)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmails() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

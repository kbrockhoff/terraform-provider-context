package context

import (
	"fmt"
	"regexp"
	"time"
)

var (
	namespaceRegex   = regexp.MustCompile(`^[a-z][a-z0-9-]{0,6}[a-z0-9]$|^[a-z]$`)
	environmentRegex = regexp.MustCompile(`^[a-z][a-z0-9-]{0,6}[a-z0-9]$|^[a-z]$`)
	dateRegex        = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	emailRegex       = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

// ValidCloudProviders contains the list of valid cloud provider identifiers
var ValidCloudProviders = map[string]bool{
	"dc":  true,
	"aws": true,
	"az":  true,
	"gcp": true,
	"oci": true,
	"ibm": true,
	"do":  true,
	"vul": true,
	"ali": true,
	"cv":  true,
}

// ValidEnvironmentTypes contains the list of valid environment types
var ValidEnvironmentTypes = map[string]bool{
	"":                true, // Allow empty
	"None":            true,
	"Ephemeral":       true,
	"Development":     true,
	"Testing":         true,
	"UAT":             true,
	"Production":      true,
	"MissionCritical": true,
}

// ValidAvailabilityLevels contains the list of valid availability levels
var ValidAvailabilityLevels = map[string]bool{
	"":            true, // Allow empty
	"preemptable": true,
	"spot":        true,
	"standard":    true,
	"dedicated":   true,
	"isolated":    true,
}

// ValidSensitivityLevels contains the list of valid data sensitivity levels
var ValidSensitivityLevels = map[string]bool{
	"":             true, // Allow empty
	"public":       true,
	"internal":     true,
	"confidential": true,
	"restricted":   true,
	"critical":     true,
}

// ValidateNamespace validates namespace format
func ValidateNamespace(namespace string) error {
	if namespace == "" {
		return nil // Optional field
	}

	if len(namespace) > 8 {
		return fmt.Errorf("namespace must be 1-8 characters, got %d: %s", len(namespace), namespace)
	}

	if !namespaceRegex.MatchString(namespace) {
		return fmt.Errorf("namespace must be lowercase alphanumeric with hyphens (1-8 chars): %s", namespace)
	}

	return nil
}

// ValidateEnvironment validates environment format
func ValidateEnvironment(environment string) error {
	if environment == "" {
		return nil // Optional field
	}

	if len(environment) > 8 {
		return fmt.Errorf("environment must be 1-8 characters, got %d: %s", len(environment), environment)
	}

	if !environmentRegex.MatchString(environment) {
		return fmt.Errorf("environment must be lowercase alphanumeric with hyphens (1-8 chars): %s", environment)
	}

	return nil
}

// ValidateCloudProvider validates cloud provider identifier
func ValidateCloudProvider(provider string) error {
	if provider == "" {
		return nil // Will use default
	}

	if !ValidCloudProviders[provider] {
		return fmt.Errorf("invalid cloud provider '%s', must be one of: dc, aws, az, gcp, oci, ibm, do, vul, ali, cv", provider)
	}

	return nil
}

// ValidateEnvironmentType validates environment type
func ValidateEnvironmentType(envType string) error {
	if !ValidEnvironmentTypes[envType] {
		return fmt.Errorf("invalid environment type '%s', must be one of: None, Ephemeral, Development, Testing, UAT, Production, MissionCritical", envType)
	}

	return nil
}

// ValidateAvailability validates availability level
func ValidateAvailability(availability string) error {
	if !ValidAvailabilityLevels[availability] {
		return fmt.Errorf("invalid availability level '%s', must be one of: preemptable, spot, standard, dedicated, isolated", availability)
	}

	return nil
}

// ValidateSensitivity validates data sensitivity level
func ValidateSensitivity(sensitivity string) error {
	if !ValidSensitivityLevels[sensitivity] {
		return fmt.Errorf("invalid sensitivity level '%s', must be one of: public, internal, confidential, restricted, critical", sensitivity)
	}

	return nil
}

// ValidateDeletionDate validates deletion date format
func ValidateDeletionDate(date string) error {
	if date == "" {
		return nil // Optional field
	}

	if !dateRegex.MatchString(date) {
		return fmt.Errorf("deletion date must be in YYYY-MM-DD format: %s", date)
	}

	// Try to parse the date
	_, err := time.Parse("2006-01-02", date)
	if err != nil {
		return fmt.Errorf("invalid deletion date: %s", date)
	}

	return nil
}

// ValidateEmail validates email format
func ValidateEmail(email string) error {
	if email == "" {
		return nil // Optional field
	}

	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format: %s", email)
	}

	return nil
}

// ValidateEmails validates a list of email addresses
func ValidateEmails(emails []string) error {
	for _, email := range emails {
		if err := ValidateEmail(email); err != nil {
			return err
		}
	}
	return nil
}

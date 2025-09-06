package core

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	MaxNamePrefixLength = 24
	MinNamePrefixLength = 2
)

var namePrefixRegex = regexp.MustCompile(`^[a-z][a-z0-9-]{0,22}[a-z0-9]$`)

// NameGenerator handles name prefix generation
type NameGenerator struct {
	Namespace   string
	Name        string
	Environment string
}

// Generate creates a name prefix following Brockhoff standards
func (ng *NameGenerator) Generate() (string, error) {
	// If only name is provided, use it directly
	if ng.Namespace == "" && ng.Environment == "" {
		if ng.Name == "" {
			return "", fmt.Errorf("name is required when namespace and environment are not provided")
		}
		return ng.validateAndTruncate(ng.Name)
	}

	// Build the full name prefix
	parts := []string{}
	if ng.Namespace != "" {
		parts = append(parts, ng.Namespace)
	}
	if ng.Name != "" {
		parts = append(parts, ng.Name)
	}
	if ng.Environment != "" {
		parts = append(parts, ng.Environment)
	}

	if len(parts) == 0 {
		return "", fmt.Errorf("at least one of namespace, name, or environment must be provided")
	}

	namePrefix := strings.Join(parts, "-")
	return ng.validateAndTruncate(namePrefix)
}

// validateAndTruncate ensures the name prefix meets requirements
func (ng *NameGenerator) validateAndTruncate(namePrefix string) (string, error) {
	// Convert to lowercase
	namePrefix = strings.ToLower(namePrefix)

	// Check minimum length
	if len(namePrefix) < MinNamePrefixLength {
		return "", fmt.Errorf("name prefix must be at least %d characters, got: %s", MinNamePrefixLength, namePrefix)
	}

	// Truncate if too long
	if len(namePrefix) > MaxNamePrefixLength {
		namePrefix = ng.intelligentTruncate(namePrefix)
	}

	// Validate against regex
	if !namePrefixRegex.MatchString(namePrefix) {
		return "", fmt.Errorf("name prefix does not match required pattern /^[a-z][a-z0-9-]{0,22}[a-z0-9]$/: %s", namePrefix)
	}

	return namePrefix, nil
}

// intelligentTruncate applies smart truncation to fit within max length
func (ng *NameGenerator) intelligentTruncate(namePrefix string) string {
	if len(namePrefix) <= MaxNamePrefixLength {
		return namePrefix
	}

	// If we have all three components, try to preserve namespace and environment
	if ng.Namespace != "" && ng.Name != "" && ng.Environment != "" {
		// Calculate available space for name
		baseLen := len(ng.Namespace) + len(ng.Environment) + 2 // +2 for hyphens
		availableForName := MaxNamePrefixLength - baseLen

		if availableForName >= 2 { // Minimum 2 chars for name
			truncatedName := ng.Name
			if len(truncatedName) > availableForName {
				truncatedName = truncatedName[:availableForName]
			}
			// Remove trailing hyphen if present
			truncatedName = strings.TrimSuffix(truncatedName, "-")
			return fmt.Sprintf("%s-%s-%s", ng.Namespace, truncatedName, ng.Environment)
		}
	}

	// Simple truncation as fallback
	result := namePrefix[:MaxNamePrefixLength]

	// Ensure we don't end with a hyphen
	for strings.HasSuffix(result, "-") && len(result) > MinNamePrefixLength {
		result = result[:len(result)-1]
	}

	// Ensure last character is alphanumeric
	if len(result) > 0 && !regexp.MustCompile(`[a-z0-9]$`).MatchString(result) {
		result = result[:len(result)-1]
	}

	return result
}

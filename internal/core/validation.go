package core

// This package re-exports from pkg/context for backward compatibility
// New code should import from github.com/kbrockhoff/terraform-provider-context/pkg/context directly

import (
	ctx "github.com/kbrockhoff/terraform-provider-context/pkg/context"
)

// Exported validation constants
var (
	ValidCloudProviders     = ctx.ValidCloudProviders
	ValidEnvironmentTypes   = ctx.ValidEnvironmentTypes
	ValidAvailabilityLevels = ctx.ValidAvailabilityLevels
	ValidSensitivityLevels  = ctx.ValidSensitivityLevels
)

// Validation functions
func ValidateNamespace(namespace string) error {
	return ctx.ValidateNamespace(namespace)
}

func ValidateEnvironment(environment string) error {
	return ctx.ValidateEnvironment(environment)
}

func ValidateCloudProvider(provider string) error {
	return ctx.ValidateCloudProvider(provider)
}

func ValidateEnvironmentType(envType string) error {
	return ctx.ValidateEnvironmentType(envType)
}

func ValidateAvailability(availability string) error {
	return ctx.ValidateAvailability(availability)
}

func ValidateSensitivity(sensitivity string) error {
	return ctx.ValidateSensitivity(sensitivity)
}

func ValidateDeletionDate(date string) error {
	return ctx.ValidateDeletionDate(date)
}

func ValidateEmail(email string) error {
	return ctx.ValidateEmail(email)
}

func ValidateEmails(emails []string) error {
	return ctx.ValidateEmails(emails)
}

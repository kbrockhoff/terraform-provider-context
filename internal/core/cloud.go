package core

// This package re-exports from pkg/context for backward compatibility
// New code should import from github.com/kbrockhoff/terraform-provider-context/pkg/context directly

import (
	ctx "github.com/kbrockhoff/terraform-provider-context/pkg/context"
)

// CloudProvider interface defines cloud-specific tag formatting rules
type CloudProvider = ctx.CloudProvider

// Cloud provider implementations
type (
	AWSProvider     = ctx.AWSProvider
	AzureProvider   = ctx.AzureProvider
	GCPProvider     = ctx.GCPProvider
	DefaultProvider = ctx.DefaultProvider
)

// GetCloudProvider returns the appropriate CloudProvider implementation
func GetCloudProvider(provider string) CloudProvider {
	return ctx.GetCloudProvider(provider)
}

package core

// This package re-exports from pkg/context for backward compatibility
// New code should import from github.com/kbrockhoff/terraform-provider-context/pkg/context directly

import (
	ctx "github.com/kbrockhoff/terraform-provider-context/pkg/context"
)

// GitInfo contains repository information
type GitInfo = ctx.GitInfo

// GetGitInfo retrieves git repository information with caching
func GetGitInfo() (*GitInfo, error) {
	return ctx.GetGitInfo()
}

// ClearGitCache clears the git information cache
func ClearGitCache() {
	ctx.ClearGitCache()
}

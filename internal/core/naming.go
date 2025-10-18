package core

// This package re-exports from pkg/context for backward compatibility
// New code should import from github.com/kbrockhoff/terraform-provider-context/pkg/context directly

import (
	ctx "github.com/kbrockhoff/terraform-provider-context/pkg/context"
)

// Name prefix constants
const (
	MaxNamePrefixLength = ctx.MaxNamePrefixLength
	MinNamePrefixLength = ctx.MinNamePrefixLength
)

// NameGenerator handles name prefix generation
type NameGenerator = ctx.NameGenerator

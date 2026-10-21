package oidc

import (
	"strings"

	"github.com/gobwas/glob"
)

// MaxPatternLength bounds a single claim condition pattern.
const MaxPatternLength = 512

// globMetaChars is the single set of characters that make a pattern a glob
// rather than an exact string.
const globMetaChars = "*?[]{}\\"

// IsGlob reports whether pattern contains glob metacharacters.
func IsGlob(pattern string) bool {
	return strings.ContainsAny(pattern, globMetaChars)
}

// Compile compiles a glob with `:` as a separator so wildcards cannot cross
// `:` segments (the delimiter in GitHub Actions `sub` claims like
// `repo:org/repo:ref:refs/heads/main`).
func Compile(pattern string) (glob.Glob, error) {
	return glob.Compile(pattern, ':')
}

// MatchPattern matches a value against a pattern: exact comparison unless the
// pattern is a glob.
func MatchPattern(pattern, value string) bool {
	if !IsGlob(pattern) {
		return pattern == value
	}

	g, err := Compile(pattern)
	if err != nil {
		return false
	}

	return g.Match(value)
}

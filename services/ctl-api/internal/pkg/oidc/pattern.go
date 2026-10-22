package oidc

import (
	"strings"

	"github.com/gobwas/glob"
)

const MaxPatternLength = 512

const globMetaChars = "*?[{"

func IsGlob(pattern string) bool {
	return strings.ContainsAny(pattern, globMetaChars)
}

// Compile compiles a glob with `:` as a separator so wildcards cannot cross
// `:` segments (the delimiter in GitHub Actions `sub` claims like
// `repo:org/repo:ref:refs/heads/main`).
func Compile(pattern string) (glob.Glob, error) {
	return glob.Compile(pattern, ':')
}

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

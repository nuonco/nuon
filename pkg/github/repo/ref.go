package github

import (
	"regexp"
	"strings"
	"unicode"
)

var commitSHARegexp = regexp.MustCompile(`^(?:[0-9a-f]{40}|[0-9a-f]{64})$`)

func IsCommitSHA(s string) bool {
	return commitSHARegexp.MatchString(s)
}

func IsZeroCommitSHA(s string) bool {
	return s != "" && strings.Trim(s, "0") == ""
}

// IsSafeGitRef rejects refs that could change the GitHub API URL they are
// interpolated into.
func IsSafeGitRef(ref string) bool {
	if ref == "" || strings.HasPrefix(ref, "/") || strings.HasSuffix(ref, "/") ||
		strings.Contains(ref, "..") || strings.ContainsAny(ref, `\?#%`) {
		return false
	}
	return !strings.ContainsFunc(ref, func(r rune) bool {
		return unicode.IsControl(r) || unicode.IsSpace(r)
	})
}

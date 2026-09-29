package ignorechanges

import (
	"fmt"
	"regexp"
)

type Decision struct {
	Ignored bool

	Reason string
}

// why: Evaluate reports whether every changed path matches pattern.
//
// The pattern is unanchored, matching Go's regexp semantics, so `docs/` covers
// any path containing that substring; callers wanting a prefix write `^docs/`.
//
// An empty pattern disables the check. An empty path list is ignorable: there
// is nothing to build. A pattern that does not compile returns an error and a
// non-ignored decision, so a bad config fails open into a normal run rather
// than silently swallowing every push.
func Evaluate(pattern string, paths []string) (Decision, error) {
	if pattern == "" {
		return Decision{}, nil
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return Decision{}, fmt.Errorf("unable to compile ignore_changes_regex %q: %w", pattern, err)
	}

	if len(paths) == 0 {
		return Decision{
			Ignored: true,
			Reason:  fmt.Sprintf("no changed files; ignored by ignore_changes_regex %q", pattern),
		}, nil
	}

	for _, path := range paths {
		if !re.MatchString(path) {
			return Decision{}, nil
		}
	}

	return Decision{
		Ignored: true,
		Reason:  fmt.Sprintf("all %d changed files match ignore_changes_regex %q", len(paths), pattern),
	}, nil
}

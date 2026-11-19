// Package ignorechanges decides whether an app branch run's changed files are
// all covered by the branch's ignore_changes_regex, and stops the run's
// workflow when they are.
package ignorechanges

import (
	"fmt"
	"regexp"

	"github.com/nuonco/nuon/pkg/config"
)

const (
	regexStatusDescription = "Ignored by app branch changes regex"
	pathsStatusDescription = "Ignored by app branch run paths"
)

// Decision is the outcome of evaluating a changed file set against a pattern.
type Decision struct {
	Ignored bool

	// Reason is user-facing: it lands on the run's error message, the stopped
	// step's description and the commit status, so it says which pattern
	// matched and how many files it covered.
	Reason string
}

// EvaluateRun applies the run path filter first, then ignore_changes_regex, and
// returns the commit status description that goes with the decision.
func EvaluateRun(filter *config.RunPathFilter, regex string, paths []string) (Decision, string, error) {
	if d := filter.Match(paths); !d.Run {
		return Decision{Ignored: true, Reason: d.Reason}, pathsStatusDescription, nil
	}
	decision, err := Evaluate(regex, paths)
	return decision, regexStatusDescription, err
}

// Evaluate reports whether every changed path matches pattern.
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

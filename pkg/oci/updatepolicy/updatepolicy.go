// Package updatepolicy implements the semver-based tag selection used by
// image-type component builds when the user has set a `update_policy`
// constraint.
//
// The runner is the single owner of this code path. ctl-api stores and
// passes the policy string through unchanged; the runner lists tags from
// the user's source registry, filters/sorts them with this package, and
// resolves the selected tag to a manifest digest the same way it would for
// a literal tag.
//
// The package is intentionally tiny and pure: it takes a list of tag
// strings and a Masterminds-compatible semver constraint, and returns the
// highest tag that parses as a valid semver and satisfies the constraint.
// Tags that do not parse as semver (e.g. "latest", "stable", "main") are
// silently skipped. A constraint that matches zero parseable tags returns
// [ErrNoMatchingTag] so the caller can fail the build loudly.
package updatepolicy

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// ErrNoMatchingTag is returned when zero tags in the input list parse as
// semver and satisfy the constraint. Callers should treat this as a build
// failure: the user's constraint doesn't match anything in the registry.
var ErrNoMatchingTag = errors.New("no tags in the source registry match the update_policy constraint")

// Validate reports whether constraint is a syntactically valid
// Masterminds-semver constraint (e.g. "~1.25.0", "^2", ">=1.0.0,<2.0.0").
// Used at the API boundary to reject malformed user input before we ever
// reach the runner.
func Validate(constraint string) error {
	if strings.TrimSpace(constraint) == "" {
		return errors.New("update_policy must not be empty")
	}
	if _, err := semver.NewConstraint(constraint); err != nil {
		return fmt.Errorf("invalid semver constraint %q: %w", constraint, err)
	}
	return nil
}

// SelectHighestMatching returns the tag from tags that:
//
//   - parses as a valid semver version, and
//   - satisfies constraint (a Masterminds-compatible constraint string),
//
// preferring the highest-precedence semver among the matches.
//
// Tags that do not parse as semver are silently skipped (e.g. "latest",
// "stable", "main"). Returns [ErrNoMatchingTag] when zero tags match.
//
// The returned string is the original tag as it appeared in the input
// list, preserving any leading "v" or other formatting the registry used
// (e.g. input "v1.25.5" → returned "v1.25.5", not "1.25.5").
func SelectHighestMatching(tags []string, constraint string) (string, error) {
	c, err := semver.NewConstraint(constraint)
	if err != nil {
		return "", fmt.Errorf("invalid semver constraint %q: %w", constraint, err)
	}

	type candidate struct {
		original string
		version  *semver.Version
	}

	candidates := make([]candidate, 0, len(tags))
	for _, t := range tags {
		v, perr := semver.NewVersion(t)
		if perr != nil {
			// Not a semver tag (e.g. "latest", "stable", branch refs) —
			// silently ignored per the package contract.
			continue
		}
		if !c.Check(v) {
			continue
		}
		candidates = append(candidates, candidate{original: t, version: v})
	}

	if len(candidates) == 0 {
		return "", ErrNoMatchingTag
	}

	// Highest version wins. Sort.Stable to keep deterministic ordering when
	// two tag strings parse to the same semver (e.g. "1.0" and "1.0.0").
	sort.SliceStable(candidates, func(i, j int) bool {
		return candidates[i].version.LessThan(candidates[j].version)
	})

	return candidates[len(candidates)-1].original, nil
}

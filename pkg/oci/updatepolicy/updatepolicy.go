package updatepolicy

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
)

var ErrNoMatchingTag = errors.New("no tags in the source registry match the update_policy constraint")

func Validate(constraint string) error {
	if strings.TrimSpace(constraint) == "" {
		return errors.New("update_policy must not be empty")
	}
	if _, err := semver.NewConstraint(constraint); err != nil {
		return fmt.Errorf("invalid semver constraint %q: %w", constraint, err)
	}
	return nil
}

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

	sort.SliceStable(candidates, func(i, j int) bool {
		return candidates[i].version.LessThan(candidates[j].version)
	})

	return candidates[len(candidates)-1].original, nil
}

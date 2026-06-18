package diff

import (
	"fmt"
	"sort"
	"strings"
)

// MapDiff diffs two map[string]string values, returning a parent node with
// per-key children. Returns nil if both maps are empty/nil.
func MapDiff(key string, old, new map[string]string) *Diff {
	if len(old) == 0 && len(new) == 0 {
		return nil
	}

	allKeys := make(map[string]struct{})
	for k := range old {
		allKeys[k] = struct{}{}
	}
	for k := range new {
		allKeys[k] = struct{}{}
	}

	sorted := make([]string, 0, len(allKeys))
	for k := range allKeys {
		sorted = append(sorted, k)
	}
	sort.Strings(sorted)

	children := make([]*Diff, 0, len(sorted))
	for _, k := range sorted {
		children = append(children, NewDiff(
			WithKey(k),
			WithStringDiff(old[k], new[k]),
		))
	}

	return NewDiff(
		WithKey(key),
		WithChildren(children...),
	)
}

// WithBoolDiff creates a DiffOption that compares two bool values.
func WithBoolDiff(old, new bool) DiffOption {
	return WithStringDiff(fmt.Sprintf("%t", old), fmt.Sprintf("%t", new))
}

// WithOptionalStringDiff creates a DiffOption that compares two *string values,
// treating nil as empty string.
func WithOptionalStringDiff(old, new *string) DiffOption {
	oldStr := ""
	if old != nil {
		oldStr = *old
	}
	newStr := ""
	if new != nil {
		newStr = *new
	}
	return WithStringDiff(oldStr, newStr)
}

// WithOptionalBoolDiff creates a DiffOption that compares two *bool values,
// treating nil as false.
func WithOptionalBoolDiff(old, new *bool) DiffOption {
	oldVal := false
	if old != nil {
		oldVal = *old
	}
	newVal := false
	if new != nil {
		newVal = *new
	}
	return WithBoolDiff(oldVal, newVal)
}

// WithStringSliceDiff creates a DiffOption that compares two string slices
// by sorting and joining them.
func WithStringSliceDiff(old, new []string) DiffOption {
	oldSorted := make([]string, len(old))
	copy(oldSorted, old)
	sort.Strings(oldSorted)

	newSorted := make([]string, len(new))
	copy(newSorted, new)
	sort.Strings(newSorted)

	return WithStringDiff(strings.Join(oldSorted, ", "), strings.Join(newSorted, ", "))
}

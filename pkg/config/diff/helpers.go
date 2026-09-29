package diff

import (
	"fmt"
	"sort"
	"strings"
)

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

func WithContentDiff(old, new string) DiffOption {
	return func(dt *Diff) {
		op := OpChange
		label := "modified"
		switch {
		case old == new:
			op = OpNoop
			label = "unchanged"
		case old == "":
			op = OpAdd
			label = "added"
		case new == "":
			op = OpRemove
			label = "removed"
		}
		dt.Diff = &DiffKey{
			Op:     op,
			Diff:   label,
			Before: old,
			After:  new,
		}
	}
}

func WithBoolDiff(old, new bool) DiffOption {
	return WithStringDiff(fmt.Sprintf("%t", old), fmt.Sprintf("%t", new))
}

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

func WithStringSliceDiff(old, new []string) DiffOption {
	oldSorted := make([]string, len(old))
	copy(oldSorted, old)
	sort.Strings(oldSorted)

	newSorted := make([]string, len(new))
	copy(newSorted, new)
	sort.Strings(newSorted)

	return WithStringDiff(strings.Join(oldSorted, ", "), strings.Join(newSorted, ", "))
}

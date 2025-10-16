package diff

import (
	"regexp"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// diffReporter implements the cmp.Reporter interface to capture structured diff information
type diffReporter struct {
	entries map[string]*DiffEntry
	paths   []cmp.Path
}

func (r *diffReporter) PushStep(ps cmp.PathStep) {
	r.paths = append(r.paths, append(r.paths[len(r.paths)-1], ps))
}

func (r *diffReporter) PopStep() {
	r.paths = r.paths[:len(r.paths)-1]
}

func (r *diffReporter) Report(rs cmp.Result) {
	if !rs.Equal() {
		// Get the current path as a string
		currentPath := r.paths[len(r.paths)-1]
		pathStr := pathToString(currentPath)

		// Get values from both sides
		vx, vy := currentPath.Last().Values()

		// Create or update diff entry
		entry, exists := r.entries[pathStr]
		if !exists {
			entry = &DiffEntry{
				Path: pathStr,
			}
			r.entries[pathStr] = entry
		}

		// Set original and applied values
		if vx.IsValid() {
			entry.Original = vx.Interface()
		}
		if vy.IsValid() {
			entry.Applied = vy.Interface()
		}

		// Determine the type of change
		if !vx.IsValid() && vy.IsValid() {
			entry.Type = EntryAdded
		} else if vx.IsValid() && !vy.IsValid() {
			entry.Type = EntryRemoved
		} else {
			entry.Type = EntryModified
		}
	}
}

// pathToString converts a cmp.Path to a string representation
// similar to what JSON path would look like
func pathToString(p cmp.Path) string {
	if len(p) == 0 {
		return ""
	}

	var parts []string
	for _, step := range p {
		switch step := step.(type) {
		case cmp.MapIndex:
			// For map keys, add the key
			key := step.Key().String()
			// Trim quotes from strings
			key = strings.Trim(key, `"`)
			parts = append(parts, key)
		case cmp.SliceIndex:
			// For slices, use array-style indexing
			parts[len(parts)-1] = parts[len(parts)-1] + "[" + step.String() + "]"
		case cmp.StructField:
			// For struct fields, add the field name
			parts = append(parts, step.String())
		}
	}

	return strings.Join(parts, ".")
}

// DetectChanges compares original and applied objects and returns structured diff entries
func DetectChanges(original, applied map[string]interface{}, ignoreFields []string) ([]DiffEntry, bool) {
	// Create a path filter function for cmp.Diff
	pathFilter := func(p cmp.Path) bool {
		if len(p) == 0 {
			return false
		}

		// Convert the path to our string format for more precise matching
		pathStr := pathToString(p)

		// Check if this path should be ignored
		for _, ignore := range ignoreFields {
			// We need exact matches or prefix matches ending with a dot
			if pathStr == ignore ||
				strings.HasPrefix(pathStr, ignore+".") {
				return true
			}
		}
		return false
	}

	// First generate a raw text diff for debugging and payload
	rawDiff := cmp.Diff(original, applied,
		cmpopts.IgnoreMapEntries(func(k string, v interface{}) bool {
			return k == "status"
		}),
		cmp.FilterPath(pathFilter, cmp.Ignore()),
	)

	// Create a reporter to capture diff information
	r := &diffReporter{
		entries: make(map[string]*DiffEntry),
		paths:   []cmp.Path{{}}, // Initialize with empty path
	}

	// Use cmp.Equal with reporter to capture differences
	equal := cmp.Equal(original, applied,
		cmpopts.IgnoreMapEntries(func(k string, v interface{}) bool {
			return k == "status"
		}),
		cmp.FilterPath(pathFilter, cmp.Ignore()),
		cmp.Reporter(r),
	)

	// Convert map to slice
	entries := make([]DiffEntry, 0, len(r.entries))
	for _, entry := range r.entries {
		// Additional check to filter out any ignored paths that might have slipped through
		shouldInclude := true
		for _, ignore := range ignoreFields {
			if entry.Path == ignore || strings.HasPrefix(entry.Path, ignore+".") {
				shouldInclude = false
				break
			}
		}

		if shouldInclude {
			entries = append(entries, *entry)
		}
	}

	// If we have changes, add the raw diff as payload to the first entry for debugging
	if len(entries) > 0 && rawDiff != "" {
		entries[0].Payload = rawDiff
	}

	return entries, !equal
}

// ParseRawResourceName parses a resource name string in the format "namespace, name, kind (apiBase)"
func ParseRawResourceName(s string) (namespace, name, kind, apiPath string) {
	if s == "" {
		return
	}

	// Use regex to parse the format: "namespace, name, kind (apiBase)"
	re := regexp.MustCompile(`^([^,]+),\s*([^,]+),\s*([^(]+)\s*\(([^)]+)\)$`)
	matches := re.FindStringSubmatch(s)

	if len(matches) == 5 {
		namespace = strings.TrimSpace(matches[1])
		name = strings.TrimSpace(matches[2])
		kind = strings.TrimSpace(matches[3])
		apiPath = strings.TrimSpace(matches[4])
	}
	return
}

package diff

import (
	"regexp"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

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
		currentPath := r.paths[len(r.paths)-1]
		pathStr := pathToString(currentPath)

		vx, vy := currentPath.Last().Values()

		entry, exists := r.entries[pathStr]
		if !exists {
			entry = &DiffEntry{
				Path: pathStr,
			}
			r.entries[pathStr] = entry
		}

		if vx.IsValid() {
			entry.Original = vx.Interface()
		}
		if vy.IsValid() {
			entry.Applied = vy.Interface()
		}

		if !vx.IsValid() && vy.IsValid() {
			entry.Type = EntryAdded
		} else if vx.IsValid() && !vy.IsValid() {
			entry.Type = EntryRemoved
		} else {
			entry.Type = EntryModified
		}
	}
}

func pathToString(p cmp.Path) string {
	if len(p) == 0 {
		return ""
	}

	var parts []string
	for _, step := range p {
		switch step := step.(type) {
		case cmp.MapIndex:
			key := step.Key().String()
			key = strings.Trim(key, `"`)
			parts = append(parts, key)
		case cmp.SliceIndex:
			parts[len(parts)-1] = parts[len(parts)-1] + "[" + step.String() + "]"
		case cmp.StructField:
			parts = append(parts, step.String())
		}
	}

	return strings.Join(parts, ".")
}

func DetectChanges(original, applied map[string]interface{}, ignoreFields []string) ([]DiffEntry, bool) {
	pathFilter := func(p cmp.Path) bool {
		if len(p) == 0 {
			return false
		}

		pathStr := pathToString(p)

		for _, ignore := range ignoreFields {
			if pathStr == ignore ||
				strings.HasPrefix(pathStr, ignore+".") {
				return true
			}
		}
		return false
	}

	rawDiff := cmp.Diff(original, applied,
		cmpopts.IgnoreMapEntries(func(k string, v interface{}) bool {
			return k == "status"
		}),
		cmp.FilterPath(pathFilter, cmp.Ignore()),
	)

	r := &diffReporter{
		entries: make(map[string]*DiffEntry),
		paths:   []cmp.Path{{}},
	}

	equal := cmp.Equal(original, applied,
		cmpopts.IgnoreMapEntries(func(k string, v interface{}) bool {
			return k == "status"
		}),
		cmp.FilterPath(pathFilter, cmp.Ignore()),
		cmp.Reporter(r),
	)

	entries := make([]DiffEntry, 0, len(r.entries))
	for _, entry := range r.entries {
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

	if len(entries) > 0 && rawDiff != "" {
		entries[0].Payload = rawDiff
	}

	return entries, !equal
}

func ParseRawResourceName(s string) (namespace, name, kind, apiPath string) {
	if s == "" {
		return
	}

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

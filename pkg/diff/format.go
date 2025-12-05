package diff

import (
	"fmt"
	"sort"
	"strings"

	"sigs.k8s.io/yaml"
)

// FormatToLineByLine converts a ResourceDiff object to use line-by-line entries
// This provides a more readable diff for UI presentation
func FormatToLineByLine(resourceDiff ResourceDiff) ResourceDiff {
	result := resourceDiff
	result.Entries = []DiffEntry{}

	// Only proceed if we have entries to format
	if len(resourceDiff.Entries) == 0 {
		return result
	}

	// For completely new resources or deleted resources, we want to show the entire YAML
	if resourceDiff.Type == EntryAdded || resourceDiff.Type == EntryRemoved {
		var obj interface{}
		if resourceDiff.Type == EntryAdded && len(resourceDiff.Entries) > 0 && resourceDiff.Entries[0].Applied != nil {
			obj = resourceDiff.Entries[0].Applied
		} else if resourceDiff.Type == EntryRemoved && len(resourceDiff.Entries) > 0 && resourceDiff.Entries[0].Original != nil {
			obj = resourceDiff.Entries[0].Original
		}

		if obj != nil {
			// Add a header entry
			result.Entries = append(result.Entries, DiffEntry{
				Type: resourceDiff.Type,
			})

			// Convert the object to YAML lines
			yamlLines, err := objectToYAMLLines(obj)
			if err == nil {
				// Add each line as a separate entry
				for _, line := range yamlLines {
					result.Entries = append(result.Entries, DiffEntry{
						Type:    resourceDiff.Type,
						Payload: line,
					})
				}
			}
		}
		return result
	}

	// For modified resources, we need to show the specific changes
	// Group entries by path prefix to organize them better
	pathGroups := groupEntriesByPathPrefix(resourceDiff.Entries)

	// Sort the path groups to maintain consistent order
	sortedPaths := make([]string, 0, len(pathGroups))
	for path := range pathGroups {
		sortedPaths = append(sortedPaths, path)
	}
	sort.Strings(sortedPaths)

	// Process each group of entries
	for _, pathPrefix := range sortedPaths {
		entries := pathGroups[pathPrefix]

		// For each entry in this group, convert to YAML lines
		for _, entry := range entries {
			// Skip entries with no changes
			if entry.Type == EntryUnchanged {
				continue
			}

			// Handle simple scalar values directly
			if isSimpleValue(entry.Original) && isSimpleValue(entry.Applied) {
				// For simple values, just add a single entry for each side
				if entry.Original != nil {
					result.Entries = append(result.Entries, DiffEntry{
						Type:    EntryRemoved,
						Payload: fmt.Sprintf("%v", entry.Original),
						Path:    entry.Path,
					})
				}
				if entry.Applied != nil {
					result.Entries = append(result.Entries, DiffEntry{
						Type:    EntryAdded,
						Payload: fmt.Sprintf("%v", entry.Applied),
						Path:    entry.Path,
					})
				}
				continue
			}

			// Handle more complex objects with YAML conversion
			if entry.Original != nil && entry.Applied != nil {
				// Handle modified entries
				originalYAML, err := objectToYAMLLines(entry.Original)
				if err != nil || len(originalYAML) == 0 {
					// Fallback for objects that don't convert well to YAML
					result.Entries = append(result.Entries, DiffEntry{
						Type:    EntryRemoved,
						Payload: fmt.Sprintf("%v", entry.Original),
						Path:    entry.Path,
					})
				} else {
					for _, line := range originalYAML {
						result.Entries = append(result.Entries, DiffEntry{
							Type:    EntryRemoved,
							Payload: line,
							Path:    entry.Path,
						})
					}
				}

				appliedYAML, err := objectToYAMLLines(entry.Applied)
				if err != nil || len(appliedYAML) == 0 {
					// Fallback for objects that don't convert well to YAML
					result.Entries = append(result.Entries, DiffEntry{
						Type:    EntryAdded,
						Payload: fmt.Sprintf("%v", entry.Applied),
						Path:    entry.Path,
					})
				} else {
					for _, line := range appliedYAML {
						result.Entries = append(result.Entries, DiffEntry{
							Type:    EntryAdded,
							Payload: line,
							Path:    entry.Path,
						})
					}
				}
			} else if entry.Original != nil {
				// Handle removed entries
				originalYAML, err := objectToYAMLLines(entry.Original)
				if err != nil || len(originalYAML) == 0 {
					result.Entries = append(result.Entries, DiffEntry{
						Type:    EntryRemoved,
						Payload: fmt.Sprintf("%v", entry.Original),
						Path:    entry.Path,
					})
				} else {
					for _, line := range originalYAML {
						result.Entries = append(result.Entries, DiffEntry{
							Type:    EntryRemoved,
							Payload: line,
							Path:    entry.Path,
						})
					}
				}
			} else if entry.Applied != nil {
				// Handle added entries
				appliedYAML, err := objectToYAMLLines(entry.Applied)
				if err != nil || len(appliedYAML) == 0 {
					result.Entries = append(result.Entries, DiffEntry{
						Type:    EntryAdded,
						Payload: fmt.Sprintf("%v", entry.Applied),
						Path:    entry.Path,
					})
				} else {
					for _, line := range appliedYAML {
						result.Entries = append(result.Entries, DiffEntry{
							Type:    EntryAdded,
							Payload: line,
							Path:    entry.Path,
						})
					}
				}
			}
		}
	}

	// If we have a raw diff payload in the first entry, use it as a fallback
	if len(result.Entries) == 0 && len(resourceDiff.Entries) > 0 && resourceDiff.Entries[0].Payload != "" {
		// Parse the raw diff and convert to line-by-line entries
		lines := strings.Split(resourceDiff.Entries[0].Payload, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			if strings.HasPrefix(line, "+") {
				result.Entries = append(result.Entries, DiffEntry{
					Type:    EntryAdded,
					Payload: strings.TrimPrefix(line, "+ "),
				})
			} else if strings.HasPrefix(line, "-") {
				result.Entries = append(result.Entries, DiffEntry{
					Type:    EntryRemoved,
					Payload: strings.TrimPrefix(line, "- "),
				})
			} else {
				// Skip unchanged lines in the raw diff to focus on changes
				// This makes the output cleaner for the line-by-line format
				continue
			}
		}
	}

	return result
}

// isSimpleValue checks if a value is a simple scalar (not a map/slice/complex object)
func isSimpleValue(v interface{}) bool {
	if v == nil {
		return true
	}

	switch v.(type) {
	case string, bool, int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64, float32, float64:
		return true
	default:
		// Check for map or slice
		_, isMap := v.(map[string]interface{})
		_, isSlice := v.([]interface{})
		return !isMap && !isSlice
	}
}

// groupEntriesByPathPrefix groups diff entries by their path prefix
// This helps organize related changes together
func groupEntriesByPathPrefix(entries []DiffEntry) map[string][]DiffEntry {
	result := make(map[string][]DiffEntry)

	for _, entry := range entries {
		// Extract the top-level path component
		pathParts := strings.Split(entry.Path, ".")
		prefix := ""
		if len(pathParts) > 0 {
			prefix = pathParts[0]
		}

		result[prefix] = append(result[prefix], entry)
	}

	return result
}

// objectToYAMLLines converts an object to YAML and returns it as individual lines
func objectToYAMLLines(obj interface{}) ([]string, error) {
	if obj == nil {
		return nil, fmt.Errorf("cannot convert nil object to YAML")
	}

	// For simple scalar values, just return a single line
	if isSimpleValue(obj) {
		return []string{fmt.Sprintf("%v", obj)}, nil
	}

	// Convert object to YAML
	yamlBytes, err := yaml.Marshal(obj)
	if err != nil {
		return nil, err
	}

	// Split into lines and remove empty lines
	lines := strings.Split(string(yamlBytes), "\n")
	var result []string
	for _, line := range lines {
		if line != "" {
			result = append(result, line)
		}
	}

	return result, nil
}

// FormatResourceDiffs converts all ResourceDiff objects in a slice to use line-by-line entries
func FormatResourceDiffs(diffs []ResourceDiff) []ResourceDiff {
	result := make([]ResourceDiff, len(diffs))

	for i, diff := range diffs {
		result[i] = FormatToLineByLine(diff)
	}

	return result
}

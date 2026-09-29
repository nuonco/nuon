package diff

import (
	"fmt"
	"sort"
	"strings"

	"sigs.k8s.io/yaml"
)

func FormatToLineByLine(resourceDiff ResourceDiff) ResourceDiff {
	result := resourceDiff
	result.Entries = []DiffEntry{}

	if len(resourceDiff.Entries) == 0 {
		return result
	}

	if resourceDiff.Type == EntryAdded || resourceDiff.Type == EntryRemoved {
		var obj interface{}
		if resourceDiff.Type == EntryAdded && len(resourceDiff.Entries) > 0 && resourceDiff.Entries[0].Applied != nil {
			obj = resourceDiff.Entries[0].Applied
		} else if resourceDiff.Type == EntryRemoved && len(resourceDiff.Entries) > 0 && resourceDiff.Entries[0].Original != nil {
			obj = resourceDiff.Entries[0].Original
		}

		if obj != nil {
			result.Entries = append(result.Entries, DiffEntry{
				Type: resourceDiff.Type,
			})

			yamlLines, err := objectToYAMLLines(obj)
			if err == nil {
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

	pathGroups := groupEntriesByPathPrefix(resourceDiff.Entries)

	sortedPaths := make([]string, 0, len(pathGroups))
	for path := range pathGroups {
		sortedPaths = append(sortedPaths, path)
	}
	sort.Strings(sortedPaths)

	for _, pathPrefix := range sortedPaths {
		entries := pathGroups[pathPrefix]

		for _, entry := range entries {
			if entry.Type == EntryUnchanged {
				continue
			}

			if isSimpleValue(entry.Original) && isSimpleValue(entry.Applied) {
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

			if entry.Original != nil && entry.Applied != nil {
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
			} else if entry.Original != nil {
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

	if len(result.Entries) == 0 && len(resourceDiff.Entries) > 0 && resourceDiff.Entries[0].Payload != "" {
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
				continue
			}
		}
	}

	return result
}

func isSimpleValue(v interface{}) bool {
	if v == nil {
		return true
	}

	switch v.(type) {
	case string, bool, int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64, float32, float64:
		return true
	default:
		_, isMap := v.(map[string]interface{})
		_, isSlice := v.([]interface{})
		return !isMap && !isSlice
	}
}

func groupEntriesByPathPrefix(entries []DiffEntry) map[string][]DiffEntry {
	result := make(map[string][]DiffEntry)

	for _, entry := range entries {
		pathParts := strings.Split(entry.Path, ".")
		prefix := ""
		if len(pathParts) > 0 {
			prefix = pathParts[0]
		}

		result[prefix] = append(result[prefix], entry)
	}

	return result
}

func objectToYAMLLines(obj interface{}) ([]string, error) {
	if obj == nil {
		return nil, fmt.Errorf("cannot convert nil object to YAML")
	}

	if isSimpleValue(obj) {
		return []string{fmt.Sprintf("%v", obj)}, nil
	}

	yamlBytes, err := yaml.Marshal(obj)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(yamlBytes), "\n")
	var result []string
	for _, line := range lines {
		if line != "" {
			result = append(result, line)
		}
	}

	return result, nil
}

func FormatResourceDiffs(diffs []ResourceDiff) []ResourceDiff {
	result := make([]ResourceDiff, len(diffs))

	for i, diff := range diffs {
		result[i] = FormatToLineByLine(diff)
	}

	return result
}

package toml

import (
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// ParseStrict attempts to parse TOML using strict mode
// Returns a TomlDocument and an error if parsing fails
func ParseStrict(text string) (*TomlDocument, error) {
	var data map[string]any
	err := toml.Unmarshal([]byte(text), &data)
	if err != nil {
		return nil, err
	}

	doc := NewTomlDocument()
	extractTablesAndKeys(data, []string{}, doc)
	return doc, nil
}

// extractTablesAndKeys recursively extracts tables and keys from parsed TOML data
func extractTablesAndKeys(data map[string]any, parentPath []string, doc *TomlDocument) {
	for key, value := range data {
		currentPath := append([]string{}, parentPath...)
		currentPath = append(currentPath, key)
		pathStr := strings.Join(currentPath, ".")

		// Check if value is a map (nested table)
		if valueMap, ok := value.(map[string]any); ok {
			// Add as table
			doc.Tables = append(doc.Tables, Table{
				Name: pathStr,
				Path: currentPath,
				Range: Range{
					Start: Position{Line: 0, Character: 0},
					End:   Position{Line: 0, Character: 0},
				},
			})

			// Set current table to the last one added
			doc.CurrentTable = pathStr

			// Recursively extract nested keys
			extractTablesAndKeys(valueMap, currentPath, doc)
		} else if valueSlice, ok := value.([]any); ok {
			// Check if it's an array of tables
			allMaps := true
			for _, item := range valueSlice {
				if _, isMap := item.(map[string]any); !isMap {
					allMaps = false
					break
				}
			}

			if allMaps {
				// Array of tables
				doc.Tables = append(doc.Tables, Table{
					Name: pathStr,
					Path: currentPath,
					Range: Range{
						Start: Position{Line: 0, Character: 0},
						End:   Position{Line: 0, Character: 0},
					},
				})

				for _, item := range valueSlice {
					if itemMap, ok := item.(map[string]any); ok {
						extractTablesAndKeys(itemMap, currentPath, doc)
					}
				}
			} else {
				// Regular array value
				doc.Keys = append(doc.Keys, Key{
					Name:  key,
					Path:  currentPath,
					Value: value,
					Range: Range{
						Start: Position{Line: 0, Character: 0},
						End:   Position{Line: 0, Character: 0},
					},
				})
				doc.Values[pathStr] = value
			}
		} else {
			// Regular key-value pair
			doc.Keys = append(doc.Keys, Key{
				Name:  key,
				Path:  currentPath,
				Value: value,
				Range: Range{
					Start: Position{Line: 0, Character: 0},
					End:   Position{Line: 0, Character: 0},
				},
			})
			doc.Values[pathStr] = value
		}
	}
}

package toml

import (
	"strings"

	"github.com/pelletier/go-toml/v2"
)

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

func extractTablesAndKeys(data map[string]any, parentPath []string, doc *TomlDocument) {
	for key, value := range data {
		currentPath := append([]string{}, parentPath...)
		currentPath = append(currentPath, key)
		pathStr := strings.Join(currentPath, ".")

		if valueMap, ok := value.(map[string]any); ok {
			doc.Tables = append(doc.Tables, Table{
				Name: pathStr,
				Path: currentPath,
				Range: Range{
					Start: Position{Line: 0, Character: 0},
					End:   Position{Line: 0, Character: 0},
				},
			})

			doc.CurrentTable = pathStr

			extractTablesAndKeys(valueMap, currentPath, doc)
		} else if valueSlice, ok := value.([]any); ok {
			allMaps := len(valueSlice) > 0
			for _, item := range valueSlice {
				if _, isMap := item.(map[string]any); !isMap {
					allMaps = false
					break
				}
			}

			if allMaps {
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

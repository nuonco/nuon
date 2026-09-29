package toml

import (
	"regexp"
	"strings"
)

var (
	tableHeaderRegex      = regexp.MustCompile(`^\s*\[\s*([A-Za-z0-9_.-]+)`)
	keyValueRegex         = regexp.MustCompile(`^\s*([A-Za-z0-9_.-]+)\s*=?`)
	commentRegex          = regexp.MustCompile(`^\s*#`)
	arrayTableHeaderRegex = regexp.MustCompile(`^\s*\[\[\s*([A-Za-z0-9_.-]+)`)
)

func ParseLoose(text string) *TomlDocument {
	doc := NewTomlDocument()
	lines := strings.Split(text, "\n")

	currentTable := ""
	inMultilineString := false
	multilineDelimiter := ""

	for lineNum, line := range lines {
		if inMultilineString {
			if strings.Contains(line, multilineDelimiter) {
				inMultilineString = false
				multilineDelimiter = ""
			}
			continue
		}

		if commentRegex.MatchString(line) {
			continue
		}

		if matches := arrayTableHeaderRegex.FindStringSubmatch(line); matches != nil {
			tableName := strings.TrimSpace(matches[1])
			currentTable = tableName

			doc.Tables = append(doc.Tables, Table{
				Name: tableName,
				Path: strings.Split(tableName, "."),
				Range: Range{
					Start: Position{Line: lineNum, Character: 0},
					End:   Position{Line: lineNum, Character: len(line)},
				},
			})
			doc.CurrentTable = currentTable
			continue
		}

		if matches := tableHeaderRegex.FindStringSubmatch(line); matches != nil {
			tableName := strings.TrimSpace(matches[1])
			currentTable = tableName

			doc.Tables = append(doc.Tables, Table{
				Name: tableName,
				Path: strings.Split(tableName, "."),
				Range: Range{
					Start: Position{Line: lineNum, Character: 0},
					End:   Position{Line: lineNum, Character: len(line)},
				},
			})
			doc.CurrentTable = currentTable
			continue
		}

		if matches := keyValueRegex.FindStringSubmatch(line); matches != nil {
			keyName := strings.TrimSpace(matches[1])
			hasEquals := strings.Contains(line, "=")

			var keyPath []string
			if currentTable != "" {
				keyPath = append(strings.Split(currentTable, "."), keyName)
			} else {
				keyPath = []string{keyName}
			}

			key := Key{
				Name: keyName,
				Path: keyPath,
				Range: Range{
					Start: Position{Line: lineNum, Character: 0},
					End:   Position{Line: lineNum, Character: len(line)},
				},
			}

			if !hasEquals {
				key.Prefix = keyName
			}

			doc.Keys = append(doc.Keys, key)
			doc.CurrentTable = currentTable

			if hasEquals {
				eqIdx := strings.Index(line, "=")
				valuePart := line[eqIdx+1:]
				if delim, open := opensMultilineString(valuePart); open {
					inMultilineString = true
					multilineDelimiter = delim
				}
			}
		}
	}

	return doc
}

func opensMultilineString(valuePart string) (string, bool) {
	trimmed := strings.TrimSpace(valuePart)
	for _, delim := range []string{`"""`, `'''`} {
		if strings.HasPrefix(trimmed, delim) {
			rest := trimmed[len(delim):]
			if strings.Contains(rest, delim) {
				return "", false
			}
			return delim, true
		}
	}
	return "", false
}

func ParseLooseWithCursor(text string, cursorPos Position) *TomlDocument {
	doc := ParseLoose(text)

	lines := strings.Split(text, "\n")
	if cursorPos.Line < len(lines) {
		line := lines[cursorPos.Line]
		if cursorPos.Character <= len(line) {
			prefix := line[:cursorPos.Character]

			if matches := keyValueRegex.FindStringSubmatch(prefix); matches != nil {
				partialKey := strings.TrimSpace(matches[1])

				for i := range doc.Keys {
					if doc.Keys[i].Range.Start.Line == cursorPos.Line {
						doc.Keys[i].Prefix = partialKey
						break
					}
				}
			}
		}
	}

	return doc
}

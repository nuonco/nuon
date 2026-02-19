package toml

import (
	"regexp"
	"strings"
)

var (
	// Regex patterns for loose parsing
	tableHeaderRegex      = regexp.MustCompile(`^\s*\[\s*([A-Za-z0-9_.-]+)`)
	keyValueRegex         = regexp.MustCompile(`^\s*([A-Za-z0-9_.-]+)\s*=?`)
	commentRegex          = regexp.MustCompile(`^\s*#`)
	arrayTableHeaderRegex = regexp.MustCompile(`^\s*\[\[\s*([A-Za-z0-9_.-]+)`)
)

// ParseLoose parses TOML using a text-based scanner that tolerates invalid input
// This mode never returns an error and works on incomplete TOML
func ParseLoose(text string) *TomlDocument {
	doc := NewTomlDocument()
	lines := strings.Split(text, "\n")

	currentTable := ""
	inMultilineString := false
	multilineDelimiter := "" // `"""` or `'''`

	for lineNum, line := range lines {
		// If inside a multiline string, check for the closing delimiter
		if inMultilineString {
			if strings.Contains(line, multilineDelimiter) {
				inMultilineString = false
				multilineDelimiter = ""
			}
			continue
		}

		// Skip comments
		if commentRegex.MatchString(line) {
			continue
		}

		// Check for array-of-tables header [[table]]
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

		// Check for table header [table]
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

		// Check for key-value or partial key
		if matches := keyValueRegex.FindStringSubmatch(line); matches != nil {
			keyName := strings.TrimSpace(matches[1])
			hasEquals := strings.Contains(line, "=")

			// Build full path
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

			// If no equals sign, it's a partial/incomplete key
			if !hasEquals {
				key.Prefix = keyName
			}

			doc.Keys = append(doc.Keys, key)
			doc.CurrentTable = currentTable

			// Check if this line opens a multiline string (""" or ''')
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

// opensMultilineString checks if a value part (after the =) opens a multiline string.
// Returns the delimiter and true if a multiline string is opened but not closed on the same line.
func opensMultilineString(valuePart string) (string, bool) {
	trimmed := strings.TrimSpace(valuePart)
	for _, delim := range []string{`"""`, `'''`} {
		if strings.HasPrefix(trimmed, delim) {
			// Check if it's also closed on the same line (after the opening delimiter)
			rest := trimmed[len(delim):]
			if strings.Contains(rest, delim) {
				return "", false
			}
			return delim, true
		}
	}
	return "", false
}

// ParseLooseWithCursor parses TOML and detects the prefix at a specific cursor position
func ParseLooseWithCursor(text string, cursorPos Position) *TomlDocument {
	doc := ParseLoose(text)

	// Extract prefix at cursor position
	lines := strings.Split(text, "\n")
	if cursorPos.Line < len(lines) {
		line := lines[cursorPos.Line]
		if cursorPos.Character <= len(line) {
			prefix := line[:cursorPos.Character]

			// Find partial key
			if matches := keyValueRegex.FindStringSubmatch(prefix); matches != nil {
				partialKey := strings.TrimSpace(matches[1])

				// Update the key on this line with the prefix
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

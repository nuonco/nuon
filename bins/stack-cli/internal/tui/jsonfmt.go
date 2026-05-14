package tui

import (
	"bytes"
	"encoding/json"
	"strings"
	"unicode"

	"charm.land/lipgloss/v2"
)

// prettyJSON returns a colorized, indented rendering of raw. On parse failure
// it falls back to the raw input dimmed with a one-line warning so callers
// don't need to branch.
//
// width caps visible columns (wraps overflowing lines); maxLines caps the
// total number of visible lines and appends a dim ellipsis when exceeded.
func prettyJSON(raw string, width, maxLines int) string {
	var indented bytes.Buffer
	if err := json.Indent(&indented, []byte(raw), "", "  "); err != nil {
		warn := dimStyle.Render("(not valid JSON — shown raw)")
		body := dimStyle.Render(raw)
		return warn + "\n" + capLines(wrap(body, width), maxLines)
	}
	colored := colorizeJSON(indented.String())
	wrapped := wrapColored(colored, width)
	return capLines(wrapped, maxLines)
}

// colorizeJSON walks the formatted JSON and styles each token. It is not a
// strict parser — it relies on json.Indent producing canonical output (no
// stray whitespace inside tokens, every key followed by `: `).
func colorizeJSON(s string) string {
	var out strings.Builder
	i := 0
	for i < len(s) {
		c := s[i]
		switch {
		case c == '"':
			// Walk to the closing quote, honoring backslash escapes.
			j := i + 1
			for j < len(s) {
				if s[j] == '\\' && j+1 < len(s) {
					j += 2
					continue
				}
				if s[j] == '"' {
					break
				}
				j++
			}
			tok := s[i : j+1]
			k := j + 1
			for k < len(s) && (s[k] == ' ' || s[k] == '\t') {
				k++
			}
			if k < len(s) && s[k] == ':' {
				out.WriteString(jsonKey.Render(tok))
			} else {
				out.WriteString(jsonString.Render(tok))
			}
			i = j + 1
		case c == '{' || c == '}' || c == '[' || c == ']' || c == ',' || c == ':':
			out.WriteString(jsonPunct.Render(string(c)))
			i++
		case c == '-' || (c >= '0' && c <= '9'):
			j := i + 1
			for j < len(s) && (isNumChar(rune(s[j]))) {
				j++
			}
			out.WriteString(jsonNumber.Render(s[i:j]))
			i = j
		case unicode.IsLetter(rune(c)):
			j := i + 1
			for j < len(s) && unicode.IsLetter(rune(s[j])) {
				j++
			}
			out.WriteString(jsonLiteral.Render(s[i:j]))
			i = j
		default:
			out.WriteByte(c)
			i++
		}
	}
	return out.String()
}

func isNumChar(r rune) bool {
	switch {
	case r >= '0' && r <= '9':
		return true
	case r == '.' || r == 'e' || r == 'E' || r == '+' || r == '-':
		return true
	}
	return false
}

// wrapColored wraps each line to width using lipgloss.Width for measurement so
// ANSI escape sequences don't get counted. Continuation lines get a small
// hanging indent so wrapped content stays visually nested.
func wrapColored(s string, width int) string {
	if width <= 4 {
		return s
	}
	var out []string
	for _, line := range strings.Split(s, "\n") {
		if lipgloss.Width(line) <= width {
			out = append(out, line)
			continue
		}
		// Naive break by visible width. We re-split on raw bytes since the
		// styled tokens are short and don't contain newlines.
		current := ""
		runes := []rune(line)
		for _, r := range runes {
			next := current + string(r)
			if lipgloss.Width(next) > width {
				out = append(out, current)
				current = "  " + string(r)
				continue
			}
			current = next
		}
		if current != "" {
			out = append(out, current)
		}
	}
	return strings.Join(out, "\n")
}

func capLines(s string, maxLines int) string {
	if maxLines <= 0 {
		return s
	}
	lines := strings.Split(s, "\n")
	if len(lines) <= maxLines {
		return s
	}
	return strings.Join(lines[:maxLines], "\n") + "\n" + dimStyle.Render("…")
}

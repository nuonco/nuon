package handlers

import (
	"regexp"
	"strings"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

var (
	fmtTableHeaderRegex      = regexp.MustCompile(`^\s*\[\s*([A-Za-z0-9_.-]+)\s*\]\s*$`)
	fmtArrayTableHeaderRegex = regexp.MustCompile(`^\s*\[\[\s*([A-Za-z0-9_.-]+)\s*\]\]\s*$`)
	fmtCommentRegex          = regexp.MustCompile(`^\s*#`)
	fmtKeyValueRegex         = regexp.MustCompile(`^\s*([A-Za-z0-9_.-]+)\s*=\s*(.*)$`)
)

func TextDocumentFormatting(ctx *glsp.Context, params *protocol.DocumentFormattingParams) ([]protocol.TextEdit, error) {
	uri := params.TextDocument.URI

	openDocumentsMutex.RLock()
	text, ok := openDocuments[uri]
	openDocumentsMutex.RUnlock()
	if !ok {
		return nil, nil
	}

	formatted := FormatToml(text)
	if formatted == text {
		return nil, nil
	}

	lines := strings.Split(text, "\n")
	lastLine := uint32(len(lines) - 1)
	lastChar := uint32(len(lines[lastLine]))

	return []protocol.TextEdit{
		{
			Range: protocol.Range{
				Start: protocol.Position{Line: 0, Character: 0},
				End:   protocol.Position{Line: lastLine, Character: lastChar},
			},
			NewText: formatted,
		},
	}, nil
}

type lineKind int

const (
	lineBlank lineKind = iota
	lineComment
	lineTableHeader
	lineArrayTableHeader
	lineKeyValue
	lineMultilineContent
	lineOther
)

type classifiedLine struct {
	kind    lineKind
	raw     string
	key     string // only for lineKeyValue
	value   string // only for lineKeyValue
	name    string // only for table/array-table headers
	section int    // which section this line belongs to
}

func FormatToml(text string) string {
	rawLines := strings.Split(text, "\n")
	classified := classifyLines(rawLines)
	classified = assignSections(classified)
	aligned := alignSections(classified)
	return buildOutput(aligned)
}

func classifyLines(rawLines []string) []classifiedLine {
	result := make([]classifiedLine, 0, len(rawLines))
	inMultiline := false
	multilineDelim := ""

	for _, raw := range rawLines {
		if inMultiline {
			result = append(result, classifiedLine{kind: lineMultilineContent, raw: raw})
			if strings.Contains(raw, multilineDelim) {
				inMultiline = false
				multilineDelim = ""
			}
			continue
		}

		trimmed := strings.TrimSpace(raw)

		if trimmed == "" {
			result = append(result, classifiedLine{kind: lineBlank, raw: raw})
			continue
		}

		if fmtCommentRegex.MatchString(raw) {
			result = append(result, classifiedLine{kind: lineComment, raw: raw})
			continue
		}

		if m := fmtArrayTableHeaderRegex.FindStringSubmatch(raw); m != nil {
			result = append(result, classifiedLine{kind: lineArrayTableHeader, raw: raw, name: m[1]})
			continue
		}

		if m := fmtTableHeaderRegex.FindStringSubmatch(raw); m != nil {
			result = append(result, classifiedLine{kind: lineTableHeader, raw: raw, name: m[1]})
			continue
		}

		if m := fmtKeyValueRegex.FindStringSubmatch(raw); m != nil {
			cl := classifiedLine{kind: lineKeyValue, raw: raw, key: m[1], value: m[2]}
			// Check if value opens a multiline string
			valTrimmed := strings.TrimSpace(m[2])
			for _, delim := range []string{`"""`, `'''`} {
				if strings.HasPrefix(valTrimmed, delim) {
					rest := valTrimmed[len(delim):]
					if !strings.Contains(rest, delim) {
						inMultiline = true
						multilineDelim = delim
					}
				}
			}
			result = append(result, cl)
			continue
		}

		result = append(result, classifiedLine{kind: lineOther, raw: raw})
	}

	return result
}

func assignSections(lines []classifiedLine) []classifiedLine {
	section := 0
	for i := range lines {
		if lines[i].kind == lineTableHeader || lines[i].kind == lineArrayTableHeader {
			section++
		}
		lines[i].section = section
	}
	return lines
}

func alignSections(lines []classifiedLine) []classifiedLine {
	// Find max key length per section
	maxKey := make(map[int]int)
	for _, l := range lines {
		if l.kind == lineKeyValue {
			if len(l.key) > maxKey[l.section] {
				maxKey[l.section] = len(l.key)
			}
		}
	}

	for i := range lines {
		if lines[i].kind == lineKeyValue {
			pad := maxKey[lines[i].section]
			lines[i].raw = lines[i].key + strings.Repeat(" ", pad-len(lines[i].key)) + " = " + strings.TrimSpace(lines[i].value)
		}
	}

	return lines
}

func buildOutput(lines []classifiedLine) string {
	var out []string
	prevKind := lineBlank // treat start-of-file as blank so we don't add a leading blank line

	for i, l := range lines {
		switch l.kind {
		case lineBlank:
			// Collapse consecutive blanks to one; skip if previous was also blank
			if prevKind != lineBlank {
				out = append(out, "")
			}
			prevKind = lineBlank
			continue

		case lineTableHeader:
			// Blank line before header (unless at start of output)
			if len(out) > 0 && prevKind != lineBlank {
				out = append(out, "")
			}
			out = append(out, "["+l.name+"]")

		case lineArrayTableHeader:
			if len(out) > 0 && prevKind != lineBlank {
				out = append(out, "")
			}
			out = append(out, "[["+l.name+"]]")

		case lineComment:
			out = append(out, strings.TrimRight(l.raw, " \t"))

		case lineKeyValue:
			out = append(out, strings.TrimRight(l.raw, " \t"))

		case lineMultilineContent:
			// Preserve exactly as-is
			out = append(out, l.raw)

		case lineOther:
			out = append(out, strings.TrimRight(l.raw, " \t"))
		}

		_ = i
		prevKind = l.kind
	}

	// Ensure final newline
	result := strings.Join(out, "\n")
	if !strings.HasSuffix(result, "\n") {
		result += "\n"
	}
	return result
}

package handlers

import (
	"strings"

	"github.com/nuonco/nuon/pkg/parser/toml"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func TextDocumentFoldingRange(ctx *glsp.Context, params *protocol.FoldingRangeParams) ([]protocol.FoldingRange, error) {
	uri := params.TextDocument.URI
	log.Infof("📁 Folding range requested for %s", uri)

	openDocumentsMutex.RLock()
	text, ok := openDocuments[uri]
	openDocumentsMutex.RUnlock()
	if !ok {
		log.Warningf("⚠️ Document not found: %s", uri)
		return nil, nil
	}

	ranges := []protocol.FoldingRange{}
	lines := strings.Split(text, "\n")
	doc := toml.ParseToml(text)

	ranges = append(ranges, buildTableFoldingRanges(doc.Tables, lines)...)
	ranges = append(ranges, findMultiLineStringRanges(lines)...)
	ranges = append(ranges, findCommentBlockRanges(lines)...)

	log.Infof("✅ Found %d folding ranges", len(ranges))
	return ranges, nil
}

func buildTableFoldingRanges(tables []toml.Table, lines []string) []protocol.FoldingRange {
	ranges := []protocol.FoldingRange{}

	for i, table := range tables {
		startLine := uint32(table.Range.Start.Line)

		var endLine uint32
		if i+1 < len(tables) {
			nextTableLine := tables[i+1].Range.Start.Line
			endLine = uint32(findLastContentLine(lines, int(startLine)+1, nextTableLine-1))
		} else {
			endLine = uint32(findLastContentLine(lines, int(startLine)+1, len(lines)-1))
		}

		if endLine > startLine {
			kind := string(protocol.FoldingRangeKindRegion)
			ranges = append(ranges, protocol.FoldingRange{
				StartLine: startLine,
				EndLine:   endLine,
				Kind:      &kind,
			})
		}
	}

	return ranges
}

func findLastContentLine(lines []string, startLine, endLine int) int {
	if startLine < 0 {
		startLine = 0
	}
	if endLine >= len(lines) {
		endLine = len(lines) - 1
	}

	lastContent := startLine - 1
	for i := startLine; i <= endLine; i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			lastContent = i
		}
	}

	if lastContent < startLine {
		return endLine
	}
	return lastContent
}

func findMultiLineStringRanges(lines []string) []protocol.FoldingRange {
	ranges := []protocol.FoldingRange{}

	inMultiLineString := false
	multiLineStart := 0
	delimiter := ""

	for i, line := range lines {
		if !inMultiLineString {
			if idx := strings.Index(line, `"""`); idx != -1 {
				rest := line[idx+3:]
				if !strings.Contains(rest, `"""`) {
					inMultiLineString = true
					multiLineStart = i
					delimiter = `"""`
				}
			} else if idx := strings.Index(line, `'''`); idx != -1 {
				rest := line[idx+3:]
				if !strings.Contains(rest, `'''`) {
					inMultiLineString = true
					multiLineStart = i
					delimiter = `'''`
				}
			}
		} else {
			if strings.Contains(line, delimiter) {
				inMultiLineString = false
				if i > multiLineStart {
					kind := string(protocol.FoldingRangeKindRegion)
					ranges = append(ranges, protocol.FoldingRange{
						StartLine: uint32(multiLineStart),
						EndLine:   uint32(i),
						Kind:      &kind,
					})
				}
				delimiter = ""
			}
		}
	}

	return ranges
}

func findCommentBlockRanges(lines []string) []protocol.FoldingRange {
	ranges := []protocol.FoldingRange{}

	inCommentBlock := false
	blockStart := 0

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		isComment := strings.HasPrefix(trimmed, "#")

		if isComment && !inCommentBlock {
			inCommentBlock = true
			blockStart = i
		} else if !isComment && inCommentBlock {
			inCommentBlock = false
			blockEnd := i - 1

			if blockEnd-blockStart >= 2 {
				kind := string(protocol.FoldingRangeKindComment)
				ranges = append(ranges, protocol.FoldingRange{
					StartLine: uint32(blockStart),
					EndLine:   uint32(blockEnd),
					Kind:      &kind,
				})
			}
		}
	}

	if inCommentBlock {
		blockEnd := len(lines) - 1
		if blockEnd-blockStart >= 2 {
			kind := string(protocol.FoldingRangeKindComment)
			ranges = append(ranges, protocol.FoldingRange{
				StartLine: uint32(blockStart),
				EndLine:   uint32(blockEnd),
				Kind:      &kind,
			})
		}
	}

	return ranges
}

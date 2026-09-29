package handlers

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func TextDocumentDidChange(ctx *glsp.Context, params *protocol.DidChangeTextDocumentParams) error {
	uri := params.TextDocument.URI
	log.Infof("📝 Document changed: %s (%d changes)", uri, len(params.ContentChanges))

	if len(params.ContentChanges) == 0 {
		log.Warningf("⚠️  No content changes received")
		return nil
	}

	openDocumentsMutex.RLock()
	currentText, ok := openDocuments[uri]
	openDocumentsMutex.RUnlock()
	if !ok {
		log.Errorf("❌ Document not found for didChange: %s", uri)
		return fmt.Errorf("document not found: %s", uri)
	}
	log.Debugf("Current document length: %d chars", len(currentText))

	for i, changeAny := range params.ContentChanges {
		text, rangePtr, ok := extractChangeContent(changeAny)
		if !ok {
			log.Errorf("❌ Could not extract content from change at index %d", i)
			return fmt.Errorf("could not extract content from change at index %d", i)
		}

		if rangePtr == nil {
			log.Debugf("🔄 Full document sync (change %d): %d chars", i, len(text))
			currentText = text
		} else {
			log.Debugf("🔧 Incremental change at %d:%d-%d:%d, new text: %d chars",
				rangePtr.Start.Line, rangePtr.Start.Character,
				rangePtr.End.Line, rangePtr.End.Character, len(text))
			currentText = applyTextChange(currentText, rangePtr, text)
		}
	}

	log.Debugf("✅ Updated document, new length: %d chars", len(currentText))
	openDocumentsMutex.Lock()
	openDocuments[uri] = currentText
	openDocumentsMutex.Unlock()

	PublishDiagnostics(ctx, uri, currentText)

	return nil
}

func extractChangeContent(changeAny interface{}) (text string, rangePtr *protocol.Range, ok bool) {
	if changeMap, isMap := changeAny.(map[string]interface{}); isMap {
		if textVal, hasText := changeMap["text"]; hasText {
			if textStr, isStr := textVal.(string); isStr {
				text = textStr
			} else {
				return "", nil, false
			}
		} else {
			return "", nil, false
		}

		if rangeVal, hasRange := changeMap["range"]; hasRange && rangeVal != nil {
			if rangeMap, isRangeMap := rangeVal.(map[string]interface{}); isRangeMap {
				rangePtr = &protocol.Range{}
				if start, hasStart := rangeMap["start"].(map[string]interface{}); hasStart {
					if line, hasLine := start["line"].(float64); hasLine {
						rangePtr.Start.Line = uint32(line)
					}
					if char, hasChar := start["character"].(float64); hasChar {
						rangePtr.Start.Character = uint32(char)
					}
				}
				if end, hasEnd := rangeMap["end"].(map[string]interface{}); hasEnd {
					if line, hasLine := end["line"].(float64); hasLine {
						rangePtr.End.Line = uint32(line)
					}
					if char, hasChar := end["character"].(float64); hasChar {
						rangePtr.End.Character = uint32(char)
					}
				}
			}
		}
		return text, rangePtr, true
	}

	v := reflect.ValueOf(changeAny)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	textFound := false
	candidates := []string{"Text", "Content", "Value", "NewText", "New"}
	for _, name := range candidates {
		f := v.FieldByName(name)
		if f.IsValid() && f.Kind() == reflect.String {
			text = f.String()
			textFound = true
			break
		}
	}
	if !textFound {
		return "", nil, false
	}

	rangeField := v.FieldByName("Range")
	if rangeField.IsValid() && !rangeField.IsNil() {
		if r, ok := rangeField.Interface().(*protocol.Range); ok {
			rangePtr = r
		}
	}

	return text, rangePtr, true
}

func applyTextChange(text string, rangePtr *protocol.Range, newText string) string {
	if rangePtr == nil {
		return newText
	}

	lines := strings.Split(text, "\n")
	startLine := int(rangePtr.Start.Line)
	startChar := int(rangePtr.Start.Character)
	endLine := int(rangePtr.End.Line)
	endChar := int(rangePtr.End.Character)

	if startLine < 0 {
		startLine = 0
	}
	if startLine >= len(lines) {
		startLine = len(lines) - 1
	}
	if endLine < 0 {
		endLine = 0
	}
	if endLine >= len(lines) {
		endLine = len(lines) - 1
	}

	if startLine == endLine {
		line := lines[startLine]
		if startChar < 0 {
			startChar = 0
		}
		if startChar > len(line) {
			startChar = len(line)
		}
		if endChar < 0 {
			endChar = 0
		}
		if endChar > len(line) {
			endChar = len(line)
		}
		before := line[:startChar]
		after := line[endChar:]
		lines[startLine] = before + newText + after
	} else {
		firstLine := lines[startLine]
		lastLine := lines[endLine]

		if startChar < 0 {
			startChar = 0
		}
		if startChar > len(firstLine) {
			startChar = len(firstLine)
		}
		if endChar < 0 {
			endChar = 0
		}
		if endChar > len(lastLine) {
			endChar = len(lastLine)
		}

		replacement := firstLine[:startChar] + newText + lastLine[endChar:]
		newLines := []string{replacement}

		if endLine+1 < len(lines) {
			newLines = append(newLines, lines[endLine+1:]...)
		}

		lines = append(lines[:startLine], newLines...)
	}

	return strings.Join(lines, "\n")
}

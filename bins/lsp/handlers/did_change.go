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

	// Get current document text
	openDocumentsMutex.RLock()
	currentText, ok := openDocuments[uri]
	openDocumentsMutex.RUnlock()
	if !ok {
		log.Errorf("❌ Document not found for didChange: %s", uri)
		return fmt.Errorf("document not found: %s", uri)
	}
	log.Debugf("Current document length: %d chars", len(currentText))

	// Process each content change
	for i, changeAny := range params.ContentChanges {
		// Extract text and range from the change (supports both struct and map types)
		text, rangePtr, ok := extractChangeContent(changeAny)
		if !ok {
			log.Errorf("❌ Could not extract content from change at index %d", i)
			return fmt.Errorf("could not extract content from change at index %d", i)
		}

		// Check if this is a full document sync (no range)
		if rangePtr == nil {
			log.Debugf("🔄 Full document sync (change %d): %d chars", i, len(text))
			currentText = text
		} else {
			// Incremental change with range - apply the delta
			log.Debugf("🔧 Incremental change at %d:%d-%d:%d, new text: %d chars",
				rangePtr.Start.Line, rangePtr.Start.Character,
				rangePtr.End.Line, rangePtr.End.Character, len(text))
			currentText = applyTextChange(currentText, rangePtr, text)
		}
	}

	// Update the in-memory document
	log.Debugf("✅ Updated document, new length: %d chars", len(currentText))
	openDocumentsMutex.Lock()
	openDocuments[uri] = currentText
	openDocumentsMutex.Unlock()

	// Trigger diagnostics
	PublishDiagnostics(ctx, uri, currentText)

	return nil
}

// extractChangeContent extracts text and range from a content change event
// Handles both typed structs and map[string]interface{} from JSON unmarshaling
func extractChangeContent(changeAny interface{}) (text string, rangePtr *protocol.Range, ok bool) {
	// Try as map first (most common case from JSON)
	if changeMap, isMap := changeAny.(map[string]interface{}); isMap {
		// Extract text
		if textVal, hasText := changeMap["text"]; hasText {
			if textStr, isStr := textVal.(string); isStr {
				text = textStr
			} else {
				return "", nil, false
			}
		} else {
			return "", nil, false
		}

		// Extract range (optional - nil means full document sync)
		if rangeVal, hasRange := changeMap["range"]; hasRange && rangeVal != nil {
			// Range is typically a map with start/end
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

	// Fallback: try as typed struct using reflection
	v := reflect.ValueOf(changeAny)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Extract text from struct
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

	// Extract range from struct (optional)
	rangeField := v.FieldByName("Range")
	if rangeField.IsValid() && !rangeField.IsNil() {
		if r, ok := rangeField.Interface().(*protocol.Range); ok {
			rangePtr = r
		}
	}

	return text, rangePtr, true
}

// applyTextChange applies an incremental text change to the document
func applyTextChange(text string, rangePtr *protocol.Range, newText string) string {
	if rangePtr == nil {
		return newText
	}

	lines := strings.Split(text, "\n")
	startLine := int(rangePtr.Start.Line)
	startChar := int(rangePtr.Start.Character)
	endLine := int(rangePtr.End.Line)
	endChar := int(rangePtr.End.Character)

	// Clamp to valid bounds
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

	// Single line change
	if startLine == endLine {
		line := lines[startLine]
		// Clamp character offsets to line bounds
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
		// Multi-line change
		firstLine := lines[startLine]
		lastLine := lines[endLine]

		// Clamp character offsets to respective line bounds
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

		// Build replacement
		replacement := firstLine[:startChar] + newText + lastLine[endChar:]
		newLines := []string{replacement}

		// Keep lines after the change
		if endLine+1 < len(lines) {
			newLines = append(newLines, lines[endLine+1:]...)
		}

		// Replace the affected lines
		lines = append(lines[:startLine], newLines...)
	}

	return strings.Join(lines, "\n")
}

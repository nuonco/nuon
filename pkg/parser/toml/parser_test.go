package toml

import (
	"testing"
)

func TestParseToml_ValidToml(t *testing.T) {
	input := `
[connected_repo]
directory = "/path/to/repo"
enabled = true
`
	doc := ParseToml(input)

	if len(doc.Tables) == 0 {
		t.Error("Expected at least one table")
	}

	if len(doc.Keys) == 0 {
		t.Error("Expected at least one key")
	}

	if doc.Tables[0].Name != "connected_repo" {
		t.Errorf("Expected table name 'connected_repo', got '%s'", doc.Tables[0].Name)
	}
}

func TestParseToml_IncompleteTable(t *testing.T) {
	input := `[connected_repo`

	doc := ParseToml(input)

	// Should not panic or error
	if doc == nil {
		t.Error("Expected non-nil document for incomplete table")
	}

	if len(doc.Tables) == 0 {
		t.Error("Expected loose parser to detect incomplete table")
	}
}

func TestParseToml_IncompleteKey(t *testing.T) {
	input := `
[connected_repo]
dir
`
	doc := ParseToml(input)

	if doc == nil {
		t.Error("Expected non-nil document")
	}

	// Should detect the partial key
	if len(doc.Keys) == 0 {
		t.Error("Expected loose parser to detect partial key")
	}

	if doc.Keys[0].Name != "dir" {
		t.Errorf("Expected key name 'dir', got '%s'", doc.Keys[0].Name)
	}
}

func TestParseToml_DanglingAssignment(t *testing.T) {
	input := `
[connected_repo]
directory =
`
	doc := ParseToml(input)

	if doc == nil {
		t.Error("Expected non-nil document")
	}

	// Should not panic
	if len(doc.Keys) == 0 {
		t.Error("Expected to detect key with dangling assignment")
	}
}

func TestParseToml_BrokenString(t *testing.T) {
	input := `
[connected_repo]
directory = "broken
`
	doc := ParseToml(input)

	// Should fallback to loose parsing without error
	if doc == nil {
		t.Error("Expected non-nil document for broken string")
	}
}

func TestParseToml_NestedTables(t *testing.T) {
	input := `
[parent.child]
key = "value"
`
	doc := ParseToml(input)

	if len(doc.Tables) == 0 {
		t.Error("Expected nested table")
	}

	// Find the nested table (it should be the last one or named 'parent.child')
	var found bool
	for _, table := range doc.Tables {
		if table.Name == "parent.child" {
			found = true
			if len(table.Path) != 2 {
				t.Errorf("Expected path length 2, got %d", len(table.Path))
			}
			break
		}
	}

	if !found {
		t.Error("Expected to find 'parent.child' table")
	}
}

func TestParseToml_ArrayOfTables(t *testing.T) {
	input := `
[[entries]]
name = "first"

[[entries]]
name = "second"
`
	doc := ParseToml(input)

	if len(doc.Tables) == 0 {
		t.Error("Expected array-of-tables")
	}
}

func TestParseToml_OnlyComments(t *testing.T) {
	input := `
# This is a comment
# Another comment
`
	doc := ParseToml(input)

	if doc == nil {
		t.Error("Expected non-nil document for comments-only file")
	}
}

func TestParseToml_EmptyFile(t *testing.T) {
	input := ``

	doc := ParseToml(input)

	if doc == nil {
		t.Error("Expected non-nil document for empty file")
	}
}

func TestParseLoose_CurrentTableDetection(t *testing.T) {
	input := `
[connected_repo]
directory = "test"
`
	doc := ParseLoose(input)

	if doc.CurrentTable != "connected_repo" {
		t.Errorf("Expected current table 'connected_repo', got '%s'", doc.CurrentTable)
	}
}

func TestParseLooseWithCursor_PrefixDetection(t *testing.T) {
	input := `
[connected_repo]
dir`
	cursorPos := Position{Line: 2, Character: 3}

	doc := ParseLooseWithCursor(input, cursorPos)

	if len(doc.Keys) == 0 {
		t.Error("Expected at least one key")
	}

	if doc.Keys[0].Prefix != "dir" {
		t.Errorf("Expected prefix 'dir', got '%s'", doc.Keys[0].Prefix)
	}
}

func TestContextAt_BasicContext(t *testing.T) {
	input := `
[connected_repo]
directory = "test"
`
	doc := ParseToml(input)
	ctx := doc.ContextAt(Position{Line: 2, Character: 0})

	if ctx.CurrentTable != "connected_repo" {
		t.Errorf("Expected current table 'connected_repo', got '%s'", ctx.CurrentTable)
	}

	if len(ctx.KeyPath) == 0 {
		t.Error("Expected non-empty key path")
	}
}

func TestParseToml_MultilineString(t *testing.T) {
	input := `
[section]
text = """
multiline
content
"""
`
	doc := ParseToml(input)

	if doc == nil {
		t.Fatal("Expected non-nil document")
	}

	// Should only have one key ("text"), not keys from inside the multiline string
	if len(doc.Keys) != 1 {
		t.Errorf("Expected 1 key, got %d: %v", len(doc.Keys), keysToNames(doc.Keys))
	}
	if len(doc.Keys) > 0 && doc.Keys[0].Name != "text" {
		t.Errorf("Expected key name 'text', got '%s'", doc.Keys[0].Name)
	}
}

func TestParseToml_MultilineStringWithKeyLikeContent(t *testing.T) {
	input := `
[[steps]]
name = "create secrets"
inline_contents = """
#!/usr/bin/env sh
password=` + "`" + `openssl rand -hex 12` + "`" + `
setNuon something
kubectl create -n whoami secret generic whoami
"""
`
	doc := ParseToml(input)

	if doc == nil {
		t.Fatal("Expected non-nil document")
	}

	// Should only detect "name" and "inline_contents", not "password", "setNuon", "kubectl" etc.
	keyNames := keysToNames(doc.Keys)
	for _, k := range keyNames {
		if k == "password" || k == "setNuon" || k == "kubectl" {
			t.Errorf("Multiline string content '%s' should not be parsed as a key", k)
		}
	}
	if len(doc.Keys) != 2 {
		t.Errorf("Expected 2 keys (name, inline_contents), got %d: %v", len(doc.Keys), keyNames)
	}
}

func TestParseToml_MultilineStringClosedOnSameLine(t *testing.T) {
	input := `
[section]
text = """short"""
other_key = "value"
`
	doc := ParseToml(input)

	if doc == nil {
		t.Fatal("Expected non-nil document")
	}

	// Both keys should be detected since the multiline string is closed on the same line
	if len(doc.Keys) != 2 {
		t.Errorf("Expected 2 keys, got %d: %v", len(doc.Keys), keysToNames(doc.Keys))
	}
}

func TestParseToml_MultilineLiteralString(t *testing.T) {
	input := `
[section]
text = '''
password = "fake"
other = true
'''
real_key = "value"
`
	doc := ParseToml(input)

	if doc == nil {
		t.Fatal("Expected non-nil document")
	}

	// Should have "text" and "real_key", not "password" or "other" from inside '''
	keyNames := keysToNames(doc.Keys)
	for _, k := range keyNames {
		if k == "password" || k == "other" {
			t.Errorf("Literal string content '%s' should not be parsed as a key", k)
		}
	}
	if len(doc.Keys) != 2 {
		t.Errorf("Expected 2 keys (text, real_key), got %d: %v", len(doc.Keys), keyNames)
	}
}

func keysToNames(keys []Key) []string {
	names := make([]string, len(keys))
	for i, k := range keys {
		names[i] = k.Name
	}
	return names
}

func TestParseToml_AlwaysUsesLooseMode(t *testing.T) {
	validInput := `
[table]
key = "value"
`
	doc := ParseToml(validInput)

	// Should always use loose parsing for position information
	if doc == nil {
		t.Error("Expected non-nil document")
	}

	// Verify it has proper position information
	if len(doc.Tables) == 0 {
		t.Error("Expected tables to be detected")
	}

	// Check that positions are preserved (not all zeros)
	hasNonZeroPosition := false
	for _, table := range doc.Tables {
		if table.Range.Start.Line > 0 {
			hasNonZeroPosition = true
			break
		}
	}
	if !hasNonZeroPosition {
		t.Error("Expected loose mode to preserve position information")
	}
}

func TestValidateToml_ValidInput(t *testing.T) {
	validInput := `
[table]
key = "value"
`
	err := ValidateToml(validInput)
	if err != nil {
		t.Errorf("Expected valid TOML to pass validation, got error: %v", err)
	}
}

func TestValidateToml_InvalidInput(t *testing.T) {
	invalidInput := `[broken`

	err := ValidateToml(invalidInput)
	if err == nil {
		t.Error("Expected invalid TOML to fail validation")
	}
}

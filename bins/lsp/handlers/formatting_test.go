package handlers

import (
	"testing"
)

func TestFormatToml_SpacingNormalization(t *testing.T) {
	input := "key=value\nother =  123\n"
	want := "key   = value\nother = 123\n"
	got := FormatToml(input)
	if got != want {
		t.Errorf("spacing normalization:\ngot:  %q\nwant: %q", got, want)
	}
}

func TestFormatToml_AlignmentWithinSection(t *testing.T) {
	input := `[server]
port = 8080
hostname = "localhost"
x = true
`
	want := `[server]
port     = 8080
hostname = "localhost"
x        = true
`
	got := FormatToml(input)
	if got != want {
		t.Errorf("alignment:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestFormatToml_AlignmentPerSection(t *testing.T) {
	input := `[a]
x = 1
longkey = 2

[b]
y = 3
`
	want := `[a]
x       = 1
longkey = 2

[b]
y = 3
`
	got := FormatToml(input)
	if got != want {
		t.Errorf("per-section alignment:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestFormatToml_TableHeaderNormalization(t *testing.T) {
	input := "  [ server ]\nport = 8080\n"
	want := "[server]\nport = 8080\n"
	got := FormatToml(input)
	if got != want {
		t.Errorf("header normalization:\ngot:  %q\nwant: %q", got, want)
	}
}

func TestFormatToml_ArrayTableHeaderNormalization(t *testing.T) {
	input := "  [[ items ]]\nname = \"a\"\n"
	want := "[[items]]\nname = \"a\"\n"
	got := FormatToml(input)
	if got != want {
		t.Errorf("array table normalization:\ngot:  %q\nwant: %q", got, want)
	}
}

func TestFormatToml_BlankLineBeforeSections(t *testing.T) {
	input := "[a]\nx = 1\n[b]\ny = 2\n"
	want := "[a]\nx = 1\n\n[b]\ny = 2\n"
	got := FormatToml(input)
	if got != want {
		t.Errorf("blank line insertion:\ngot:  %q\nwant: %q", got, want)
	}
}

func TestFormatToml_BlankLineCollapsing(t *testing.T) {
	input := "[a]\nx = 1\n\n\n\ny = 2\n"
	want := "[a]\nx = 1\n\ny = 2\n"
	got := FormatToml(input)
	if got != want {
		t.Errorf("blank line collapsing:\ngot:  %q\nwant: %q", got, want)
	}
}

func TestFormatToml_TrailingWhitespace(t *testing.T) {
	input := "key = value   \n# comment   \n"
	want := "key = value\n# comment\n"
	got := FormatToml(input)
	if got != want {
		t.Errorf("trailing whitespace:\ngot:  %q\nwant: %q", got, want)
	}
}

func TestFormatToml_CommentPreservation(t *testing.T) {
	input := "# This is a comment\nkey = value\n# Another comment\n"
	want := "# This is a comment\nkey = value\n# Another comment\n"
	got := FormatToml(input)
	if got != want {
		t.Errorf("comment preservation:\ngot:  %q\nwant: %q", got, want)
	}
}

func TestFormatToml_MultilineStringPreservation(t *testing.T) {
	input := "key = \"\"\"\n  hello\n  world\n\"\"\"\n"
	want := "key = \"\"\"\n  hello\n  world\n\"\"\"\n"
	got := FormatToml(input)
	if got != want {
		t.Errorf("multiline string:\ngot:  %q\nwant: %q", got, want)
	}
}

func TestFormatToml_AlreadyFormatted(t *testing.T) {
	input := "[server]\nport = 8080\nhost = \"localhost\"\n"
	got := FormatToml(input)
	if got != input {
		t.Errorf("already formatted should be unchanged:\ngot:  %q\nwant: %q", got, input)
	}
}

func TestFormatToml_EmptyFile(t *testing.T) {
	got := FormatToml("")
	if got != "\n" {
		t.Errorf("empty file:\ngot:  %q\nwant: %q", got, "\n")
	}
}

func TestFormatToml_CommentOnlyFile(t *testing.T) {
	input := "# just a comment\n"
	got := FormatToml(input)
	if got != input {
		t.Errorf("comment only:\ngot:  %q\nwant: %q", got, input)
	}
}

func TestFormatToml_TopLevelKeysAndSections(t *testing.T) {
	input := "name=\"myapp\"\nversion  =  \"1.0\"\n[server]\nport=8080\n"
	want := "name    = \"myapp\"\nversion = \"1.0\"\n\n[server]\nport = 8080\n"
	got := FormatToml(input)
	if got != want {
		t.Errorf("top-level + sections:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestFormatToml_FinalNewline(t *testing.T) {
	input := "key = value"
	got := FormatToml(input)
	if got != "key = value\n" {
		t.Errorf("final newline:\ngot:  %q\nwant: %q", got, "key = value\n")
	}
}

func TestFormatToml_InlineCommentOnKeyValue(t *testing.T) {
	input := "key = value # inline comment\n"
	want := "key = value # inline comment\n"
	got := FormatToml(input)
	if got != want {
		t.Errorf("inline comment:\ngot:  %q\nwant: %q", got, want)
	}
}

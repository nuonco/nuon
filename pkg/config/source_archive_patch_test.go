package config

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"hegel.dev/go/hegel"
)

var sourceArchiveHunkHeader = regexp.MustCompile(`^@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@\n$`)

func applySourceArchivePatch(before, patch string) (string, error) {
	if patch == "" {
		return before, nil
	}
	lines := slices.Collect(strings.Lines(patch))
	if len(lines) < 3 || !strings.HasPrefix(lines[0], "--- ") || !strings.HasPrefix(lines[1], "+++ ") {
		return "", fmt.Errorf("missing unified patch headers")
	}
	source := slices.Collect(strings.Lines(before))
	var result []string
	cursor := 0
	for index := 2; index < len(lines); {
		header := sourceArchiveHunkHeader.FindStringSubmatch(lines[index])
		if header == nil {
			return "", fmt.Errorf("invalid hunk header %q", lines[index])
		}
		var bounds [4]int
		for i, field := range header[1:] {
			if field == "" {
				bounds[i] = 1
				continue
			}
			value, err := strconv.Atoi(field)
			if err != nil {
				return "", err
			}
			bounds[i] = value
		}
		oldStart, oldCount, newStart, newCount := bounds[0], bounds[1], bounds[2], bounds[3]
		if oldCount > 0 {
			oldStart--
		}
		if newCount > 0 {
			newStart--
		}
		if oldStart < cursor || oldStart > len(source) {
			return "", fmt.Errorf("old range starts outside remaining source")
		}
		result = append(result, source[cursor:oldStart]...)
		cursor = oldStart
		if newStart != len(result) {
			return "", fmt.Errorf("new range starts at %d, not %d", newStart, len(result))
		}
		index++
		consumed, emitted := 0, 0
		for consumed < oldCount || emitted < newCount {
			if index >= len(lines) || !strings.HasSuffix(lines[index], "\n") {
				return "", fmt.Errorf("incomplete hunk")
			}
			line := lines[index]
			prefix, value := line[0], line[1:]
			index++
			if index < len(lines) && lines[index] == "\\ No newline at end of file\n" {
				value = strings.TrimSuffix(value, "\n")
				index++
			}
			switch prefix {
			case ' ', '-':
				if cursor >= len(source) || source[cursor] != value {
					return "", fmt.Errorf("source line %d does not match %q", cursor, value)
				}
				cursor++
				consumed++
				if prefix == ' ' {
					result = append(result, value)
					emitted++
				}
			case '+':
				result = append(result, value)
				emitted++
			default:
				return "", fmt.Errorf("invalid hunk line %q", line)
			}
			if consumed > oldCount || emitted > newCount {
				return "", fmt.Errorf("hunk exceeds declared range")
			}
		}
	}
	result = append(result, source[cursor:]...)
	return strings.Join(result, ""), nil
}

func TestSourceArchivePatchInterpreter(t *testing.T) {
	for _, pair := range []struct {
		before, patch, after string
	}{
		{"", "--- a/file\n+++ b/file\n@@ -0,0 +1 @@\n+é\r\n", "é\r\n"},
		{"old", "--- a/file\n+++ b/file\n@@ -1 +1 @@\n-old\n\\ No newline at end of file\n+new\n\\ No newline at end of file\n", "new"},
		{"old\n", "--- a/file\n+++ b/file\n@@ -1 +0,0 @@\n-old\n", ""},
		{"a\nb\nc\nd\ne\n", "--- a/file\n+++ b/file\n@@ -1 +1 @@\n-a\n+A\n@@ -5 +5 @@\n-e\n+E\n", "A\nb\nc\nd\nE\n"},
		{"--- file\n", "--- a/file\n+++ b/file\n@@ -1 +1,2 @@\n --- file\n++++ file\n", "--- file\n+++ file\n"},
	} {
		got, err := applySourceArchivePatch(pair.before, pair.patch)
		require.NoError(t, err)
		require.Equal(t, pair.after, got)
	}
	for _, patch := range []string{
		"@@ -1 +1 @@\n-a\n+b\n",
		"--- a/file\n+++ b/file\n@@ -1 +1 @@\n-not-a\n+b\n",
		"--- a/file\n+++ b/file\n@@ -1 +1 @@\n-a\n",
		"--- a/file\n+++ b/file\n@@ -1 +2 @@\n-a\n+b\n",
		"--- a/file\n+++ b/file\n@@ -1 +1 @@\n-a\n+b",
	} {
		_, err := applySourceArchivePatch("a\n", patch)
		require.Error(t, err, "%q", patch)
	}
}

func TestSourceArchivePatchPreservesExactText(t *testing.T) {
	for _, pair := range []struct {
		name, before, after string
	}{
		{"empty_to_line", "", "hello\n"},
		{"line_to_empty", "hello\n", ""},
		{"missing_final_newline", "old", "new"},
		{"remove_final_newline", "same\n", "same"},
		{"add_final_newline", "same", "same\n"},
		{"empty_to_newline", "", "\n"},
		{"crlf_unicode_headers", "--- old\r\n@@ -1 +1 @@\r\n\u20ac", "+++ new\r\n\\ No newline at end of file\r\n\U00010400"},
		{"separate_hunks", "old\n" + strings.Repeat("context\n", 16) + "tail", "new\n" + strings.Repeat("context\n", 16) + "TAIL"},
	} {
		t.Run(pair.name, func(t *testing.T) {
			old, head := NewSourceArchive(), NewSourceArchive()
			require.NoError(t, old.AddFile("file.txt", []byte(pair.before)))
			require.NoError(t, head.AddFile("file.txt", []byte(pair.after)))
			file := head.Diff(old).Files[0]
			require.False(t, file.PatchTruncated)
			require.NotEmpty(t, file.Patch)
			got, err := applySourceArchivePatch(pair.before, file.Patch)
			require.NoError(t, err, "%q", file.Patch)
			require.Equal(t, pair.after, got)
		})
	}
}

func sourceArchiveText() hegel.Generator[string] {
	return hegel.Composite(func(t hegel.TestCase) string {
		line := hegel.OneOf[string](
			hegel.Text().MaxSize(24).ExcludeCharacters("\r\n"),
			hegel.SampledFrom([]string{"", "--- a/file", "+++ b/file", "@@ -1 +1 @@", "\\ No newline at end of file", "\u20ac\U00010400\u00e9", "\x00"}),
		)
		lines := hegel.Draw(t, hegel.Lists(line).MaxSize(14))
		ending := hegel.Draw(t, hegel.SampledFrom([]string{"\n", "\r\n"}))
		text := strings.Join(lines, ending)
		if len(lines) > 0 && hegel.Draw(t, hegel.Booleans()) {
			text += ending
		}
		return text
	})
}

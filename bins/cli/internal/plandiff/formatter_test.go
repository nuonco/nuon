package plandiff

import (
	"strings"
	"testing"
)

func TestFormatSummary(t *testing.T) {
	tests := []struct {
		name     string
		summary  Summary
		contains []string
		excludes []string
	}{
		{
			name:     "empty summary",
			summary:  Summary{},
			contains: []string{"No changes"},
		},
		{
			name:     "create only",
			summary:  Summary{Create: 3},
			contains: []string{"3 to create"},
			excludes: []string{"to update", "to delete"},
		},
		{
			name:     "add only",
			summary:  Summary{Add: 2},
			contains: []string{"2 to add"},
		},
		{
			name:     "update only",
			summary:  Summary{Update: 1},
			contains: []string{"1 to update"},
		},
		{
			name:     "change only",
			summary:  Summary{Change: 4},
			contains: []string{"4 to change"},
		},
		{
			name:     "replace only",
			summary:  Summary{Replace: 1},
			contains: []string{"1 to replace"},
		},
		{
			name:     "delete only",
			summary:  Summary{Delete: 2},
			contains: []string{"2 to delete"},
		},
		{
			name:     "destroy only",
			summary:  Summary{Destroy: 3},
			contains: []string{"3 to destroy"},
		},
		{
			name:     "read only",
			summary:  Summary{Read: 1},
			contains: []string{"1 to read"},
		},
		{
			name:     "noop only",
			summary:  Summary{NoOp: 5},
			contains: []string{"5 unchanged"},
		},
		{
			name: "mixed terraform style",
			summary: Summary{
				Create:  2,
				Update:  1,
				Delete:  1,
				Replace: 1,
			},
			contains: []string{"2 to create", "1 to update", "1 to delete", "1 to replace", "Plan:"},
		},
		{
			name: "mixed k8s style",
			summary: Summary{
				Add:     3,
				Change:  2,
				Destroy: 1,
			},
			contains: []string{"3 to add", "2 to change", "1 to destroy"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := FormatSummary(tt.summary)

			for _, want := range tt.contains {
				if !strings.Contains(output, want) {
					t.Errorf("FormatSummary output missing %q\nOutput: %s", want, output)
				}
			}

			for _, exclude := range tt.excludes {
				if strings.Contains(output, exclude) {
					t.Errorf("FormatSummary output should not contain %q\nOutput: %s", exclude, output)
				}
			}
		})
	}
}

func TestColorizeAction(t *testing.T) {
	tests := []struct {
		action string
	}{
		{"create"},
		{"add"},
		{"added"},
		{"update"},
		{"change"},
		{"changed"},
		{"replace"},
		{"delete"},
		{"destroy"},
		{"destroyed"},
		{"read"},
		{"no-op"},
		{"unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			result := ColorizeAction(tt.action)
			if result == "" {
				t.Errorf("ColorizeAction(%q) returned empty string", tt.action)
			}
		})
	}
}

func TestColorizeActionSymbol(t *testing.T) {
	tests := []struct {
		action string
		want   string
	}{
		{"create", "+"},
		{"add", "+"},
		{"added", "+"},
		{"update", "~"},
		{"change", "~"},
		{"changed", "~"},
		{"replace", "±"},
		{"delete", "-"},
		{"destroy", "-"},
		{"destroyed", "-"},
		{"read", "←"},
		{"no-op", " "},
		{"unknown", " "},
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			result := ColorizeActionSymbol(tt.action)
			if !strings.Contains(result, tt.want) {
				t.Errorf("ColorizeActionSymbol(%q) = %q, want symbol %q", tt.action, result, tt.want)
			}
		})
	}
}

func TestFormatDiff(t *testing.T) {
	tests := []struct {
		name     string
		before   string
		after    string
		contains []string
	}{
		{
			name:     "empty both",
			before:   "",
			after:    "",
			contains: nil,
		},
		{
			name:     "only before",
			before:   "old value",
			after:    "",
			contains: []string{"- old value"},
		},
		{
			name:     "only after",
			before:   "",
			after:    "new value",
			contains: []string{"+ new value"},
		},
		{
			name:     "both values",
			before:   "old",
			after:    "new",
			contains: []string{"- old", "+ new"},
		},
		{
			name:     "multiline before",
			before:   "line1\nline2",
			after:    "",
			contains: []string{"- line1", "- line2"},
		},
		{
			name:     "multiline after",
			before:   "",
			after:    "line1\nline2",
			contains: []string{"+ line1", "+ line2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDiff(tt.before, tt.after)

			if len(tt.contains) == 0 && result != "" {
				t.Errorf("FormatDiff expected empty, got %q", result)
				return
			}

			for _, want := range tt.contains {
				if !strings.Contains(result, want) {
					t.Errorf("FormatDiff output missing %q\nOutput: %s", want, result)
				}
			}
		})
	}
}

func TestFormatUnifiedDiff(t *testing.T) {
	tests := []struct {
		name     string
		before   string
		after    string
		contains []string
	}{
		{
			name:     "empty both",
			before:   "",
			after:    "",
			contains: nil,
		},
		{
			name:     "added line",
			before:   "unchanged",
			after:    "unchanged\nnew line",
			contains: []string{"unchanged", "+ new line"},
		},
		{
			name:     "removed line",
			before:   "unchanged\nold line",
			after:    "unchanged",
			contains: []string{"- old line", "unchanged"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatUnifiedDiff(tt.before, tt.after)

			if len(tt.contains) == 0 && result != "" {
				t.Errorf("FormatUnifiedDiff expected empty, got %q", result)
				return
			}

			for _, want := range tt.contains {
				if !strings.Contains(result, want) {
					t.Errorf("FormatUnifiedDiff output missing %q\nOutput: %s", want, result)
				}
			}
		})
	}
}

func TestIndentString(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		indent int
		want   string
	}{
		{
			name:   "empty string",
			input:  "",
			indent: 4,
			want:   "",
		},
		{
			name:   "single line",
			input:  "hello",
			indent: 2,
			want:   "  hello",
		},
		{
			name:   "multiple lines",
			input:  "line1\nline2",
			indent: 4,
			want:   "    line1\n    line2",
		},
		{
			name:   "with empty line",
			input:  "line1\n\nline2",
			indent: 2,
			want:   "  line1\n\n  line2",
		},
		{
			name:   "zero indent",
			input:  "hello",
			indent: 0,
			want:   "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IndentString(tt.input, tt.indent)
			if got != tt.want {
				t.Errorf("IndentString(%q, %d) = %q, want %q", tt.input, tt.indent, got, tt.want)
			}
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{
			name:   "shorter than max",
			input:  "hello",
			maxLen: 10,
			want:   "hello",
		},
		{
			name:   "equal to max",
			input:  "hello",
			maxLen: 5,
			want:   "hello",
		},
		{
			name:   "longer than max",
			input:  "hello world",
			maxLen: 8,
			want:   "hello...",
		},
		{
			name:   "max len 0",
			input:  "hello",
			maxLen: 0,
			want:   "hello",
		},
		{
			name:   "max len 3 or less",
			input:  "hello",
			maxLen: 3,
			want:   "hel",
		},
		{
			name:   "negative max",
			input:  "hello",
			maxLen: -1,
			want:   "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TruncateString(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("TruncateString(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestFormatResourceHeader(t *testing.T) {
	tests := []struct {
		name         string
		resourceType string
		resourceName string
		action       string
		contains     []string
	}{
		{
			name:         "create action",
			resourceType: "aws_instance",
			resourceName: "example",
			action:       "create",
			contains:     []string{"+", "aws_instance", "example", "create"},
		},
		{
			name:         "delete action",
			resourceType: "Deployment",
			resourceName: "my-app",
			action:       "destroy",
			contains:     []string{"-", "Deployment", "my-app", "destroy"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatResourceHeader(tt.resourceType, tt.resourceName, tt.action)

			for _, want := range tt.contains {
				if !strings.Contains(result, want) {
					t.Errorf("FormatResourceHeader output missing %q\nOutput: %s", want, result)
				}
			}
		})
	}
}

func TestFormatSectionHeader(t *testing.T) {
	result := FormatSectionHeader("Test Section")
	if !strings.Contains(result, "Test Section") {
		t.Errorf("FormatSectionHeader missing title")
	}
	if !strings.Contains(result, "─") {
		t.Errorf("FormatSectionHeader missing underline")
	}
}

func TestFormatKeyValue(t *testing.T) {
	tests := []struct {
		name   string
		key    string
		value  any
		indent int
		want   string
	}{
		{
			name:   "string value",
			key:    "name",
			value:  "test",
			indent: 0,
			want:   "name: test",
		},
		{
			name:   "int value",
			key:    "count",
			value:  42,
			indent: 2,
			want:   "  count: 42",
		},
		{
			name:   "bool value",
			key:    "enabled",
			value:  true,
			indent: 4,
			want:   "    enabled: true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatKeyValue(tt.key, tt.value, tt.indent)
			if got != tt.want {
				t.Errorf("FormatKeyValue = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatChangeBlock(t *testing.T) {
	tests := []struct {
		name     string
		before   any
		after    any
		contains []string
	}{
		{
			name:     "both nil",
			before:   nil,
			after:    nil,
			contains: nil,
		},
		{
			name:     "string values",
			before:   "old",
			after:    "new",
			contains: []string{"before:", "old", "after:", "new"},
		},
		{
			name:     "map value",
			before:   map[string]any{"key1": "val1", "key2": "val2"},
			after:    nil,
			contains: []string{"before:", "2 fields"},
		},
		{
			name:     "slice value",
			before:   nil,
			after:    []any{"a", "b", "c"},
			contains: []string{"after:", "3 items"},
		},
		{
			name:     "long string truncated",
			before:   nil,
			after:    strings.Repeat("a", 150),
			contains: []string{"after:", "..."},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatChangeBlock(tt.before, tt.after)

			if len(tt.contains) == 0 && result != "" {
				t.Errorf("FormatChangeBlock expected empty, got %q", result)
				return
			}

			for _, want := range tt.contains {
				if !strings.Contains(result, want) {
					t.Errorf("FormatChangeBlock output missing %q\nOutput: %s", want, result)
				}
			}
		})
	}
}

func TestIsDestructiveAction(t *testing.T) {
	tests := []struct {
		action string
		want   bool
	}{
		{"delete", true},
		{"destroy", true},
		{"destroyed", true},
		{"replace", true},
		{"create", false},
		{"update", false},
		{"add", false},
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			if got := IsDestructiveAction(tt.action); got != tt.want {
				t.Errorf("IsDestructiveAction(%q) = %v, want %v", tt.action, got, tt.want)
			}
		})
	}
}

func TestIsCreativeAction(t *testing.T) {
	tests := []struct {
		action string
		want   bool
	}{
		{"create", true},
		{"add", true},
		{"added", true},
		{"delete", false},
		{"update", false},
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			if got := IsCreativeAction(tt.action); got != tt.want {
				t.Errorf("IsCreativeAction(%q) = %v, want %v", tt.action, got, tt.want)
			}
		})
	}
}

func TestIsModifyAction(t *testing.T) {
	tests := []struct {
		action string
		want   bool
	}{
		{"update", true},
		{"change", true},
		{"changed", true},
		{"replace", true},
		{"create", false},
		{"delete", false},
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			if got := IsModifyAction(tt.action); got != tt.want {
				t.Errorf("IsModifyAction(%q) = %v, want %v", tt.action, got, tt.want)
			}
		})
	}
}

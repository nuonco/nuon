package labels

import (
	"testing"
)

func TestParseLabelsQuery(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Labels
	}{
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: nil,
		},
		{
			name:     "single label",
			input:    "env:prod",
			expected: Labels{"env": "prod"},
		},
		{
			name:     "multiple labels",
			input:    "env:prod,team:platform",
			expected: Labels{"env": "prod", "team": "platform"},
		},
		{
			name:     "value with colon",
			input:    "url:http://example.com",
			expected: Labels{"url": "http://example.com"},
		},
		{
			name:     "whitespace around parts",
			input:    " env : prod , team : platform ",
			expected: Labels{"env": "prod", "team": "platform"},
		},
		{
			name:     "entry without separator is wildcard",
			input:    "env:prod,badentry,team:platform",
			expected: Labels{"env": "prod", "badentry": "*", "team": "platform"},
		},
		{
			name:     "all entries without separator are wildcards",
			input:    "env,team,region",
			expected: Labels{"env": "*", "team": "*", "region": "*"},
		},
		{
			name:     "equals separator",
			input:    "env=prod,team=platform",
			expected: Labels{"env": "prod", "team": "platform"},
		},
		{
			name:     "explicit wildcard value",
			input:    "env:*",
			expected: Labels{"env": "*"},
		},
		{
			name:     "empty value",
			input:    "env:",
			expected: Labels{"env": ""},
		},
		{
			name:     "trailing comma",
			input:    "env:prod,",
			expected: Labels{"env": "prod"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseLabelsQuery(tt.input)
			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
				return
			}
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d labels, got %d: %v", len(tt.expected), len(result), result)
				return
			}
			for k, v := range tt.expected {
				if result[k] != v {
					t.Errorf("expected %s=%s, got %s=%s", k, v, k, result[k])
				}
			}
		})
	}
}

func TestWithLabels_EmptyIsNoop(t *testing.T) {
	// WithLabels with empty labels should return a function that doesn't modify the db
	scope := WithLabels("labels", nil)
	if scope == nil {
		t.Fatal("expected non-nil scope function")
	}

	scope2 := WithLabels("labels", Labels{})
	if scope2 == nil {
		t.Fatal("expected non-nil scope function")
	}
}

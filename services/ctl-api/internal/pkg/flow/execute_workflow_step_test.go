package flow

import (
	"testing"
)

func TestGetcloneStepName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No retry suffix",
			input:    "Step Name",
			expected: "Step Name (retry 1)",
		},
		{
			name:     "Retry suffix with valid number",
			input:    "Step Name (retry 1)",
			expected: "Step Name (retry 2)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getCloneStepName(tt.input)
			if result != tt.expected {
				t.Errorf("incrementRetryName(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRemoveRetryFromStepName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No retry suffix",
			input:    "Step Name",
			expected: "Step Name",
		},
		{
			name:     "Retry suffix with number",
			input:    "Step Name (retry 1)",
			expected: "Step Name",
		},
		{
			name:     "Retry suffix with higher number",
			input:    "Step Name (retry 5)",
			expected: "Step Name",
		},
		{
			name:     "Retry suffix with spaces",
			input:    "Step Name  (retry 3)",
			expected: "Step Name",
		},
		{
			name:     "Multiple retry patterns",
			input:    "Step Name (retry 2) (retry 1)",
			expected: "Step Name (retry 2)",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Only retry suffix",
			input:    "(retry 1)",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeRetryFromStepName(tt.input)
			if result != tt.expected {
				t.Errorf("removeRetryFromStepName(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

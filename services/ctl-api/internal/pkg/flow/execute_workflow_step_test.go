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

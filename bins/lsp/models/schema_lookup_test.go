package models

import (
	"testing"
)

func TestDetectSchemaType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single comment with schema type",
			input:    "#helm\n",
			expected: "helm",
		},
		{
			name:     "comment with schema type and whitespace",
			input:    "  #helm  \n",
			expected: "helm",
		},
		{
			name:     "multiple comments, uses first",
			input:    "#helm\n#terraform\n",
			expected: "helm",
		},
		{
			name:     "comment then TOML content",
			input:    "#helm\n\n[public_repo]\nusername = \"test\"\n",
			expected: "helm",
		},
		{
			name:     "empty lines before comment",
			input:    "\n\n#terraform\n",
			expected: "terraform",
		},
		{
			name:     "no comment returns empty string",
			input:    "[public_repo]\nusername = \"test\"\n",
			expected: "",
		},
		{
			name:     "empty comment returns empty string",
			input:    "#\n[public_repo]\n",
			expected: "",
		},
		{
			name:     "only empty lines returns empty string",
			input:    "\n\n\n",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectSchemaType(tt.input)
			if result != tt.expected {
				t.Errorf("DetectSchemaType() = %q, want %q", result, tt.expected)
			}
		})
	}
}

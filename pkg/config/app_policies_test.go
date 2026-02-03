package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAppPolicy_SetNameFromSourceFile(t *testing.T) {
	tests := []struct {
		name         string
		sourceFile   string
		existingName string
		wantName     string
	}{
		{
			name:       "derives name from rego file",
			sourceFile: "policies/block-mutable-tags.rego",
			wantName:   "block-mutable-tags",
		},
		{
			name:       "derives name from yml file",
			sourceFile: "policies/no-pod-exec.yml",
			wantName:   "no-pod-exec",
		},
		{
			name:       "derives name from yaml file",
			sourceFile: "policies/restrict-secrets.yaml",
			wantName:   "restrict-secrets",
		},
		{
			name:       "handles file without directory",
			sourceFile: "my-policy.rego",
			wantName:   "my-policy",
		},
		{
			name:         "does not overwrite existing name",
			sourceFile:   "policies/block-mutable-tags.rego",
			existingName: "custom-name",
			wantName:     "custom-name",
		},
		{
			name:       "handles empty source file",
			sourceFile: "",
			wantName:   "",
		},
		{
			name:       "handles deeply nested path",
			sourceFile: "a/b/c/d/my-policy.rego",
			wantName:   "my-policy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := &AppPolicy{
				SourceFile: tt.sourceFile,
				Name:       tt.existingName,
			}
			policy.SetNameFromSourceFile()
			if policy.Name != tt.wantName {
				t.Errorf("SetNameFromSourceFile() = %q, want %q", policy.Name, tt.wantName)
			}
		})
	}
}

func TestAppPolicy_SetNameFromContents(t *testing.T) {
	tests := []struct {
		name         string
		contents     string
		existingName string
		wantName     string
	}{
		{
			name:     "derives name from relative rego file",
			contents: "./block-mutable-tags.rego",
			wantName: "block-mutable-tags",
		},
		{
			name:     "derives name from relative yml file",
			contents: "./no-pod-exec.yml",
			wantName: "no-pod-exec",
		},
		{
			name:     "derives name from parent directory path",
			contents: "../policies/my-policy.rego",
			wantName: "my-policy",
		},
		{
			name:     "derives name from absolute path",
			contents: "/app/policies/restrict-secrets.yaml",
			wantName: "restrict-secrets",
		},
		{
			name:         "does not overwrite existing name",
			contents:     "./block-mutable-tags.rego",
			existingName: "custom-name",
			wantName:     "custom-name",
		},
		{
			name:     "does not derive from URL",
			contents: "https://example.com/policy.rego",
			wantName: "",
		},
		{
			name:     "does not derive from inline content",
			contents: "package nuon\n\ndefault allow := false",
			wantName: "",
		},
		{
			name:     "handles empty contents",
			contents: "",
			wantName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := &AppPolicy{
				Contents: tt.contents,
				Name:     tt.existingName,
			}
			policy.SetNameFromContents()
			require.Equal(t, tt.wantName, policy.Name)
		})
	}
}

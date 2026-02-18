package cloudformation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCustomNestedStackS3Key(t *testing.T) {
	tests := []struct {
		name         string
		orgID        string
		appID        string
		contentsHash string
		templateURL  string
		expected     string
	}{
		{
			name:         "yaml extension",
			orgID:        "org123",
			appID:        "app456",
			contentsHash: "abc123",
			templateURL:  "https://example.com/template.yaml",
			expected:     "stacks/org123/app456/abc123.yaml",
		},
		{
			name:         "json extension",
			orgID:        "org123",
			appID:        "app456",
			contentsHash: "abc123",
			templateURL:  "https://example.com/template.json",
			expected:     "stacks/org123/app456/abc123.json",
		},
		{
			name:         "yml extension",
			orgID:        "org123",
			appID:        "app456",
			contentsHash: "abc123",
			templateURL:  "https://example.com/template.yml",
			expected:     "stacks/org123/app456/abc123.yml",
		},
		{
			name:         "no recognized extension defaults to yaml",
			orgID:        "org123",
			appID:        "app456",
			contentsHash: "abc123",
			templateURL:  "https://example.com/template",
			expected:     "stacks/org123/app456/abc123.yaml",
		},
		{
			name:         "case insensitive extension",
			orgID:        "org123",
			appID:        "app456",
			contentsHash: "abc123",
			templateURL:  "https://example.com/template.JSON",
			expected:     "stacks/org123/app456/abc123.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CustomNestedStackS3Key(tt.orgID, tt.appID, tt.contentsHash, tt.templateURL)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCustomNestedStackTemplateURL(t *testing.T) {
	tests := []struct {
		name         string
		baseURL      string
		orgID        string
		appID        string
		contentsHash string
		templateURL  string
		expected     string
	}{
		{
			name:         "base URL without trailing slash",
			baseURL:      "https://s3.amazonaws.com/my-bucket",
			orgID:        "org123",
			appID:        "app456",
			contentsHash: "abc123",
			templateURL:  "https://example.com/template.yaml",
			expected:     "https://s3.amazonaws.com/my-bucket/stacks/org123/app456/abc123.yaml",
		},
		{
			name:         "base URL with trailing slash",
			baseURL:      "https://s3.amazonaws.com/my-bucket/",
			orgID:        "org123",
			appID:        "app456",
			contentsHash: "abc123",
			templateURL:  "https://example.com/template.json",
			expected:     "https://s3.amazonaws.com/my-bucket/stacks/org123/app456/abc123.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CustomNestedStackTemplateURL(tt.baseURL, tt.orgID, tt.appID, tt.contentsHash, tt.templateURL)
			assert.Equal(t, tt.expected, result)
		})
	}
}

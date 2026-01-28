package dockerhub

import "testing"

func TestNormalizeReference(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "official image - single name",
			input:    "nginx",
			expected: "docker.io/library/nginx",
		},
		{
			name:     "official image - alpine",
			input:    "alpine",
			expected: "docker.io/library/alpine",
		},
		{
			name:     "user image - two parts",
			input:    "myuser/myimage",
			expected: "docker.io/myuser/myimage",
		},
		{
			name:     "docker.io already qualified",
			input:    "docker.io/library/nginx",
			expected: "docker.io/library/nginx",
		},
		{
			name:     "docker.io user image",
			input:    "docker.io/myuser/myimage",
			expected: "docker.io/myuser/myimage",
		},
		{
			name:     "docker.io official image needs library prefix",
			input:    "docker.io/nginx",
			expected: "docker.io/library/nginx",
		},
		{
			name:     "registry-1.docker.io official image",
			input:    "registry-1.docker.io/nginx",
			expected: "registry-1.docker.io/library/nginx",
		},
		{
			name:     "index.docker.io official image",
			input:    "index.docker.io/nginx",
			expected: "index.docker.io/library/nginx",
		},
		{
			name:     "gcr.io registry",
			input:    "gcr.io/myproject/myimage",
			expected: "gcr.io/myproject/myimage",
		},
		{
			name:     "ghcr.io registry",
			input:    "ghcr.io/owner/repo",
			expected: "ghcr.io/owner/repo",
		},
		{
			name:     "ECR registry",
			input:    "123456789012.dkr.ecr.us-east-1.amazonaws.com/myrepo",
			expected: "123456789012.dkr.ecr.us-east-1.amazonaws.com/myrepo",
		},
		{
			name:     "quay.io registry",
			input:    "quay.io/myorg/myimage",
			expected: "quay.io/myorg/myimage",
		},
		{
			name:     "localhost registry",
			input:    "localhost/myimage",
			expected: "localhost/myimage",
		},
		{
			name:     "localhost with port",
			input:    "localhost:5000/myimage",
			expected: "localhost:5000/myimage",
		},
		{
			name:     "registry with port",
			input:    "myregistry:5000/myimage",
			expected: "myregistry:5000/myimage",
		},
		{
			name:     "nested path in registry",
			input:    "gcr.io/myproject/subdir/myimage",
			expected: "gcr.io/myproject/subdir/myimage",
		},
		{
			name:     "https prefix stripped",
			input:    "https://docker.io/library/nginx",
			expected: "docker.io/library/nginx",
		},
		{
			name:     "http prefix stripped",
			input:    "http://localhost:5000/myimage",
			expected: "localhost:5000/myimage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeReference(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeReference(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsDockerHubRegistry(t *testing.T) {
	tests := []struct {
		host     string
		expected bool
	}{
		{"docker.io", true},
		{"registry-1.docker.io", true},
		{"index.docker.io", true},
		{"gcr.io", false},
		{"ghcr.io", false},
		{"quay.io", false},
		{"localhost", false},
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			result := IsDockerHubRegistry(tt.host)
			if result != tt.expected {
				t.Errorf("IsDockerHubRegistry(%q) = %v, want %v", tt.host, result, tt.expected)
			}
		})
	}
}

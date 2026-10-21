package oidc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		value   string
		want    bool
	}{
		{
			name:    "exact match",
			pattern: "repo:acme/app:ref:refs/heads/main",
			value:   "repo:acme/app:ref:refs/heads/main",
			want:    true,
		},
		{
			name:    "exact mismatch",
			pattern: "repo:acme/app:ref:refs/heads/main",
			value:   "repo:acme/app:ref:refs/heads/dev",
			want:    false,
		},
		{
			name:    "glob within segment",
			pattern: "repo:acme/*:ref:refs/heads/main",
			value:   "repo:acme/app:ref:refs/heads/main",
			want:    true,
		},
		{
			name:    "glob cannot cross colon segments",
			pattern: "repo:acme/*",
			value:   "repo:acme/app:ref:refs/heads/main",
			want:    false,
		},
		{
			name:    "glob cannot smuggle extra segments mid-pattern",
			pattern: "repo:acme/*:ref:refs/heads/main",
			value:   "repo:acme/app:extra:ref:refs/heads/main",
			want:    false,
		},
		{
			name:    "trailing multi-segment wildcard",
			pattern: "repo:acme/app:*:*",
			value:   "repo:acme/app:ref:refs/heads/main",
			want:    true,
		},
		{
			name:    "glob matches nested branch paths since slash is not a separator",
			pattern: "repo:acme/app:ref:refs/heads/*",
			value:   "repo:acme/app:ref:refs/heads/feature/x",
			want:    true,
		},
		{
			name:    "question mark is a metacharacter",
			pattern: "repo:acme/ap?:ref:refs/heads/main",
			value:   "repo:acme/app:ref:refs/heads/main",
			want:    true,
		},
		{
			name:    "empty value does not match wildcard-free pattern",
			pattern: "repo:acme/app",
			value:   "",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, MatchPattern(tt.pattern, tt.value))
		})
	}
}

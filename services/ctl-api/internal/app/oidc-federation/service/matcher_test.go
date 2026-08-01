package service

import (
	"strings"
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
			require.Equal(t, tt.want, matchPattern(tt.pattern, tt.value))
		})
	}
}

func TestMatchClaims(t *testing.T) {
	claims := map[string]any{
		"sub":              "repo:acme/app:ref:refs/heads/main",
		"repository_owner": "acme",
		"ref":              "refs/heads/main",
		"actor_id":         float64(12345),
	}

	tests := []struct {
		name       string
		conditions map[string]string
		want       bool
	}{
		{
			name:       "single condition matches",
			conditions: map[string]string{"sub": "repo:acme/app:ref:refs/heads/main"},
			want:       true,
		},
		{
			name: "all conditions must match",
			conditions: map[string]string{
				"sub":              "repo:acme/*:ref:refs/heads/main",
				"repository_owner": "acme",
			},
			want: true,
		},
		{
			name: "one failing condition fails the policy",
			conditions: map[string]string{
				"sub":              "repo:acme/*:ref:refs/heads/main",
				"repository_owner": "other",
			},
			want: false,
		},
		{
			name:       "missing claim fails",
			conditions: map[string]string{"sub": "repo:*", "environment": "prod"},
			want:       false,
		},
		{
			name:       "non-string claim fails",
			conditions: map[string]string{"sub": "repo:acme/app:ref:refs/heads/main", "actor_id": "12345"},
			want:       false,
		},
		{
			name:       "empty conditions never match",
			conditions: map[string]string{},
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, matchClaims(tt.conditions, claims))
		})
	}
}

func TestValidateClaimConditions(t *testing.T) {
	tests := []struct {
		name       string
		conditions map[string]string
		wantErr    string
	}{
		{
			name:       "valid single sub condition",
			conditions: map[string]string{"sub": "repo:acme/app:*:*"},
		},
		{
			name:       "valid multiple conditions",
			conditions: map[string]string{"sub": "repo:acme/app:*:*", "repository_owner": "acme"},
		},
		{
			name:       "empty conditions rejected",
			conditions: map[string]string{},
			wantErr:    "at least one claim condition",
		},
		{
			name:       "sub required",
			conditions: map[string]string{"repository_owner": "acme"},
			wantErr:    "`sub` claim is required",
		},
		{
			name:       "reserved claim rejected",
			conditions: map[string]string{"sub": "repo:*", "iss": "https://evil.example.com"},
			wantErr:    "validated automatically",
		},
		{
			name:       "empty pattern rejected",
			conditions: map[string]string{"sub": ""},
			wantErr:    "cannot be empty",
		},
		{
			name:       "oversized pattern rejected",
			conditions: map[string]string{"sub": strings.Repeat("a", 513)},
			wantErr:    "exceeds 512 characters",
		},
		{
			name:       "invalid glob rejected",
			conditions: map[string]string{"sub": "repo:[acme"},
			wantErr:    "not a valid pattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateClaimConditions(tt.conditions)
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}

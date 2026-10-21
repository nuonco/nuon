package oidc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMatchClaims(t *testing.T) {
	claims, err := DecodeClaims(map[string]any{
		"sub":              "repo:acme/app:ref:refs/heads/main",
		"repository_owner": "acme",
		"ref":              "refs/heads/main",
		"actor_id":         float64(12345),
		"aud":              "nuon",
		"exp":              float64(1700000000),
	})
	require.NoError(t, err)

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
			require.Equal(t, tt.want, ParseConditions(tt.conditions).Match(claims))
		})
	}
}

func TestDecodeClaims(t *testing.T) {
	claims, err := DecodeClaims(map[string]any{
		"iss":      "https://idp.example.com",
		"sub":      "acme",
		"aud":      []any{"a", "b"},
		"exp":      float64(1700000000),
		"actor_id": float64(1),
		"ref":      "main",
	})
	require.NoError(t, err)
	require.Equal(t, Audience{"a", "b"}, claims.Aud)
	require.Equal(t, int64(1700000000), claims.Exp)
	require.Equal(t, []string{"actor_id", "aud", "exp", "iss", "ref", "sub"}, claims.Names())

	single, err := DecodeClaims(map[string]any{"aud": "a"})
	require.NoError(t, err)
	require.Equal(t, Audience{"a"}, single.Aud)

	_, err = DecodeClaims(map[string]any{"sub": float64(1)})
	require.Error(t, err)
}

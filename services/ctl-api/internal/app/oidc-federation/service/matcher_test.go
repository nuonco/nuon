package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidc"
)

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
			err := validateClaimConditions("https://idp.example.com", tt.conditions)
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}

const (
	testGitHubIssuer = "https://token.actions.githubusercontent.com"
	testGitLabIssuer = "https://gitlab.example.com"
)

func TestValidateClaimConditionsGitHubIssuer(t *testing.T) {
	sub := func(p string) map[string]string { return map[string]string{"sub": p} }

	accepted := []struct {
		name       string
		conditions map[string]string
	}{
		{"exact ref", sub("repo:acme/app:ref:refs/heads/main")},
		{"wildcard after owner", sub("repo:acme/app:**")},
		{"wildcard repo", sub("repo:acme/*:ref:refs/heads/main")},
		{"immutable ids", sub("repo:acme@123/app@456:ref:refs/heads/main")},
		{"broad sub with owner id", map[string]string{"sub": "repo:*/app:**", "repository_owner_id": "123456"}},
	}

	rejected := []struct {
		name       string
		conditions map[string]string
	}{
		{"double star", sub("**")},
		{"repo double star", sub("repo:**")},
		{"all wildcard segments", sub("repo:*:*:*")},
		{"wildcard owner and repo", sub("repo:*/*:ref:*")},
		{"wildcard owner", sub("repo:*/app:ref:refs/heads/main")},
		{"partial wildcard owner", sub("repo:acme*/app:ref:refs/heads/main")},
		{"brace owner", sub("repo:{acme,evil}/app:**")},
		{"wildcard owner id", map[string]string{"sub": "repo:*/app:**", "repository_owner_id": "12*"}},
		{"non-numeric owner id", map[string]string{"sub": "repo:*/app:**", "repository_owner_id": "acme"}},
	}

	for _, tt := range accepted {
		t.Run("accepted/"+tt.name, func(t *testing.T) {
			require.NoError(t, validateClaimConditions(testGitHubIssuer, tt.conditions))
		})
	}
	for _, tt := range rejected {
		t.Run("rejected/"+tt.name, func(t *testing.T) {
			require.Error(t, validateClaimConditions(testGitHubIssuer, tt.conditions))
		})
	}

	t.Run("trailing slash tolerated", func(t *testing.T) {
		require.Error(t, validateClaimConditions(testGitHubIssuer+"/", sub("repo:*/app:ref:refs/heads/main")))
		require.NoError(t, validateClaimConditions(testGitHubIssuer+"/", sub("repo:acme/app:**")))
	})

	t.Run("accepted patterns without owner id never match a foreign repo", func(t *testing.T) {
		for _, tt := range accepted {
			if _, ok := tt.conditions["repository_owner_id"]; ok {
				continue
			}
			require.False(t, oidc.MatchPattern(tt.conditions["sub"], "repo:evil-org/app:ref:refs/heads/main"), tt.name)
		}
	})
}

func TestValidateClaimConditionsOtherIssuer(t *testing.T) {
	tests := []struct {
		name    string
		sub     string
		wantErr bool
	}{
		{"double star rejected", "**", true},
		{"single star rejected", "*", true},
		{"specific value accepted", "project_path:acme/app:ref_type:branch:ref:main", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateClaimConditions(testGitLabIssuer, map[string]string{"sub": tt.sub})
			if tt.wantErr {
				require.ErrorContains(t, err, "too broad")
				return
			}
			require.NoError(t, err)
		})
	}
}

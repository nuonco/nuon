package oidc_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidc"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidc/github"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

func TestValidateConditions(t *testing.T) {
	v := validator.New()
	providers := oidc.NewProviders(github.Provider{})

	tests := map[string]struct {
		issuer     string
		conditions map[string]string
		wantErr    string
	}{
		"valid single sub condition": {
			issuer:     "https://idp.example.com",
			conditions: map[string]string{"sub": "repo:acme/app:*:*"},
		},
		"valid multiple conditions": {
			issuer:     "https://idp.example.com",
			conditions: map[string]string{"sub": "repo:acme/app:*:*", "repository_owner": "acme"},
		},
		"any non-reserved claim accepted for other issuers": {
			issuer:     "https://idp.example.com",
			conditions: map[string]string{"sub": "acme-ci", "custom_claim": "x"},
		},
		"empty conditions rejected": {
			issuer:     "https://idp.example.com",
			conditions: map[string]string{},
			wantErr:    "`sub` claim is required",
		},
		"sub required": {
			issuer:     "https://idp.example.com",
			conditions: map[string]string{"repository_owner": "acme"},
			wantErr:    "`sub` claim is required",
		},
		"reserved claim rejected": {
			issuer:     "https://idp.example.com",
			conditions: map[string]string{"sub": "repo:*", "iss": "https://evil.example.com"},
			wantErr:    `claim "iss" cannot be used`,
		},
		"empty claim name rejected": {
			issuer:     "https://idp.example.com",
			conditions: map[string]string{"sub": "acme-ci", "": "x"},
			wantErr:    "cannot be used as a condition",
		},
		"empty pattern rejected": {
			issuer:     "https://idp.example.com",
			conditions: map[string]string{"sub": "acme-ci", "team": ""},
			wantErr:    `claim "team" cannot be empty`,
		},
		"max length pattern accepted": {
			issuer:     "https://idp.example.com",
			conditions: map[string]string{"sub": strings.Repeat("a", 512)},
		},
		"oversized pattern rejected": {
			issuer:     "https://idp.example.com",
			conditions: map[string]string{"sub": strings.Repeat("a", 513)},
			wantErr:    "exceeds 512 characters",
		},
		"invalid glob rejected": {
			issuer:     "https://idp.example.com",
			conditions: map[string]string{"sub": "repo:[acme"},
			wantErr:    "not a valid pattern",
		},
		"double star too broad": {
			issuer:     "https://idp.example.com",
			conditions: map[string]string{"sub": "**"},
			wantErr:    "too broad",
		},
		"single star too broad": {
			issuer:     "https://idp.example.com",
			conditions: map[string]string{"sub": "*"},
			wantErr:    "too broad",
		},
		"specific value accepted": {
			issuer:     "https://idp.example.com",
			conditions: map[string]string{"sub": "project_path:acme/app:ref_type:branch:ref:main"},
		},
		"github exact ref": {
			issuer:     github.IssuerURL,
			conditions: map[string]string{"sub": "repo:acme/app:ref:refs/heads/main"},
		},
		"github wildcard after owner": {
			issuer:     github.IssuerURL,
			conditions: map[string]string{"sub": "repo:acme/app:**"},
		},
		"github wildcard repo": {
			issuer:     github.IssuerURL,
			conditions: map[string]string{"sub": "repo:acme/*:ref:refs/heads/main"},
		},
		"github immutable ids": {
			issuer:     github.IssuerURL,
			conditions: map[string]string{"sub": "repo:acme@123/app@456:ref:refs/heads/main"},
		},
		"github wildcard owner with owner id": {
			issuer:     github.IssuerURL,
			conditions: map[string]string{"sub": "repo:*/app:**", "repository_owner_id": "123456"},
		},
		"github trailing slash issuer pinned": {
			issuer:     github.IssuerURL + "/",
			conditions: map[string]string{"sub": "repo:acme/app:**"},
		},
		"github trailing slash issuer unpinned": {
			issuer:     github.IssuerURL + "/",
			conditions: map[string]string{"sub": "repo:*/app:ref:refs/heads/main"},
			wantErr:    "must pin an owner",
		},
		"github double star": {
			issuer:     github.IssuerURL,
			conditions: map[string]string{"sub": "**"},
			wantErr:    "must pin an owner",
		},
		"github wildcard owner": {
			issuer:     github.IssuerURL,
			conditions: map[string]string{"sub": "repo:*/app:ref:refs/heads/main"},
			wantErr:    "must pin an owner",
		},
		"github partial wildcard owner": {
			issuer:     github.IssuerURL,
			conditions: map[string]string{"sub": "repo:acme*/app:ref:refs/heads/main"},
			wantErr:    "must pin an owner",
		},
		"github brace owner": {
			issuer:     github.IssuerURL,
			conditions: map[string]string{"sub": "repo:{acme,evil}/app:**"},
			wantErr:    "must pin an owner",
		},
		"github wildcard owner id": {
			issuer:     github.IssuerURL,
			conditions: map[string]string{"sub": "repo:*/app:**", "repository_owner_id": "12*"},
			wantErr:    "must pin an owner",
		},
		"github negative owner id": {
			issuer:     github.IssuerURL,
			conditions: map[string]string{"sub": "repo:*/app:**", "repository_owner_id": "-1"},
			wantErr:    "must pin an owner",
		},
		"github non-numeric owner id": {
			issuer:     github.IssuerURL,
			conditions: map[string]string{"sub": "repo:*/app:**", "repository_owner_id": "acme"},
			wantErr:    "must pin an owner",
		},
		"github unknown claim rejected": {
			issuer:     github.IssuerURL,
			conditions: map[string]string{"sub": "repo:acme/app:**", "repository_ownr": "acme"},
			wantErr:    `claim "repository_ownr" is not a GitHub Actions token claim`,
		},
		"github known claim accepted": {
			issuer:     github.IssuerURL,
			conditions: map[string]string{"sub": "repo:acme/app:**", "workflow_ref": "acme/app/.github/workflows/ci.yml@refs/heads/main", "jti": "*"},
		},
		"github reserved claim rejected": {
			issuer:     github.IssuerURL,
			conditions: map[string]string{"sub": "repo:acme/app:**", "aud": "nuon"},
			wantErr:    `claim "aud" cannot be used`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := oidc.ValidateConditions(v, providers, tt.issuer, tt.conditions)
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}

			require.ErrorContains(t, validator.FormatValidationError(err), tt.wantErr)
		})
	}
}

func TestAcceptedOwnerPinnedPatternsNeverMatchForeignOwner(t *testing.T) {
	v := validator.New()
	providers := oidc.NewProviders(github.Provider{})

	patterns := []string{
		"repo:acme/app:ref:refs/heads/main",
		"repo:acme/app:**",
		"repo:acme/*:ref:refs/heads/main",
		"repo:acme/app:*:*",
		"repo:acme@123/app@456:ref:refs/heads/main",
		"repo:acme@123/*:**",
	}

	foreign := github.NewSubject("evil-org", "app", "ref:refs/heads/main")
	foreignIDs := foreign.WithIDs("999", "456")

	for _, pattern := range patterns {
		t.Run(pattern, func(t *testing.T) {
			require.NoError(t, oidc.ValidateConditions(v, providers, github.IssuerURL, map[string]string{"sub": pattern}))
			require.False(t, oidc.MatchPattern(pattern, foreign.String()))
			require.False(t, oidc.MatchPattern(pattern, foreignIDs.String()))
		})
	}
}

func TestClaimNames(t *testing.T) {
	names := oidc.ClaimNames(github.Claims{})

	for _, want := range []string{
		"sub", "aud", "iss", "exp", "iat", "nbf", "jti", "actor", "actor_id", "base_ref", "environment",
		"environment_node_id", "event_name", "head_ref", "job_workflow_ref", "job_workflow_sha", "ref",
		"ref_protected", "ref_type", "repository", "repository_id", "repository_owner", "repository_owner_id",
		"repository_visibility", "run_attempt", "run_id", "run_number", "runner_environment", "sha", "workflow",
		"workflow_ref", "workflow_sha", "enterprise", "enterprise_id",
	} {
		require.Contains(t, names, want)
	}
	require.NotContains(t, names, "")
}

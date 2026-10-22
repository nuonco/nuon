package github_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidc"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidc/github"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

func TestProviderProbes(t *testing.T) {
	probes := github.Provider{}.Probes()

	require.Contains(t, probes, "repo:nuonprobeqz7a/nuonprobekx3b:ref:refs/heads/main")
	require.Contains(t, probes, "repo:nuonprobeqz7a@1/nuonprobekx3b@2:ref:refs/heads/main")
}

func TestProviderValidateConditions(t *testing.T) {
	v := validator.New()

	tests := map[string]struct {
		conditions map[string]string
		wantErr    string
	}{
		"literal owner": {
			conditions: map[string]string{"sub": "repo:acme/app:**"},
		},
		"owner id": {
			conditions: map[string]string{"sub": "repo:*/app:**", "repository_owner_id": "42"},
		},
		"unpinned": {
			conditions: map[string]string{"sub": "repo:*/app:**"},
			wantErr:    "must pin an owner",
		},
		"unknown claim": {
			conditions: map[string]string{"sub": "repo:acme/app:**", "nope": "x"},
			wantErr:    `claim "nope" is not a GitHub Actions token claim`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := github.Provider{}.ValidateConditions(v, oidc.ParseConditions(tt.conditions))
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}

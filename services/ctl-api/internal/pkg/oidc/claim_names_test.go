package oidc_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidc"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidc/github"
)

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

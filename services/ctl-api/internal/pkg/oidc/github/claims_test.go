package github

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeClaims(t *testing.T) {
	claims, err := DecodeClaims(map[string]any{
		"sub":                 "repo:acme/app:ref:refs/heads/main",
		"aud":                 "nuon",
		"repository_owner_id": "123",
		"ref":                 "refs/heads/main",
		"custom":              "x",
	})
	require.NoError(t, err)
	require.Equal(t, "repo:acme/app:ref:refs/heads/main", claims.Sub)
	require.Equal(t, "123", claims.RepositoryOwnerID)
	require.Equal(t, "refs/heads/main", claims.Ref)
	require.Equal(t, map[string]any{"custom": "x"}, claims.Extra)
}

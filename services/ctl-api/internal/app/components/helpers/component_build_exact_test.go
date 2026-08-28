package helpers

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestValidateExactBuildConfig(t *testing.T) {
	ref := "0123456789abcdef0123456789abcdef01234567"
	req := CreateComponentBuildFromConfigConnectionRequest{
		ComponentConfigConnectionID: "config-id",
		ResolvedGitCommitSHA:        &ref,
	}
	valid := func() *app.ComponentConfigConnection {
		return &app.ComponentConfigConnection{
			ID:                "config-id",
			ComponentID:       "component-id",
			Component:         app.Component{ID: "component-id"},
			Type:              app.ComponentTypeHelmChart,
			VCSConnectionType: app.VCSConnectionTypeConnectedRepo,
		}
	}

	t.Run("accepts exact config and resolved ref", func(t *testing.T) {
		require.NoError(t, validateExactBuildConfig(req, valid()))
	})
	t.Run("rejects missing config", func(t *testing.T) {
		require.ErrorContains(t, validateExactBuildConfig(req, nil), "no config found")
	})
	t.Run("rejects config without concrete component config", func(t *testing.T) {
		config := valid()
		config.Type = app.ComponentTypeUnknown
		require.ErrorContains(t, validateExactBuildConfig(req, config), "no config found")
	})
	t.Run("rejects mismatched component", func(t *testing.T) {
		config := valid()
		config.ComponentID = "other-component"
		require.ErrorContains(t, validateExactBuildConfig(req, config), "valid owning component")
	})
	t.Run("rejects VCS config without resolved ref", func(t *testing.T) {
		missingRef := req
		missingRef.ResolvedGitCommitSHA = nil
		require.ErrorContains(t, validateExactBuildConfig(missingRef, valid()), "resolved Git commit SHA is required")
	})
	t.Run("rejects mutable branch", func(t *testing.T) {
		branch := "main"
		mutable := req
		mutable.ResolvedGitCommitSHA = &branch
		require.ErrorContains(t, validateExactBuildConfig(mutable, valid()), "resolved Git commit SHA is required")
	})
	t.Run("does not require a ref for non-VCS config", func(t *testing.T) {
		config := valid()
		config.VCSConnectionType = app.VCSConnectionTypeNone
		noRef := req
		noRef.ResolvedGitCommitSHA = nil
		require.NoError(t, validateExactBuildConfig(noRef, config))
	})
}

func TestValidateExactBuildRequest(t *testing.T) {
	require.Error(t, validateExactBuildRequest(CreateComponentBuildFromConfigConnectionRequest{}))
	require.NoError(t, validateExactBuildRequest(CreateComponentBuildFromConfigConnectionRequest{
		ComponentConfigConnectionID: "config-id",
	}))
}

func TestNormalizeExactBuildRequest(t *testing.T) {
	sha := "  ABCDEF0123456789ABCDEF0123456789ABCDEF01  "
	req := normalizeExactBuildRequest(CreateComponentBuildFromConfigConnectionRequest{ResolvedGitCommitSHA: &sha})
	require.Equal(t, "abcdef0123456789abcdef0123456789abcdef01", *req.ResolvedGitCommitSHA)
}

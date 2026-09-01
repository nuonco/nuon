package helpers

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestPreviewGitBaseUsesVCSBranchNotAppBranchName(t *testing.T) {
	branch := &app.AppBranch{Name: "seed"}

	require.Equal(t, "main", previewGitBase(branch, &app.AppBranchConfig{
		PublicGitVCSConfig: &app.PublicGitVCSConfig{Branch: "main"},
	}))

	require.Equal(t, "main", previewGitBase(branch, &app.AppBranchConfig{
		ConnectedGithubVCSConfig: &app.ConnectedGithubVCSConfig{Branch: "main"},
	}))

	require.Equal(t, "seed", previewGitBase(branch, &app.AppBranchConfig{}))
	require.Equal(t, "seed", previewGitBase(branch, nil))
}

package activities

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestRunCommitSHA(t *testing.T) {
	sha := "abc123"
	require.Equal(t, sha, runCommitSHA(&app.AppBranchRun{
		VCSConnectionCommit: &app.VCSConnectionCommit{SHA: sha},
		HeadSHA:             "ignored",
	}))
	require.Equal(t, "head", runCommitSHA(&app.AppBranchRun{HeadSHA: "head"}))
	require.Equal(t, "", runCommitSHA(&app.AppBranchRun{}))
}

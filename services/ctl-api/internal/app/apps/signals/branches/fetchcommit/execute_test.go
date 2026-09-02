package fetchcommit

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestPreviewCommitRef(t *testing.T) {
	require.Empty(t, previewCommitRef(&app.AppBranchRun{
		RunType: app.AppBranchRunTypeManual,
		HeadSHA: "manual-sha",
	}))
	require.Equal(t, "head-sha", previewCommitRef(&app.AppBranchRun{
		RunType: app.AppBranchRunTypeGitPreview,
		HeadSHA: "head-sha",
		Preview: &app.AppBranchRunPreview{GitRef: "feature/payments"},
	}))
	require.Equal(t, "feature/payments", previewCommitRef(&app.AppBranchRun{
		RunType: app.AppBranchRunTypeGitPreview,
		Preview: &app.AppBranchRunPreview{GitRef: "feature/payments"},
	}))
}

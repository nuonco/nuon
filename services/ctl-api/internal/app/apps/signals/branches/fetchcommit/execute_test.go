package fetchcommit

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestPreviewCommitRef(t *testing.T) {
	manual := &app.AppBranchRun{
		RunType: app.AppBranchRunTypeManual,
		HeadSHA: "manual-sha",
	}
	require.Empty(t, previewCommitRef(manual))
	pin, ok := pinnedManualCommit(manual)
	require.True(t, ok)
	require.Equal(t, "manual-sha", pin.ref)
	require.Zero(t, pin.prNumber)

	tagged := &app.AppBranchRun{
		RunType:  app.AppBranchRunTypeManual,
		Metadata: app.AppBranchRunMetadata{Tag: "v1.2.3", GitRef: "v1.2.3"},
	}
	pin, ok = pinnedManualCommit(tagged)
	require.True(t, ok)
	require.Equal(t, "v1.2.3", pin.ref)

	pr := 42
	pull := &app.AppBranchRun{
		RunType:  app.AppBranchRunTypeManual,
		PRNumber: &pr,
	}
	pin, ok = pinnedManualCommit(pull)
	require.True(t, ok)
	require.Equal(t, 42, pin.prNumber)
	require.Empty(t, pin.ref)

	_, ok = pinnedManualCommit(&app.AppBranchRun{RunType: app.AppBranchRunTypeManual})
	require.False(t, ok)
	_, ok = pinnedManualCommit(&app.AppBranchRun{RunType: app.AppBranchRunTypeGit, HeadSHA: "push-sha"})
	require.False(t, ok)

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

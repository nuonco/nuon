package apps

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

func TestCollectPreviewInputFailuresReportsBranchAndInstallTogether(t *testing.T) {
	_, err := collectPreviewInputFailures(
		PreviewBranchRunOptions{Mode: "apply", InstallID: "inl-missing"},
		"",
		fmt.Errorf(`app branch "daily" not found`),
		fmt.Errorf(`install "inl-missing" not found`),
	)

	require.Error(t, err)
	var userErr *ui.CLIUserError
	require.ErrorAs(t, err, &userErr)
	require.Contains(t, userErr.Msg, `app branch "daily" not found`)
	require.Contains(t, userErr.Msg, `install "inl-missing" not found`)
}

func TestCollectPreviewInputFailuresReturnsResolvedBranch(t *testing.T) {
	branchID, err := collectPreviewInputFailures(
		PreviewBranchRunOptions{Mode: "apply", InstallID: "inl-1"},
		"abr-1",
		nil,
		nil,
	)

	require.NoError(t, err)
	require.Equal(t, "abr-1", branchID)
}

func TestCollectPreviewInputFailuresIncludesInvalidMode(t *testing.T) {
	_, err := collectPreviewInputFailures(
		PreviewBranchRunOptions{Mode: "nope"},
		"abr-1",
		nil,
		nil,
	)

	require.Error(t, err)
	require.ErrorContains(t, err, `invalid mode "nope"`)
}

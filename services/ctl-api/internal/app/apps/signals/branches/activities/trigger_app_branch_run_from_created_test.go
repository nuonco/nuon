package activities

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSkipFirstAppBranchRun(t *testing.T) {
	t.Parallel()

	skip, reason := skipFirstAppBranchRun(true, true, false)
	require.True(t, skip)
	require.Equal(t, "run_exists", reason)

	skip, reason = skipFirstAppBranchRun(false, false, false)
	require.True(t, skip)
	require.Equal(t, "no_vcs", reason)

	skip, reason = skipFirstAppBranchRun(false, true, false)
	require.False(t, skip)
	require.Empty(t, reason)

	skip, reason = skipFirstAppBranchRun(false, false, true)
	require.False(t, skip)
	require.Empty(t, reason)
}

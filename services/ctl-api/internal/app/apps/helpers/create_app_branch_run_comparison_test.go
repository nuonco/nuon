package helpers

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestShouldCreateComparison(t *testing.T) {
	require.True(t, shouldCreateComparison(app.AppBranchRunTypeGit, false))
	require.False(t, shouldCreateComparison(app.AppBranchRunTypeGitPreview, false))
	require.True(t, shouldCreateComparison(app.AppBranchRunTypeManual, false))
	require.False(t, shouldCreateComparison(app.AppBranchRunTypeManual, true))
}

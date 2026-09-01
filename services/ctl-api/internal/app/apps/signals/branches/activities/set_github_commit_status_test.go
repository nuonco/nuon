package activities

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestCommitStatusContext(t *testing.T) {
	require.Equal(t, "nuon/acme/payments/production", CommitStatusContext("acme", "payments", "production", false, ""))
	require.Equal(t, "nuon/acme/payments/production preview", CommitStatusContext("acme", "payments", "production", true, ""))
	require.Equal(t, "nuon/acme/payments/production preview (build and validate)", CommitStatusContext("acme", "payments", "production", true, app.AppBranchRunPreviewModeBuildOnly))
	require.Equal(t, "nuon/acme/payments/production preview (plan-only)", CommitStatusContext("acme", "payments", "production", true, app.AppBranchRunPreviewModePlanOnly))
	require.Equal(t, "nuon/acme/payments/production preview (apply)", CommitStatusContext("acme", "payments", "production", true, app.AppBranchRunPreviewModeApply))
	require.Equal(t, "nuon/acme/production", CommitStatusContext("acme", "", "production", false, ""))
	require.Equal(t, "nuon", CommitStatusContext("", "", "", false, ""))
	require.Equal(t, "nuon preview", CommitStatusContext("", "", "", true, ""))
	require.Len(t, CommitStatusContext(strings.Repeat("x", 300), "payments", "production", false, ""), maxCommitStatusContextLen)
	require.Len(t, CommitStatusContext(strings.Repeat("x", 300), "payments", "production", true, app.AppBranchRunPreviewModeBuildOnly), maxCommitStatusContextLen)
}

package links

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAppBranchRunUILink(t *testing.T) {
	require.Equal(
		t,
		"https://app.example.com/org-1/apps/app-1/branches/branch-1/runs/workflow-1",
		AppBranchRunUILink("https://app.example.com/", "org-1", "app-1", "branch-1", "workflow-1"),
	)
	require.Empty(t, AppBranchRunUILink("", "org-1", "app-1", "branch-1", "workflow-1"))
}

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

func TestComponentBuildUILink(t *testing.T) {
	require.Equal(
		t,
		"https://app.example.com/org-1/apps/app-1/components/component-1/builds/build-1",
		ComponentBuildUILink("https://app.example.com/", "org-1", "app-1", "component-1", "build-1"),
	)
	require.Empty(t, ComponentBuildUILink("https://app.example.com", "org-1", "app-1", "component-1", ""))
}

func TestSandboxBuildUILink(t *testing.T) {
	require.Equal(
		t,
		"https://app.example.com/org-1/apps/app-1/sandbox/builds/build-1",
		SandboxBuildUILink("https://app.example.com/", "org-1", "app-1", "build-1"),
	)
	require.Empty(t, SandboxBuildUILink("https://app.example.com", "org-1", "app-1", ""))
}

func TestInstallAppBranchRunsUILink(t *testing.T) {
	require.Equal(
		t,
		"https://app.example.com/org-1/installs/install-1/app-branch-runs",
		InstallAppBranchRunsUILink("https://app.example.com/", "org-1", "install-1"),
	)
	require.Empty(t, InstallAppBranchRunsUILink("", "org-1", "install-1"))
	require.Empty(t, InstallAppBranchRunsUILink("https://app.example.com", "org-1", ""))
}

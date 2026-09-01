package preview

import (
	"context"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
	"github.com/stretchr/testify/require"
)

func TestCurrentRequestPrefersPullRequest(t *testing.T) {
	sources := &models.HelpersListPreviewSourcesResult{
		PullRequests: []*models.HelpersPreviewSourcePR{
			{PrNumber: 42, HeadRef: "feature", HeadSha: "pr-sha"},
		},
		Branches: []*models.HelpersPreviewSourceBranch{
			{Name: "feature", Sha: "branch-sha"},
		},
	}

	req := currentRequest("feature", sources, "")

	require.NotNil(t, req)
	require.Equal(t, models.AppAppBranchRunPreviewSourcePr, req.Source)
	require.Equal(t, int64(42), req.PrNumber)
	require.Equal(t, "pr-sha", req.HeadSha)
}

func TestWizardUsesPreviewDefaultsAndCurrentSource(t *testing.T) {
	load := func(context.Context, string) (*Data, error) { return nil, nil }
	m := initialModel(context.Background(), nil, load, Options{
		BranchID:   "branch-id",
		CurrentRef: "feature",
	})
	data := &Data{
		BranchName: "default",
		ConfigID:   "config-id",
		PreviewConfig: &models.AppAppBranchPreviewConfig{
			Mode:      models.AppAppBranchRunPreviewModeApply,
			InstallID: "install-2",
		},
		Sources: &models.HelpersListPreviewSourcesResult{
			PullRequests: []*models.HelpersPreviewSourcePR{
				{PrNumber: 42, HeadRef: "feature", HeadSha: "pr-sha"},
			},
		},
		Installs: []*models.AppInstall{
			{ID: "install-1", Name: "First"},
			{ID: "install-2", Name: "Default"},
		},
	}

	updated, _ := m.Update(branchLoadedMsg{data: data})
	m = updated.(model)
	require.Equal(t, stepMode, m.step)
	require.Equal(t, 1, m.cursor)

	updated, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = updated.(model)
	require.Equal(t, stepSourceScope, m.step)
	require.Equal(t, "current", m.items[m.cursor].value)

	updated, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = updated.(model)
	require.Equal(t, stepInstall, m.step)
	require.Equal(t, "install-2", m.items[m.cursor].value)

	updated, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = updated.(model)
	require.NotNil(t, m.result.Request)
	require.Equal(t, "config-id", m.result.ConfigID)
	require.Equal(t, models.AppAppBranchRunPreviewModeApply, m.result.Request.Mode)
	require.Equal(t, models.AppAppBranchRunPreviewSourcePr, m.result.Request.Source)
	require.Equal(t, int64(42), m.result.Request.PrNumber)
	require.Equal(t, "install-2", m.result.Request.InstallID)
}

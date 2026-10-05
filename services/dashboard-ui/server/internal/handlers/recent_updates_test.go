package handlers

import (
	"fmt"
	"testing"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func TestSelectRecentRuns(t *testing.T) {
	workflows := []*models.AppWorkflow{
		branchWorkflow("wf-old", "2026-09-01T10:00:00Z", "run-old", "app-1", "branch-1", "main"),
		branchWorkflow("wf-new", "2026-09-01T12:00:00Z", "run-new", "app-2", "branch-2", "release"),
		branchWorkflow("wf-new", "2026-09-01T12:00:00Z", "run-new", "app-2", "branch-2", "release"),
		{ID: "wf-install", CreatedAt: "2026-09-01T13:00:00Z"},
	}

	refs := selectRecentRuns(workflows)
	if len(refs) != 2 {
		t.Fatalf("len = %d, want 2", len(refs))
	}
	if refs[0].workflow.ID != "wf-new" || refs[1].workflow.ID != "wf-old" {
		t.Fatalf("order = %s, %s", refs[0].workflow.ID, refs[1].workflow.ID)
	}
}

func TestSelectRecentRunsCapsAtLimit(t *testing.T) {
	workflows := make([]*models.AppWorkflow, 0, recentUpdatesLimit+2)
	for i := 0; i < recentUpdatesLimit+2; i++ {
		workflows = append(workflows, branchWorkflow(
			fmt.Sprintf("wf-%02d", i),
			fmt.Sprintf("2026-09-01T10:%02d:00Z", i),
			fmt.Sprintf("run-%02d", i),
			"app-1",
			"branch-1",
			"main",
		))
	}
	if len(selectRecentRuns(workflows)) != recentUpdatesLimit {
		t.Fatalf("len = %d, want %d", len(selectRecentRuns(workflows)), recentUpdatesLimit)
	}
}

func branchWorkflow(workflowID, createdAt, runID, appID, branchID, branchName string) *models.AppWorkflow {
	return &models.AppWorkflow{
		ID:        workflowID,
		CreatedAt: createdAt,
		OwnerID:   branchID,
		AppBranchRuns: []*models.AppAppBranchRun{{
			ID:        runID,
			CreatedAt: createdAt,
			AppBranch: &models.AppAppBranch{
				ID:    branchID,
				Name:  branchName,
				AppID: appID,
			},
		}},
	}
}

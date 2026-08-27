package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestParseIncludeDiff(t *testing.T) {
	require.Equal(t, map[string]bool{"git": true, "full": true}, parseIncludeDiff("git,full"))
	require.Equal(t, map[string]bool{"config": true}, parseIncludeDiff(" CONFIG "))
	require.Empty(t, parseIncludeDiff(""))
	require.Empty(t, parseIncludeDiff(",,"))
}

func TestComparisonRunSummary(t *testing.T) {
	require.Nil(t, comparisonRunSummary(nil))

	workflowID := "wf-base"
	pr := 32
	created := time.Date(2026, 8, 27, 12, 0, 0, 0, time.UTC)
	summary := comparisonRunSummary(&app.AppBranchRun{
		ID:         "run-base",
		WorkflowID: &workflowID,
		Status:     "success",
		CreatedAt:  created,
		PRNumber:   &pr,
		BaseBranch: "main",
		EventType:  "pull_request",
		VCSConnectionCommit: &app.VCSConnectionCommit{
			SHA:     "abc123",
			Message: "chore: update",
		},
	})
	require.NotNil(t, summary)
	require.Equal(t, "run-base", summary.ID)
	require.Equal(t, "wf-base", summary.WorkflowID)
	require.Equal(t, "abc123", summary.VCSConnectionCommit.SHA)
	require.Equal(t, 32, *summary.PRNumber)
}

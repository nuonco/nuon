package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

const (
	watchDefaultTimeout = 30 * time.Second
	watchMaxTimeout     = 60 * time.Second
	watchPollInterval   = 3 * time.Second
)

type mcpWatchWorkflowInput struct {
	WorkflowID      string            `json:"workflow_id" jsonschema:"workflow ID to watch"`
	Cursor          string            `json:"cursor,omitempty" jsonschema:"cursor from the previous watch_workflow response; returns when the workflow fingerprint differs"`
	LastKnownStatus string            `json:"last_known_status,omitempty" jsonschema:"the status you already know about; used when cursor is omitted and returns when status differs"`
	TimeoutSeconds  int               `json:"timeout_seconds,omitempty" jsonschema:"max seconds to wait for a change (default 30, max 60)"`
	Reported        map[string]string `json:"reported,omitempty" jsonschema:"step id to status already shown; pass reported from the previous watch_workflow response"`
}

type mcpWatchWorkflowResult struct {
	Changed      bool               `json:"changed"`
	Cursor       string             `json:"cursor"`
	StepProgress string             `json:"step_progress"`
	Reported     map[string]string  `json:"reported,omitempty"`
	Workflow     mcpWorkflowSummary `json:"workflow"`
	NextAction   *mcpNextAction     `json:"next_action,omitempty"`
}

func (s *service) mcpWatchWorkflow(ctx context.Context, req *mcp.CallToolRequest, in mcpWatchWorkflowInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Read(ctx)
	if err != nil {
		return nil, nil, err
	}

	timeout := watchDefaultTimeout
	if in.TimeoutSeconds > 0 {
		timeout = time.Duration(in.TimeoutSeconds) * time.Second
		if timeout > watchMaxTimeout {
			timeout = watchMaxTimeout
		}
	}

	deadline := time.Now().Add(timeout)
	var notified string
	var progress float64

	for {
		summary, err := s.fetchWorkflowSummary(ctx, orgID, in.WorkflowID)
		if err != nil {
			return nil, nil, err
		}

		cursor := mcpWorkflowCursor(*summary)
		if cursor != notified {
			progress++
			notifyWatchProgress(ctx, req, *summary, in.Reported, progress)
			notified = cursor
		}

		_, _, hasNew := mcpStepProgress(*summary, in.Reported)
		snapshotOnly := in.Cursor == "" && in.LastKnownStatus == ""
		// A cursor change that only repeats an in-progress step, or that only
		// adds a pending step, is not worth returning. Keep polling until a
		// step starts or finishes so that success is not dropped between calls.
		if snapshotOnly || hasNew || !mcpShouldKeepWatching(*summary) || time.Now().After(deadline) {
			return apiPkg.MCPJSONResult(mcpWatchResult(*summary, cursor, mcpWatchChanged(in, summary.Status, cursor), in.Reported))
		}

		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		case <-time.After(watchPollInterval):
		}
	}
}

func (s *service) fetchWorkflowSummary(ctx context.Context, orgID, workflowID string) (*mcpWorkflowSummary, error) {
	var workflow app.Workflow
	err := s.db.WithContext(ctx).
		Preload("Steps", func(db *gorm.DB) *gorm.DB {
			return db.Order("group_idx, group_retry_idx, idx, created_at asc")
		}).
		Preload("Steps.Approval", func(db *gorm.DB) *gorm.DB {
			return db.Omit("contents")
		}).
		Preload("Steps.Approval.Response").
		Where("id = ? AND org_id = ?", workflowID, orgID).
		First(&workflow).Error
	if err != nil {
		return nil, fmt.Errorf("unable to find workflow %q: %w", workflowID, err)
	}

	summary := summarizeMCPWorkflow(workflow)
	s.attachStackSetups(ctx, orgID, workflow.Steps, summary.Steps)
	return &summary, nil
}

func mcpWatchChanged(in mcpWatchWorkflowInput, status, cursor string) bool {
	if in.Cursor != "" {
		return cursor != in.Cursor
	}
	return in.LastKnownStatus != "" && status != in.LastKnownStatus
}

func mcpWatchResult(summary mcpWorkflowSummary, cursor string, changed bool, reported map[string]string) mcpWatchWorkflowResult {
	stepProgress, nextReported, _ := mcpStepProgress(summary, reported)
	for i := range summary.Steps {
		summary.Steps[i].ExecutionTime = ""
	}
	result := mcpWatchWorkflowResult{
		Changed:      changed,
		Cursor:       cursor,
		StepProgress: stepProgress,
		Reported:     nextReported,
		Workflow:     summary,
	}
	if mcpShouldKeepWatching(summary) {
		result.NextAction = &mcpNextAction{
			Action: "watch_workflow",
			Label:  "Watch workflow",
			Tool:   "watch_workflow",
			Arguments: map[string]any{
				"workflow_id": summary.ID,
				"cursor":      cursor,
				"reported":    nextReported,
			},
		}
	}
	return result
}

func mcpStepProgress(summary mcpWorkflowSummary, reported map[string]string) (string, map[string]string, bool) {
	next := make(map[string]string, len(reported)+len(summary.Steps))
	for id, status := range reported {
		next[id] = status
	}
	if len(summary.Steps) == 0 {
		if summary.StatusDescription != "" {
			return summary.StatusDescription, next, false
		}
		return "Generating steps", next, false
	}
	var b strings.Builder
	shown := 0
	hasNew := false
	for i, step := range summary.Steps {
		if !mcpShowStepProgress(step.Status, next[step.ID]) {
			continue
		}
		if !mcpActiveStepStatus(step.Status) || next[step.ID] != step.Status {
			hasNew = true
		}
		if shown > 0 {
			b.WriteByte('\n')
		}
		shown++
		fmt.Fprintf(&b, "%d. %s — %s", i+1, step.Name, step.Status)
		next[step.ID] = step.Status
	}
	if len(next) == 0 {
		next = nil
	}
	return b.String(), next, hasNew
}

func mcpShowStepProgress(status, reported string) bool {
	switch app.Status(status) {
	case app.StatusPending, app.StatusQueued, app.StatusNotAttempted:
		return false
	case app.StatusInProgress, app.StatusRetrying, app.StatusCheckPlan, app.StatusPlanning, app.StatusApplying:
		return true
	default:
		return status != reported
	}
}

func mcpActiveStepStatus(status string) bool {
	switch app.Status(status) {
	case app.StatusInProgress, app.StatusRetrying, app.StatusCheckPlan, app.StatusPlanning, app.StatusApplying:
		return true
	default:
		return false
	}
}

func notifyWatchProgress(ctx context.Context, req *mcp.CallToolRequest, summary mcpWorkflowSummary, reported map[string]string, progress float64) {
	if req == nil || req.Params == nil || req.Session == nil {
		return
	}
	token := req.Params.GetProgressToken()
	if token == nil {
		return
	}
	message, _, _ := mcpStepProgress(summary, reported)
	if message == "" {
		return
	}
	_ = req.Session.NotifyProgress(ctx, &mcp.ProgressNotificationParams{
		ProgressToken: token,
		Progress:      progress,
		Message:       message,
	})
}

func mcpShouldKeepWatching(summary mcpWorkflowSummary) bool {
	switch app.Status(summary.Status) {
	case app.StatusSuccess, app.StatusError, app.StatusCancelled, app.StatusNotAttempted, app.StatusDiscarded:
		return false
	}
	return len(summary.PendingApprovals) == 0 && len(summary.NextActions) == 0
}

func mcpWorkflowCursor(summary mcpWorkflowSummary) string {
	var b strings.Builder
	fmt.Fprintf(&b, "status=%s\nsteps=%d/%d\n", summary.Status, summary.CompletedSteps, summary.TotalSteps)
	for _, step := range summary.Steps {
		fmt.Fprintf(&b, "step=%s status=%s\n", step.ID, step.Status)
		if step.CompositeError != nil {
			fmt.Fprintf(&b, "err=%s %s\n", step.CompositeError.Type, step.CompositeError.Message)
		}
		if step.StackSetup != nil {
			fmt.Fprintf(&b, "stack=%s link=%t\n", step.StackSetup.Status, step.StackSetup.QuickLinkURL != "")
		}
		for _, action := range step.NextActions {
			fmt.Fprintf(&b, "step-action=%s\n", action.Action)
		}
	}
	for _, approval := range summary.PendingApprovals {
		fmt.Fprintf(&b, "approval=%s %s %d %d %d %d\n",
			approval.ApprovalID,
			approval.ChangesState,
			approval.ChangesCreate,
			approval.ChangesUpdate,
			approval.ChangesDelete,
			approval.ChangesReplace,
		)
	}
	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}

func mcpWatchContinuation(workflowID string) *mcpNextAction {
	if workflowID == "" {
		return nil
	}
	return &mcpNextAction{
		Action: "watch_workflow",
		Label:  "Watch workflow",
		Tool:   "watch_workflow",
		Arguments: map[string]any{
			"workflow_id": workflowID,
		},
	}
}

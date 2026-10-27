package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpGetWorkflowStepLogsInput struct {
	StepID       string `json:"step_id" jsonschema:"workflow step ID to get logs for"`
	Severity     string `json:"severity,omitempty" jsonschema:"filter by severity (e.g. ERROR, WARN, INFO)"`
	BodyContains string `json:"body_contains,omitempty" jsonschema:"substring filter on log body"`
	Cursor       string `json:"cursor,omitempty" jsonschema:"pagination cursor from a previous response's next_cursor"`
	Limit        int    `json:"limit,omitempty" jsonschema:"max records to return (default 100, max 200)"`
}

type mcpStepLogResult struct {
	StepName        string        `json:"step_name"`
	LogStreamID     string        `json:"log_stream_id,omitempty"`
	ReturnedRecords int           `json:"returned_records"`
	HasMore         bool          `json:"has_more"`
	NextCursor      string        `json:"next_cursor,omitempty"`
	Message         string        `json:"message,omitempty"`
	Logs            []mcpLogEntry `json:"logs"`
}

func (s *service) mcpGetWorkflowStepLogs(ctx context.Context, _ *mcp.CallToolRequest, in mcpGetWorkflowStepLogsInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Read(ctx)
	if err != nil {
		return nil, nil, err
	}

	var step app.WorkflowStep
	if err := s.db.WithContext(ctx).
		Where(app.WorkflowStep{OrgID: orgID}).
		Where("id = ?", in.StepID).
		First(&step).Error; err != nil {
		return nil, nil, fmt.Errorf("unable to find step %q: %w", in.StepID, err)
	}

	if step.StepTargetID == "" || !stepTargetHasLogs(step.StepTargetType) {
		return apiPkg.MCPJSONResult(mcpStepLogResult{
			StepName: step.Name,
			Message:  "This step has no log stream.",
			Logs:     []mcpLogEntry{},
		})
	}

	logStreamID, err := s.resolveStepLogStream(ctx, orgID, step.StepTargetType, step.StepTargetID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apiPkg.MCPJSONResult(mcpStepLogResult{
				StepName: step.Name,
				Message:  "This step has no log stream yet.",
				Logs:     []mcpLogEntry{},
			})
		}
		return nil, nil, err
	}

	page, err := s.queryMCPLogs(ctx, orgID, logStreamID, mcpLogFilters{
		Severity:     in.Severity,
		BodyContains: in.BodyContains,
		Cursor:       in.Cursor,
		Limit:        in.Limit,
	})
	if err != nil {
		return nil, nil, err
	}

	return apiPkg.MCPJSONResult(mcpStepLogResult{
		StepName:        step.Name,
		LogStreamID:     page.LogStreamID,
		ReturnedRecords: page.ReturnedRecords,
		HasMore:         page.HasMore,
		NextCursor:      page.NextCursor,
		Logs:            page.Logs,
	})
}

func stepTargetHasLogs(targetType string) bool {
	switch targetType {
	case string(app.WorkflowStepTargetTypeInstallDeploys),
		string(app.WorkflowStepTargetTypeInstallActionWorkflowRuns),
		string(app.WorkflowStepTargetTypeInstallSandboxRuns):
		return true
	default:
		return false
	}
}

func (s *service) resolveStepLogStream(ctx context.Context, orgID, targetType, targetID string) (string, error) {
	var ownerType string
	switch targetType {
	case string(app.WorkflowStepTargetTypeInstallDeploys):
		ownerType = "install_deploys"
	case string(app.WorkflowStepTargetTypeInstallActionWorkflowRuns):
		ownerType = "install_action_workflow_runs"
	case string(app.WorkflowStepTargetTypeInstallSandboxRuns):
		ownerType = "install_sandbox_runs"
	default:
		return "", fmt.Errorf("step target type %q does not support logs", targetType)
	}

	return s.resolveOwnerLogStream(ctx, orgID, ownerType, targetID)
}

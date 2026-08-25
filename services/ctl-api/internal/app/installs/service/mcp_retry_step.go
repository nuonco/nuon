package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
)

type mcpRetryStepInput struct {
	WorkflowID string `json:"workflow_id" jsonschema:"workflow ID"`
	StepID     string `json:"step_id" jsonschema:"step ID to retry"`
}

func (s *service) mcpRetryStep(ctx context.Context, _ *mcp.CallToolRequest, in mcpRetryStepInput) (*mcp.CallToolResult, any, error) {
	if err := requireWriteScope(ctx); err != nil {
		return nil, nil, err
	}

	orgID := keys.OrgIDFromContext(ctx)

	var workflow app.Workflow
	if err := s.db.WithContext(ctx).
		Where("id = ? AND org_id = ?", in.WorkflowID, orgID).
		First(&workflow).Error; err != nil {
		return nil, nil, fmt.Errorf("unable to find workflow %q: %w", in.WorkflowID, err)
	}

	var step app.WorkflowStep
	if err := s.db.WithContext(ctx).
		Where(app.WorkflowStep{OrgID: orgID}).
		Where("id = ? AND owner_id = ?", in.StepID, workflow.ID).
		First(&step).Error; err != nil {
		return nil, nil, fmt.Errorf("unable to find step %q: %w", in.StepID, err)
	}

	resp, err := s.flowsClient.RetryStep(ctx, &flowclient.RetryStepRequest{
		InstallWorkflowID: workflow.ID,
		StepID:            step.ID,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("retry step: %w", err)
	}

	return apiPkg.MCPJSONResult(map[string]any{
		"workflow_id": workflow.ID,
		"step_id":     step.ID,
		"retryable":   resp.Retryable,
	})
}

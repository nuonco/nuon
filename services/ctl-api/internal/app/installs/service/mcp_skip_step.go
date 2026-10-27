package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
)

type mcpSkipStepInput struct {
	WorkflowID string `json:"workflow_id" jsonschema:"workflow ID"`
	StepID     string `json:"step_id" jsonschema:"step ID to skip"`
}

func (s *service) mcpSkipStep(ctx context.Context, _ *mcp.CallToolRequest, in mcpSkipStepInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Write(ctx)
	if err != nil {
		return nil, nil, err
	}
	if in.WorkflowID == "" {
		return nil, nil, fmt.Errorf("workflow_id is required")
	}
	if in.StepID == "" {
		return nil, nil, fmt.Errorf("step_id is required")
	}

	var workflow app.Workflow
	if err := s.db.WithContext(ctx).
		Where("id = ? AND org_id = ?", in.WorkflowID, orgID).
		First(&workflow).Error; err != nil {
		return nil, nil, fmt.Errorf("unable to find workflow %q: %w", in.WorkflowID, err)
	}

	var step app.WorkflowStep
	if err := s.db.WithContext(ctx).
		Where("id = ? AND install_workflow_id = ? AND org_id = ?", in.StepID, workflow.ID, orgID).
		First(&step).Error; err != nil {
		return nil, nil, fmt.Errorf("unable to find step %q: %w", in.StepID, err)
	}

	resp, err := s.flowsClient.SkipStep(ctx, &flowclient.SkipStepRequest{
		InstallWorkflowID: workflow.ID,
		StepID:            step.ID,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("skip step: %w", err)
	}

	out := map[string]any{
		"workflow_id": workflow.ID,
		"step_id":     step.ID,
		"skippable":   resp.Skippable,
	}
	if next := mcpWatchContinuation(workflow.ID); next != nil {
		out["next_action"] = next
	}
	return apiPkg.MCPJSONResult(out)
}

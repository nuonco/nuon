package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
)

type mcpCancelWorkflowInput struct {
	WorkflowID string `json:"workflow_id" jsonschema:"workflow ID to cancel"`
}

func (s *service) mcpCancelWorkflow(ctx context.Context, _ *mcp.CallToolRequest, in mcpCancelWorkflowInput) (*mcp.CallToolResult, any, error) {
	if err := requireWriteScope(ctx); err != nil {
		return nil, nil, err
	}

	orgID := keys.OrgIDFromContext(ctx)

	var wf app.Workflow
	if err := s.db.WithContext(ctx).
		Where("id = ? AND org_id = ?", in.WorkflowID, orgID).
		First(&wf).Error; err != nil {
		return nil, nil, fmt.Errorf("unable to find workflow %q: %w", in.WorkflowID, err)
	}

	if !generics.SliceContains(wf.Status.Status, []app.Status{
		app.StatusInProgress,
		app.StatusPending,
		app.AwaitingApproval,
		app.Status("awaiting-approval"),
		app.StatusFailedPendingRetry,
	}) {
		return nil, nil, fmt.Errorf("workflow is not cancelable (status: %s)", wf.Status.Status)
	}

	if wf.Status.Status == app.StatusPending {
		if err := s.cancelWorkflow(ctx, wf.ID); err != nil {
			return nil, nil, fmt.Errorf("unable to cancel workflow: %w", err)
		}
		return apiPkg.MCPJSONResult(map[string]string{
			"workflow_id": wf.ID,
			"status":      "cancelled",
		})
	}

	if _, err := s.flowsClient.CancelWorkflow(ctx, &flowclient.CancelWorkflowRequest{
		InstallWorkflowID: wf.ID,
	}); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if dbErr := s.cancelWorkflow(ctx, wf.ID); dbErr != nil {
				return nil, nil, fmt.Errorf("unable to cancel orphaned workflow: %w", dbErr)
			}
			return apiPkg.MCPJSONResult(map[string]string{
				"workflow_id": wf.ID,
				"status":      "cancelled",
			})
		}
		return nil, nil, fmt.Errorf("unable to cancel workflow: %w", err)
	}

	return apiPkg.MCPJSONResult(map[string]string{
		"workflow_id": wf.ID,
		"status":      "cancelling",
	})
}

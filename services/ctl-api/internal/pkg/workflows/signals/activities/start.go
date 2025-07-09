package activities

import (
	"context"
	"fmt"

	enumsv1 "go.temporal.io/api/enums/v1"
	tclient "go.temporal.io/sdk/client"
	"go.uber.org/zap"

	"go.temporal.io/sdk/temporal"

	"github.com/powertoolsdev/mono/pkg/workflows"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

// TODO(sdboyer) this entire file will probably be removed in favor of a generic system

type StartEventLoopRequest struct {
	WorkflowID   string
	WorkflowType string
	ObjectID     string
	Namespace    string
}

// @temporal-gen activity
func (a *Activities) StartEventLoop(ctx context.Context, req *StartEventLoopRequest) error {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return fmt.Errorf("unable to get org id from context: %w", err)
	}

	org := app.Org{}
	res := a.db.WithContext(ctx).First(&org, "id = ?", orgID)
	if res.Error != nil {
		return fmt.Errorf("unable to get org: %w", res.Error)
	}

	sandboxMode := org.SandboxMode
	if org.OrgType == app.OrgTypeIntegration {
		return nil
	}

	opts := tclient.StartWorkflowOptions{
		ID:        req.WorkflowID,
		TaskQueue: workflows.APITaskQueue,
		// Memo is non-indexed metadata available when listing workflows
		Memo: map[string]interface{}{
			"id":         req.ObjectID,
			"started-by": "workflow",
		},
		WorkflowIDReusePolicy: enumsv1.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 0,
		},
	}

	ereq := eventloop.EventLoopRequest{
		ID:          req.ObjectID,
		SandboxMode: sandboxMode,
		Version:     a.cfg.Version,
	}

	if req.WorkflowType == "" {
		req.WorkflowType = "EventLoop"
	}

	wkflowRun, err := a.client.ExecuteWorkflowInNamespace(ctx,
		req.Namespace,
		opts,
		req.WorkflowType,
		ereq)
	if err != nil {
		return err
	}

	a.l.Debug("started event loop",
		zap.String("workflow-id", wkflowRun.GetID()),
		zap.String("run-id", wkflowRun.GetID()),
		zap.String("id", req.ObjectID),
		zap.Error(err),
	)
	return nil
}

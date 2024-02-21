package hooks

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/workflows"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/releases/worker"
	enumsv1 "go.temporal.io/api/enums/v1"
	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.uber.org/zap"
)

func (a *Hooks) startEventLoop(ctx context.Context, releaseID string, orgType app.OrgType) error {
	if orgType == app.OrgTypeIntegration {
		return nil
	}

	workflowID := worker.EventLoopWorkflowID(releaseID)
	opts := tclient.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: workflows.APITaskQueue,
		// Memo is non-indexed metadata available when listing workflows
		Memo: map[string]interface{}{
			"release-id": releaseID,
			"started-by": "api",
		},
		WorkflowIDReusePolicy: enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 0,
		},
	}
	req := worker.ReleaseEventLoopRequest{
		ReleaseID:   releaseID,
		SandboxMode: orgType == app.OrgTypeSandbox,
	}
	wkflowRun, err := a.client.ExecuteWorkflowInNamespace(ctx,
		defaultNamespace,
		opts,
		worker.EventLoopWorkflowName,
		req)
	if err != nil {
		return fmt.Errorf("unable to create release event loop: %w", err)
	}
	a.l.Debug("started release event loop",
		zap.String("workflow-id", wkflowRun.GetID()),
		zap.String("run-id", wkflowRun.GetID()),
		zap.String("release-id", releaseID),
		zap.Error(err),
	)

	return nil
}

func (a *Hooks) Created(ctx context.Context, releaseID string, orgType app.OrgType) {
	if err := a.startEventLoop(ctx, releaseID, orgType); err != nil {
		a.l.Error("error starting event loop",
			zap.String("release-id", releaseID),
			zap.Error(err),
		)
		return
	}

	a.sendSignal(ctx, releaseID, worker.Signal{
		Operation: worker.OperationPollDependencies,
	})
	a.sendSignal(ctx, releaseID, worker.Signal{
		Operation: worker.OperationProvision,
	})
}

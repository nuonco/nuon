package hooks

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/workflows"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/worker"
	enumsv1 "go.temporal.io/api/enums/v1"
	tclient "go.temporal.io/sdk/client"
	"go.uber.org/zap"
)

func (a *Hooks) startEventLoop(ctx context.Context, appID string, orgType app.OrgType) error {
	if orgType == app.OrgTypeIntegration {
		return nil
	}

	workflowID := worker.EventLoopWorkflowID(appID)
	opts := tclient.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: workflows.APITaskQueue,
		// Memo is non-indexed metadata available when listing workflows
		Memo: map[string]interface{}{
			"app-id":     appID,
			"started-by": "api",
		},
		WorkflowIDReusePolicy: enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}

	req := worker.AppEventLoopRequest{
		AppID:       appID,
		SandboxMode: orgType == app.OrgTypeSandbox,
	}
	wkflowRun, err := a.client.ExecuteWorkflowInNamespace(ctx,
		defaultNamespace,
		opts,
		worker.EventLoopWorkflowName,
		req)
	if err != nil {
		return err
	}

	a.l.Debug("started app event loop",
		zap.String("workflow-id", wkflowRun.GetID()),
		zap.String("run-id", wkflowRun.GetID()),
		zap.String("app-id", appID),
		zap.Error(err),
	)
	return nil
}

func (a *Hooks) Created(ctx context.Context, appID string, orgType app.OrgType) {
	if err := a.startEventLoop(ctx, appID, orgType); err != nil {
		a.l.Error("error starting event loop",
			zap.String("app-id", appID),
			zap.Error(err),
		)
		return
	}

	a.sendSignal(ctx, appID, worker.Signal{
		Operation: worker.OperationPollDependencies,
	})
	a.sendSignal(ctx, appID, worker.Signal{
		Operation: worker.OperationProvision,
	})
}

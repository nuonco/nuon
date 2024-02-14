package hooks

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/workflows"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/worker/signals"
	enumsv1 "go.temporal.io/api/enums/v1"
	tclient "go.temporal.io/sdk/client"
	"go.uber.org/zap"
)

func (a *Hooks) startEventLoop(ctx context.Context, componentID string, orgType app.OrgType) error {
	if orgType == app.OrgTypeIntegration {
		return nil
	}

	workflowID := signals.EventLoopWorkflowID(componentID)
	opts := tclient.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: workflows.APITaskQueue,
		// Memo is non-indexed metadata available when listing workflows
		Memo: map[string]interface{}{
			"component-id": componentID,
			"started-by":   "api",
		},
		WorkflowIDReusePolicy: enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}

	req := signals.ComponentEventLoopRequest{
		ComponentID: componentID,
		SandboxMode: orgType == app.OrgTypeSandbox,
	}
	wkflowRun, err := a.client.ExecuteWorkflowInNamespace(ctx,
		defaultNamespace,
		opts,
		signals.EventLoopWorkflowName,
		req)
	if err != nil {
		return err
	}
	a.l.Debug("started component event loop",
		zap.String("workflow-id", wkflowRun.GetID()),
		zap.String("run-id", wkflowRun.GetID()),
		zap.String("component-id", componentID),
		zap.Error(err),
	)

	return nil
}

func (a *Hooks) Created(ctx context.Context, componentID string, orgType app.OrgType) {
	if err := a.startEventLoop(ctx, componentID, orgType); err != nil {
		a.l.Error("error starting event loop",
			zap.String("component-id", componentID),
			zap.Error(err),
		)
		return
	}

	a.sendSignal(ctx, componentID, signals.Signal{
		Operation: signals.OperationPollDependencies,
	})
	a.sendSignal(ctx, componentID, signals.Signal{
		Operation: signals.OperationProvision,
	})
}

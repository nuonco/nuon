package hooks

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/workflows"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/signals"
	enumsv1 "go.temporal.io/api/enums/v1"
	tclient "go.temporal.io/sdk/client"
	"go.uber.org/zap"
)

func (a *Hooks) startEventLoop(ctx context.Context, installID string, sandboxMode bool) error {
	workflowID := signals.EventLoopWorkflowID(installID)
	opts := tclient.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: workflows.APITaskQueue,
		// Memo is non-indexed metadata available when listing workflows
		Memo: map[string]interface{}{
			"install-id": installID,
			"started-by": "api",
		},
		WorkflowIDReusePolicy: enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}

	req := signals.InstallEventLoopRequest{
		InstallID:   installID,
		SandboxMode: sandboxMode,
	}
	wkflowRun, err := a.client.ExecuteWorkflowInNamespace(ctx,
		defaultNamespace,
		opts,
		signals.EventLoopWorkflowName,
		req)
	if err != nil {
		return err
	}

	a.l.Debug("started install event loop",
		zap.String("workflow-id", wkflowRun.GetID()),
		zap.String("run-id", wkflowRun.GetID()),
		zap.String("install-id", installID),
		zap.Error(err),
	)
	return nil
}

func (a *Hooks) Created(ctx context.Context, installID string, sandboxMode bool) {
	if err := a.startEventLoop(ctx, installID, sandboxMode); err != nil {
		a.l.Error("unable to start event loop",
			zap.String("install-id", installID),
			zap.Error(err),
		)
		return
	}

	a.sendSignal(ctx, installID, signals.Signal{
		Operation: signals.OperationPollDependencies,
	})
	a.sendSignal(ctx, installID, signals.Signal{
		Operation: signals.OperationProvision,
	})
	a.sendSignal(ctx, installID, signals.Signal{
		Operation: signals.OperationDeployComponents,
	})
}

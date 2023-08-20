package hooks

import (
	"context"
	"log"

	"github.com/powertoolsdev/mono/pkg/workflows"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/worker"
	enumsv1 "go.temporal.io/api/enums/v1"
	tclient "go.temporal.io/sdk/client"
	"go.uber.org/zap"
)

func (a *Hooks) Created(ctx context.Context, appID string) {
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
	wkflowRun, err := a.client.ExecuteWorkflowInNamespace(ctx,
		defaultNamespace,
		opts,
		worker.EventLoopWorkflowName,
		appID)
	if err != nil {
		log.Fatalln("error creating app event loop", err)
		return
	}
	a.l.Debug("started app event loop",
		zap.String("workflow-id", wkflowRun.GetID()),
		zap.String("run-id", wkflowRun.GetID()),
		zap.String("app-id", appID),
		zap.Error(err),
	)

	a.sendSignal(ctx, appID, worker.Signal{
		DryRun:    a.cfg.DevEnableWorkersDryRun,
		Operation: worker.OperationProvision,
	})
}

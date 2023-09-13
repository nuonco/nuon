package hooks

import (
	"context"
	"log"

	"github.com/powertoolsdev/mono/pkg/workflows"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/releases/worker"
	enumsv1 "go.temporal.io/api/enums/v1"
	tclient "go.temporal.io/sdk/client"
	"go.uber.org/zap"
)

func (a *Hooks) Created(ctx context.Context, releaseID string) {
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
	}
	wkflowRun, err := a.client.ExecuteWorkflowInNamespace(ctx,
		defaultNamespace,
		opts,
		worker.EventLoopWorkflowName,
		releaseID)
	if err != nil {
		log.Fatalln("error creating release event loop", err)
		return
	}
	a.l.Debug("started release event loop",
		zap.String("workflow-id", wkflowRun.GetID()),
		zap.String("run-id", wkflowRun.GetID()),
		zap.String("release-id", releaseID),
		zap.Error(err),
	)

	a.sendSignal(ctx, releaseID, worker.Signal{
		DryRun:    a.cfg.DevEnableWorkersDryRun,
		Operation: worker.OperationPollDependencies,
	})
	a.sendSignal(ctx, releaseID, worker.Signal{
		DryRun:    a.cfg.DevEnableWorkersDryRun,
		Operation: worker.OperationProvision,
	})
}

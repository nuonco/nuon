package hooks

import (
	"context"
	"log"

	"github.com/powertoolsdev/mono/pkg/workflows"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker"
	enumsv1 "go.temporal.io/api/enums/v1"
	tclient "go.temporal.io/sdk/client"
	"go.uber.org/zap"
)

func (a *Hooks) Restart(ctx context.Context, installID string) {
	workflowID := worker.EventLoopWorkflowID(installID)
	opts := tclient.StartWorkflowOptions{
		ID:	   workflowID,
		TaskQueue: workflows.APITaskQueue,
		// Memo is non-indexed metadata available when listing workflows
		Memo: map[string]interface{}{
			"install-id": installID,
			"started-by": "api",
		},
		WorkflowIDReusePolicy: enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}
	wkflowRun, err := a.client.ExecuteWorkflowInNamespace(ctx,
		defaultNamespace,
		opts,
		worker.EventLoopWorkflowName,
		installID)
	if err != nil {
		log.Fatalln("error creating install event loop", err)
		return
	}
	a.l.Debug("started install event loop",
		zap.String("workflow-id", wkflowRun.GetID()),
		zap.String("run-id", wkflowRun.GetID()),
		zap.String("install-id", installID),
		zap.Error(err),
	)

	a.sendSignal(ctx, installID, worker.Signal{
		DryRun:    a.cfg.DevEnableWorkersDryRun,
		Operation: worker.OperationPollDependencies,
	})
}

package hooks

import (
	"context"
	"log"

	"github.com/powertoolsdev/mono/pkg/workflows"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker"
	enumsv1 "go.temporal.io/api/enums/v1"
	tclient "go.temporal.io/sdk/client"
	"go.uber.org/zap"
)

func (o *Hooks) Created(ctx context.Context, orgID string) {
	o.l.Debug("org created hook", zap.String("id", orgID))
	o.l.Info("org created hook (info)", zap.String("id", orgID))

	workflowID := worker.EventLoopWorkflowID(orgID)
	signal := worker.Signal{
		DryRun:    o.cfg.EnableWorkersDryRun,
		Operation: worker.OperationProvision,
	}
	opts := tclient.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: workflows.APITaskQueue,

		// Memo is non-indexed metadata available when listing workflows
		Memo: map[string]interface{}{
			"org-id":     orgID,
			"started-by": "api",
		},
		WorkflowIDReusePolicy: enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}
	wkflowRun, err := o.client.SignalWithStartWorkflowInNamespace(ctx,
		defaultNamespace,
		workflowID,
		orgID,
		signal,
		opts,
		worker.EventLoopWorkflowName,
		orgID)
	if err != nil {
		log.Fatalln("error creating event loop + signaling to provision", err)
		return
	}
	o.l.Debug("started event loop",
		zap.String("workflow-id", wkflowRun.GetID()),
		zap.String("workflow-id", wkflowRun.GetID()),
		zap.Error(err),
	)
}

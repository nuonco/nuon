package hooks

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/workflows"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker"
	enumsv1 "go.temporal.io/api/enums/v1"
	tclient "go.temporal.io/sdk/client"
	"go.uber.org/zap"
)

func (o *Hooks) startEventLoop(ctx context.Context, orgID string, sandboxMode bool) error {
	workflowID := worker.EventLoopWorkflowID(orgID)
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
	req := worker.OrgEventLoopRequest{
		OrgID:       orgID,
		SandboxMode: sandboxMode,
	}
	wkflowRun, err := o.client.ExecuteWorkflowInNamespace(ctx,
		defaultNamespace,
		opts,
		worker.EventLoopWorkflowName,
		req)
	if err != nil {
		return err
	}

	o.l.Debug("started org event loop",
		zap.String("workflow-id", wkflowRun.GetID()),
		zap.String("run-id", wkflowRun.GetID()),
		zap.String("org-id", orgID),
		zap.Error(err),
	)
	return nil
}

func (o *Hooks) Created(ctx context.Context, orgID string, sandboxMode bool) {
	if err := o.startEventLoop(ctx, orgID, sandboxMode); err != nil {
		o.l.Error("error starting event loop",
			zap.String("org-id", orgID),
			zap.Error(err),
		)
		return
	}

	o.sendSignal(ctx, orgID, worker.Signal{
		Operation: worker.OperationProvision,
	})
}

package hooks

import (
	"context"

	enumsv1 "go.temporal.io/api/enums/v1"
	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/workflows"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/signals"
)

func (o *Hooks) startEventLoop(ctx context.Context, orgID string, orgType app.OrgType) error {
	if orgType == app.OrgTypeIntegration {
		return nil
	}

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
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 0,
		},
	}
	req := worker.OrgEventLoopRequest{
		OrgID:       orgID,
		SandboxMode: orgType == app.OrgTypeSandbox,
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

func (o *Hooks) Created(ctx context.Context, orgID string, orgType app.OrgType) {
	if err := o.startEventLoop(ctx, orgID, orgType); err != nil {
		o.l.Error("error starting event loop",
			zap.String("org-id", orgID),
			zap.Error(err),
		)
		return
	}

	o.sendSignal(ctx, orgID, signals.Signal{
		Operation: signals.OperationCreated,
	})
	o.sendSignal(ctx, orgID, signals.Signal{
		Operation: signals.OperationProvision,
	})
}

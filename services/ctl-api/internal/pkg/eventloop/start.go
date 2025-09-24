package eventloop

import (
	"context"

	enumsv1 "go.temporal.io/api/enums/v1"
	tclient "go.temporal.io/sdk/client"

	"go.temporal.io/sdk/temporal"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/pkg/workflows"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (a *evClient) startEventLoop(ctx context.Context, id string, signal Signal) error {
	org, err := signal.GetOrg(ctx, id, a.db)
	if err != nil {
		a.mw.Incr("event_loop_signal", metrics.ToStatusTag("unable_to_get_org"))
		return err
	}

	orgTyp := app.OrgTypeUnknown
	if org != nil {
		orgTyp = org.OrgType
	}
	if orgTyp == app.OrgTypeIntegration {
		return nil
	}

	sandboxMode := false
	if org != nil {
		sandboxMode = org.SandboxMode
	}

	workflowID := signal.WorkflowID(id)
	opts := tclient.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: workflows.APITaskQueue,
		// Memo is non-indexed metadata available when listing workflows
		Memo: map[string]interface{}{
			"id":         id,
			"started-by": "api",
		},
		WorkflowIDReusePolicy: enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 0,
		},
	}

	req := EventLoopRequest{
		ID:          id,
		SandboxMode: sandboxMode,
		Version:     a.cfg.Version,
	}
	wkflowRun, err := a.client.ExecuteWorkflowInNamespace(ctx,
		signal.Namespace(),
		opts,
		signal.WorkflowName(),
		req)
	if err != nil {
		return err
	}

	a.l.Debug("started event loop",
		zap.String("workflow-id", wkflowRun.GetID()),
		zap.String("run-id", wkflowRun.GetID()),
		zap.String("id", id),
		zap.Error(err),
	)
	return nil
}

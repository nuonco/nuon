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

func (a *Hooks) Restart(ctx context.Context, orgID string) {
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
	wkflowRun, err := a.client.ExecuteWorkflowInNamespace(ctx,
		defaultNamespace,
		opts,
		worker.EventLoopWorkflowName,
		orgID)
	if err != nil {
		log.Fatalln("error creating org event loop", err)
		return
	}
	a.l.Debug("started org event loop",
		zap.String("workflow-id", wkflowRun.GetID()),
		zap.String("run-id", wkflowRun.GetID()),
		zap.String("org-id", orgID),
		zap.Error(err),
	)
}

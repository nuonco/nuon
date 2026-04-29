package enqueuer

import (
	"time"

	enumsv1 "go.temporal.io/api/enums/v1"
	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"

	"github.com/nuonco/nuon/pkg/workflows"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue"
)

func (e *Enqueuer) queueStartOperation(q *app.Queue) tclient.WithStartWorkflowOperation {
	wkflowReq := queue.QueueWorkflowRequest{
		QueueID: q.ID,
		Version: e.cfg.Version,
	}
	startOpts := tclient.StartWorkflowOptions{
		ID:        q.Workflow.ID,
		TaskQueue: workflows.APITaskQueue,
		Memo: map[string]any{
			"id":           q.ID,
			"owner-id":     q.OwnerID,
			"owner-type":   q.OwnerType,
			"idle-timeout": time.Duration(q.IdleTimeout).String(),
		},
		WorkflowIDConflictPolicy: enumsv1.WORKFLOW_ID_CONFLICT_POLICY_USE_EXISTING,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 0,
		},
	}
	return e.tClient.NewWithStartWorkflowOperation(startOpts, "Queue", wkflowReq)
}

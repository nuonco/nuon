package worker

import (
	"fmt"
	"math"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/general/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/general/worker/activities"

	enumsv1 "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/workflow"
)

func (w *Workflows) reconcile(ctx workflow.Context, _ signals.RequestSignal) error {
	workflowId := fmt.Sprintf("reconcile-event-loops")
	cwo := workflow.ChildWorkflowOptions{
		WorkflowID:            workflowId,
		WorkflowIDReusePolicy: enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
		ParentClosePolicy:     enumsv1.PARENT_CLOSE_POLICY_TERMINATE,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	workflow.ExecuteChildWorkflow(ctx, w.ReconcileEventLoops)
	return nil
}

func (w *Workflows) ReconcileEventLoops(ctx workflow.Context) error {
	// this is the meat of the work here - executes this work immediately
	response, err := activities.AwaitEnsureEventLoop(ctx, activities.EnsureEventLoopsRequest{})
	if err != nil {
		return fmt.Errorf("unable to start EnsureEventLoop: %w", err)
	}

	for _, res := range response {
		w.logger.Debug(fmt.Sprintf("%+v\n", res))
		// create paginated requests and fire off the Page activity
		pages := int(math.Ceil(float64(res.RowCount / activities.DefaultPageSize)))
		for page := 0; page <= pages; page++ {
			pageReq := activities.EnsureEventLoopsPageRequest{
				Namespace: res.Namespace,
				Offset:    int(int64(page) * activities.DefaultPageSize),
				Limit:     int(activities.DefaultPageSize),
			}
			_, err := activities.AwaitEnsureEventLoopPage(ctx, pageReq)
			if err != nil {
				w.logger.Error(fmt.Sprintf("%v", err))
			}
		}

	}
	return nil
}

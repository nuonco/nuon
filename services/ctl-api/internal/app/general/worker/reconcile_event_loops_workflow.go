package worker

import (
	"time"

	enumsv1 "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/general/worker/activities"
)

const (
	// the health check runs at minute 17 of every hour
	// https://crontab.guru/#17_*/1_*_*_*
	reconcileWorkflowCronTab string = "17 */1 * * *"

	// default ping waypoint timeout
	defaultPingWaypointTimeout time.Duration = time.Second * 10
)

func (w *Workflows) startReconcileEventLoopsWorkflowCron(ctx workflow.Context, req activities.EnsureEventLoopsRequest) {
	workflowId := "reconcile-event-loops"
	cwo := workflow.ChildWorkflowOptions{
		WorkflowID:            workflowId,
		CronSchedule:          reconcileWorkflowCronTab,
		WorkflowIDReusePolicy: enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
		ParentClosePolicy:     enumsv1.PARENT_CLOSE_POLICY_TERMINATE,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	workflow.ExecuteChildWorkflow(ctx, w.SendReconcileSignal, req)
}

// invoked by a workflow cron
func (w *Workflows) SendReconcileSignal(ctx workflow.Context) error {
	err := activities.AwaitSendReconcileSignal(ctx, activities.EnsureEventLoopsRequest{})
	if err != nil {
		return err
	}
	return nil
}

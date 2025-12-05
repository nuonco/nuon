package worker

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/general/worker/activities"
	"go.temporal.io/sdk/workflow"

	enumsv1 "go.temporal.io/api/enums/v1"
)

const (
	purgeStaleDataWorkflowName string = "general-purge-stale-data"
)

func (w *Workflows) startPurgeStaleDataWorkflow(ctx workflow.Context) {
	cwo := workflow.ChildWorkflowOptions{
		WorkflowID:            purgeStaleDataWorkflowName,
		CronSchedule:          w.cfg.EventLoopGeneralPurgeStaleDataCron,
		WorkflowIDReusePolicy: enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
		ParentClosePolicy:     enumsv1.PARENT_CLOSE_POLICY_TERMINATE,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)
	workflow.ExecuteChildWorkflow(ctx, w.PurgeStaleData)
}

func (w *Workflows) PurgeStaleData(ctx workflow.Context) error {

	err := activities.AwaitPurgeStaleTemporalPayloads(ctx, activities.PurgeStaleTemporalPayloadsRequest{
		DurationAgo: w.cfg.EventLoopGeneralPurgeStaleDataDurationAgo,
	})
	if err != nil {
		return err
	}

	return nil
}

package worker

import (
	"fmt"

	enumsv1 "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/actions"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

func actionWorkflowTriggerWorkflowID(installID string, actionWorkflowID string) string {
	return fmt.Sprintf("event-loop-%s-action-workflow-trigger-%s", installID, actionWorkflowID)
}

func (w *Workflows) ActionWorkflowTriggers(ctx workflow.Context, sreq signals.RequestSignal) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}
	l.Info("starting children event loops for action crons")

	workflows, err := activities.AwaitGetActionWorkflowsByInstallID(ctx, sreq.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get action workflow run")
	}

	for _, workflow := range workflows {
		cfg, err := activities.AwaitGetActionWorkflowLatestConfigByActionWorkflowID(ctx, workflow.ActionWorkflowID)
		if err != nil {
			return errors.Wrap(err, "unable to get action workflow config")
		}

		// sync the workflow cron
		if cfg.CronTrigger == nil {
			continue
		}

		if err := w.startActionWorkflowCronTrigger(ctx, sreq, workflow, cfg.CronTrigger); err != nil {
			return errors.Wrap(err, "unable to sync action workflow trigger")
		}
	}

	if err := workflow.Await(ctx, func() bool {
		return ctx.Err() != nil
	}); err != nil {
		if temporal.IsCanceledError(err) {
			return workflow.NewContinueAsNewError(ctx, workflow.GetInfo(ctx).WorkflowType.Name, sreq)
		}

		return workflow.NewContinueAsNewError(ctx, workflow.GetInfo(ctx).WorkflowType.Name, sreq)
	}

	return workflow.NewContinueAsNewError(ctx, workflow.GetInfo(ctx).WorkflowType.Name, sreq)
}

func (w *Workflows) startActionWorkflowCronTrigger(ctx workflow.Context, sreq signals.RequestSignal, iw *app.InstallActionWorkflow, trigger *app.ActionWorkflowTriggerConfig) error {
	cwo := workflow.ChildWorkflowOptions{
		WorkflowID:            actionWorkflowTriggerWorkflowID(sreq.ID, iw.ID),
		CronSchedule:          trigger.CronSchedule,
		WorkflowIDReusePolicy: enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
		ParentClosePolicy:     enumsv1.PARENT_CLOSE_POLICY_TERMINATE,
	}

	req := signals.NewRequestSignal(sreq.EventLoopRequest, &signals.Signal{
		Type: signals.OperationExecuteActionWorkflow,
		InstallActionWorkflowTrigger: signals.InstallActionWorkflowTriggerSubSignal{
			InstallActionWorkflowID: iw.ID,
			TriggerType:             app.ActionWorkflowTriggerTypeCron,
			TriggeredByType:         "cron",
			RunEnvVars: map[string]string{
				"TRIGGER_TYPE": "cron",
			},
		},
	})

	dctx := workflow.WithChildOptions(ctx, cwo)
	var wkflows actions.Workflows
	workflow.ExecuteChildWorkflow(dctx, wkflows.ExecuteActionWorkflow, req)

	return nil
}

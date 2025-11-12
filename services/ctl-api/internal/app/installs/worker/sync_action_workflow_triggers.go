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

	if !w.cfg.ActionCronsEnabled {
		return nil
	}

	install, err := activities.AwaitGetByInstallID(ctx, sreq.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get install")
	}

	fullAppConfig, err := activities.AwaitGetAppConfigByID(ctx, install.AppConfigID)
	if err != nil {
		return errors.Wrap(err, "unable to get app config")
	}

	awcMap := make(map[string]app.ActionWorkflowConfig, len(fullAppConfig.ActionWorkflowConfigs))
	for _, awc := range fullAppConfig.ActionWorkflowConfigs {
		awcMap[awc.ActionWorkflowID] = awc
	}

	workflows, err := activities.AwaitGetActionWorkflowsByInstallID(ctx, sreq.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get install action workflows")
	}

	for _, workflow := range workflows {
		cfg, ok := awcMap[workflow.ActionWorkflowID]
		if !ok {
			// skip action workflows that are not part of current app config
			continue
		}

		// sync the workflow cron
		if cfg.CronTrigger == nil {
			continue
		}

		if err := w.startActionWorkflowCronTrigger(ctx, sreq, workflow.ID, cfg.CronTrigger); err != nil {
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

func (w *Workflows) startActionWorkflowCronTrigger(ctx workflow.Context, sreq signals.RequestSignal, iawID string, trigger *app.ActionWorkflowTriggerConfig) error {
	cwo := workflow.ChildWorkflowOptions{
		WorkflowID:            actionWorkflowTriggerWorkflowID(sreq.ID, iawID),
		CronSchedule:          trigger.CronSchedule,
		WorkflowIDReusePolicy: enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
		ParentClosePolicy:     enumsv1.PARENT_CLOSE_POLICY_TERMINATE,
	}

	req := signals.NewRequestSignal(sreq.EventLoopRequest, &signals.Signal{
		Type: signals.OperationExecuteActionWorkflow,
		InstallActionWorkflowTrigger: signals.InstallActionWorkflowTriggerSubSignal{
			InstallActionWorkflowID: iawID,
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

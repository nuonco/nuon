package worker

import (
	"fmt"

	enumsv1 "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
)

func actionWorkflowTriggerWorkflowID(installID string, actionWorkflowName string) string {
	actionWorkflowName = generics.SystemName(actionWorkflowName)
	return fmt.Sprintf("action-workflow-trigger-%s-%s", installID, actionWorkflowName)
}

// @temporal-gen workflow
// @execution-timeout 1m
// @task-timeout 30s
func (w *Workflows) SyncActionWorkflowTriggers(ctx workflow.Context, sreq signals.RequestSignal) error {
	workflows, err := activities.AwaitGetActionWorkflowsByInstallID(ctx, sreq.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get action workflow run")
	}

	for _, workflow := range workflows {
		cfg, err := activities.AwaitGetActionWorkflowLatestConfigByActionWorkflowID(ctx, workflow.ID)
		if err != nil {
			return errors.Wrap(err, "unable to get action workflow config")
		}

		for _, trigger := range cfg.CronTriggers {
			if err := w.syncActionWorkflowCronTrigger(ctx, sreq, workflow, &trigger); err != nil {
				return errors.Wrap(err, "unable to sync action workflow trigger")
			}
		}
	}

	return nil
}

func (w *Workflows) syncActionWorkflowCronTrigger(ctx workflow.Context, sreq signals.RequestSignal, aw *app.ActionWorkflow, trigger *app.ActionWorkflowTriggerConfig) error {
	cwo := workflow.ChildWorkflowOptions{
		WorkflowID:            actionWorkflowTriggerWorkflowID(sreq.ID, aw.ID),
		CronSchedule:          trigger.CronSchedule,
		WorkflowIDReusePolicy: enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
		ParentClosePolicy:     enumsv1.PARENT_CLOSE_POLICY_TERMINATE,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	workflow.ExecuteChildWorkflow(ctx, w.CronActionWorkflow, &CronActionWorkflowRequest{
		InstallID:        sreq.ID,
		ActionWorkflowID: aw.ID,
		RunEnvVars: map[string]*string{
			"TRIGGER": generics.ToPtr("cron"),
		},
	})

	return nil
}

package worker

import (
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

// Run all workflow actions defined for a lifecycle hook
type LifecycleActionWorkflowsRequest struct {
	InstallID string `validate:"required" json:"install_id"`
	Trigger   app.ActionWorkflowTriggerType

	RunEnvVars map[string]*string `validate:"required" json:"data"`
}

func LifecycleActionWorkflowsID(req *LifecycleActionWorkflowsRequest) string {
	return fmt.Sprintf("action-workflows-lifecycle-%s-%s", req.Trigger, req.InstallID)
}

// @temporal-gen workflow
// @execution-timeout 1h
// @task-timeout 30s
// @id-callback LifecycleActionWorkflowsID
func (w *Workflows) LifecycleActionWorkflows(ctx workflow.Context, req *LifecycleActionWorkflowsRequest) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}
	l.Info("executing actions with trigger " + string(req.Trigger))

	workflows, err := activities.AwaitGetActionWorkflowsByInstallID(ctx, req.InstallID)
	if err != nil {
		return errors.Wrap(err, "unable to get action workflow run")
	}

	for _, workflow := range workflows {
		cfg, err := activities.AwaitGetActionWorkflowLatestConfigByActionWorkflowID(ctx, workflow.ID)
		if err != nil {
			return errors.Wrap(err, "unable to get action workflow config")
		}

		for _, trigger := range cfg.LifecycleTriggers {
			if trigger.Type != req.Trigger {
				continue
			}

			l.Info("executing action " + workflow.Name)
			if err := w.lifecycleActionWorkflow(ctx, req.InstallID, workflow.ID, req.Trigger, req.RunEnvVars); err != nil {
				return errors.Wrap(err, "unable to sync action workflow trigger")
			}
		}
	}

	return nil
}

func (w *Workflows) lifecycleActionWorkflow(ctx workflow.Context, installID, actionWorkflowID string, triggerType app.ActionWorkflowTriggerType, runEnvVars map[string]*string) error {
	actionWorkflowRun, err := activities.AwaitCreateActionWorkflowRun(ctx, &activities.CreateActionWorkflowRunRequest{
		InstallID:        installID,
		ActionWorkflowID: actionWorkflowID,
		TriggerType:      triggerType,
		RunEnvVars:       runEnvVars,
	})
	if err != nil {
		return errors.Wrap(err, "unable to create action workflow config")
	}

	if err := w.actionWorkflowRun(ctx, installID, actionWorkflowRun.ID); err != nil {
		return errors.Wrap(err, "unable to create action workflow run")
	}

	return nil
}

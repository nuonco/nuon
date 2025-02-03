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

	TriggerType     app.ActionWorkflowTriggerType `json:"trigger_type"`
	TriggeredByID   string                        `json:"triggered_by_id"`
	TriggeredByType string                        `json:"triggered_by_type"`

	RunEnvVars map[string]*string `validate:"required" json:"run_env_vars"`
}

func LifecycleActionWorkflowsID(req *LifecycleActionWorkflowsRequest) string {
	return fmt.Sprintf("action-workflows-lifecycle-%s-%s", req.TriggerType, req.InstallID)
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
	l.Info("executing actions with trigger " + string(req.TriggerType))

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
			if trigger.Type != req.TriggerType {
				continue
			}

			l.Info("executing action " + workflow.Name)
			if err := w.lifecycleActionWorkflow(ctx,
				req.InstallID,
				workflow.ID,
				req.TriggerType,
				req.RunEnvVars,
				req.TriggeredByID,
				req.TriggeredByType,
			); err != nil {
				return errors.Wrap(err, "unable to sync action workflow trigger")
			}
		}
	}

	return nil
}

func (w *Workflows) lifecycleActionWorkflow(ctx workflow.Context,
	installID,
	actionWorkflowID string,
	triggerType app.ActionWorkflowTriggerType,
	runEnvVars map[string]*string,
	triggeredByID string,
	triggeredByType string,
) error {
	actionWorkflowRun, err := activities.AwaitCreateActionWorkflowRun(ctx, &activities.CreateActionWorkflowRunRequest{
		InstallID:        installID,
		ActionWorkflowID: actionWorkflowID,
		TriggerType:      triggerType,
		TriggeredByID:    triggeredByID,
		TriggeredByType:  triggeredByType,
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

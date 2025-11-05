package actions

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

	installActionWorkflows, err := activities.AwaitGetActionWorkflowsByInstallID(ctx, req.InstallID)
	if err != nil {
		return errors.Wrap(err, "unable to get action workflow run")
	}

	for _, installWorkflow := range installActionWorkflows {
		cfg, err := activities.AwaitGetActionWorkflowLatestConfigByActionWorkflowID(ctx, installWorkflow.ActionWorkflowID)
		if err != nil {
			return errors.Wrap(err, "unable to get action workflow config")
		}

		for _, trigger := range cfg.LifecycleTriggers {
			if trigger.Type != req.TriggerType {
				continue
			}

			l.Info("executing action " + installWorkflow.ActionWorkflow.Name)
			if err := w.lifecycleActionWorkflow(ctx,
				req.InstallID,
				installWorkflow.ID,
				installWorkflow.ActionWorkflowID,
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
	installActionWorkflowID,
	actionWorkflowID string,
	triggerType app.ActionWorkflowTriggerType,
	runEnvVars map[string]*string,
	triggeredByID string,
	triggeredByType string,
) error {
	actionWorkflowRun, err := activities.AwaitCreateActionWorkflowRun(ctx, &activities.CreateActionWorkflowRunRequest{
		InstallActionWorkflowID: installActionWorkflowID,
		ActionWorkflowID:        actionWorkflowID,
		InstallID:               installID,
		TriggerType:             triggerType,
		TriggeredByID:           triggeredByID,
		TriggeredByType:         triggeredByType,
		RunEnvVars:              runEnvVars,
	})
	if err != nil {
		return errors.Wrap(err, "unable to create action workflow config")
	}

	if err := w.executeActionWorkflowRun(ctx, installID, actionWorkflowRun.ID); err != nil {
		return errors.Wrap(err, "unable to create action workflow run")
	}

	return nil
}

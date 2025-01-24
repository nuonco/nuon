package workflow

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"

	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"

	"github.com/nuonco/nuon-runner-go/models"
)

func (h *handler) Fetch(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	h.state = &handlerState{}
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	// fetch the plan json
	l.Info("fetching actions job plan")
	planJSON, err := h.apiClient.GetJobPlanJSON(ctx, job.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get job plan")
	}

	// parse the plan
	var plan plantypes.ActionWorkflowRunPlan
	if err := json.Unmarshal([]byte(planJSON), &plan); err != nil {
		return errors.Wrap(err, "unable to parse action workflow run plan")
	}
	h.state.plan = &plan

	// fetch the run object
	run, err := h.apiClient.GetInstallActionWorkflowRun(ctx,
		plan.InstallID,
		plan.ID,
	)
	if err != nil {
		return errors.Wrap(err, "unable to get action workflow run")
	}
	h.state.run = run

	// fetch the workflow config
	l.Info("fetching actions workflow config")
	cfg, err := h.apiClient.GetActionWorkflowConfig(ctx, run.ActionWorkflowConfigID)
	if err != nil {
		return errors.Wrap(err, "unable to get action workflow run")
	}
	h.state.workflowCfg = cfg

	return nil
}

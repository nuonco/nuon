package jobloop

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cockroachdb/errors"
	pkgctx "github.com/nuonco/nuon/bins/runner/internal/pkg/ctx"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"go.uber.org/zap"
)

func (h *jobLoop) execActionSandboxStep(ctx context.Context, job *models.AppRunnerJob) error {
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

	// fetch the run object
	run, err := h.apiClient.GetInstallActionWorkflowRun(ctx,
		plan.InstallID,
		plan.ID,
	)
	if err != nil {
		return errors.Wrap(err, "unable to get action workflow run")
	}

	isAdhoc := run.ActionWorkflowConfigID == ""

	var cfg *models.AppActionWorkflowConfig
	if !isAdhoc {
		l.Info("fetching actions workflow config")
		cfg, err = h.apiClient.GetActionWorkflowConfig(ctx, run.ActionWorkflowConfigID)
		if err != nil {
			return errors.Wrap(err, "unable to get action workflow config")
		}
	}

	for idx, step := range run.Steps {
		var stepName string
		var actionWorkflowID string

		if isAdhoc {
			if step.AdhocConfig != nil {
				stepName = step.AdhocConfig.Name
			} else {
				stepName = "adhoc step"
			}
			actionWorkflowID = run.ID
		} else {
			stepCfg := cfg.Steps[idx]
			stepName = stepCfg.Name
			actionWorkflowID = cfg.ActionWorkflowID
		}

		l = l.With(
			zap.String("workflow_step_name", stepName),
			zap.String("step_run_id", step.ID),
		)

		l.Info(fmt.Sprintf("executing step %s (%d of %d)", stepName, idx+1, len(run.Steps)))

		_, err := h.apiClient.UpdateInstallActionWorkflowRunStep(ctx, plan.InstallID, actionWorkflowID, step.ID, &models.ServiceUpdateInstallActionWorkflowRunStepRequest{
			Status:            models.AppInstallActionWorkflowRunStepStatusFinished,
			ExecutionDuration: int64(time.Second * 5),
		})
		if err != nil {
			return errors.Wrap(err, "unable to update step status")
		}
	}

	return nil
}

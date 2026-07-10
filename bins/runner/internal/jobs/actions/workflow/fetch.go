package workflow

import (
	"context"

	"github.com/pkg/errors"

	pkgplantypes "github.com/nuonco/nuon/bins/runner/internal/pkg/plantypes"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func (h *handler) Fetch(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	h.state = &handlerState{}
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	// fetch the plan json
	l.Info("fetching actions job plan")
	cp, err := h.apiClient.GetJobCompositePlan(ctx, job.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get job plan")
	}

	composite, err := plantypes.CompositePlanFromAny(cp)
	if err != nil {
		return errors.Wrap(err, "unable to parse composite plan")
	}
	if composite.ActionWorkflowRunPlan == nil {
		return errors.New("composite plan missing action workflow run plan")
	}
	plan := composite.ActionWorkflowRunPlan
	h.state.plan = plan

	h.state.auth = &pkgplantypes.PlanAuth{
		AWSAuth:   plan.AWSAuth,
		AzureAuth: plan.AzureAuth,
		GCPAuth:   plan.GCPAuth,
	}

	// fetch the run object
	run, err := h.apiClient.GetInstallActionWorkflowRun(ctx,
		plan.InstallID,
		plan.ID,
	)
	if err != nil {
		return errors.Wrap(err, "unable to get action workflow run")
	}
	h.state.run = run

	// fetch the workflow config (skip for adhoc runs)
	if run.ActionWorkflowConfigID != "" {
		l.Info("fetching actions workflow config")
		cfg, err := h.apiClient.GetActionWorkflowConfig(ctx, run.ActionWorkflowConfigID)
		if err != nil {
			return errors.Wrap(err, "unable to get action workflow config")
		}
		h.state.workflowCfg = cfg
	}

	return nil
}

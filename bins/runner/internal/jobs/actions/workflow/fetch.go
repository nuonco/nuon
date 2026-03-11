package workflow

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"

	pkgctx "github.com/nuonco/nuon/bins/runner/internal/pkg/ctx"
	pkgplantypes "github.com/nuonco/nuon/bins/runner/internal/pkg/plantypes"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"

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

	l.Info("fetching composite plan for the job")
	compositePlan, err := h.apiClient.GetJobCompositePlan(ctx, job.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get job plan")
	}

	h.state.auth, err = pkgplantypes.PlanAuthFromSDK(&compositePlan.PlanAuth.PlantypesPlanAuth)
	if err != nil {
		return errors.Wrap(err, "unable to build plan auth for job")
	}

	if h.state.plan.ClusterInfo != nil {
		h.state.plan.ClusterInfo.WithAWSAuth(h.state.auth.AWSAuth)
		h.state.plan.ClusterInfo.WithAzureAuth(h.state.auth.AzureAuth)
		h.state.plan.ClusterInfo.WithGCPAuth(h.state.auth.GCPAuth != nil)
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

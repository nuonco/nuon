package helm

import (
	"context"
	"time"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"github.com/pkg/errors"

	pkgplantypes "github.com/nuonco/nuon/bins/runner/internal/pkg/plantypes"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
)

func (h *handler) Fetch(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	h.state = &handlerState{}

	l.Info("fetching helm job plan")
	cp, err := h.apiClient.GetJobCompositePlan(ctx, job.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get job plan")
	}

	composite, err := plantypes.CompositePlanFromAny(cp)
	if err != nil {
		return errors.Wrap(err, "unable to parse composite plan")
	}
	if composite.DeployPlan == nil {
		return errors.New("composite plan missing deploy plan")
	}
	plan := composite.DeployPlan
	h.state.plan = plan

	h.state.auth = &pkgplantypes.PlanAuth{
		AWSAuth:   plan.HelmDeployPlan.AWSAuth,
		AzureAuth: plan.HelmDeployPlan.AzureAuth,
		GCPAuth:   plan.HelmDeployPlan.GCPAuth,
	}

	l.Info("fetching app config")
	appCfg, err := h.apiClient.GetAppConfig(ctx, plan.AppID, plan.AppConfigID)
	if err != nil {
		return errors.Wrap(err, "unable to get app config")
	}
	h.state.appCfg = appCfg

	for _, cfg := range appCfg.ComponentConfigConnections {
		if cfg.ComponentID != plan.ComponentID {
			continue
		}

		h.state.helmCfg = cfg.Helm
	}
	if h.state.helmCfg == nil {
		return errors.New("unable to find helm config")
	}

	h.state.jobID = job.ID
	h.state.jobExecutionID = jobExecution.ID

	h.state.timeout = time.Duration(job.ExecutionTimeout)

	return nil
}

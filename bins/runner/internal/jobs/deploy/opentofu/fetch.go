package opentofu

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

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

	l.Info("fetching opentofu job plan")
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
		AWSAuth:   plan.OpentofuDeployPlan.AWSAuth,
		AzureAuth: plan.OpentofuDeployPlan.AzureAuth,
		GCPAuth:   plan.OpentofuDeployPlan.GCPAuth,
	}

	l.Info("fetching app config")
	appCfg, err := h.apiClient.GetAppConfig(ctx, plan.AppID, plan.AppConfigID)
	if err != nil {
		return errors.Wrap(err, "unable to get app config")
	}
	h.state.appCfg = appCfg

	l.Info("fetching opentofu config")
	for _, cfg := range appCfg.ComponentConfigConnections {
		if cfg.ComponentID != plan.ComponentID {
			continue
		}

		h.state.opentofuCfg = cfg.OpentofuModule
	}
	if h.state.opentofuCfg == nil {
		return errors.New("unable to find opentofu config")
	}

	h.state.jobID = job.ID
	h.state.jobExecutionID = jobExecution.ID

	h.state.timeout = time.Duration(job.ExecutionTimeout)
	l.Info("setting opentofu operation timeout", zap.String("duration", h.state.timeout.String()))

	return nil
}

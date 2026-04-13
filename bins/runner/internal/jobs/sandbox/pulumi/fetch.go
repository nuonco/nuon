package pulumi

import (
	"context"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	pkgctx "github.com/nuonco/nuon/bins/runner/internal/pkg/ctx"
	pkgplantypes "github.com/nuonco/nuon/bins/runner/internal/pkg/plantypes"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
)

func (h *handler) Fetch(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	h.state = &handlerState{}

	l.Info("fetching sandbox pulumi job plan")
	planJSON, err := h.apiClient.GetJobPlanJSON(ctx, job.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get job plan")
	}

	var plan plantypes.DeployPlan
	if err := json.Unmarshal([]byte(planJSON), &plan); err != nil {
		return errors.Wrap(err, "unable to parse deploy plan")
	}
	h.state.plan = &plan

	if plan.PulumiDeployPlan == nil {
		return errors.New("deploy plan does not contain a pulumi deploy plan")
	}

	h.state.auth = &pkgplantypes.PlanAuth{
		AWSAuth:   plan.PulumiDeployPlan.AWSAuth,
		AzureAuth: plan.PulumiDeployPlan.AzureAuth,
		GCPAuth:   plan.PulumiDeployPlan.GCPAuth,
	}

	l.Info("fetching app config")
	appCfg, err := h.apiClient.GetAppConfig(ctx, plan.AppID, plan.AppConfigID)
	if err != nil {
		return errors.Wrap(err, "unable to get app config")
	}

	l.Info("fetching pulumi config")
	for _, cfg := range appCfg.ComponentConfigConnections {
		if cfg.ComponentID != plan.ComponentID {
			continue
		}
		h.state.pulumiCfg = cfg.Pulumi
	}
	if h.state.pulumiCfg == nil {
		return errors.New("unable to find pulumi config")
	}

	h.state.jobID = job.ID
	h.state.jobExecutionID = jobExecution.ID

	h.state.timeout = time.Duration(job.ExecutionTimeout)
	l.Info("setting sandbox pulumi operation timeout", zap.String("duration", h.state.timeout.String()))

	return nil
}

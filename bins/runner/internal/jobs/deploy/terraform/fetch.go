package terraform

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nuonco/nuon-runner-go/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
)

func (h *handler) Fetch(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	h.state = &handlerState{}

	l.Info("fetching terraform job plan")
	planJSON, err := h.apiClient.GetJobPlanJSON(ctx, job.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get job plan")
	}

	// parse the plan
	var plan plantypes.DeployPlan
	if err := json.Unmarshal([]byte(planJSON), &plan); err != nil {
		return errors.Wrap(err, "unable to parse sandbox workflow run plan")
	}
	h.state.plan = &plan

	l.Info("fetching app config")
	appCfg, err := h.apiClient.GetAppConfig(ctx, plan.AppID, plan.AppConfigID)
	if err != nil {
		return errors.Wrap(err, "unable to get app config")
	}
	h.state.appCfg = appCfg

	l.Info("fetching terraform config")
	for _, cfg := range appCfg.ComponentConfigConnections {
		if cfg.ComponentID != plan.ComponentID {
			continue
		}

		h.state.terraformCfg = cfg.TerraformModule
	}
	if h.state.terraformCfg == nil {
		return errors.New("unable to find terraform config")
	}

	h.state.jobID = job.ID
	h.state.jobExecutionID = jobExecution.ID

	h.state.timeout = time.Duration(job.ExecutionTimeout)
	l.Info("setting terraform operation timeout", zap.String("duration", h.state.timeout.String()))

	return nil
}

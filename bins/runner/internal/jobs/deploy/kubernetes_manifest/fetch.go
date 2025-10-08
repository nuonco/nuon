package kubernetes_manifest

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

	l.Info("fetching kubernetes manifest job plan")
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

	h.state.jobID = job.ID
	h.state.jobExecutionID = jobExecution.ID

	h.state.timeout = time.Duration(job.ExecutionTimeout)

	h.state.kubeClient, err = h.getClient(ctx)
	if err != nil {
		l.Debug("unable to initialize kube client", zap.String("jobID", job.ID))
		return errors.Wrap(err, "unable to intialize kubeclient")
	}

	return nil
}

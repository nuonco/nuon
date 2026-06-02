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

	l.Info("fetching pulumi sandbox job plan")
	planJSON, err := h.apiClient.GetJobPlanJSON(ctx, job.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get job plan")
	}

	var plan plantypes.SandboxRunPlan
	if err := json.Unmarshal([]byte(planJSON), &plan); err != nil {
		return errors.Wrap(err, "unable to parse sandbox run plan")
	}
	h.state.plan = &plan

	if plan.PulumiBackend == nil {
		return errors.New("sandbox run plan does not contain a pulumi backend")
	}

	h.state.auth = &pkgplantypes.PlanAuth{
		AWSAuth:   plan.AWSAuth,
		AzureAuth: plan.AzureAuth,
		GCPAuth:   plan.GCPAuth,
	}

	h.state.jobID = job.ID
	h.state.jobExecutionID = jobExecution.ID
	h.state.timeout = time.Duration(job.ExecutionTimeout)
	l.Info("setting sandbox pulumi operation timeout", zap.String("duration", h.state.timeout.String()))

	return nil
}

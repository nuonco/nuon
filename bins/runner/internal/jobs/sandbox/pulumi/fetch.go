package pulumi

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

	l.Info("fetching pulumi sandbox job plan")
	cp, err := h.apiClient.GetJobCompositePlan(ctx, job.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get job plan")
	}

	composite, err := plantypes.CompositePlanFromAny(cp)
	if err != nil {
		return errors.Wrap(err, "unable to parse composite plan")
	}
	if composite.SandboxRunPlan == nil {
		return errors.New("composite plan missing sandbox run plan")
	}
	plan := composite.SandboxRunPlan
	h.state.plan = plan

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

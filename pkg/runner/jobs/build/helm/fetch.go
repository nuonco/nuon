package helm

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"github.com/pkg/errors"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
)

func (h *handler) Fetch(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	h.state = &handlerState{}

	l.Info("fetching job plan")
	cp, err := h.apiClient.GetJobCompositePlan(ctx, job.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get job plan")
	}

	l.Info("parsing job plan")
	composite, err := plantypes.CompositePlanFromAny(cp)
	if err != nil {
		return errors.Wrap(err, "unable to parse composite plan")
	}
	if composite.BuildPlan == nil {
		return errors.New("composite plan missing build plan")
	}
	plan := composite.BuildPlan

	h.state.plan = plan
	h.state.jobID = job.ID
	h.state.jobExecutionID = jobExecution.ID
	h.state.cfg = plan.HelmBuildPlan
	h.state.regCfg = plan.Dst
	h.state.resultTag = job.OwnerID
	return nil
}

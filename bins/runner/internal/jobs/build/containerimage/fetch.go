package containerimage

import (
	"context"
	"encoding/json"

	"github.com/nuonco/nuon-runner-go/models"
	"github.com/pkg/errors"

	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
)

func (h *handler) Fetch(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	h.state = &handlerState{}

	l.Info("fetching job plan")
	planJSON, err := h.apiClient.GetJobPlanJSON(ctx, job.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get job plan")
	}

	// parse the plan
	l.Info("parsing job plan")
	var plan plantypes.BuildPlan
	if err := json.Unmarshal([]byte(planJSON), &plan); err != nil {
		return errors.Wrap(err, "unable to parse sandbox workflow run plan")
	}

	h.state.plan = &plan
	h.state.jobID = job.ID
	h.state.jobExecutionID = jobExecution.ID
	h.state.cfg = plan.ContainerImagePullPlan
	h.state.regCfg = plan.Dst
	h.state.resultTag = job.OwnerID
	return nil
}

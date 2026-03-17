package sandbox

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func (h *handler) Fetch(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	h.state = &handlerState{}

	planJSON, err := h.apiClient.GetJobPlanJSON(ctx, job.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get job plan")
	}

	var plan plantypes.BuildPlan
	if err := json.Unmarshal([]byte(planJSON), &plan); err != nil {
		return errors.Wrap(err, "unable to parse sandbox build plan")
	}

	h.state.plan = &plan
	h.state.jobID = job.ID
	h.state.jobExecutionID = jobExecution.ID
	h.state.cfg = plan.TerraformBuildPlan
	h.state.regCfg = plan.Dst
	h.state.resultTag = job.OwnerID
	return nil
}

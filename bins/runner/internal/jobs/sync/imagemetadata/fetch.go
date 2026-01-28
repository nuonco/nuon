package imagemetadata

import (
	"context"
	"encoding/json"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"github.com/pkg/errors"
)

func (h *handler) Fetch(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	h.state = &handlerState{}

	planJSON, err := h.apiClient.GetJobPlanJSON(ctx, job.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get job plan")
	}

	var plan FetchImageMetadataPlan
	if err := json.Unmarshal([]byte(planJSON), &plan); err != nil {
		return errors.Wrap(err, "unable to parse fetch image metadata plan")
	}
	h.state.plan = &plan

	h.state.jobID = job.ID
	h.state.jobExecutionID = jobExecution.ID
	return nil
}

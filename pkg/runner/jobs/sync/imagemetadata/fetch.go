package imagemetadata

import (
	"context"
	"encoding/json"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"github.com/pkg/errors"
)

func (h *handler) Fetch(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	h.state = &handlerState{}

	cp, err := h.apiClient.GetJobCompositePlan(ctx, job.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get job plan")
	}

	composite, err := plantypes.CompositePlanFromAny(cp)
	if err != nil {
		return errors.Wrap(err, "unable to parse composite plan")
	}
	if composite.FetchImageMetadataPlan == nil {
		return errors.New("composite plan missing fetch image metadata plan")
	}

	planJSON, err := json.Marshal(composite.FetchImageMetadataPlan)
	if err != nil {
		return errors.Wrap(err, "unable to marshal fetch image metadata plan")
	}

	var plan FetchImageMetadataPlan
	if err := json.Unmarshal(planJSON, &plan); err != nil {
		return errors.Wrap(err, "unable to parse fetch image metadata plan")
	}
	h.state.plan = &plan

	h.state.jobID = job.ID
	h.state.jobExecutionID = jobExecution.ID
	return nil
}

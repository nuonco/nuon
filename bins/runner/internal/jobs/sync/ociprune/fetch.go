package ociprune

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
)

func (h *handler) Fetch(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	h.state = &handlerState{
		jobID:          job.ID,
		jobExecutionID: jobExecution.ID,
	}

	cp, err := h.apiClient.GetJobCompositePlan(ctx, job.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get job plan")
	}

	composite, err := plantypes.CompositePlanFromAny(cp)
	if err != nil {
		return errors.Wrap(err, "unable to parse composite plan")
	}
	if composite.PruneOCIPlan == nil {
		return errors.New("composite plan missing prune oci plan")
	}
	h.state.plan = composite.PruneOCIPlan
	return nil
}

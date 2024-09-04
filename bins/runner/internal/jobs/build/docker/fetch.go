package docker

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-runner-go/models"

	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/plan"
)

func (h *handler) Fetch(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	if h.state != nil {
		return fmt.Errorf("another job is still in flight")
	}
	h.state = &handlerState{}

	plan, err := plan.FetchPlan(ctx, h.apiClient, job)
	if err != nil {
		return fmt.Errorf("unable to fetch plan: %w", err)
	}

	h.state.plan = plan
	h.state.jobID = job.ID
	h.state.jobExecutionID = jobExecution.ID

	return nil
}

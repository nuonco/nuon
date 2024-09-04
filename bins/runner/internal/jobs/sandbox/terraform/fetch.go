package terraform

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon-runner-go/models"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/plan"
)

func (h *handler) Fetch(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	h.state = &handlerState{}

	plan, err := plan.FetchPlan(ctx, h.apiClient, job)
	if err != nil {
		return fmt.Errorf("unable to fetch plan: %w", err)
	}

	h.state.plan = plan
	h.state.jobID = job.ID
	h.state.jobExecutionID = jobExecution.ID

	h.state.timeout = time.Duration(job.ExecutionTimeout)
	h.log.Info("setting helm operation timeout", zap.String("duration", h.state.timeout.String()))

	return nil
}

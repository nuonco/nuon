package helm

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon-runner-go/models"
	"go.uber.org/zap"

	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/plan"
)

func (h *handler) Fetch(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	h.state = &handlerState{}

	plan, err := plan.FetchPlan(ctx, h.apiClient, job)
	if err != nil {
		return fmt.Errorf("unable to fetch plan: %w", err)
	}

	h.state.plan = plan
	h.state.jobID = job.ID
	h.state.jobExecutionID = jobExecution.ID
	h.state.outputs = map[string]interface{}{}

	h.state.timeout = time.Duration(job.ExecutionTimeout)
	l.Info("setting helm operation timeout", zap.String("duration", h.state.timeout.String()))

	return nil
}

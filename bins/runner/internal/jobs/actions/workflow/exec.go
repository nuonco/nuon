package workflow

import (
	"context"
	"fmt"

	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"

	"github.com/nuonco/nuon-runner-go/models"
)

func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	for idx, step := range h.state.run.Steps {
		stepCfg := h.state.workflowCfg.Steps[idx]
		stepPlan := h.state.plan.Steps[idx]

		l.Info(fmt.Sprintf("executing step %s (%d of %d)", stepCfg.Name, idx+1, len(h.state.run.Steps)))
		if err := h.executeWorkflowStep(ctx, step, stepCfg, stepPlan); err != nil {
			return err
		}
	}

	return nil
}

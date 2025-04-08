package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"

	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"

	"github.com/nuonco/nuon-runner-go/models"
)

func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	execCtx, cancel := context.WithTimeout(ctx, time.Duration(h.state.workflowCfg.Timeout))
	defer cancel()

	for idx, step := range h.state.run.Steps {
		stepCfg := h.state.workflowCfg.Steps[idx]
		stepPlan := h.state.plan.Steps[idx]

		l.Info(fmt.Sprintf("executing step %s (%d of %d)", stepCfg.Name, idx+1, len(h.state.run.Steps)))
		err := h.executeWorkflowStep(ctx, execCtx, step, stepCfg, stepPlan)
		if err == nil {
			continue
		}

		remainingSteps := h.state.run.Steps[idx+1:]
		remainingCfgs := h.state.workflowCfg.Steps[idx+1:]
		if len(remainingSteps) > 0 {
			if err := h.noopWorkflowSteps(ctx, remainingSteps, remainingCfgs); err != nil {
				l.Warn(fmt.Sprintf("unable to mark %d remaining steps as NOOP after step.%d errored", idx, idx-1))
			}
		}

		return errors.Wrap(err, fmt.Sprintf("action workflow failed on step %d", idx))
	}

	return nil
}

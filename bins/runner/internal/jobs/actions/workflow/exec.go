package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"

	pkgctx "github.com/nuonco/nuon/bins/runner/internal/pkg/ctx"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	// For adhoc runs, workflowCfg is nil - use a default timeout
	timeout := 5 * time.Minute // default timeout
	if h.state.workflowCfg != nil {
		timeout = time.Duration(h.state.workflowCfg.Timeout)
	}
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for idx, step := range h.state.run.Steps {
		var stepCfg *models.AppActionWorkflowStepConfig
		var stepName string

		if h.state.workflowCfg != nil {
			stepCfg = h.state.workflowCfg.Steps[idx]
			stepName = stepCfg.Name
		} else if step.AdhocConfig != nil {
			// For adhoc runs, convert the adhoc config to a regular step config
			stepCfg = &models.AppActionWorkflowStepConfig{
				Command:        step.AdhocConfig.Command,
				InlineContents: step.AdhocConfig.InlineContents,
				Name:           step.AdhocConfig.Name,
				EnvVars:        step.AdhocConfig.EnvVars,
			}
			stepName = step.AdhocConfig.Name
		} else {
			stepName = "adhoc step"
		}
		stepPlan := h.state.plan.Steps[idx]

		l.Info(fmt.Sprintf("executing step %s (%d of %d)", stepName, idx+1, len(h.state.run.Steps)))
		err := h.executeWorkflowStep(ctx, execCtx, step, stepCfg, stepPlan)
		if err == nil {
			continue
		}

		remainingSteps := h.state.run.Steps[idx+1:]
		if len(remainingSteps) > 0 {
			// Build remaining configs for error handling
			var remainingCfgs []*models.AppActionWorkflowStepConfig
			if h.state.workflowCfg != nil {
				remainingCfgs = h.state.workflowCfg.Steps[idx+1:]
			} else {
				// For adhoc runs, build configs from step AdHocConfigs
				for _, s := range remainingSteps {
					if s.AdhocConfig != nil {
						remainingCfgs = append(remainingCfgs, &models.AppActionWorkflowStepConfig{
							Command:        s.AdhocConfig.Command,
							InlineContents: s.AdhocConfig.InlineContents,
							Name:           s.AdhocConfig.Name,
							EnvVars:        s.AdhocConfig.EnvVars,
						})
					}
				}
			}

			if err := h.noopWorkflowSteps(ctx, remainingSteps, remainingCfgs); err != nil {
				l.Warn(fmt.Sprintf("unable to mark %d remaining steps as NOOP after step.%d errored", idx, idx-1))
			}
		}

		return errors.Wrap(err, fmt.Sprintf("action workflow failed on step %d", idx))
	}

	return nil
}

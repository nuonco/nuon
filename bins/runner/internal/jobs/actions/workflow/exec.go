package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	pkgctx "github.com/nuonco/nuon/bins/runner/internal/pkg/ctx"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/op"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

// resolveStepConfig builds the step config for execution, preferring interpolated
// values from the plan (where Go templates have been rendered) over raw config values.
func resolveStepConfig(
	configStepCfg *models.AppActionWorkflowStepConfig,
	step *models.AppInstallActionWorkflowRunStep,
	stepPlan *plantypes.ActionWorkflowRunStepPlan,
) (*models.AppActionWorkflowStepConfig, string) {
	if configStepCfg != nil {
		if stepPlan.InterpolatedCommand != "" {
			configStepCfg.Command = stepPlan.InterpolatedCommand
		}
		if stepPlan.InterpolatedInlineContents != "" {
			configStepCfg.InlineContents = stepPlan.InterpolatedInlineContents
		}
		return configStepCfg, configStepCfg.Name
	}

	if step.AdhocConfig != nil {
		command := stepPlan.InterpolatedCommand
		if command == "" {
			command = step.AdhocConfig.Command
		}
		inlineContents := stepPlan.InterpolatedInlineContents
		if inlineContents == "" {
			inlineContents = step.AdhocConfig.InlineContents
		}
		return &models.AppActionWorkflowStepConfig{
			Command:        command,
			InlineContents: inlineContents,
			Name:           step.AdhocConfig.Name,
			EnvVars:        step.AdhocConfig.EnvVars,
		}, step.AdhocConfig.Name
	}

	return nil, "adhoc step"
}

func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	// Tag this handler's logger with semantic-convention attributes so every
	// emitted record (including from helpers further down the call tree) carries
	// them automatically.
	installID := ""
	if h.state.plan != nil {
		installID = h.state.plan.InstallID
	}
	actionRunID := ""
	actionWorkflowID := ""
	if h.state.run != nil {
		actionRunID = h.state.run.ID
		actionWorkflowID = h.state.run.InstallActionWorkflowID
	}
	l = l.With(
		zap.String("service.name", "runner.action"),
		zap.String("nuon.tool", "action"),
		zap.String("action.operation", string(job.Operation)),
		zap.String("action.run_id", actionRunID),
		zap.String("action.workflow_id", actionWorkflowID),
		zap.String("install.id", installID),
	)
	ctx = pkgctx.SetLogger(ctx, l)

	// For adhoc runs, workflowCfg is nil - use a default timeout
	timeout := 5 * time.Minute // default timeout
	if h.state.plan != nil && h.state.plan.Timeout > 0 {
		timeout = h.state.plan.Timeout
	} else if h.state.workflowCfg != nil && h.state.workflowCfg.Timeout > 0 {
		timeout = time.Duration(h.state.workflowCfg.Timeout)
	}
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for idx, step := range h.state.run.Steps {
		var configStepCfg *models.AppActionWorkflowStepConfig
		if h.state.workflowCfg != nil {
			configStepCfg = h.state.workflowCfg.Steps[idx]
		}
		stepPlan := h.state.plan.Steps[idx]
		stepCfg, stepName := resolveStepConfig(configStepCfg, step, stepPlan)
		if stepCfg != nil {
			stepCfg.Idx = int64(idx)
		}

		l.Info(fmt.Sprintf("executing step %s (%d of %d)", stepName, idx+1, len(h.state.run.Steps)))
		opExecCtx, end := op.Tool(execCtx, "action", "exec_step")
		err := h.executeWorkflowStep(ctx, opExecCtx, step, stepCfg, stepPlan)
		end(err)
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

package plan

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

func (p *Planner) createStepPlan(ctx workflow.Context,
	step *app.InstallActionWorkflowRunStep,
	stateMap map[string]any,
	installID string,
	isAdHoc bool,
) (*plantypes.ActionWorkflowRunStepPlan, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, err
	}

	var stepConfig *app.ActionWorkflowStepConfig
	if isAdHoc {
		cfg := app.ActionWorkflowStepConfig(*step.AdHocConfig)
		stepConfig = &cfg
	} else {
		stepConfig = &step.Step
	}

	plan := &plantypes.ActionWorkflowRunStepPlan{
		ID: step.ID,
		Attrs: map[string]string{
			"step.name": stepConfig.Name,
			"step.id":   stepConfig.ID,
		},
		InterpolatedEnvVars: make(map[string]string, 0),
	}

	if !isAdHoc {
		l.Debug("creating git source for config")
		gitSource, err := activities.AwaitGetActionWorkflowStepGitSourceByStepID(ctx, stepConfig.ID)
		if err != nil {
			l.Error("unable to  configure git source for step", zap.Error(err))
			return nil, errors.Wrap(err, "unable to get git source")
		}
		plan.GitSource = gitSource
	}

	for k, v := range stepConfig.EnvVars {
		renderedVal, err := render.Render(*v, stateMap)
		if err != nil {
			l.Error("error rendering env-var",
				zap.String("env-var", *v),
				zap.Error(err))
			return nil, err
		}

		plan.InterpolatedEnvVars[k] = renderedVal
	}

	if stepConfig.InlineContents != "" {
		l.Debug("rendering inline contents")
		renderedVal, err := render.Render(stepConfig.InlineContents, stateMap)
		if err != nil {
			l.Error("error rendering inline contents",
				zap.String("input", stepConfig.InlineContents),
				zap.Any("state", stateMap),
				zap.Error(err),
			)
			return nil, err
		}

		l.Debug("successfully rendered inline contents", zap.String("rendered", renderedVal))
		plan.InterpolatedInlineContents = renderedVal
	}

	if stepConfig.Command != "" {
		l.Debug("rendering command")
		renderedVal, err := render.Render(stepConfig.Command, stateMap)
		if err != nil {
			l.Error("error rendering command",
				zap.String("command", stepConfig.Command),
				zap.Any("state", stateMap),
				zap.Error(err),
			)
			return nil, err
		}

		l.Debug("successfully rendered command", zap.String("rendered", renderedVal))
		stepConfig.Command = renderedVal
	}

	return plan, nil
}

package plan

import (
	"go.uber.org/zap"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// RenderActionWorkflowStepPlan renders an action workflow step plan with no Temporal dependency.
func (p *Planner) RenderActionWorkflowStepPlan(l *zap.Logger,
	step *app.InstallActionWorkflowRunStep,
	stateMap map[string]any,
	gitSource *plantypes.GitSource,
) (*plantypes.ActionWorkflowRunStepPlan, error) {
	plan := &plantypes.ActionWorkflowRunStepPlan{
		ID: step.ID,
		Attrs: map[string]string{
			"step.name": step.Step.Name,
			"step.id":   step.Step.ID,
		},
		InterpolatedEnvVars: make(map[string]string, 0),
		GitSource:           gitSource,
	}

	for k, v := range step.Step.EnvVars {
		renderedVal, err := render.RenderV2(*v, stateMap)
		if err != nil {
			l.Error("error rendering env-var",
				zap.String("env-var", *v),
				zap.Error(err))
			return nil, err
		}

		plan.InterpolatedEnvVars[k] = renderedVal
	}

	if step.Step.InlineContents != "" {
		l.Debug("rendering inline contents")
		renderedVal, err := render.RenderV2(step.Step.InlineContents, stateMap)
		if err != nil {
			l.Error("error rendering inline contents",
				zap.String("input", step.Step.InlineContents),
				zap.Any("state", stateMap),
				zap.Error(err),
			)
			return nil, err
		}

		l.Debug("successfully rendered inline contents", zap.String("rendered", renderedVal))
		plan.InterpolatedInlineContents = renderedVal
	}

	if step.Step.Command != "" {
		l.Debug("rendering command")
		renderedVal, err := render.RenderV2(step.Step.Command, stateMap)
		if err != nil {
			l.Error("error rendering command",
				zap.String("command", step.Step.Command),
				zap.Any("state", stateMap),
				zap.Error(err),
			)
			return nil, err
		}

		l.Debug("successfully rendered command", zap.String("rendered", renderedVal))
		plan.InterpolatedCommand = renderedVal
	}

	return plan, nil
}

package plan

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

func (p *Planner) createAdhocStepPlan(ctx workflow.Context,
	step *app.InstallActionWorkflowRunStep,
	stateMap map[string]any,
	installID string,
) (*plantypes.ActionWorkflowRunStepPlan, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, err
	}

	plan := &plantypes.ActionWorkflowRunStepPlan{
		ID: step.ID,
		Attrs: map[string]string{
			"step.name": "adhoc",
			"step.id":   step.Step.ID,
		},
		InterpolatedEnvVars: make(map[string]string, 0),
		GitSource:           &plantypes.GitSource{},
	}

	adhocCfg := step.AdHocConfig
	for k, v := range adhocCfg.EnvVars {
		renderedVal, err := render.Render(*v, stateMap)
		if err != nil {
			l.Error("error rendering env-var",
				zap.String("env-var", *v),
				zap.Error(err))
			return nil, err
		}

		plan.InterpolatedEnvVars[k] = renderedVal
	}

	if adhocCfg.InlineContents != "" {
		l.Debug("rendering inline contents")
		renderedVal, err := render.Render(adhocCfg.InlineContents, stateMap)
		if err != nil {
			l.Error("error rendering inline contents",
				zap.String("input", adhocCfg.InlineContents),
				zap.Any("state", stateMap),
				zap.Error(err),
			)
			return nil, err
		}

		l.Debug("successfully rendered inline contents", zap.String("rendered", renderedVal))
		plan.InterpolatedInlineContents = renderedVal
	}

	if adhocCfg.Command != "" {
		l.Debug("rendering command")
		renderedVal, err := render.Render(adhocCfg.Command, stateMap)
		if err != nil {
			l.Error("error rendering command",
				zap.String("command", adhocCfg.Command),
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

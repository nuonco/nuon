package plan

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
	"github.com/powertoolsdev/mono/pkg/render"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

func (p *planner) createStepPlan(ctx workflow.Context, step app.InstallActionWorkflowRunStep, installID string) (*plantypes.ActionWorkflowRunStepPlan, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, err
	}

	plan := &plantypes.ActionWorkflowRunStepPlan{
		ID: step.ID,
		Attrs: map[string]string{
			"step.name": step.Step.Name,
			"step.id":   step.Step.ID,
		},
		InterpolatedEnvVars: make(map[string]string, 0),
	}

	// step 1 - fetch token for repo
	l.Debug("creating git source for config")
	gitSource, err := activities.AwaitGetGitSourceByStepID(ctx, step.Step.ID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get git source")
	}

	plan.GitSource = gitSource

	// step 2 - interpolate all variables in the set
	l.Debug("fetching install state")
	state, err := activities.AwaitGetInstallStateByInstallID(ctx, installID)
	if err != nil {
		l.Error("unable to get install state", zap.Error(err))
		return nil, errors.Wrap(err, "unable to get install state")
	}

	stateMap, err := state.AsMap()
	if err != nil {
		return nil, errors.Wrap(err, "unable to convert state to map")
	}

	for k, v := range step.Step.EnvVars {
		renderedVal, err := render.Render(*v, stateMap)
		if err != nil {
			l.Error(fmt.Sprintf("error rendering %s from intermediate data", *v),
				zap.Any("intermediate-data", stateMap))
			return nil, err
		}

		plan.InterpolatedEnvVars[k] = renderedVal
	}

	if step.Step.InlineContents != "" {
		l.Debug("rendering inline contents")
		renderedVal, err := render.Render(step.Step.InlineContents, stateMap)
		if err != nil {
			return nil, err
		}

		l.Debug("successfully rendered inline contents", zap.String("rendered", renderedVal))
		step.Step.InlineContents = renderedVal
	}
	if step.Step.Command != "" {
		l.Debug("rendering command")
		renderedVal, err := render.Render(step.Step.Command, stateMap)
		if err != nil {
			return nil, err
		}

		l.Debug("successfully rendered command", zap.String("rendered", renderedVal))
		step.Step.Command = renderedVal
	}

	return plan, nil
}

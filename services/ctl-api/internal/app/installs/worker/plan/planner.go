package plan

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

type planner struct{}

func (p *planner) createPlan(ctx workflow.Context, runID string) (*plantypes.ActionWorkflowRunPlan, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, err
	}

	l.Info("creating plan for executing action workflow")
	run, err := activities.AwaitGetInstallActionWorkflowRunByRunID(ctx, runID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get run")
	}

	envVars, err := p.getEnvVars(ctx, run)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get env vars")
	}

	plan := &plantypes.ActionWorkflowRunPlan{
		InstallID: run.InstallID,
		ID:        runID,
		Attrs: map[string]string{
			"action.name": run.ActionWorkflowConfig.ActionWorkflow.Name,
			"action.id":   run.ActionWorkflowConfig.ActionWorkflow.ID,
		},
		Steps:   make([]*plantypes.ActionWorkflowRunStepPlan, 0),
		EnvVars: envVars,
	}

	for idx, stepCfg := range run.Steps {
		l.Debug(fmt.Sprintf("creating plan for step %d", idx))
		stepPlan, err := p.createStepPlan(ctx, stepCfg, run.InstallID)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("unable to create plan for step %d", idx))
		}

		plan.Steps = append(plan.Steps, stepPlan)
	}

	l.Info("successfully created plan")
	return plan, nil
}

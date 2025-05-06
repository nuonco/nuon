package plan

import (
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	awscredentials "github.com/powertoolsdev/mono/pkg/aws/credentials"
	"github.com/powertoolsdev/mono/pkg/generics"
	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

func (p *Planner) createActionWorkflowRunPlan(ctx workflow.Context, runID string) (*plantypes.ActionWorkflowRunPlan, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, err
	}

	l.Info("creating plan for executing action workflow")
	run, err := activities.AwaitGetInstallActionWorkflowRunByRunID(ctx, runID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get run")
	}

	// step 2 - interpolate all variables in the set
	l.Debug("fetching install state")
	state, err := activities.AwaitGetInstallStateByInstallID(ctx, run.InstallID)
	if err != nil {
		l.Error("unable to get install state", zap.Error(err))
		return nil, errors.Wrap(err, "unable to get install state")
	}

	stateMap, err := state.AsMap()
	if err != nil {
		return nil, errors.Wrap(err, "unable to convert state to map")
	}

	stack, err := activities.AwaitGetInstallStackByInstallID(ctx, run.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install stack")
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
	if stack.InstallStackOutputs.AWSStackOutputs != nil {
		plan.AWSAuth = &awscredentials.Config{
			Region: stack.InstallStackOutputs.AWSStackOutputs.Region,
			AssumeRole: &awscredentials.AssumeRoleConfig{
				SessionName: fmt.Sprintf("install-action-workflow-%s", run.ID),
				RoleARN:     stack.InstallStackOutputs.AWSStackOutputs.MaintenanceIAMRoleARN,
			},
		}
	}

	if !generics.SliceContains(run.TriggerType, []app.ActionWorkflowTriggerType{
		app.ActionWorkflowTriggerTypePreSandboxRun,
	}) {
		clusterInfo, err := p.getKubeClusterInfo(ctx, stack, state)
		if err != nil {
			return plan, errors.Wrap(err, "unable to get cluster information")
		}

		plan.ClusterInfo = clusterInfo
	}

	for idx, stepCfg := range run.Steps {
		l.Debug(fmt.Sprintf("creating plan for step %d", idx))
		stepPlan, err := p.createStepPlan(ctx, &stepCfg, stateMap, run.InstallID)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("unable to create plan for step %d", idx))
		}

		plan.Steps = append(plan.Steps, stepPlan)
	}

	l.Info("successfully created plan")
	return plan, nil
}

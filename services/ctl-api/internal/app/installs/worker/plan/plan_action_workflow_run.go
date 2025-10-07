package plan

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	awscredentials "github.com/powertoolsdev/mono/pkg/aws/credentials"
	"github.com/powertoolsdev/mono/pkg/config/refs"
	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/helpers"
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

	org, err := activities.AwaitGetOrgByInstallID(ctx, run.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install id")
	}

	install, err := activities.AwaitGetByInstallID(ctx, run.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	appCfg, err := activities.AwaitGetAppConfigByID(ctx, install.AppConfigID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get app config")
	}

	// step 2 - interpolate all variables in the set
	l.Debug("fetching install state")
	state, err := activities.AwaitGetInstallStateByInstallID(ctx, run.InstallID)
	if err != nil {
		l.Error("unable to get install state", zap.Error(err))
		return nil, errors.Wrap(err, "unable to get install state")
	}

	stateMap, err := state.WorkflowSafeAsMap(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to convert state to map")
	}

	stack, err := activities.AwaitGetInstallStackByInstallID(ctx, run.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install stack")
	}

	builtInEnvVars, err := p.getBuiltinEnvVars(ctx, run)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get env vars")
	}

	overrideEnvVars, err := p.getOverrideEnvVars(ctx, run)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get override env vars")
	}

	plan := &plantypes.ActionWorkflowRunPlan{
		InstallID: run.InstallID,
		ID:        runID,
		Attrs: map[string]string{
			"action.name": run.ActionWorkflowConfig.ActionWorkflow.Name,
			"action.id":   run.ActionWorkflowConfig.ActionWorkflow.ID,
		},
		Steps:           make([]*plantypes.ActionWorkflowRunStepPlan, 0),
		BuiltinEnvVars:  builtInEnvVars,
		OverrideEnvVars: overrideEnvVars,
	}

	if !org.SandboxMode && stack.InstallStackOutputs.AWSStackOutputs != nil {
		role := stack.InstallStackOutputs.AWSStackOutputs.MaintenanceIAMRoleARN

		if !run.ActionWorkflowConfig.BreakGlassRoleARN.Empty() {
			roleArn, ok := stack.InstallStackOutputs.AWSStackOutputs.BreakGlassRoleARNs[run.ActionWorkflowConfig.BreakGlassRoleARN.ValueString()]
			if !ok {
				l.Error(fmt.Sprintf(
					"break glass role %s not provisioned in install stack",
					run.ActionWorkflowConfig.BreakGlassRoleARN.ValueString(),
				))
				return nil, fmt.Errorf("break glass role not provisioned in install stack")
			}
			role = roleArn
		}

		plan.AWSAuth = &awscredentials.Config{
			Region: stack.InstallStackOutputs.AWSStackOutputs.Region,
			AssumeRole: &awscredentials.AssumeRoleConfig{
				SessionName: fmt.Sprintf("install-action-workflow-%s", run.ID),
				RoleARN:     role,
			},
		}
	}

	clusterInfo, err := p.getKubeClusterInfo(ctx, stack, state)
	if err == nil {
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

	if org.SandboxMode {
		targetRefs := helpers.GetActionReferences(appCfg, run.ActionWorkflowConfig.ActionWorkflow.Name)

		plan.SandboxMode = &plantypes.SandboxMode{
			Enabled: true,
			Outputs: refs.GetFakeRefs(targetRefs),
		}
	}

	l.Info("successfully created plan")
	return plan, nil
}

// TODO(ja): make this a method on the run struct?
func hstoreToMap(hstore pgtype.Hstore) map[string]string {
	result := make(map[string]string)
	for key, value := range hstore {
		result[key] = *value
	}
	return result
}

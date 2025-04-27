package plan

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	awscredentials "github.com/powertoolsdev/mono/pkg/aws/credentials"
	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
	"github.com/powertoolsdev/mono/pkg/render"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
)

func (p *Planner) createTerraformDeployPlan(ctx workflow.Context, req *CreateDeployPlanRequest) (*plantypes.TerraformDeployPlan, error) {
	install, err := activities.AwaitGetByInstallID(ctx, req.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	stack, err := activities.AwaitGetInstallStackByInstallID(ctx, req.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install stack")
	}

	installDeploy, err := activities.AwaitGetDeployByDeployID(ctx, req.InstallDeployID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install deploy")
	}

	installComp, err := activities.AwaitGetInstallComponentByID(ctx, installDeploy.InstallComponentID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install component")
	}

	state, err := activities.AwaitGetInstallState(ctx, &activities.GetInstallStateRequest{
		InstallID: install.ID,
	})
	stateData, err := state.AsMap()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get state")
	}

	compBuild, err := activities.AwaitGetComponentBuildByComponentBuildID(ctx, installDeploy.ComponentBuildID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get component build")
	}

	cfg := compBuild.ComponentConfigConnection.TerraformModuleComponentConfig

	if err := render.RenderStruct(cfg, stateData); err != nil {
		return nil, errors.Wrap(err, "unable to render config")
	}

	envVars := generics.ToStringMap(cfg.EnvVars)
	vars := generics.ToStringMapAny(cfg.Variables)

	if err := render.RenderMap(&envVars, stateData); err != nil {
		return nil, errors.Wrap(err, "unable to render environment variables")
	}

	if err := render.RenderMap(&vars, stateData); err != nil {
		return nil, errors.Wrap(err, "unable to render environment variables")
	}

	roleARN := stack.InstallStackOutputs.AWSStackOutputs.MaintenanceIAMRoleARN
	return &plantypes.TerraformDeployPlan{
		Vars:      vars,
		EnvVars:   envVars,
		VarsFiles: cfg.VariablesFiles,
		State:     state,

		TerraformBackend: &plantypes.TerraformBackend{
			WorkspaceID: installComp.TerraformWorkspace.ID,
		},
		AzureAuth: nil,
		AWSAuth: &awscredentials.Config{
			Region: stack.InstallStackOutputs.AWSStackOutputs.Region,
			AssumeRole: &awscredentials.AssumeRoleConfig{
				SessionName: fmt.Sprintf("install-deploy-%s", req.InstallDeployID),
				RoleARN:     roleARN,
			},
		},
		Hooks: &plantypes.TerraformDeployHooks{
			Enabled: false,
		},
	}, nil
}

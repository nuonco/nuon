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
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

func (p *Planner) createTerraformDeployPlan(ctx workflow.Context, req *CreateDeployPlanRequest) (*plantypes.TerraformDeployPlan, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, err
	}

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

	cfg := installDeploy.ComponentBuild.ComponentConfigConnection.TerraformModuleComponentConfig
	envVars := generics.ToStringMap(cfg.EnvVars)
	vars := generics.ToStringMapAny(cfg.Variables)

	l.Info("rendering environment variables")
	if err := render.RenderMap(&envVars, stateData); err != nil {
		return nil, errors.Wrap(err, "unable to render environment variables")
	}

	l.Info("rendering terraform variables")
	if err := render.RenderMap(&vars, stateData); err != nil {
		return nil, errors.Wrap(err, "unable to render environment variables")
	}

	roleARN := stack.InstallStackOutputs.AWSStackOutputs.MaintenanceIAMRoleARN
	return &plantypes.TerraformDeployPlan{

		Vars:    vars,
		EnvVars: envVars,
		State:   state,

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

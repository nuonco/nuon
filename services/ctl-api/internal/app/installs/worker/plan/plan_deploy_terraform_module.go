package plan

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	awscredentials "github.com/powertoolsdev/mono/pkg/aws/credentials"
	"github.com/powertoolsdev/mono/pkg/kube"
	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
	"github.com/powertoolsdev/mono/pkg/render"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

func (p *Planner) createTerraformDeployPlan(ctx workflow.Context, req *CreateDeployPlanRequest) (*plantypes.TerraformDeployPlan, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get logger")
	}

	install, err := activities.AwaitGetByInstallID(ctx, req.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	org, err := activities.AwaitGetOrgByInstallID(ctx, req.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install id")
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
		l.Error("error rendering terraform config",
			zap.Error(err),
			zap.Any("state", stateData),
		)
		return nil, errors.Wrap(err, "unable to render config")
	}

	envVars := generics.ToStringMap(cfg.EnvVars)
	vars := generics.ToStringMapAny(cfg.Variables)

	if err := render.RenderMap(&envVars, stateData); err != nil {
		l.Error("error rendering env-vars",
			zap.Any("env-vars", envVars),
			zap.Error(err),
			zap.Any("state", stateData),
		)
		return nil, errors.Wrap(err, "unable to render environment variables")
	}

	if err := render.RenderMap(&vars, stateData); err != nil {
		l.Error("error rendering vars",
			zap.Any("vars", vars),
			zap.Error(err),
			zap.Any("state", stateData),
		)
		return nil, errors.Wrap(err, "unable to render environment variables")
	}

	roleARN := stack.InstallStackOutputs.AWSStackOutputs.MaintenanceIAMRoleARN

	var clusterInfo *kube.ClusterInfo
	if !org.SandboxMode {
		clusterInfo, err = p.getKubeClusterInfo(ctx, stack, state)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get cluster information")
		}
	}

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
		ClusterInfo: clusterInfo,
		Hooks: &plantypes.TerraformDeployHooks{
			Enabled: false,
		},
	}, nil
}

package plan

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/config/refs"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	operationroles "github.com/nuonco/nuon/services/ctl-api/internal/pkg/operation-roles"
)

func (p *Planner) createDeployPlan(ctx workflow.Context, req *CreateDeployPlanRequest) (*plantypes.DeployPlan, *operationroles.RoleSelection, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, nil, err
	}

	deploy, err := activities.AwaitGetDeployByDeployID(ctx, req.InstallDeployID)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to get install deploy")
	}

	build, err := activities.AwaitGetComponentBuildByComponentBuildID(ctx, deploy.ComponentBuildID)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to get component build")
	}

	install, err := activities.AwaitGetByInstallID(ctx, req.InstallID)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to get install")
	}

	installDeploy, err := activities.AwaitGetDeployByDeployID(ctx, req.InstallDeployID)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to get install deploy")
	}

	appCfg, err := activities.AwaitGetAppConfigByID(ctx, install.AppConfigID)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to get app config")
	}

	stack, err := activities.AwaitGetInstallStackByInstallID(ctx, req.InstallID)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to get install stack")
	}

	installState, err := activities.AwaitGetInstallState(ctx, &activities.GetInstallStateRequest{
		InstallID: install.ID,
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to get install state")
	}

	// this is here just to reduce duplicated return values from all component types, we're caling same function within
	// every component to build cloud auth, its non expensive and deterministic so we can be sure that in both calls
	// we get same information.
	// unless, we change response models like appcfg, installdeploy, build, stack etc somewhere in middle, which should
	// not happen ideally.
	roleSelection, _, err := p.getRoleForDeploy(l, appCfg, installDeploy, build, stack, installState)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to get role for deploy")
	}

	ociConfig, err := p.getInstallRegistryRepositoryConfig(ctx, installDeploy, build, appCfg, stack, installState)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to get install registry repository config")
	}

	plan := &plantypes.DeployPlan{
		Src:    ociConfig,
		SrcTag: deploy.ID,

		AppID:         install.AppID,
		AppConfigID:   appCfg.ID,
		InstallID:     install.ID,
		ComponentName: installDeploy.ComponentName,
		ComponentID:   installDeploy.ComponentID,
	}

	switch build.ComponentConfigConnection.Type {
	case app.ComponentTypeDockerBuild, app.ComponentTypeExternalImage:
		l.Info("generating noop plan")
		plan.NoopDeployPlan = p.createNoopDeployPlan()
	case app.ComponentTypeTerraformModule:
		l.Info("generating terraform plan")
		tfPlan, err := p.createTerraformDeployPlan(ctx, req, appCfg, stack, installState, installDeploy)
		if err != nil {
			l.Info("error generating terraform plan", zap.Error(err))
			return nil, nil, errors.Wrap(err, "unable to create terraform deploy plan")
		}
		plan.TerraformDeployPlan = tfPlan
	case app.ComponentTypeHelmChart:
		l.Info("generating helm plan")
		helmPlan, err := p.createHelmDeployPlan(ctx, req, appCfg, stack, installState, installDeploy)
		if err != nil {
			l.Error("error generating helm plan", zap.Error(err))
			return nil, nil, errors.Wrap(err, "unable to helm deploy plan")
		}
		plan.HelmDeployPlan = helmPlan
	case app.ComponentTypeKubernetesManifest:
		l.Info("generating kubernetes manifest plan")
		kubernetesManifestPlan, err := p.createKubernetesManifestDeployPlan(ctx, req, appCfg, stack, installState, installDeploy)
		if err != nil {
			l.Error("error generating kubernetes manifest plan", zap.Error(err))
			return nil, nil, errors.Wrap(err, "unable to kubernets manifest deploy plan")
		}
		plan.KubernetesManifestDeployPlan = kubernetesManifestPlan
	case app.ComponentTypePulumi:
		l.Info("generating pulumi plan")
		pulumiPlan, err := p.createPulumiDeployPlan(ctx, req, appCfg, stack, installState, installDeploy)
		if err != nil {
			l.Error("error generating pulumi plan", zap.Error(err))
			return nil, nil, errors.Wrap(err, "unable to create pulumi deploy plan")
		}
		plan.PulumiDeployPlan = pulumiPlan
	}

	if install.SandboxMode.Bool {
		targetRefs := helpers.GetComponentReferences(appCfg, installDeploy.ComponentName)

		plan.SandboxMode = &plantypes.SandboxMode{
			Enabled: true,
			Outputs: refs.GetFakeRefs(targetRefs),
		}

		switch build.ComponentConfigConnection.Type {
		case app.ComponentTypeHelmChart:
			plan.SandboxMode.Helm = p.createHelmDeploySandboxMode(ctx, plan.HelmDeployPlan)
		case app.ComponentTypeKubernetesManifest:
			sandboxPlan, err := p.createKubernetesManifestDeployPlanSandboxMode(plan.KubernetesManifestDeployPlan)
			if err != nil {
				return nil, nil, errors.Wrap(err, "unable to create sandbox plan")
			}

			plan.SandboxMode.KubernetesManifest = sandboxPlan
		case app.ComponentTypeTerraformModule:
			sandboxPlan, err := p.createTerraformDeploySandboxMode(ctx, plan.TerraformDeployPlan)
			if err != nil {
				return nil, nil, errors.Wrap(err, "unable to create sandbox plan")
			}

			plan.SandboxMode.Terraform = sandboxPlan
		case app.ComponentTypePulumi:
			plan.SandboxMode.Pulumi = p.createPulumiDeploySandboxMode()
		}
	}

	return plan, roleSelection, nil
}

func (p *Planner) getRoleForDeploy(
	l *zap.Logger,
	appCfg *app.AppConfig,
	installDeploy *app.InstallDeploy,
	compBuild *app.ComponentBuild,
	stack *app.InstallStack,
	installState *state.State,
) (*operationroles.RoleSelection, app.OperationType, error) {
	return operationroles.GetRoleForDeploy(l, appCfg, installDeploy, &compBuild.ComponentConfigConnection, stack, installState)
}

func (p *Planner) getAuthForDeploy(
	ctx workflow.Context,
	installDeploy *app.InstallDeploy,
	compBuild *app.ComponentBuild,
	appCfg *app.AppConfig,
	stack *app.InstallStack,
	installState *state.State,
	sessionName string,
) (*CloudAuth, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, err
	}

	roleSelection, operation, err := p.getRoleForDeploy(
		l,
		appCfg,
		installDeploy,
		compBuild,
		stack,
		installState,
	)
	if err != nil {
		return nil, err
	}

	l.Info("selected role for component deploy plan",
		zap.String("role_name", roleSelection.RoleName),
		zap.String("role_arn", roleSelection.RoleARN),
		zap.String("source", string(roleSelection.Source)),
		zap.String("operation", string(operation)),
		zap.String("deploy_type", string(installDeploy.Type)),
	)

	return getCloudAuth(roleSelection, &stack.InstallStackOutputs, sessionName)
}

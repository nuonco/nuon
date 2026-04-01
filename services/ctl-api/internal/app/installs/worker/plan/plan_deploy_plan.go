package plan

import (
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/config/refs"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/principal"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	operationroles "github.com/nuonco/nuon/services/ctl-api/internal/pkg/operation-roles"
)

func (p *Planner) createDeployPlan(ctx workflow.Context, req *CreateDeployPlanRequest) (*plantypes.DeployPlan, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, err
	}

	deploy, err := activities.AwaitGetDeployByDeployID(ctx, req.InstallDeployID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install deploy")
	}

	build, err := activities.AwaitGetComponentBuildByComponentBuildID(ctx, deploy.ComponentBuildID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get component build")
	}

	install, err := activities.AwaitGetByInstallID(ctx, req.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	installDeploy, err := activities.AwaitGetDeployByDeployID(ctx, req.InstallDeployID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install deploy")
	}

	appCfg, err := activities.AwaitGetAppConfigByID(ctx, install.AppConfigID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get app config")
	}

	stack, err := activities.AwaitGetInstallStackByInstallID(ctx, req.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install stack")
	}

	installState, err := activities.AwaitGetInstallState(ctx, &activities.GetInstallStateRequest{
		InstallID: install.ID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install state")
	}

	ociConfig, err := p.getInstallRegistryRepositoryConfig(ctx, installDeploy, build, appCfg, stack, installState)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install registry repository config")
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
			return nil, errors.Wrap(err, "unable to create terraform deploy plan")
		}
		plan.TerraformDeployPlan = tfPlan
	case app.ComponentTypeHelmChart:
		l.Info("generating helm plan")
		helmPlan, err := p.createHelmDeployPlan(ctx, req, appCfg, stack, installState, installDeploy)
		if err != nil {
			l.Error("error generating helm plan", zap.Error(err))
			return nil, errors.Wrap(err, "unable to helm deploy plan")
		}
		plan.HelmDeployPlan = helmPlan
	case app.ComponentTypeKubernetesManifest:
		l.Info("generating kubernetes manifest plan")
		kubernetesManifestPlan, err := p.createKubernetesManifestDeployPlan(ctx, req, appCfg, stack, installState, installDeploy)
		if err != nil {
			l.Error("error generating kubernetes manifest plan", zap.Error(err))
			return nil, errors.Wrap(err, "unable to kubernets manifest deploy plan")
		}
		plan.KubernetesManifestDeployPlan = kubernetesManifestPlan
	}

	// the following section is for sandbox mode only
	org, err := activities.AwaitGetOrgByInstallID(ctx, deploy.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get org")
	}
	if org.SandboxMode {
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
				return nil, errors.Wrap(err, "unable to create sandbox plan")
			}

			plan.SandboxMode.KubernetesManifest = sandboxPlan
		case app.ComponentTypeTerraformModule:
			sandboxPlan, err := p.createTerraformDeploySandboxMode(ctx, plan.TerraformDeployPlan)
			if err != nil {
				return nil, errors.Wrap(err, "unable to create sandbox plan")
			}

			plan.SandboxMode.Terraform = sandboxPlan
		}
	}

	return plan, nil
}

func (p *Planner) getRoleForDeploy(
	l *zap.Logger,
	appCfg *app.AppConfig,
	installDeploy *app.InstallDeploy,
	compBuild *app.ComponentBuild,
	stack *app.InstallStack,
	installState *state.State,
) (*operationroles.RoleSelection, app.OperationType, error) {
	operation := app.OperationDeploy
	if installDeploy.Type == app.InstallDeployTypeTeardown {
		operation = app.OperationTeardown
	}

	selectionCtx := &operationroles.SelectionContext{
		Operation:     operation,
		PrincipalType: principal.TypeComponent,
		PrincipalName: compBuild.ComponentConfigConnection.Component.Name,
		RuntimeRole:   installDeploy.Role,
		EntityRoles: operationroles.EntityOperationRoleMapFromHstore(
			compBuild.ComponentConfigConnection.OperationRoles,
		),
		MatrixRules:  appCfg.OperationRoleConfig.Rules,
		DefaultRole:  appCfg.PermissionsConfig.MaintenanceRole.Name,
		AppConfig:    appCfg,
		StackOutputs: &stack.InstallStackOutputs,
		InstallState: installState,
	}

	roleSelection, err := operationroles.SelectRole(selectionCtx, l)
	if err != nil {
		l.Warn("dynamic role selection failed, falling back to default role",
			zap.Error(err),
			zap.String("default_role", selectionCtx.DefaultRole),
		)

		var fallbackErr error
		roleSelection, fallbackErr = operationroles.GetDefaultRoleSelection(selectionCtx)
		if fallbackErr != nil {
			return nil, "", fmt.Errorf("unable to get default role: %w", fallbackErr)
		}

		l.Warn("using default role for component deploy",
			zap.String("role_name", roleSelection.RoleName),
			zap.String("role_arn", roleSelection.RoleARN),
		)
	}

	return roleSelection, operation, nil
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

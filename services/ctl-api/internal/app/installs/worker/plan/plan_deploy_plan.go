package plan

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/config/refs"
	"github.com/nuonco/nuon/pkg/generics"
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
	roleSelection, _, err := p.getRoleForDeploy(ctx, l, appCfg, installDeploy, build, stack, installState)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to get role for deploy")
	}

	ociConfig, err := p.getInstallRegistryRepositoryConfig(ctx, installDeploy, build, appCfg, stack, installState, roleSelection)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to get install registry repository config")
	}

	// Address the install-registry artifact by its content (manifest
	// digest) rather than the synthetic install-deploy ID. The sync plan
	// copies image-type builds into the install registry under their
	// ResolvedTag; using the digest as SrcTag here is correct because
	// oras.Copy resolves both tags and digests, and the digest is the
	// immutable identity of the artifact.
	//
	// For non-image builds and image builds without SourceDigest, the sync
	// plan still tags the install-registry copy with the install-deploy ID,
	// so we fall back to that here.
	srcTag := deploy.ID
	srcDigest := ""
	if build.SourceDigest != "" {
		srcTag = build.SourceDigest
		srcDigest = build.SourceDigest
	}

	plan := &plantypes.DeployPlan{
		Src:       ociConfig,
		SrcTag:    srcTag,
		SrcDigest: srcDigest,

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
		tfPlan, err := p.createTerraformDeployPlan(ctx, req, appCfg, stack, installState, installDeploy, roleSelection)
		if err != nil {
			l.Info("error generating terraform plan", zap.Error(err))
			return nil, nil, errors.Wrap(err, "unable to create terraform deploy plan")
		}
		plan.TerraformDeployPlan = tfPlan
	case app.ComponentTypeHelmChart:
		l.Info("generating helm plan")
		helmPlan, err := p.createHelmDeployPlan(ctx, req, appCfg, stack, installState, installDeploy, roleSelection)
		if err != nil {
			l.Error("error generating helm plan", zap.Error(err))
			return nil, nil, errors.Wrap(err, "unable to helm deploy plan")
		}
		plan.HelmDeployPlan = helmPlan
	case app.ComponentTypeKubernetesManifest:
		l.Info("generating kubernetes manifest plan")
		kubernetesManifestPlan, err := p.createKubernetesManifestDeployPlan(ctx, req, appCfg, stack, installState, installDeploy, roleSelection)
		if err != nil {
			l.Error("error generating kubernetes manifest plan", zap.Error(err))
			return nil, nil, errors.Wrap(err, "unable to kubernets manifest deploy plan")
		}
		plan.KubernetesManifestDeployPlan = kubernetesManifestPlan
	case app.ComponentTypePulumi:
		l.Info("generating pulumi plan")
		pulumiPlan, err := p.createPulumiDeployPlan(ctx, req, appCfg, stack, installState, installDeploy, roleSelection)
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
	ctx workflow.Context,
	l *zap.Logger,
	appCfg *app.AppConfig,
	installDeploy *app.InstallDeploy,
	compBuild *app.ComponentBuild,
	stack *app.InstallStack,
	installState *state.State,
) (*operationroles.RoleSelection, app.OperationType, error) {
	flw := p.installWorkflowForRoleDefault(ctx, l, generics.FromPtrStr(installDeploy.InstallWorkflowID))
	return operationroles.GetRoleForDeploy(l, appCfg, installDeploy, &compBuild.ComponentConfigConnection, stack, installState, flw)
}

// installWorkflowForRoleDefault returns the parent install workflow used to
// derive a step's lowest-precedence default role, or nil when workflow-type
// defaulting should not apply. It returns nil when the global config flag is
// off (WORKFLOW_DEFAULT_ROLE_ENABLED) or the workflow can't be resolved, so the
// operation-roles package falls back to the maintenance role and planning never
// fails on role defaulting.
func (p *Planner) installWorkflowForRoleDefault(
	ctx workflow.Context,
	l *zap.Logger,
	installWorkflowID string,
) *app.Workflow {
	enabled, err := activities.AwaitWorkflowDefaultRoleEnabled(ctx, activities.WorkflowDefaultRoleEnabledRequest{})
	if err != nil {
		l.Warn("unable to check workflow-default-role config; using maintenance role",
			zap.Error(err),
		)
		return nil
	}
	if !enabled {
		return nil
	}

	if installWorkflowID == "" {
		l.Warn("install workflow ID missing for role default; using maintenance role")
		return nil
	}

	flw, err := activities.AwaitGetInstallWorkflowByID(ctx, installWorkflowID)
	if err != nil {
		l.Warn("unable to get install workflow for role default; using maintenance role",
			zap.String("install_workflow_id", installWorkflowID),
			zap.Error(err),
		)
		return nil
	}

	return flw
}

// getAuthForDeploy builds cloud auth from an already-resolved role selection.
// The role is selected once per deploy/sync plan and threaded through so the
// auth embedded in the runner plan can never diverge from the recorded role
// selection, and the role-default activities only run once per plan.
func (p *Planner) getAuthForDeploy(
	ctx workflow.Context,
	roleSelection *operationroles.RoleSelection,
	stack *app.InstallStack,
	sessionName string,
) (*CloudAuth, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, err
	}

	l.Info("using selected role for component deploy auth",
		zap.String("role_name", roleSelection.RoleName),
		zap.String("role_arn", roleSelection.RoleARN),
		zap.String("source", string(roleSelection.Source)),
		zap.String("session_name", sessionName),
	)

	return getCloudAuth(roleSelection, &stack.InstallStackOutputs, sessionName)
}

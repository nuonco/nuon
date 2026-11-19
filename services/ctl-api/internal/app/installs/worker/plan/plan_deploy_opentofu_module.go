package plan

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/kube"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/deployerrors"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	operationroles "github.com/nuonco/nuon/services/ctl-api/internal/pkg/operation-roles"
)

func (p *Planner) createOpentofuDeployPlan(
	ctx workflow.Context,
	req *CreateDeployPlanRequest,
	appCfg *app.AppConfig,
	stack *app.InstallStack,
	state *state.State,
	installDeploy *app.InstallDeploy,
	roleSelection *operationroles.RoleSelection,
) (*plantypes.OpentofuDeployPlan, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get logger")
	}

	installComp, err := activities.AwaitGetInstallComponentByID(
		ctx,
		installDeploy.InstallComponentID,
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install component")
	}

	stateData, err := state.WorkflowSafeAsMap(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get state")
	}

	compBuild, err := activities.AwaitGetComponentBuildByComponentBuildID(
		ctx,
		installDeploy.ComponentBuildID,
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get component build")
	}

	return p.RenderOpentofuDeployPlan(l, &RenderOpentofuDeployPlanInput{
		Stack:         stack,
		State:         state,
		StateData:     stateData,
		InstallDeploy: installDeploy,
		CompBuild:     compBuild,
		WorkspaceID:   installComp.TerraformWorkspace.ID,
		RoleSelection: roleSelection,
		ResolveClusterInfo: func(cloudAuth *CloudAuth) (*kube.ClusterInfo, error) {
			return p.resolveKubernetesContext(ctx, &compBuild.ComponentConfigConnection, appCfg, stack, state, cloudAuth)
		},
	})
}

type RenderOpentofuDeployPlanInput struct {
	Stack         *app.InstallStack
	State         *state.State
	StateData     map[string]any
	InstallDeploy *app.InstallDeploy
	CompBuild     *app.ComponentBuild
	WorkspaceID   string
	RoleSelection *operationroles.RoleSelection

	ResolveClusterInfo func(*CloudAuth) (*kube.ClusterInfo, error)
}

func (p *Planner) RenderOpentofuDeployPlan(
	l *zap.Logger,
	in *RenderOpentofuDeployPlanInput,
) (*plantypes.OpentofuDeployPlan, error) {
	cfg := in.CompBuild.ComponentConfigConnection.OpentofuModuleComponentConfig
	if err := render.RenderStruct(cfg, in.StateData); err != nil {
		l.Error("error rendering opentofu config",
			zap.Error(err),
			zap.Any("state", in.StateData),
		)
		return nil, deployerrors.NewDeployPlanRenderFailed(err, "unable to render config")
	}
	vars := generics.ToStringMapAny(cfg.Variables)
	if err := render.RenderMap(&vars, in.StateData); err != nil {
		l.Error("error rendering vars",
			zap.Any("vars", vars),
			zap.Error(err),
			zap.Any("state", in.StateData),
		)
		return nil, deployerrors.NewDeployPlanRenderFailed(err, "unable to render opentofu variables")
	}

	envVars := generics.ToStringMap(cfg.EnvVars)

	if err := render.RenderMap(&envVars, in.StateData); err != nil {
		l.Error("error rendering env-vars",
			zap.Any("env-vars", envVars),
			zap.Error(err),
			zap.Any("state", in.StateData),
		)
		return nil, deployerrors.NewDeployPlanRenderFailed(err, "unable to render environment variables")
	}

	varsFiles := []string(cfg.VariablesFiles)
	tfVarsOverride, err := p.installComponentOverride(
		in.State, in.StateData,
		config.TFVarsOverrideInputName(in.InstallDeploy.ComponentName),
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to render opentofu vars override")
	}
	if tfVarsOverride != "" {
		varsFiles = append(varsFiles, tfVarsOverride)
	}

	cloudAuth, err := p.AuthForDeploy(
		l,
		in.RoleSelection,
		in.Stack,
		fmt.Sprintf("component-deploy-%s", in.InstallDeploy.ID),
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get auth for deploy")
	}

	clusterInfo, err := in.ResolveClusterInfo(cloudAuth)
	if err != nil {
		l.Warn("unable to resolve kubernetes context, this usually means this was not a kubernetes application")
	}

	return &plantypes.OpentofuDeployPlan{
		Vars:      vars,
		EnvVars:   envVars,
		VarsFiles: varsFiles,
		State:     in.State,

		OpentofuBackend: &plantypes.TerraformBackend{
			WorkspaceID: in.WorkspaceID,
		},
		AzureAuth:   cloudAuth.Azure,
		AWSAuth:     cloudAuth.AWS,
		GCPAuth:     cloudAuth.GCP,
		ClusterInfo: clusterInfo,
		Hooks: &plantypes.TerraformDeployHooks{
			Enabled: false,
		},
	}, nil
}

func (p *Planner) createOpentofuDeploySandboxMode(
	ctx workflow.Context,
	req *plantypes.OpentofuDeployPlan,
) (*plantypes.TerraformSandboxMode, error) {
	return p.createTerraformDeploySandboxMode(ctx, &plantypes.TerraformDeployPlan{
		TerraformBackend: &plantypes.TerraformBackend{
			WorkspaceID: req.OpentofuBackend.WorkspaceID,
		},
	})
}

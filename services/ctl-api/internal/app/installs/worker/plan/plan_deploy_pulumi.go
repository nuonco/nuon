package plan

import (
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/kube"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	operationroles "github.com/nuonco/nuon/services/ctl-api/internal/pkg/operation-roles"
)

func (p *Planner) createPulumiDeployPlan(
	ctx workflow.Context,
	req *CreateDeployPlanRequest,
	appCfg *app.AppConfig,
	stack *app.InstallStack,
	state *state.State,
	installDeploy *app.InstallDeploy,
	roleSelection *operationroles.RoleSelection,
) (*plantypes.PulumiDeployPlan, error) {
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

	return p.RenderPulumiDeployPlan(l, &RenderPulumiDeployPlanInput{
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
		HasUpdatePlansFeature: func() (bool, error) {
			return activities.AwaitHasFeatureByFeature(ctx, string(app.OrgFeaturePulumiUpdatePlans))
		},
	})
}

// RenderPulumiDeployPlanInput carries the already-loaded data a pulumi deploy
// plan is rendered from.
type RenderPulumiDeployPlanInput struct {
	Stack         *app.InstallStack
	State         *state.State
	StateData     map[string]any
	InstallDeploy *app.InstallDeploy
	CompBuild     *app.ComponentBuild
	WorkspaceID   string
	RoleSelection *operationroles.RoleSelection

	// ResolveClusterInfo resolves the kubernetes cluster at this exact point in
	// the plan render.
	ResolveClusterInfo func(*CloudAuth) (*kube.ClusterInfo, error)
	// HasUpdatePlansFeature checks the feature at this exact point in the plan
	// render.
	HasUpdatePlansFeature func() (bool, error)
}

// RenderPulumiDeployPlan is the pure core of createPulumiDeployPlan.
func (p *Planner) RenderPulumiDeployPlan(
	l *zap.Logger,
	in *RenderPulumiDeployPlanInput,
) (*plantypes.PulumiDeployPlan, error) {
	cfg := in.CompBuild.ComponentConfigConnection.PulumiComponentConfig
	if err := render.RenderStruct(cfg, in.StateData); err != nil {
		l.Error("error rendering pulumi config",
			zap.Error(err),
			zap.Any("state", in.StateData),
		)
		return nil, errors.Wrap(err, "unable to render config")
	}

	configMap := generics.ToStringMap(cfg.Config)
	if err := render.RenderMap(&configMap, in.StateData); err != nil {
		l.Error("error rendering pulumi config map",
			zap.Any("config", configMap),
			zap.Error(err),
			zap.Any("state", in.StateData),
		)
		return nil, errors.Wrap(err, "unable to render pulumi config")
	}

	envVars := generics.ToStringMap(cfg.EnvVars)
	if err := render.RenderMap(&envVars, in.StateData); err != nil {
		l.Error("error rendering env-vars",
			zap.Any("env-vars", envVars),
			zap.Error(err),
			zap.Any("state", in.StateData),
		)
		return nil, errors.Wrap(err, "unable to render environment variables")
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

	updatePlans, err := in.HasUpdatePlansFeature()
	if err != nil {
		return nil, errors.Wrap(err, "unable to check pulumi-update-plans feature")
	}

	return &plantypes.PulumiDeployPlan{
		Config:        configMap,
		EnvVars:       envVars,
		Runtime:       cfg.Runtime,
		PulumiVersion: cfg.Version,
		StackName:     fmt.Sprintf("nuon-%s", in.InstallDeploy.InstallID),
		WorkspaceID:   in.WorkspaceID,
		AzureAuth:     cloudAuth.Azure,
		AWSAuth:       cloudAuth.AWS,
		GCPAuth:       cloudAuth.GCP,
		ClusterInfo:   clusterInfo,
		State:         in.State,
		Destroy:       in.InstallDeploy.Type == app.InstallDeployTypeTeardown,
		UpdatePlans:   updatePlans,
	}, nil
}

func (p *Planner) createPulumiDeploySandboxMode() *plantypes.PulumiSandboxMode {
	fakePlan := `{
  "stdout": "Previewing update (sandbox):\n\n    + pulumi:pulumi:Stack sandbox-stack create\n    + cloud:storage:Bucket app-bucket create\n    + cloud:compute:Instance app-server create\n\nResources:\n    + 3 to create",
  "stderr": "",
  "change_summary": {"create": 3},
  "resource_changes": [
    {
      "urn": "urn:pulumi:sandbox::app::cloud:storage:Bucket::app-bucket",
      "type": "cloud:storage:Bucket",
      "name": "app-bucket",
      "action": "create",
      "new_inputs": {"name": "app-bucket-sandbox", "location": "US", "forceDestroy": true}
    },
    {
      "urn": "urn:pulumi:sandbox::app::cloud:compute:Instance::app-server",
      "type": "cloud:compute:Instance",
      "name": "app-server",
      "action": "create",
      "new_inputs": {"machineType": "e2-medium", "zone": "us-central1-a", "bootDisk": {"initializeParams": {"image": "debian-cloud/debian-11"}}}
    },
    {
      "urn": "urn:pulumi:sandbox::app::cloud:dns:RecordSet::app-dns",
      "type": "cloud:dns:RecordSet",
      "name": "app-dns",
      "action": "create",
      "new_inputs": {"name": "app.sandbox.example.com", "type": "A", "ttl": 300}
    }
  ]
}`
	return &plantypes.PulumiSandboxMode{
		PlanContents:        fakePlan,
		PlanDisplayContents: fakePlan,
	}
}

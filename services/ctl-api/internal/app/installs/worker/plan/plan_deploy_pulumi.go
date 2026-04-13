package plan

import (
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

func (p *Planner) createPulumiDeployPlan(
	ctx workflow.Context,
	req *CreateDeployPlanRequest,
	appCfg *app.AppConfig,
	stack *app.InstallStack,
	state *state.State,
	installDeploy *app.InstallDeploy,
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

	cfg := compBuild.ComponentConfigConnection.PulumiComponentConfig
	if err := render.RenderStruct(cfg, stateData); err != nil {
		l.Error("error rendering pulumi config",
			zap.Error(err),
			zap.Any("state", stateData),
		)
		return nil, errors.Wrap(err, "unable to render config")
	}

	configMap := generics.ToStringMap(cfg.Config)
	if err := render.RenderMap(&configMap, stateData); err != nil {
		l.Error("error rendering pulumi config map",
			zap.Any("config", configMap),
			zap.Error(err),
			zap.Any("state", stateData),
		)
		return nil, errors.Wrap(err, "unable to render pulumi config")
	}

	envVars := generics.ToStringMap(cfg.EnvVars)
	if err := render.RenderMap(&envVars, stateData); err != nil {
		l.Error("error rendering env-vars",
			zap.Any("env-vars", envVars),
			zap.Error(err),
			zap.Any("state", stateData),
		)
		return nil, errors.Wrap(err, "unable to render environment variables")
	}

	cloudAuth, err := p.getAuthForDeploy(
		ctx,
		installDeploy,
		compBuild,
		appCfg,
		stack,
		state,
		fmt.Sprintf("component-deploy-%s", installDeploy.ID),
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get auth for deploy")
	}

	clusterInfo, err := p.getKubeClusterInfo(ctx, stack, state, cloudAuth)
	if err != nil {
		l.Warn("unable to get cluster information, this usually means this was not a kubernetes application")
	}

	return &plantypes.PulumiDeployPlan{
		Config:        configMap,
		EnvVars:       envVars,
		Runtime:       cfg.Runtime,
		PulumiVersion: cfg.Version,
		StackName:     fmt.Sprintf("nuon-%s", installDeploy.InstallID),
		WorkspaceID:   installComp.TerraformWorkspace.ID,
		AzureAuth:     cloudAuth.Azure,
		AWSAuth:       cloudAuth.AWS,
		GCPAuth:       cloudAuth.GCP,
		ClusterInfo:   clusterInfo,
		State:         state,
		Destroy:       installDeploy.Type == app.InstallDeployTypeTeardown,
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

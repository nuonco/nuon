package plan

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
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

	ociConfig, err := p.getInstallRegistryRepositoryConfig(ctx, req.InstallID, req.InstallDeployID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install registry repository config")
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

	appCfg, err := activities.AwaitGetAppConfigByID(ctx, build.ComponentConfigConnection.AppConfigID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get app config")
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
		tfPlan, err := p.createTerraformDeployPlan(ctx, req)
		if err != nil {
			return nil, errors.Wrap(err, "unable to create terraform deploy plan")
		}
		plan.TerraformDeployPlan = tfPlan
	case app.ComponentTypeHelmChart:
		l.Info("generating helm plan")
		helmPlan, err := p.createHelmDeployPlan(ctx, req)
		if err != nil {
			return nil, errors.Wrap(err, "unable to helm deploy plan")
		}
		plan.HelmDeployPlan = helmPlan
	}

	return plan, nil
}

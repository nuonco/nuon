package plan

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/config/refs"
	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/helpers"
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

	appCfg, err := activities.AwaitGetAppConfigByID(ctx, install.AppConfigID)
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
			l.Info("error generating terraform plan", zap.Error(err))
			return nil, errors.Wrap(err, "unable to create terraform deploy plan")
		}
		plan.TerraformDeployPlan = tfPlan
	case app.ComponentTypeHelmChart:
		l.Info("generating helm plan")
		helmPlan, err := p.createHelmDeployPlan(ctx, req)
		if err != nil {
			l.Error("error generating helm plan", zap.Error(err))
			return nil, errors.Wrap(err, "unable to helm deploy plan")
		}
		plan.HelmDeployPlan = helmPlan
	case app.ComponentTypeKubernetesManifest:
		l.Info("generating kubernetes manifest plan")
		kubernetesManifestPlan, err := p.createKubernetesManifestDeployPlan(ctx, req)
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

package plan

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

func (p *Planner) createComponentBuildPlan(ctx workflow.Context, req *CreateComponentBuildPlanRequest) (*plantypes.BuildPlan, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, err
	}

	cmp, err := activities.AwaitGetComponentByComponentID(ctx, req.ComponentID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get component")
	}

	build, err := activities.AwaitGetComponentBuildWithConfigByID(ctx, req.ComponentBuildID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get component build")
	}

	gitSrc, err := activities.AwaitGetBuildGitSourceByBuildID(ctx, req.ComponentBuildID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get gitSrc")
	}

	dstCfg, err := activities.AwaitGetComponentOCIRegistryRepositoryByComponentID(ctx, cmp.ID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get destination component config")
	}

	plan := &plantypes.BuildPlan{
		ComponentID:      cmp.ID,
		ComponentBuildID: build.ID,

		Src:    gitSrc,
		Dst:    dstCfg,
		DstTag: build.ID,
	}

	switch build.ComponentConfigConnection.Type {
	case app.ComponentTypeDockerBuild:
		l.Info("generating docker build plan")
		subPlan, err := p.createDockerBuildPlan(ctx, build)
		if err != nil {
			return nil, errors.Wrap(err, "unable to create docker build plan")
		}
		plan.DockerBuildPlan = subPlan

	case app.ComponentTypeExternalImage:
		l.Info("generating container image build plan")
		subPlan, err := p.createContainerImageBuildPlan(ctx, build)
		if err != nil {
			return nil, errors.Wrap(err, "unable to create docker build plan")
		}
		plan.ContainerImagePullPlan = subPlan

	case app.ComponentTypeTerraformModule:
		l.Info("generating terraform build plan")
		tfPlan, err := p.createTerraformBuildPlan(ctx, build)
		if err != nil {
			return nil, errors.Wrap(err, "unable to create terraform deploy plan")
		}
		plan.TerraformBuildPlan = tfPlan

	case app.ComponentTypeHelmChart:
		l.Info("generating helm plan")
		helmPlan, err := p.createHelmBuildPlan(ctx, build)
		if err != nil {
			return nil, errors.Wrap(err, "unable to helm deploy plan")
		}
		plan.HelmBuildPlan = helmPlan
	}

	return plan, nil
}

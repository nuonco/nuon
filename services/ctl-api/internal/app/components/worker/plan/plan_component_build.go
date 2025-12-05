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

		helmCompCfg, err := activities.AwaitGetHelmComponentConfigByComponentConfigConnectionID(ctx, build.ComponentConfigConnectionID)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get helm component config")
		}

		helmPlan, err := p.createHelmBuildPlan(ctx, build, helmCompCfg)
		if err != nil {
			return nil, errors.Wrap(err, "unable to helm deploy plan")
		}
		plan.HelmBuildPlan = helmPlan
	}

	org, err := activities.AwaitGetOrgByID(ctx, build.OrgID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get org")
	}
	if org.SandboxMode {
		plan.SandboxMode = &plantypes.SandboxMode{
			Enabled: true,
			Outputs: map[string]any{
				"image": map[string]interface{}{
					"tag":           "v1.2.3",
					"repository":    "nuon/app-service",
					"media_type":    "application/vnd.docker.distribution.manifest.v2+json",
					"digest":        "sha256:a123b456c789d012e345f678g901h234i567j890k123l456m789n012o345p",
					"size":          28437192,
					"urls":          []string{"registry.example.com/nuon/app-service:v1.2.3"},
					"annotations":   map[string]string{"org.opencontainers.image.created": "2024-04-29T10:15:30Z"},
					"artifact_type": "application/vnd.docker.container.image.v1+json",
					"platform": map[string]any{
						"architecture": "arm64",
						"os":           "linux",
						"os_version":   "10.0",
						"variant":      "v8",
						"os_features":  []string{"sse4", "aes"},
					},
				},
			},
		}
	}

	return plan, nil
}

package build

import (
	"context"
	"fmt"

	buildv1 "github.com/powertoolsdev/protos/components/generated/types/build/v1"
	vcsv1 "github.com/powertoolsdev/protos/components/generated/types/vcs/v1"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
	"github.com/powertoolsdev/workers-executors/internal/planners/waypoint/configs"
)

func (p *planner) getDockerPlan(ctx context.Context, cfg *buildv1.Config_DockerCfg) (*planv1.WaypointPlan, error) {
	plan := p.getBasePlan()

	var (
		gitSource *planv1.GitSource
		err       error
	)
	switch vcsCfg := cfg.DockerCfg.VcsCfg.Cfg.(type) {
	case *vcsv1.Config_PrivateGithubConfig:
		gitSource, err = p.getPrivateGitSource(ctx, vcsCfg.PrivateGithubConfig)
	case *vcsv1.Config_PublicGithubConfig:
		gitSource, err = p.getPublicGitSource(ctx, vcsCfg.PublicGithubConfig)
	default:
		return nil, fmt.Errorf("unsupported vcs config: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("unable to get git source: %w", err)
	}
	plan.GitSource = gitSource

	builder, err := configs.NewDockerBuild(p.V,
		configs.WithComponent(p.Component),
		configs.WithEcrRef(plan.EcrRepositoryRef),
		configs.WithWaypointRef(plan.WaypointRef),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create waypoint config builder: %w", err)
	}

	waypointCfg, cfgFmt, err := builder.Render()
	if err != nil {
		return nil, fmt.Errorf("unable to render waypoint config: %w", err)
	}
	plan.WaypointRef.HclConfig = string(waypointCfg)
	plan.WaypointRef.HclConfigFormat = cfgFmt.String()
	return plan, nil
}

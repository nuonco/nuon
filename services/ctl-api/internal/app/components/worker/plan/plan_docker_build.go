package plan

import (
	"go.temporal.io/sdk/workflow"

	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (p *Planner) createDockerBuildPlan(ctx workflow.Context, bld *app.ComponentBuild) (*plantypes.DockerBuildPlan, error) {
	return &plantypes.DockerBuildPlan{
		BuildArgs:  map[string]*string{},
		Target:     bld.ComponentConfigConnection.DockerBuildComponentConfig.Target,
		Dockerfile: bld.ComponentConfigConnection.DockerBuildComponentConfig.Dockerfile,
	}, nil
}

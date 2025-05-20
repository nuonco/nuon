package plan

import (
	"go.temporal.io/sdk/workflow"

	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (p *Planner) createTerraformBuildPlan(ctx workflow.Context, bld *app.ComponentBuild) (*plantypes.TerraformBuildPlan, error) {
	return &plantypes.TerraformBuildPlan{
		Labels: map[string]string{
			"component_id":       bld.ComponentID,
			"component_build_id": bld.ID,
		},
	}, nil
}

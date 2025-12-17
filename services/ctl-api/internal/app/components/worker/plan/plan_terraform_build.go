package plan

import (
	"go.temporal.io/sdk/workflow"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (p *Planner) createTerraformBuildPlan(ctx workflow.Context, bld *app.ComponentBuild) (*plantypes.TerraformBuildPlan, error) {
	return &plantypes.TerraformBuildPlan{
		Labels: map[string]string{
			"component_id":       bld.ComponentID,
			"component_build_id": bld.ID,
		},
	}, nil
}

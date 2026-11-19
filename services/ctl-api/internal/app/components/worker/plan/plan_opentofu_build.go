package plan

import (
	"go.temporal.io/sdk/workflow"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/buildsource"
)

func (p *Planner) createOpentofuBuildPlan(_ workflow.Context, bld *app.ComponentBuild) (*plantypes.OpentofuBuildPlan, error) {
	inputs, err := buildsource.ComponentInputs(&bld.ComponentConfigConnection)
	if err != nil {
		return nil, err
	}
	plan := inputs.OpentofuBuildPlan
	plan.Labels = map[string]string{
		"component_id":       bld.ComponentID,
		"component_build_id": bld.ID,
	}

	return plan, nil
}

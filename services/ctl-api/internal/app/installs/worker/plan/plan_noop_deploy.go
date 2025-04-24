package plan

import plantypes "github.com/powertoolsdev/mono/pkg/plans/types"

func (p *Planner) createNoopDeployPlan() *plantypes.NoopDeployPlan {
	return &plantypes.NoopDeployPlan{}
}

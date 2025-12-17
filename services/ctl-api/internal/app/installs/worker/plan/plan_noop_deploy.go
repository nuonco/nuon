package plan

import plantypes "github.com/nuonco/nuon/pkg/plans/types"

func (p *Planner) createNoopDeployPlan() *plantypes.NoopDeployPlan {
	return &plantypes.NoopDeployPlan{}
}

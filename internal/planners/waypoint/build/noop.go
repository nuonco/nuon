package build

import (
	"context"

	buildv1 "github.com/powertoolsdev/protos/components/generated/types/build/v1"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
)

func (p *planner) getNoopPlan(ctx context.Context, cfg *buildv1.Config_Noop) (*planv1.WaypointPlan, error) {
	return nil, nil
}

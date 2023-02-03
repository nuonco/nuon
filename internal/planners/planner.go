package planners

import (
	"context"

	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
)

type Planner interface {
	GetPlanRef() *planv1.PlanRef

	// TODO(jm): make this interface return a generic *planv1.Plan type that uses a one of
	GetPlan(context.Context) (*planv1.WaypointPlan, error)
}

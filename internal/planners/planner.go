package planners

import (
	"context"

	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
)

type Planner interface {
	Plan(context.Context) (*planv1.Plan, error)
	Prefix() string
}

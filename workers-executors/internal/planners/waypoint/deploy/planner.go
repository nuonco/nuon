package deploy

import (
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/workers-executors/internal/planners"
	"github.com/powertoolsdev/workers-executors/internal/planners/waypoint"
)

type planner struct {
	*waypoint.Planner
}

var _ planners.Planner = (*planner)(nil)

func New(v *validator.Validate, opts ...waypoint.PlannerOption) (*planner, error) {
	wpPln, err := waypoint.New(v, opts...)
	if err != nil {
		return nil, err
	}

	return &planner{wpPln}, nil
}

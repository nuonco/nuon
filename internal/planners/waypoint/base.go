package waypoint

import (
	"github.com/go-playground/validator/v10"
	componentv1 "github.com/powertoolsdev/protos/components/generated/types/component/v1"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
)

type PlannerOption func(*Planner) error

type Planner struct {
	V *validator.Validate `validate:"required"`

	Metadata    *planv1.Metadata       `validate:"required"`
	OrgMetadata *planv1.OrgMetadata    `validate:"required"`
	Component   *componentv1.Component `validate:"required"`
}

func New(v *validator.Validate, opts ...PlannerOption) (*Planner, error) {
	pln := &Planner{
		V: v,
	}

	for _, opt := range opts {
		if err := opt(pln); err != nil {
			return nil, err
		}
	}

	if err := pln.V.Struct(pln); err != nil {
		return nil, err
	}

	return pln, nil
}

func WithMetadata(metadata *planv1.Metadata) PlannerOption {
	return func(plan *Planner) error {
		plan.Metadata = metadata
		return nil
	}
}

func WithOrgMetadata(orgMetadata *planv1.OrgMetadata) PlannerOption {
	return func(plan *Planner) error {
		plan.OrgMetadata = orgMetadata
		return nil
	}
}

func WithComponent(comp *componentv1.Component) PlannerOption {
	return func(plan *Planner) error {
		plan.Component = comp
		return nil
	}
}

package build

import (
	"github.com/go-playground/validator/v10"
	componentv1 "github.com/powertoolsdev/protos/components/generated/types/component/v1"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
	"github.com/powertoolsdev/workers-executors/internal/planners"
)

type plannerOption func(*planner) error

type planner struct {
	v *validator.Validate `validate:"required"`

	Metadata    *planv1.Metadata       `validate:"required"`
	OrgMetadata *planv1.OrgMetadata    `validate:"required"`
	Component   *componentv1.Component `validate:"required"`
}

var _ planners.Planner = (*planner)(nil)

func New(v *validator.Validate, opts ...plannerOption) (*planner, error) {
	pln := &planner{
		v: v,
	}

	for _, opt := range opts {
		if err := opt(pln); err != nil {
			return nil, err
		}
	}

	if err := pln.v.Struct(pln); err != nil {
		return nil, err
	}

	return pln, nil
}

func WithMetadata(metadata *planv1.Metadata) plannerOption {
	return func(plan *planner) error {
		plan.Metadata = metadata
		return nil
	}
}

func WithOrgMetadata(orgMetadata *planv1.OrgMetadata) plannerOption {
	return func(plan *planner) error {
		plan.OrgMetadata = orgMetadata
		return nil
	}
}

func WithComponent(comp *componentv1.Component) plannerOption {
	return func(plan *planner) error {
		plan.Component = comp
		return nil
	}
}
